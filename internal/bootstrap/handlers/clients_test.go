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
	"sync"
	"testing"

	"github.com/edgexfoundry/go-mod-registry/v2/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestClientsBootstrapHandler(t *testing.T) {
	configuration := &common.ConfigurationStruct{
		Service: common.ServiceInfo{},
	}

	lc := logger.NewMockClient()
	var registryClient registry.Client = nil

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		bootstrapContainer.RegistryClientInterfaceName: func(get di.Get) interface{} {
			return registryClient
		},
	})

	coreDataClientInfo := config.ClientInfo{
		Host:     "localhost",
		Port:     48080,
		Protocol: "http",
	}

	commandClientInfo := config.ClientInfo{
		Host:     "localhost",
		Port:     48081,
		Protocol: "http",
	}

	notificationsClientInfo := config.ClientInfo{
		Host:     "localhost",
		Port:     48082,
		Protocol: "http",
	}

	startupTimer := startup.NewStartUpTimer("unit-test")

	tests := []struct {
		Name                    string
		CoreDataClientInfo      *config.ClientInfo
		CommandClientInfo       *config.ClientInfo
		NotificationsClientInfo *config.ClientInfo
	}{
		{
			Name:                    "All Clients",
			CoreDataClientInfo:      &coreDataClientInfo,
			CommandClientInfo:       &commandClientInfo,
			NotificationsClientInfo: &notificationsClientInfo,
		},
		{
			Name:                    "No Clients",
			CoreDataClientInfo:      nil,
			CommandClientInfo:       nil,
			NotificationsClientInfo: nil,
		},
		{
			Name:                    "Only Core Data Clients",
			CoreDataClientInfo:      &coreDataClientInfo,
			CommandClientInfo:       nil,
			NotificationsClientInfo: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			configuration.Clients = make(map[string]config.ClientInfo)

			if test.CoreDataClientInfo != nil {
				configuration.Clients[CoreDataClientName] = coreDataClientInfo
			}

			if test.CommandClientInfo != nil {
				configuration.Clients[CoreCommandClientName] = commandClientInfo
			}

			if test.NotificationsClientInfo != nil {
				configuration.Clients[NotificationsClientName] = notificationsClientInfo
			}

			dic.Update(di.ServiceConstructorMap{
				container.ConfigurationName: func(get di.Get) interface{} {
					return configuration
				},
			})

			success := NewClients().BootstrapHandler(context.Background(), &sync.WaitGroup{}, startupTimer, dic)
			require.True(t, success)

			eventClient := container.EventClientFrom(dic.Get)
			valueDescriptorClient := container.ValueDescriptorClientFrom(dic.Get)
			commandClient := container.CommandClientFrom(dic.Get)
			notificationsClient := container.NotificationsClientFrom(dic.Get)

			if test.CoreDataClientInfo != nil {
				assert.NotNil(t, eventClient)
				assert.NotNil(t, valueDescriptorClient)
			} else {
				assert.Nil(t, eventClient)
				assert.Nil(t, valueDescriptorClient)
			}

			if test.CommandClientInfo != nil {
				assert.NotNil(t, commandClient)
			} else {
				assert.Nil(t, commandClient)
			}

			if test.NotificationsClientInfo != nil {
				assert.NotNil(t, notificationsClient)
			} else {
				assert.Nil(t, notificationsClient)
			}
		})
	}
}
