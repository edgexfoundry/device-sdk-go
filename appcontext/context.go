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
//

package appcontext

import (
	"context"
	"fmt"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/command"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"

	"github.com/google/uuid"
)

// AppFunction is a type alias for func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{})
type AppFunction = func(edgexcontext *Context, params ...interface{}) (bool, interface{})

// Context ...
type Context struct {
	// This is the ID used to track the EdgeX event through entire EdgeX framework.
	CorrelationID string
	// OutputData is used for specifying the data that is to be outputted. Leverage the .Complete() function to set.
	OutputData []byte
	// This holds the configuration for your service. This is the preferred way to access your custom application settings that have been set in the configuration.
	Configuration *common.ConfigurationStruct
	// LoggingClient is exposed to allow logging following the preferred logging strategy within EdgeX.
	LoggingClient logger.LoggingClient
	// EventClient exposes Core Data's EventClient API
	EventClient coredata.EventClient
	// ValueDescriptorClient exposes Core Data's ValueDescriptor API
	ValueDescriptorClient coredata.ValueDescriptorClient
	// CommandClient exposes Core Commands' Command API
	CommandClient command.CommandClient
	// NotificationsClient exposes Support Notification's Notifications API
	NotificationsClient notifications.NotificationsClient
	// RetryData holds the data to be stored for later retry when the pipeline function returns an error
	RetryData []byte
	// SecretProvider exposes the support for getting and storing secrets
	SecretProvider interfaces.SecretProvider
	// ResponseContentType is used for holding custom response type for HTTP trigger
	ResponseContentType string
}

// Complete is optional and provides a way to return the specified data.
// In the case of an HTTP Trigger, the data will be returned as the http response.
// In the case of the message bus trigger, the data will be placed on the specified
// message bus publish topic and host in the configuration.
func (appContext *Context) Complete(output []byte) {
	appContext.OutputData = output
}

// SetRetryData sets the RetryData to the specified payload to be stored for later retry
// when the pipeline function returns an error.
func (appContext *Context) SetRetryData(payload []byte) {
	appContext.RetryData = payload
}

// PushToCoreData pushes the provided value as an event to CoreData using the device name and reading name that have been set. If validation is turned on in
// CoreServices then your deviceName and readingName must exist in the CoreMetadata and be properly registered in EdgeX.
func (appContext *Context) PushToCoreData(deviceName string, readingName string, value interface{}) (*dtos.Event, error) {
	appContext.LoggingClient.Debug("Pushing to CoreData")
	if appContext.EventClient == nil {
		return nil, fmt.Errorf("unable to Push To CoreData: '%s' is missing from Clients configuration", common.CoreDataClientName)
	}

	now := time.Now().UnixNano()
	val, err := util.CoerceType(value)
	if err != nil {
		return nil, err
	}

	// Temporary use V1 Reading until V2 EventClient is available
	// TODO: Change to use dtos.Reading
	v1Reading := models.Reading{
		Value:     string(val),
		ValueType: v2.ValueTypeString,
		Origin:    now,
		Device:    deviceName,
		Name:      readingName,
	}

	readings := make([]models.Reading, 0, 1)
	readings = append(readings, v1Reading)

	// Temporary use V1 Event until V2 EventClient is available
	// TODO: Change to use dtos.Event
	v1Event := &models.Event{
		Device:   deviceName,
		Origin:   now,
		Readings: readings,
	}

	correlation := uuid.New().String()
	ctx := context.WithValue(context.Background(), clients.CorrelationHeader, correlation)
	result, err := appContext.EventClient.Add(ctx, v1Event) // TODO: Update to use V2 EventClient
	if err != nil {
		return nil, err
	}
	v1Event.ID = result

	// TODO: Remove once V2 EventClient is available
	v2Reading := dtos.BaseReading{
		Versionable:   commonDTO.NewVersionable(),
		Id:            v1Reading.Id,
		Created:       v1Reading.Created,
		Origin:        v1Reading.Origin,
		DeviceName:    v1Reading.Device,
		ResourceName:  v1Reading.Name,
		ProfileName:   "",
		ValueType:     v1Reading.ValueType,
		SimpleReading: dtos.SimpleReading{Value: v1Reading.Value},
	}

	// TODO: Remove once V2 EventClient is available
	v2Event := dtos.Event{
		Versionable: commonDTO.NewVersionable(),
		Id:          result,
		DeviceName:  v1Event.Device,
		Origin:      v1Event.Origin,
		Readings:    []dtos.BaseReading{v2Reading},
	}
	return &v2Event, nil
}

// GetSecrets retrieves secrets from a secret store.
// path specifies the type or location of the secrets to retrieve.
// keys specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the
// specified path will be returned.
func (appContext *Context) GetSecrets(path string, keys ...string) (map[string]string, error) {
	return appContext.SecretProvider.GetSecrets(path, keys...)
}
