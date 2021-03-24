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

package interfaces

import (
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/command"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
)

// AppFunction is a type alias for a application pipeline function.
// appCtx is a reference to the AppFunctionContext below.
// data is the data to be operated on by the function.
// bool return value indicates if the pipeline should continue executing (true) or not (false)
// interface{} is either the data to pass to the next function (continue executing) or
// an error (stop executing due to error) or nil (done executing)
type AppFunction = func(appCxt AppFunctionContext, data interface{}) (bool, interface{})

// AppFunctionContext defines the interface for an Edgex Application Service Context provided to
// App Functions when executing in the Functions Pipeline.
type AppFunctionContext interface {
	// CorrelationID returns the correlation ID associated with the context.
	CorrelationID() string
	// InputContentType returns the content type of the data that initiated the pipeline execution. Only useful when
	// the TargetType for the pipeline is []byte, otherwise the data with be the type specified by TargetType.
	InputContentType() string
	// SetResponseData sets the response data that will be returned to the trigger when pipeline execution is complete.
	SetResponseData(data []byte)
	// ResponseData returns the data that will be returned to the trigger when pipeline execution is complete.
	ResponseData() []byte
	// SetResponseContentType sets the content type that will be returned to the trigger when pipeline
	// execution is complete.
	SetResponseContentType(string)
	// ResponseContentType returns the content type that will be returned to the trigger when pipeline
	// execution is complete.
	ResponseContentType() string
	// SetRetryData set the data that is to be retried later as part of the Store and Forward capability.
	// Used when there was failure sending the data to an external source.
	SetRetryData(data []byte)
	// GetSecret returns the secret data from the secret store (secure or insecure) for the specified path.
	// An error is returned if the path is not found or any of the keys (if specified) are not found.
	// Omit keys if all secret data for the specified path is required.
	GetSecret(path string, keys ...string) (map[string]string, error)
	// SecretsLastUpdated returns that timestamp for when the secrets in the SecretStore where last updated.
	// Useful when a connection to external source needs to be redone when the credentials have been updated.
	SecretsLastUpdated() time.Time
	// LoggingClient returns the Logger client
	LoggingClient() logger.LoggingClient
	// EventClient returns the Event client. Note if Core Data is not specified in the Clients configuration,
	// this will return nil.
	EventClient() coredata.EventClient
	// CommandClient returns the Command client. Note if Support Command is not specified in the Clients configuration,
	// this will return nil.
	CommandClient() command.CommandClient
	// NotificationsClient returns the Notifications client. Note if Support Notifications is not specified in the
	// Clients configuration, this will return nil.
	NotificationsClient() notifications.NotificationsClient
	// PushToCoreData is a convenience function for adding new Event/Reading(s) to core data and
	// back onto the EdgeX MessageBus. This function uses the Event client and will result in an error if
	// Core Data is not specified in the Clients configuration
	PushToCoreData(deviceName string, readingName string, value interface{}) (*dtos.Event, error)
}
