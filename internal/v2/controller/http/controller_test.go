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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/secret"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/v2/dtos/requests"
)

var expectedCorrelationId = uuid.New().String()

func TestConfigureStandardRoutes(t *testing.T) {
	router := mux.NewRouter()
	target := NewV2HttpController(router, logger.NewMockClient(), nil, nil)
	target.ConfigureStandardRoutes()

	var routes []*mux.Route
	walkFunc := func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		routes = append(routes, route)
		return nil
	}
	err := router.Walk(walkFunc)

	require.NoError(t, err)
	assert.Len(t, routes, 5)

	paths := make(map[string]string)
	for _, route := range routes {
		url, err := route.URLPath("", "")
		require.NoError(t, err)
		methods, err := route.GetMethods()
		require.NoError(t, err)
		require.Len(t, methods, 1)
		paths[url.Path] = methods[0]
	}

	assert.Contains(t, paths, contractsV2.ApiPingRoute)
	assert.Contains(t, paths, contractsV2.ApiConfigRoute)
	assert.Contains(t, paths, contractsV2.ApiMetricsRoute)
	assert.Contains(t, paths, contractsV2.ApiVersionRoute)
	assert.Contains(t, paths, internal.ApiV2SecretsRoute)

	assert.Equal(t, http.MethodGet, paths[contractsV2.ApiPingRoute])
	assert.Equal(t, http.MethodGet, paths[contractsV2.ApiConfigRoute])
	assert.Equal(t, http.MethodGet, paths[contractsV2.ApiMetricsRoute])
	assert.Equal(t, http.MethodGet, paths[contractsV2.ApiVersionRoute])
	assert.Equal(t, http.MethodPost, paths[internal.ApiV2SecretsRoute])
}

func TestPingRequest(t *testing.T) {
	target := NewV2HttpController(nil, logger.NewMockClient(), nil, nil)

	recorder := doRequest(t, http.MethodGet, contractsV2.ApiPingRoute, target.Ping, nil)

	actual := common.PingResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	_, err = time.Parse(time.UnixDate, actual.Timestamp)
	assert.NoError(t, err)

	require.Equal(t, contractsV2.ApiVersion, actual.ApiVersion)
}

func TestVersionRequest(t *testing.T) {
	expectedAppVersion := "1.2.5"
	expectedSdkVersion := "1.3.1"

	internal.ApplicationVersion = expectedAppVersion
	internal.SDKVersion = expectedSdkVersion

	target := NewV2HttpController(nil, logger.NewMockClient(), nil, nil)

	recorder := doRequest(t, http.MethodGet, contractsV2.ApiVersion, target.Version, nil)

	actual := common.VersionSdkResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, contractsV2.ApiVersion, actual.ApiVersion)
	assert.Equal(t, expectedAppVersion, actual.Version)
	assert.Equal(t, expectedSdkVersion, actual.SdkVersion)
}

func TestMetricsRequest(t *testing.T) {
	target := NewV2HttpController(nil, logger.NewMockClient(), nil, nil)

	recorder := doRequest(t, http.MethodGet, contractsV2.ApiMetricsRoute, target.Metrics, nil)

	actual := common.MetricsResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, contractsV2.ApiVersion, actual.ApiVersion)
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

	target := NewV2HttpController(nil, logger.NewMockClient(), &expectedConfig, nil)

	recorder := doRequest(t, http.MethodGet, contractsV2.ApiConfigRoute, target.Config, nil)

	actualResponse := common.ConfigResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
	require.NoError(t, err)

	assert.Equal(t, contractsV2.ApiVersion, actualResponse.ApiVersion)

	// actualResponse.Config is an interface{} so need to re-marshal/un-marshal into sdkCommon.ConfigurationStruct
	configJson, err := json.Marshal(actualResponse.Config)
	require.NoError(t, err)
	require.Less(t, 0, len(configJson))

	actualConfig := sdkCommon.ConfigurationStruct{}
	err = json.Unmarshal(configJson, &actualConfig)
	require.NoError(t, err)

	assert.Equal(t, expectedConfig, actualConfig)
}

func TestSecretsRequest(t *testing.T) {
	expectedRequestId := "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	config := &sdkCommon.ConfigurationStruct{}

	lc := logger.NewMockClient()

	mockProvider := &mocks.SecretProvider{}
	mockProvider.On("StoreSecrets", "/mqtt", map[string]string{"password": "password", "username": "username"}).Return(nil)
	mockProvider.On("StoreSecrets", "/no", map[string]string{"password": "password", "username": "username"}).Return(errors.New("Invalid w/o Vault"))

	target := NewV2HttpController(nil, lc, config, mockProvider)
	assert.NotNil(t, target)

	validRequest := requests.SecretsRequest{
		BaseRequest: common.BaseRequest{RequestId: expectedRequestId},
		Path:        "mqtt",
		Secrets: []requests.SecretsKeyValue{
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
	noSecrets.Secrets = []requests.SecretsKeyValue{}
	missingSecretKey := validRequest
	missingSecretKey.Secrets = []requests.SecretsKeyValue{
		{Key: "", Value: "username"},
	}
	missingSecretValue := validRequest
	missingSecretValue.Secrets = []requests.SecretsKeyValue{
		{Key: "username", Value: ""},
	}
	noSecretStore := validRequest
	noSecretStore.Path = "no"

	tests := []struct {
		Name               string
		Request            requests.SecretsRequest
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

			req, err := http.NewRequest(http.MethodPost, internal.ApiV2SecretsRoute, reader)
			require.NoError(t, err)
			req.Header.Set(internal.CorrelationHeaderKey, expectedCorrelationId)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(target.Secrets)
			handler.ServeHTTP(recorder, req)

			actualResponse := common.BaseResponse{}
			err = json.Unmarshal(recorder.Body.Bytes(), &actualResponse)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, contractsV2.ApiVersion, actualResponse.ApiVersion, "Api Version not as expected")
			assert.Equal(t, testCase.ExpectedStatusCode, int(actualResponse.StatusCode), "BaseResponse status code not as expected")

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
