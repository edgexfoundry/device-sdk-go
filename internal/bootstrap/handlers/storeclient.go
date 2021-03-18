//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"fmt"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/interfaces"
)

// Database contains references to dependencies required by the database bootstrap implementation.
type Database struct {
}

// NewDatabase create a new instance of Database
func NewDatabase() *Database {
	return &Database{}
}

// BootstrapHandler creates the new interfaces.StoreClient use for database access by Store & Forward capability
func (_ *Database) BootstrapHandler(
	_ context.Context,
	_ *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	config := container.ConfigurationFrom(dic.Get)

	// Only need the database client if Store and Forward is enabled
	if !config.Writable.StoreAndForward.Enabled {
		dic.Update(di.ServiceConstructorMap{
			container.StoreClientName: func(get di.Get) interface{} {
				return nil
			},
		})
		return true
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	storeClient, err := InitializeStoreClient(secretProvider, config, startupTimer, lc)
	if err != nil {
		lc.Error(err.Error())
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		container.StoreClientName: func(get di.Get) interface{} {
			return storeClient
		},
	})

	return true
}

// InitializeStoreClient initializes the database client for Store and Forward. This is not a receiver function so that
// it can be called directly when configuration has changed and store and forward has been enabled for the first time
func InitializeStoreClient(
	secretProvider bootstrapInterfaces.SecretProvider,
	config *common.ConfigurationStruct,
	startupTimer startup.Timer,
	logger logger.LoggingClient) (interfaces.StoreClient, error) {
	var err error

	secrets, err := secretProvider.GetSecrets(config.Database.Type)
	if err != nil {
		return nil, fmt.Errorf("unable to get Database Credentials for Store and Forward: %s", err.Error())
	}

	credentials := bootstrapConfig.Credentials{
		Username: secrets[secret.UsernameKey],
		Password: secrets[secret.PasswordKey],
	}

	var storeClient interfaces.StoreClient
	for startupTimer.HasNotElapsed() {
		if storeClient, err = store.NewStoreClient(config.Database, credentials); err != nil {
			logger.Warn("unable to initialize Database for Store and Forward: %s", err.Error())
			startupTimer.SleepForInterval()
			continue
		}
		break
	}

	if err != nil {
		return nil, fmt.Errorf("initialize Database for Store and Forward failed: %s", err.Error())
	}

	return storeClient, err
}
