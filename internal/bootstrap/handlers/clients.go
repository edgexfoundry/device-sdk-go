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
	"fmt"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/command"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/urlclient"
)

// Clients contains references to dependencies required by the Clients bootstrap implementation.
type Clients struct {
}

// NewClients create a new instance of Clients
func NewClients() *Clients {
	return &Clients{}
}

// BootstrapHandler setups all the clients that have be specified in the configuration
func (_ *Clients) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	logger := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	registryClient := bootstrapContainer.RegistryFrom(dic.Get)

	var eventClient coredata.EventClient
	var valueDescriptorClient coredata.ValueDescriptorClient
	var commandClient command.CommandClient
	var notificationsClient notifications.NotificationsClient

	// Need when passing all Clients to other components
	clientMonitor, err := time.ParseDuration(config.Service.ClientMonitor)
	if err != nil {
		logger.Warn(
			fmt.Sprintf(
				"Service.ClientMonitor failed to parse: %s, use the default value: %v",
				err,
				internal.ClientMonitorDefault,
			),
		)
		// fall back to default value
		clientMonitor = internal.ClientMonitorDefault
	}

	interval := int(clientMonitor / time.Millisecond)

	// Use of these client interfaces is optional, so they are not required to be configured. For instance if not
	// sending commands, then don't need to have the Command client in the configuration.
	if _, ok := config.Clients[common.CoreDataClientName]; ok {
		eventClient = coredata.NewEventClient(
			urlclient.New(
				ctx,
				wg,
				registryClient,
				clients.CoreDataServiceKey,
				clients.ApiEventRoute,
				interval,
				config.Clients[common.CoreDataClientName].Url()+clients.ApiEventRoute,
			),
		)

		valueDescriptorClient = coredata.NewValueDescriptorClient(
			urlclient.New(
				ctx,
				wg,
				registryClient,
				clients.CoreDataServiceKey,
				clients.ApiValueDescriptorRoute,
				interval,
				config.Clients[common.CoreDataClientName].Url()+clients.ApiValueDescriptorRoute,
			),
		)
	}

	if _, ok := config.Clients[common.CoreCommandClientName]; ok {
		commandClient = command.NewCommandClient(
			urlclient.New(
				ctx,
				wg,
				registryClient,
				clients.CoreCommandServiceKey,
				clients.ApiDeviceRoute,
				interval,
				config.Clients[common.CoreCommandClientName].Url()+clients.ApiDeviceRoute,
			),
		)
	}

	if _, ok := config.Clients[common.NotificationsClientName]; ok {
		notificationsClient = notifications.NewNotificationsClient(
			urlclient.New(
				ctx,
				wg,
				registryClient,
				clients.SupportNotificationsServiceKey,
				clients.ApiNotificationRoute,
				interval,
				config.Clients[common.NotificationsClientName].Url()+clients.ApiNotificationRoute,
			),
		)
	}

	// Note that all the clients are optional so some or all these clients may be nil
	// Code that uses them must verify the client was defined and created prior to using it.
	// This information is provided in the documentation.
	dic.Update(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return eventClient
		},
		container.ValueDescriptorClientName: func(get di.Get) interface{} {
			return valueDescriptorClient
		},
		container.CommandClientName: func(get di.Get) interface{} {
			return commandClient
		},
		container.NotificationsClientName: func(get di.Get) interface{} {
			return notificationsClient
		},
	})

	return true
}
