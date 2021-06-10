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
	"os"
	"strings"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	v2clients "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDtos "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var target *Context
var dic *di.Container
var baseUrl = "http://localhost:"

func TestMain(m *testing.M) {
	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
	target = NewContext("", dic, "")

	os.Exit(m.Run())
}

func TestContext_EventClient(t *testing.T) {
	actual := target.EventClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return v2clients.NewEventClient(baseUrl + "59880")
		},
	})

	actual = target.EventClient()
	assert.NotNil(t, actual)
}

func TestContext_CommandClient(t *testing.T) {
	actual := target.CommandClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.CommandClientName: func(get di.Get) interface{} {
			return v2clients.NewCommandClient(baseUrl + "59882")
		},
	})

	actual = target.CommandClient()
	assert.NotNil(t, actual)
}

func TestContext_DeviceServiceClient(t *testing.T) {
	actual := target.DeviceServiceClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.DeviceServiceClientName: func(get di.Get) interface{} {
			return v2clients.NewDeviceServiceClient(baseUrl + "59881")
		},
	})

	actual = target.DeviceServiceClient()
	assert.NotNil(t, actual)

}

func TestContext_DeviceProfileClient(t *testing.T) {
	actual := target.DeviceProfileClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.DeviceProfileClientName: func(get di.Get) interface{} {
			return v2clients.NewDeviceProfileClient(baseUrl + "59881")
		},
	})

	actual = target.DeviceProfileClient()
	assert.NotNil(t, actual)
}

func TestContext_DeviceClient(t *testing.T) {
	actual := target.DeviceClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.DeviceClientName: func(get di.Get) interface{} {
			return v2clients.NewDeviceClient(baseUrl + "59881")
		},
	})

	actual = target.DeviceClient()
	assert.NotNil(t, actual)

}

func TestContext_NotificationClient(t *testing.T) {
	actual := target.NotificationClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.NotificationClientName: func(get di.Get) interface{} {
			return v2clients.NewNotificationClient(baseUrl + "59860")
		},
	})

	actual = target.NotificationClient()
	assert.NotNil(t, actual)

}

func TestContext_SubscriptionClient(t *testing.T) {
	actual := target.SubscriptionClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.SubscriptionClientName: func(get di.Get) interface{} {
			return v2clients.NewSubscriptionClient(baseUrl + "59860")
		},
	})

	actual = target.SubscriptionClient()
	assert.NotNil(t, actual)
}

func TestContext_LoggingClient(t *testing.T) {
	actual := target.LoggingClient()
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
	expected := common.ContentTypeXML
	target.inputContentType = expected

	actual := target.InputContentType()

	assert.Equal(t, expected, actual)
}

func TestContext_SetInputContentType(t *testing.T) {
	expected := common.ContentTypeCBOR

	target.SetInputContentType(expected)
	actual := target.inputContentType

	assert.Equal(t, expected, actual)
}

func TestContext_ResponseContentType(t *testing.T) {
	expected := common.ContentTypeJSON
	target.responseContentType = expected

	actual := target.ResponseContentType()

	assert.Equal(t, expected, actual)
}

func TestContext_SetResponseContentType(t *testing.T) {
	expected := common.ContentTypeText

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

func TestContext_ApplyValues_No_Placeholders(t *testing.T) {
	data := map[string]string{
		"key1": "val",
		"key2": "val2",
	}

	input := uuid.NewString()

	target.contextData = data

	res, err := target.ApplyValues(input)

	require.NoError(t, err)
	require.Equal(t, res, input)
}

func TestContext_ApplyValues_Placeholders(t *testing.T) {
	data := map[string]string{
		"key1": "val",
		"key2": "val2",
	}

	input := "{key1}-{key2}"

	target.contextData = data

	res, err := target.ApplyValues(input)

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s-%s", data["key1"], data["key2"]), res)
}

func TestContext_ApplyValues_MissingPlaceholder(t *testing.T) {
	data := map[string]string{
		"key1": "val",
		"key2": "val2",
	}

	input := "{key1}-{key2}-{key3}"

	target.contextData = data

	res, err := target.ApplyValues(input)

	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("failed to replace all context placeholders in input ('%s-%s-{key3}' after replacements)", data["key1"], data["key2"]), err.Error())
	require.Equal(t, "", res)
}

func TestContext_PushToCore(t *testing.T) {
	mockClient := clientMocks.EventClient{}
	mockClient.On("Add", mock.Anything, mock.Anything).Return(commonDtos.BaseWithIdResponse{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return &mockClient
		},
	})

	event := dtos.NewEvent("MyProfile", "MyDevice", "MyResource")
	err := event.AddSimpleReading("MyResource", common.ValueTypeInt32, int32(1234))
	require.NoError(t, err)

	_, err = target.PushToCore(event)
	require.NoError(t, err)
}

func TestContext_PushToCore_error(t *testing.T) {
	dic.Update(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return nil
		},
	})

	_, err := target.PushToCore(dtos.Event{})
	require.Error(t, err)
}

func TestContext_GetDeviceResource(t *testing.T) {
	mockClient := clientMocks.DeviceProfileClient{}
	mockClient.On("DeviceResourceByProfileNameAndResourceName", mock.Anything, mock.Anything, mock.Anything).Return(responses.DeviceResourceResponse{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DeviceProfileClientName: func(get di.Get) interface{} {
			return &mockClient
		},
	})

	_, err := target.GetDeviceResource("MyProfile", "MyResource")
	require.NoError(t, err)
}

func TestContext_GetDeviceResource_Error(t *testing.T) {
	dic.Update(di.ServiceConstructorMap{
		container.DeviceProfileClientName: func(get di.Get) interface{} {
			return nil
		},
	})

	_, err := target.GetDeviceResource("MyProfile", "MyResource")
	require.Error(t, err)
}
