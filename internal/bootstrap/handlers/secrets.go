//
// Copyright (c) 2020 Intel Corporation
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
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security"
)

// Secrets contains references to dependencies required by the Secrets bootstrap implementation.
type Secrets struct {
}

// NewDatabase create a new instance of Database
func NewSecrets() *Secrets {
	return &Secrets{}
}

// BootstrapHandler creates the SecretProvider based on configuration.
func (_ *Secrets) BootstrapHandler(
	ctx context.Context,
	_ *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	logger := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	var secretProvider security.SecretProvider

	secretProvider = security.NewSecretProvider(logger, config)
	ok := secretProvider.Initialize(ctx)
	if !ok {
		logger.Error("unable to initialize secret provider")
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		container.SecretProviderName: func(get di.Get) interface{} {
			return secretProvider
		},
	})

	return true
}
