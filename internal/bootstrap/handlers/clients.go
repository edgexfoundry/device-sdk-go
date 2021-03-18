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

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/command"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/urlclient/local"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
)

const (
	CoreCommandClientName   = "Command"
	CoreDataClientName      = "CoreData"
	NotificationsClientName = "Notifications"
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
	_ context.Context,
	_ *sync.WaitGroup,
	_ startup.Timer,
	dic *di.Container) bool {

	config := container.ConfigurationFrom(dic.Get)

	var eventClient coredata.EventClient
	var valueDescriptorClient coredata.ValueDescriptorClient
	var commandClient command.CommandClient
	var notificationsClient notifications.NotificationsClient

	// Use of these client interfaces is optional, so they are not required to be configured. For instance if not
	// sending commands, then don't need to have the Command client in the configuration.
	if _, ok := config.Clients[CoreDataClientName]; ok {
		eventClient = coredata.NewEventClient(
			local.New(config.Clients[CoreDataClientName].Url() + clients.ApiEventRoute))

		valueDescriptorClient = coredata.NewValueDescriptorClient(
			local.New(config.Clients[CoreDataClientName].Url() + clients.ApiValueDescriptorRoute))
	}

	if _, ok := config.Clients[CoreCommandClientName]; ok {
		commandClient = command.NewCommandClient(
			local.New(config.Clients[CoreCommandClientName].Url() + clients.ApiDeviceRoute))
	}

	if _, ok := config.Clients[NotificationsClientName]; ok {
		notificationsClient = notifications.NewNotificationsClient(
			local.New(config.Clients[NotificationsClientName].Url() + clients.ApiNotificationRoute))
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
