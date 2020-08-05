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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/internal/common"
)

func TestPingRequest(t *testing.T) {
	target := NewV2Controller(logger.NewMockClient(), nil)

	recorder := doGetRequest(t, contractsV2.ApiPingRoute, target.Ping)

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

	target := NewV2Controller(logger.NewMockClient(), nil)

	recorder := doGetRequest(t, contractsV2.ApiVersion, target.Version)

	actual := common.VersionSdkResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, contractsV2.ApiVersion, actual.ApiVersion)
	assert.Equal(t, expectedAppVersion, actual.Version)
	assert.Equal(t, expectedSdkVersion, actual.SdkVersion)
}

func TestMetricsRequest(t *testing.T) {
	target := NewV2Controller(logger.NewMockClient(), nil)

	recorder := doGetRequest(t, contractsV2.ApiMetricsRoute, target.Metrics)

	actual := common.MetricsResponse{}
	err := json.Unmarshal(recorder.Body.Bytes(), &actual)
	require.NoError(t, err)

	assert.Equal(t, contractsV2.ApiVersion, actual.ApiVersion)
	assert.NotZero(t, actual.MemAlloc)
	assert.NotZero(t, actual.MemFrees)
	assert.NotZero(t, actual.MemLiveObjects)
	assert.NotZero(t, actual.MemMallocs)
	assert.NotZero(t, actual.MemSys)
	assert.NotZero(t, actual.MemTotalAlloc)
	assert.NotNil(t, actual.CpuBusyAvg)
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

	target := NewV2Controller(logger.NewMockClient(), &expectedConfig)

	recorder := doGetRequest(t, contractsV2.ApiConfigRoute, target.Config)

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

func doGetRequest(t *testing.T, api string, handler http.HandlerFunc) *httptest.ResponseRecorder {
	req, err := http.NewRequest(http.MethodGet, api, nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	assert.Equal(t, clients.ContentTypeJSON, recorder.HeaderMap.Get(clients.ContentType))
	assert.NotEmpty(t, recorder.HeaderMap.Get(clients.CorrelationHeader))

	require.NotEmpty(t, recorder.Body.String())

	return recorder
}
