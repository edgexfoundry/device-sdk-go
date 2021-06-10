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

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
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

	var eventClient interfaces.EventClient
	var commandClient interfaces.CommandClient
	var notificationClient interfaces.NotificationClient
	var subscriptionClient interfaces.SubscriptionClient
	var deviceServiceClient interfaces.DeviceServiceClient
	var deviceProfileClient interfaces.DeviceProfileClient
	var deviceClient interfaces.DeviceClient

	// Use of these client interfaces is optional, so they are not required to be configured. For instance if not
	// sending commands, then don't need to have the Command client in the configuration.
	if val, ok := config.Clients[common.CoreDataServiceKey]; ok {
		eventClient = clients.NewEventClient(val.Url())
	}

	if val, ok := config.Clients[common.CoreCommandServiceKey]; ok {
		commandClient = clients.NewCommandClient(val.Url())
	}

	if val, ok := config.Clients[common.CoreMetaDataServiceKey]; ok {
		deviceServiceClient = clients.NewDeviceServiceClient(val.Url())
		deviceProfileClient = clients.NewDeviceProfileClient(val.Url())
		deviceClient = clients.NewDeviceClient(val.Url())
	}

	if val, ok := config.Clients[common.SupportNotificationsServiceKey]; ok {
		notificationClient = clients.NewNotificationClient(val.Url())
		subscriptionClient = clients.NewSubscriptionClient(val.Url())
	}

	// Note that all the clients are optional so some or all these clients may be nil
	// Code that uses them must verify the client was defined and created prior to using it.
	// This information is provided in the documentation.
	dic.Update(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return eventClient
		},
		container.CommandClientName: func(get di.Get) interface{} {
			return commandClient
		},
		container.DeviceServiceClientName: func(get di.Get) interface{} {
			return deviceServiceClient
		},
		container.DeviceProfileClientName: func(get di.Get) interface{} {
			return deviceProfileClient
		},
		container.DeviceClientName: func(get di.Get) interface{} {
			return deviceClient
		},
		container.NotificationClientName: func(get di.Get) interface{} {
			return notificationClient
		},
		container.SubscriptionClientName: func(get di.Get) interface{} {
			return subscriptionClient
		},
	})

	return true
}
