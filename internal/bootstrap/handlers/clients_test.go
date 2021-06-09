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

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
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
		Service: config.ServiceInfo{},
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
		Port:     59880,
		Protocol: "http",
	}

	metadataClientInfo := config.ClientInfo{
		Host:     "localhost",
		Port:     59881,
		Protocol: "http",
	}

	commandClientInfo := config.ClientInfo{
		Host:     "localhost",
		Port:     59882,
		Protocol: "http",
	}

	notificationClientInfo := config.ClientInfo{
		Host:     "localhost",
		Port:     59860,
		Protocol: "http",
	}

	startupTimer := startup.NewStartUpTimer("unit-test")

	tests := []struct {
		Name                   string
		CoreDataClientInfo     *config.ClientInfo
		CommandClientInfo      *config.ClientInfo
		MetadataClientInfo     *config.ClientInfo
		NotificationClientInfo *config.ClientInfo
	}{
		{
			Name:                   "All Clients",
			CoreDataClientInfo:     &coreDataClientInfo,
			CommandClientInfo:      &commandClientInfo,
			MetadataClientInfo:     &metadataClientInfo,
			NotificationClientInfo: &notificationClientInfo,
		},
		{
			Name:                   "No Clients",
			CoreDataClientInfo:     nil,
			CommandClientInfo:      nil,
			MetadataClientInfo:     nil,
			NotificationClientInfo: nil,
		},
		{
			Name:                   "Only Core Data Clients",
			CoreDataClientInfo:     &coreDataClientInfo,
			CommandClientInfo:      nil,
			MetadataClientInfo:     nil,
			NotificationClientInfo: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			configuration.Clients = make(map[string]config.ClientInfo)

			if test.CoreDataClientInfo != nil {
				configuration.Clients[clients.CoreDataServiceKey] = coreDataClientInfo
			}

			if test.CommandClientInfo != nil {
				configuration.Clients[clients.CoreCommandServiceKey] = commandClientInfo
			}

			if test.MetadataClientInfo != nil {
				configuration.Clients[clients.CoreMetaDataServiceKey] = metadataClientInfo
			}

			if test.NotificationClientInfo != nil {
				configuration.Clients[clients.SupportNotificationsServiceKey] = notificationClientInfo
			}

			dic.Update(di.ServiceConstructorMap{
				container.ConfigurationName: func(get di.Get) interface{} {
					return configuration
				},
			})

			success := NewClients().BootstrapHandler(context.Background(), &sync.WaitGroup{}, startupTimer, dic)
			require.True(t, success)

			eventClient := container.EventClientFrom(dic.Get)
			commandClient := container.CommandClientFrom(dic.Get)
			deviceServiceClient := container.DeviceServiceClientFrom(dic.Get)
			deviceProfileClient := container.DeviceProfileClientFrom(dic.Get)
			deviceClient := container.DeviceClientFrom(dic.Get)
			notificationClient := container.NotificationClientFrom(dic.Get)
			subscriptionClient := container.SubscriptionClientFrom(dic.Get)

			if test.CoreDataClientInfo != nil {
				assert.NotNil(t, eventClient)
			} else {
				assert.Nil(t, eventClient)
			}

			if test.CommandClientInfo != nil {
				assert.NotNil(t, commandClient)
			} else {
				assert.Nil(t, commandClient)
			}

			if test.MetadataClientInfo != nil {
				assert.NotNil(t, deviceServiceClient)
				assert.NotNil(t, deviceProfileClient)
				assert.NotNil(t, deviceClient)
			} else {
				assert.Nil(t, deviceServiceClient)
				assert.Nil(t, deviceProfileClient)
				assert.Nil(t, deviceClient)
			}

			if test.NotificationClientInfo != nil {
				assert.NotNil(t, notificationClient)
				assert.NotNil(t, subscriptionClient)
			} else {
				assert.Nil(t, notificationClient)
				assert.Nil(t, subscriptionClient)
			}
		})
	}
}
