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
//

package appfunction

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/urlclient/local"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	v2clients "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
)

var target *Context
var dic *di.Container

func TestMain(m *testing.M) {

	dic = di.NewContainer(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return coredata.NewEventClient(local.New(clients.ApiEventRoute))
		},
		container.NotificationsClientName: func(get di.Get) interface{} {
			return notifications.NewNotificationsClient(local.New(clients.ApiNotificationRoute))
		},
		container.CommandClientName: func(get di.Get) interface{} {
			return v2clients.NewCommandClient(clients.ApiCommandRoute)
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
	target = NewContext("", dic, "")

	os.Exit(m.Run())
}

func TestContext_PushToCoreData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("newId"))
		if r.Method != http.MethodPost {
			t.Errorf("expected http method is POST, active http method is : %s", r.Method)
		}
		url := clients.ApiEventRoute
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}
	}))

	defer ts.Close()

	eventClient := coredata.NewEventClient(local.New(ts.URL + clients.ApiEventRoute))
	dic.Update(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return eventClient
		},
	})

	expectedEvent := &dtos.Event{
		Versionable: common.NewVersionable(),
		DeviceName:  "device-name",
		Readings: []dtos.BaseReading{
			{
				DeviceName:   "device-name",
				ResourceName: "device-resource",
				ValueType:    v2.ValueTypeString,
				SimpleReading: dtos.SimpleReading{
					Value: "value",
				},
			},
		},
	}
	actualEvent, err := target.PushToCoreData("device-name", "device-resource", "value")
	require.NoError(t, err)

	assert.NotNil(t, actualEvent)
	assert.Equal(t, expectedEvent.ApiVersion, actualEvent.ApiVersion)
	assert.Equal(t, expectedEvent.DeviceName, actualEvent.DeviceName)
	assert.True(t, len(expectedEvent.Readings) == 1)
	assert.Equal(t, expectedEvent.Readings[0].DeviceName, actualEvent.Readings[0].DeviceName)
	assert.Equal(t, expectedEvent.Readings[0].ResourceName, actualEvent.Readings[0].ResourceName)
	assert.Equal(t, expectedEvent.Readings[0].Value, actualEvent.Readings[0].Value)
	assert.Equal(t, expectedEvent.Readings[0].ValueType, actualEvent.Readings[0].ValueType)
}

func TestContext_CommandClient(t *testing.T) {
	actual := target.CommandClient()
	assert.NotNil(t, actual)
}

func TestContext_EventClient(t *testing.T) {
	actual := target.EventClient()
	assert.NotNil(t, actual)
}

func TestContext_LoggingClient(t *testing.T) {
	actual := target.LoggingClient()
	assert.NotNil(t, actual)
}

func TestContext_NotificationsClient(t *testing.T) {
	actual := target.NotificationsClient()
	assert.NotNil(t, actual)
}

func TestContext_CorrelationID(t *testing.T) {
	expected := "123-3456"
	target.correlationID = expected

	actual := target.CorrelationID()

	assert.Equal(t, expected, actual)
}

func TestContext_SetCorrelationID(t *testing.T) {
	expected := "567-098"

	target.SetCorrelationID(expected)
	actual := target.correlationID

	assert.Equal(t, expected, actual)
}

func TestContext_InputContentType(t *testing.T) {
	expected := clients.ContentTypeXML
	target.inputContentType = expected

	actual := target.InputContentType()

	assert.Equal(t, expected, actual)
}

func TestContext_SetInputContentType(t *testing.T) {
	expected := clients.ContentTypeCBOR

	target.SetInputContentType(expected)
	actual := target.inputContentType

	assert.Equal(t, expected, actual)
}

func TestContext_ResponseContentType(t *testing.T) {
	expected := clients.ContentTypeJSON
	target.responseContentType = expected

	actual := target.ResponseContentType()

	assert.Equal(t, expected, actual)
}

func TestContext_SetResponseContentType(t *testing.T) {
	expected := clients.ContentTypeText

	target.SetResponseContentType(expected)
	actual := target.responseContentType

	assert.Equal(t, expected, actual)
}

func TestContext_SetResponseData(t *testing.T) {
	expected := []byte("response data")

	target.SetResponseData(expected)
	actual := target.responseData

	assert.Equal(t, expected, actual)
}

func TestContext_ResponseData(t *testing.T) {
	expected := []byte("response data")
	target.responseData = expected

	actual := target.ResponseData()

	assert.Equal(t, expected, actual)
}

func TestContext_SetRetryData(t *testing.T) {
	expected := []byte("retry data")

	target.SetRetryData(expected)
	actual := target.retryData

	assert.Equal(t, expected, actual)
}

func TestContext_RetryData(t *testing.T) {
	expected := []byte("retry data")
	target.retryData = expected

	actual := target.RetryData()

	assert.Equal(t, expected, actual)
}

func TestContext_GetSecret(t *testing.T) {
	// setup mock secret client
	expected := map[string]string{
		"username": "TEST_USER",
		"password": "TEST_PASS",
	}

	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("GetSecret", "mqtt").Return(expected, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSecretProvider
		},
	})

	actual, err := target.GetSecret("mqtt")
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestContext_SecretsLastUpdated(t *testing.T) {
	expected := time.Now()
	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("SecretsLastUpdated").Return(expected, nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSecretProvider
		},
	})

	actual := target.SecretsLastUpdated()
	assert.Equal(t, expected, actual)
}

func TestContext_AddValue(t *testing.T) {
	k := uuid.NewString()
	v := uuid.NewString()

	target.AddValue(k, v)

	res, found := target.contextData[strings.ToLower(k)]

	require.True(t, found, "item should be present in context map")
	require.Equal(t, v, res, "and it should be what we put there")
}

func TestContext_GetValue(t *testing.T) {
	k := uuid.NewString()
	v := uuid.NewString()

	target.contextData[strings.ToLower(k)] = v

	res, found := target.GetValue(k)

	require.True(t, found, "indicate item found in context map")
	require.Equal(t, v, res, "and it should be what we put there")
}

func TestContext_GetValue_NotPresent(t *testing.T) {
	k := uuid.NewString()

	res, found := target.GetValue(k)

	require.False(t, found, "should indicate item not found in context map")
	require.Equal(t, "", res, "and default string is returned")
}

func TestContext_RemoveValue(t *testing.T) {
	k := uuid.NewString()
	v := uuid.NewString()

	target.contextData[k] = v

	target.RemoveValue(k)

	_, found := target.contextData[strings.ToLower(k)]

	require.False(t, found, "item should not be present in context map")
}

func TestContext_RemoveValue_Not_Present(t *testing.T) {
	k := uuid.NewString()

	_, found := target.contextData[strings.ToLower(k)]

	require.False(t, found, "item should not be present in context map")

	target.RemoveValue(k)
}

func TestContext_GetAllValues(t *testing.T) {
	orig := map[string]string{
		"key1": "val",
		"key2": "val2",
	}

	target.contextData = orig

	res := target.GetAllValues()

	// pointers used to compare underlying memory
	require.NotSame(t, &orig, &res, "Returned map should be a copy")

	for k, v := range orig {
		assert.Equal(t, v, res[k], fmt.Sprintf("Source and result do not match at key %s", k))
	}
}
