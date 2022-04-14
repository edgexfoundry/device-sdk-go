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

package http

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

var expectedCorrelationId = uuid.New().String()

func TestPingRequest(t *testing.T) {
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	serviceName := uuid.NewString()
	target := NewRestController(mux.NewRouter(), dic, serviceName)

	recorder := doRequest(t, http.MethodGet, common.ApiPingRoute, target.Ping, nil)

	actual := commonDTO.PingResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	_, err = time.Parse(time.UnixDate, actual.Timestamp)
	assert.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actual.ApiVersion)
	assert.Equal(t, serviceName, actual.ServiceName)
}

func TestVersionRequest(t *testing.T) {
	expectedServiceVersion := "1.2.5"
	expectedSdkVersion := "1.3.1"
	serviceName := uuid.NewString()

	sdkCommon.ServiceVersion = expectedServiceVersion
	sdkCommon.SDKVersion = expectedSdkVersion

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	target := NewRestController(mux.NewRouter(), dic, serviceName)

	recorder := doRequest(t, http.MethodGet, common.ApiVersion, target.Version, nil)

	actual := commonDTO.VersionSdkResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actual.ApiVersion)
	assert.Equal(t, expectedServiceVersion, actual.Version)
	assert.Equal(t, expectedSdkVersion, actual.SdkVersion)
	assert.Equal(t, serviceName, actual.ServiceName)
}

func TestMetricsRequest(t *testing.T) {
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	serviceName := uuid.NewString()

	target := NewRestController(mux.NewRouter(), dic, serviceName)

	recorder := doRequest(t, http.MethodGet, common.ApiMetricsRoute, target.Metrics, nil)

	actual := commonDTO.MetricsResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actual.ApiVersion)
	assert.Equal(t, serviceName, actual.ServiceName)
	// Since when -race flag is use some values may come back as 0 we need to use the max value to detect change
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemAlloc)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemFrees)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemLiveObjects)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemMallocs)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemSys)
	assert.NotEqual(t, uint64(math.MaxUint64), actual.Metrics.MemTotalAlloc)
	assert.NotEqual(t, 0, actual.Metrics.CpuBusyAvg)
}

func TestConfigRequest(t *testing.T) {
	expectedConfig := config.ConfigurationStruct{
		Writable: config.WritableInfo{
			LogLevel: "DEBUG",
		},
		Registry: bootstrapConfig.RegistryInfo{
			Host: "localhost",
			Port: 8500,
			Type: "consul",
		},
	}

	serviceName := uuid.NewString()

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &expectedConfig
		},
	})

	target := NewRestController(mux.NewRouter(), dic, serviceName)

	recorder := doRequest(t, http.MethodGet, common.ApiConfigRoute, target.Config, nil)

	actualResponse := commonDTO.ConfigResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion)
	assert.Equal(t, serviceName, actualResponse.ServiceName)

	// actualResponse.Config is an interface{} so need to re-marshal/un-marshal into sdkCommon.ConfigurationStruct
	configJson, err := json.Marshal(actualResponse.Config)
	require.NoError(t, err)
	require.Less(t, 0, len(configJson))

	actualConfig := config.ConfigurationStruct{}
	err = json.Unmarshal(configJson, &actualConfig)
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, actualConfig)
}

func TestConfigRequest_CustomConfig(t *testing.T) {
	expectedConfig := config.ConfigurationStruct{
		Writable: config.WritableInfo{
			LogLevel: "DEBUG",
		},
		Registry: bootstrapConfig.RegistryInfo{
			Host: "localhost",
			Port: 8500,
			Type: "consul",
		},
	}

	serviceName := uuid.NewString()

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &expectedConfig
		},
	})

	expectedCustomConfig := TestCustomConfig{
		"test custom config",
	}

	type fullConfig struct {
		config.ConfigurationStruct
		CustomConfiguration TestCustomConfig
	}

	expectedFullConfig := fullConfig{
		expectedConfig,
		expectedCustomConfig,
	}

	target := NewRestController(mux.NewRouter(), dic, serviceName)
	target.SetCustomConfigInfo(&expectedCustomConfig)

	recorder := doRequest(t, http.MethodGet, common.ApiConfigRoute, target.Config, nil)

	actualResponse := commonDTO.ConfigResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)

	assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion)
	assert.Equal(t, serviceName, actualResponse.ServiceName)

	// actualResponse.Config is an interface{} so need to re-marshal/un-marshal into config.ConfigurationStruct
	configJson, err := json.Marshal(actualResponse.Config)
	require.NoError(t, err)
	require.Less(t, 0, len(configJson))

	actualConfig := fullConfig{}
	err = json.Unmarshal(configJson, &actualConfig)
	require.NoError(t, err)

	assert.Equal(t, expectedFullConfig, actualConfig)
}

type TestCustomConfig struct {
	Sample string
}

func (t TestCustomConfig) UpdateFromRaw(_ interface{}) bool {
	return true
}

func TestSecretRequest(t *testing.T) {
	expectedRequestId := "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	config := &config.ConfigurationStruct{}

	mockProvider := &mocks.SecretProvider{}
	mockProvider.On("StoreSecret", "mqtt", map[string]string{"password": "password", "username": "username"}).Return(nil)
	mockProvider.On("StoreSecret", "no", map[string]string{"password": "password", "username": "username"}).Return(errors.New("Invalid w/o Vault"))

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return config
		},
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockProvider
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	target := NewRestController(mux.NewRouter(), dic, uuid.NewString())
	assert.NotNil(t, target)

	validRequest := commonDTO.SecretRequest{
		BaseRequest: commonDTO.NewBaseRequest(),
		Path:        "mqtt",
		SecretData: []commonDTO.SecretDataKeyValue{
			{Key: "username", Value: "username"},
			{Key: "password", Value: "password"},
		},
	}
	validRequest.RequestId = expectedRequestId

	NoPath := validRequest
	NoPath.Path = ""
	validPathWithSlash := validRequest
	validPathWithSlash.Path = "mqtt"
	validNoRequestId := validRequest
	validNoRequestId.RequestId = ""
	badRequestId := validRequest
	badRequestId.RequestId = "bad requestId"
	noSecret := validRequest
	noSecret.SecretData = []commonDTO.SecretDataKeyValue{}
	missingSecretKey := validRequest
	missingSecretKey.SecretData = []commonDTO.SecretDataKeyValue{
		{Key: "", Value: "username"},
	}
	missingSecretValue := validRequest
	missingSecretValue.SecretData = []commonDTO.SecretDataKeyValue{
		{Key: "username", Value: ""},
	}
	noSecretStore := validRequest
	noSecretStore.Path = "no"

	tests := []struct {
		Name               string
		Request            commonDTO.SecretRequest
		ExpectedRequestId  string
		SecretPath         string
		SecretStoreEnabled string
		ErrorExpected      bool
		ExpectedStatusCode int
	}{
		{"Valid - sub-path no trailing slash, SecretPath has trailing slash", validRequest, expectedRequestId, "my-secret/", "true", false, http.StatusCreated},
		{"Valid - sub-path only with trailing slash", validPathWithSlash, expectedRequestId, "my-secret", "true", false, http.StatusCreated},
		{"Valid - both trailing slashes", validPathWithSlash, expectedRequestId, "my-secret/", "true", false, http.StatusCreated},
		{"Valid - no requestId", validNoRequestId, "", "", "true", false, http.StatusCreated},
		{"Invalid - no path", NoPath, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - bad requestId", badRequestId, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - no secret", noSecret, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - missing secret key", missingSecretKey, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - missing secret value", missingSecretValue, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - No Secret Store", noSecretStore, "", "", "false", true, http.StatusInternalServerError},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			_ = os.Setenv(secret.EnvSecretStore, testCase.SecretStoreEnabled)

			jsonData, err := json.Marshal(testCase.Request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))

			req, err := http.NewRequest(http.MethodPost, common.ApiSecretRoute, reader)
			require.NoError(t, err)
			req.Header.Set(common.CorrelationHeader, expectedCorrelationId)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(target.Secret)
			handler.ServeHTTP(recorder, req)

			actualResponse := commonDTO.BaseResponse{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, common.ApiVersion, actualResponse.ApiVersion, "Api Version not as expected")
			assert.Equal(t, testCase.ExpectedStatusCode, actualResponse.StatusCode, "BaseResponse status code not as expected")

			if testCase.ErrorExpected {
				assert.NotEmpty(t, actualResponse.Message, "Message is empty")
				return // Test complete for error cases
			}

			assert.Equal(t, testCase.ExpectedRequestId, actualResponse.RequestId, "RequestID not as expected")
			assert.Empty(t, actualResponse.Message, "Message not empty, as expected")
		})
	}
}

func doRequest(t *testing.T, method string, api string, handler http.HandlerFunc, body io.Reader) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, api, body)
	require.NoError(t, err)
	req.Header.Set(common.CorrelationHeader, expectedCorrelationId)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	expectedStatusCode := http.StatusOK
	if method == http.MethodPost {
		expectedStatusCode = http.StatusMultiStatus
	}

	assert.Equal(t, expectedStatusCode, recorder.Code, "Wrong status code")
	assert.Equal(t, common.ContentTypeJSON, recorder.Header().Get(common.ContentType), "Content type not set or not JSON")
	assert.Equal(t, expectedCorrelationId, recorder.Header().Get(common.CorrelationHeader), "CorrelationHeader not as expected")

	require.NotEmpty(t, recorder.Body.String(), "Response body is empty")

	return recorder
}
