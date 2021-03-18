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

package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	contracts "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedCorrelationId = uuid.New().String()
var dic *di.Container

func TestMain(m *testing.M) {
	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func TestPingRequest(t *testing.T) {
	target := NewController(nil, dic)

	recorder := doRequest(t, http.MethodGet, contracts.ApiPingRoute, target.Ping, nil)

	actual := common.PingResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	_, err = time.Parse(time.UnixDate, actual.Timestamp)
	assert.NoError(t, err)

	require.Equal(t, contracts.ApiVersion, actual.ApiVersion)
}

func TestVersionRequest(t *testing.T) {
	expectedAppVersion := "1.2.5"
	expectedSdkVersion := "1.3.1"

	internal.ApplicationVersion = expectedAppVersion
	internal.SDKVersion = expectedSdkVersion

	target := NewController(nil, dic)

	recorder := doRequest(t, http.MethodGet, contracts.ApiVersion, target.Version, nil)

	actual := common.VersionSdkResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, contracts.ApiVersion, actual.ApiVersion)
	assert.Equal(t, expectedAppVersion, actual.Version)
	assert.Equal(t, expectedSdkVersion, actual.SdkVersion)
}

func TestMetricsRequest(t *testing.T) {
	target := NewController(nil, dic)

	recorder := doRequest(t, http.MethodGet, contracts.ApiMetricsRoute, target.Metrics, nil)

	actual := common.MetricsResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, contracts.ApiVersion, actual.ApiVersion)
	assert.NotZero(t, actual.Metrics.MemAlloc)
	assert.NotZero(t, actual.Metrics.MemFrees)
	assert.NotZero(t, actual.Metrics.MemLiveObjects)
	assert.NotZero(t, actual.Metrics.MemMallocs)
	assert.NotZero(t, actual.Metrics.MemSys)
	assert.NotZero(t, actual.Metrics.MemTotalAlloc)
	assert.NotNil(t, actual.Metrics.CpuBusyAvg)
}

func TestConfigRequest(t *testing.T) {
	expectedConfig := sdkCommon.ConfigurationStruct{
		Writable: sdkCommon.WritableInfo{
			LogLevel: "DEBUG",
		},
		Registry: bootstrapConfig.RegistryInfo{
			Host: "localhost",
			Port: 8500,
			Type: "consul",
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &expectedConfig
		},
	})

	target := NewController(nil, dic)

	recorder := doRequest(t, http.MethodGet, contracts.ApiConfigRoute, target.Config, nil)

	actualResponse := common.ConfigResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)

	assert.Equal(t, contracts.ApiVersion, actualResponse.ApiVersion)

	// actualResponse.Config is an interface{} so need to re-marshal/un-marshal into sdkCommon.ConfigurationStruct
	configJson, err := json.Marshal(actualResponse.Config)
	require.NoError(t, err)
	require.Less(t, 0, len(configJson))

	actualConfig := sdkCommon.ConfigurationStruct{}
	err = json.Unmarshal(configJson, &actualConfig)
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, actualConfig)
}

func TestAddSecretRequest(t *testing.T) {
	expectedRequestId := "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &sdkCommon.ConfigurationStruct{}
		},
	})

	mockProvider := &mocks.SecretProvider{}
	mockProvider.On("StoreSecrets", "/mqtt", map[string]string{"password": "password", "username": "username"}).Return(nil)
	mockProvider.On("StoreSecrets", "/no", map[string]string{"password": "password", "username": "username"}).Return(errors.New("Invalid w/o Vault"))

	target := NewController(nil, dic)
	assert.NotNil(t, target)

	validRequest := common.SecretRequest{
		BaseRequest: common.BaseRequest{RequestId: expectedRequestId, Versionable: common.NewVersionable()},
		Path:        "mqtt",
		SecretData: []common.SecretDataKeyValue{
			{Key: "username", Value: "username"},
			{Key: "password", Value: "password"},
		},
	}

	NoPath := validRequest
	NoPath.Path = ""
	validPathWithSlash := validRequest
	validPathWithSlash.Path = "/mqtt"
	validNoRequestId := validRequest
	validNoRequestId.RequestId = ""
	badRequestId := validRequest
	badRequestId.RequestId = "bad requestId"
	noSecrets := validRequest
	noSecrets.SecretData = []common.SecretDataKeyValue{}
	missingSecretKey := validRequest
	missingSecretKey.SecretData = []common.SecretDataKeyValue{
		{Key: "", Value: "username"},
	}
	missingSecretValue := validRequest
	missingSecretValue.SecretData = []common.SecretDataKeyValue{
		{Key: "username", Value: ""},
	}
	noSecretStore := validRequest
	noSecretStore.Path = "no"

	tests := []struct {
		Name               string
		Request            common.SecretRequest
		ExpectedRequestId  string
		SecretsPath        string
		SecretStoreEnabled string
		ErrorExpected      bool
		ExpectedStatusCode int
	}{
		{"Valid - sub-path no trailing slash, SecretsPath has trailing slash", validRequest, expectedRequestId, "my-secrets/", "true", false, http.StatusCreated},
		{"Valid - sub-path only with trailing slash", validPathWithSlash, expectedRequestId, "my-secrets", "true", false, http.StatusCreated},
		{"Valid - both trailing slashes", validPathWithSlash, expectedRequestId, "my-secrets/", "true", false, http.StatusCreated},
		{"Valid - no requestId", validNoRequestId, "", "", "true", false, http.StatusCreated},
		{"Invalid - no path", NoPath, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - bad requestId", badRequestId, "", "", "true", true, http.StatusBadRequest},
		{"Invalid - no secrets", noSecrets, "", "", "true", true, http.StatusBadRequest},
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

			req, err := http.NewRequest(http.MethodPost, internal.ApiAddSecretRoute, reader)
			require.NoError(t, err)
			req.Header.Set(internal.CorrelationHeaderKey, expectedCorrelationId)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(target.AddSecret)
			handler.ServeHTTP(recorder, req)

			actualResponse := common.BaseResponse{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contracts.ApiVersion, actualResponse.ApiVersion, "Api Version not as expected")
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
	req.Header.Set(internal.CorrelationHeaderKey, expectedCorrelationId)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	expectedStatusCode := http.StatusOK
	if method == http.MethodPost {
		expectedStatusCode = http.StatusMultiStatus
	}

	assert.Equal(t, expectedStatusCode, recorder.Code, "Wrong status code")
	assert.Equal(t, clients.ContentTypeJSON, recorder.HeaderMap.Get(clients.ContentType), "Content type not set or not JSON")
	assert.Equal(t, expectedCorrelationId, recorder.HeaderMap.Get(internal.CorrelationHeaderKey), "CorrelationHeader not as expected")

	require.NotEmpty(t, recorder.Body.String(), "Response body is empty")

	return recorder
}
