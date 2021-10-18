//
// Copyright (c) 2021 Intel Corporation
// Copyright (c) 2021 One Track Consulting
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

package appfunction

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

// NewContext creates, initializes and return a new Context with implements the interfaces.AppFunctionContext interface
func NewContext(correlationID string, dic *di.Container, inputContentType string) *Context {
	return &Context{
		correlationID: correlationID,
		// Dic is public, so we can confirm it is set correctly
		Dic:                  dic,
		inputContentType:     inputContentType,
		contextData:          make(map[string]string),
		valuePlaceholderSpec: regexp.MustCompile("{[^}]*}"),
	}
}

// Context contains the data functions that implement the interfaces.AppFunctionContext
type Context struct {
	// Dic is public, so we can confirm it is set correctly
	Dic                  *di.Container
	correlationID        string
	inputContentType     string
	responseData         []byte
	retryData            []byte
	responseContentType  string
	contextData          map[string]string
	valuePlaceholderSpec *regexp.Regexp
}

// Clone returns a copy of the context that can be manipulated independently.
func (appContext *Context) Clone() interfaces.AppFunctionContext {
	contextCopy := make(map[string]string, len(appContext.contextData))

	for k, v := range appContext.contextData {
		contextCopy[k] = v
	}

	return &Context{
		Dic:                  appContext.Dic,
		correlationID:        appContext.correlationID,
		inputContentType:     appContext.inputContentType,
		responseData:         appContext.responseData,
		retryData:            appContext.retryData,
		responseContentType:  appContext.responseContentType,
		contextData:          contextCopy,
		valuePlaceholderSpec: appContext.valuePlaceholderSpec,
	}
}

// SetCorrelationID sets the correlationID. This function is not part of the AppFunctionContext interface,
// so it is internal SDK use only
func (appContext *Context) SetCorrelationID(id string) {
	appContext.correlationID = id
}

// CorrelationID returns context's the correlation ID
func (appContext *Context) CorrelationID() string {
	return appContext.correlationID
}

// SetInputContentType sets the inputContentType. This function is not part of the AppFunctionContext interface,
// so it is internal SDK use only
func (appContext *Context) SetInputContentType(contentType string) {
	appContext.inputContentType = contentType
}

// InputContentType returns the context's inputContentType
func (appContext *Context) InputContentType() string {
	return appContext.inputContentType
}

// SetResponseData provides a way to return the specified data as a response to the trigger that initiated
// the execution of the function pipeline. In the case of an HTTP Trigger, the data will be returned as the http response.
// In the case of a message bus trigger, the data will be published to the configured message bus publish topic.
func (appContext *Context) SetResponseData(output []byte) {
	appContext.responseData = output
}

// ResponseData returns the context's responseData.
func (appContext *Context) ResponseData() []byte {
	return appContext.responseData
}

// SetResponseContentType sets the context's responseContentType
func (appContext *Context) SetResponseContentType(contentType string) {
	appContext.responseContentType = contentType
}

// ResponseContentType returns the context's responseContentType
func (appContext *Context) ResponseContentType() string {
	return appContext.responseContentType
}

// SetRetryData sets the context's retryData to the specified payload to be stored for later retry
// when the pipeline function returns an error.
func (appContext *Context) SetRetryData(payload []byte) {
	appContext.retryData = payload
}

// RetryData returns the context's retryData. This function is not part of the AppFunctionContext interface,
// so it is internal SDK use only
func (appContext *Context) RetryData() []byte {
	return appContext.retryData
}

// GetSecret returns the secret data from the secret store (secure or insecure) for the specified path.
func (appContext *Context) GetSecret(path string, keys ...string) (map[string]string, error) {
	secretProvider := bootstrapContainer.SecretProviderFrom(appContext.Dic.Get)
	return secretProvider.GetSecret(path, keys...)
}

// SecretsLastUpdated returns that timestamp for when the secrets in the SecretStore where last updated.
func (appContext *Context) SecretsLastUpdated() time.Time {
	secretProvider := bootstrapContainer.SecretProviderFrom(appContext.Dic.Get)
	return secretProvider.SecretsLastUpdated()
}

// LoggingClient returns the Logging client from the dependency injection container
func (appContext *Context) LoggingClient() logger.LoggingClient {
	return bootstrapContainer.LoggingClientFrom(appContext.Dic.Get)
}

// EventClient returns the Event client, which may be nil, from the dependency injection container
func (appContext *Context) EventClient() clients.EventClient {
	return container.EventClientFrom(appContext.Dic.Get)
}

// CommandClient returns the Command client, which may be nil, from the dependency injection container
func (appContext *Context) CommandClient() clients.CommandClient {
	return container.CommandClientFrom(appContext.Dic.Get)
}

// DeviceServiceClient returns the DeviceService client, which may be nil, from the dependency injection container
func (appContext *Context) DeviceServiceClient() clients.DeviceServiceClient {
	return container.DeviceServiceClientFrom(appContext.Dic.Get)
}

// DeviceProfileClient returns the DeviceProfile client, which may be nil, from the dependency injection container
func (appContext *Context) DeviceProfileClient() clients.DeviceProfileClient {
	return container.DeviceProfileClientFrom(appContext.Dic.Get)
}

// DeviceClient returns the Device client, which may be nil, from the dependency injection container
func (appContext *Context) DeviceClient() clients.DeviceClient {
	return container.DeviceClientFrom(appContext.Dic.Get)
}

// NotificationClient returns the Notification client, which may be nil, from the dependency injection container
func (appContext *Context) NotificationClient() clients.NotificationClient {
	return container.NotificationClientFrom(appContext.Dic.Get)
}

// SubscriptionClient returns the Subscription client, which may be nil, from the dependency injection container
func (appContext *Context) SubscriptionClient() clients.SubscriptionClient {
	return container.SubscriptionClientFrom(appContext.Dic.Get)
}

// AddValue stores a value for access within other functions in pipeline
func (appContext *Context) AddValue(key string, value string) {
	appContext.contextData[strings.ToLower(key)] = value
}

// RemoveValue deletes a value stored in the context at the given key
func (appContext *Context) RemoveValue(key string) {
	delete(appContext.contextData, strings.ToLower(key))
}

// PushToCore pushes a new event to Core Data.
func (appContext *Context) PushToCore(event dtos.Event) (common.BaseWithIdResponse, error) {
	client := appContext.EventClient()
	if client == nil {
		return common.BaseWithIdResponse{}, errors.New("EventClient not initialized. Core Metadata is missing from clients configuration")
	}

	request := requests.NewAddEventRequest(event)
	return client.Add(context.Background(), request)
}

// GetDeviceResource retrieves the DeviceResource for given profileName and resourceName.
func (appContext *Context) GetDeviceResource(profileName string, resourceName string) (dtos.DeviceResource, error) {
	client := appContext.DeviceProfileClient()
	if client == nil {
		return dtos.DeviceResource{}, errors.New("DeviceProfileClient not initialized. Core Metadata is missing from clients configuration")
	}

	response, err := client.DeviceResourceByProfileNameAndResourceName(context.Background(), profileName, resourceName)
	if err != nil {
		return dtos.DeviceResource{}, err
	}

	return response.Resource, nil
}

// GetValue attempts to retrieve a value stored in the context at the given key
func (appContext *Context) GetValue(key string) (string, bool) {
	val, found := appContext.contextData[strings.ToLower(key)]
	return val, found
}

// GetAllValues returns a read-only copy of all data stored in the context
func (appContext *Context) GetAllValues() map[string]string {
	out := make(map[string]string)

	for k, v := range appContext.contextData {
		out[k] = v
	}
	return out
}

// ApplyValues looks in the provided string for placeholders of the form
// '{any-value-key}' and attempts to replace with the value stored under
// the key in context storage.  An error will be returned if any placeholders
// are not matched to a value in the context.
func (appContext *Context) ApplyValues(format string) (string, error) {
	attempts := make(map[string]bool)

	result := format

	for _, placeholder := range appContext.valuePlaceholderSpec.FindAllString(format, -1) {
		if _, tried := attempts[placeholder]; tried {
			continue
		}

		key := strings.TrimRight(strings.TrimLeft(placeholder, "{"), "}")

		value, found := appContext.GetValue(key)

		attempts[placeholder] = found

		if found {
			result = strings.Replace(result, placeholder, value, -1)
		}
	}

	for _, succeeded := range attempts {
		if !succeeded {
			return "", fmt.Errorf("failed to replace all context placeholders in input ('%s' after replacements)", result)
		}
	}

	return result, nil
}

// PipelineId returns the ID of the pipeline that is executing
func (appContext *Context) PipelineId() string {
	id, _ := appContext.GetValue(interfaces.PIPELINEID)
	return id
}
