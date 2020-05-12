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

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/command"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
)

// ValueDescriptorClientName contains the name of the ValueDescriptorClient's implementation in the DIC.
var ValueDescriptorClientName = di.TypeInstanceToName((*coredata.ValueDescriptorClient)(nil))

// ValueDescriptorClientFrom helper function queries the DIC and returns the ValueDescriptorClient's implementation.
func ValueDescriptorClientFrom(get di.Get) coredata.ValueDescriptorClient {
	if get(ValueDescriptorClientName) == nil {
		return nil
	}

	return get(ValueDescriptorClientName).(coredata.ValueDescriptorClient)
}

// EventClientName contains the name of the EventClient's implementation in the DIC.
var EventClientName = di.TypeInstanceToName((*coredata.EventClient)(nil))

// ValueDescriptorClientFrom helper function queries the DIC and returns the ValueDescriptorClient's implementation.
func EventClientFrom(get di.Get) coredata.EventClient {
	if get(EventClientName) == nil {
		return nil
	}

	return get(EventClientName).(coredata.EventClient)
}

// NotificationsClientName contains the name of the NotificationsClientInfo's implementation in the DIC.
var NotificationsClientName = di.TypeInstanceToName((*notifications.NotificationsClient)(nil))

// NotificationsClientFrom helper function queries the DIC and returns the NotificationsClientInfo's implementation.
func NotificationsClientFrom(get di.Get) notifications.NotificationsClient {
	if get(NotificationsClientName) == nil {
		return nil
	}

	return get(NotificationsClientName).(notifications.NotificationsClient)
}

// CommandClientName contains the name of the CommandClientInfo's implementation in the DIC.
var CommandClientName = di.TypeInstanceToName((*command.CommandClient)(nil))

// NotificationsClientFrom helper function queries the DIC and returns the NotificationsClientInfo's implementation.
func CommandClientFrom(get di.Get) command.CommandClient {
	if get(CommandClientName) == nil {
		return nil
	}

	return get(CommandClientName).(command.CommandClient)
}
