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

package webserver

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/telemetry"
)

var logClient logger.LoggingClient
var config *common.ConfigurationStruct

func TestMain(m *testing.M) {
	logClient = logger.NewMockClient()
	config = &common.ConfigurationStruct{}
	m.Run()
}

func TestAddRoute(t *testing.T) {
	routePath := "/testRoute"
	testHandler := func(_ http.ResponseWriter, _ *http.Request) {}
	sp := &mocks.SecretProvider{}

	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	err := webserver.AddRoute(routePath, testHandler)
	assert.NoError(t, err, "Not expecting an error")

	// Malformed path no slash
	routePath = "testRoute"
	err = webserver.AddRoute(routePath, testHandler)
	assert.Error(t, err, "Expecting an error")
}

func TestEncode(t *testing.T) {
	sp := &mocks.SecretProvider{}
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	writer := httptest.NewRecorder()
	var junkData interface{}
	// something that will always fail to marshal
	junkData = math.Inf(1)
	webserver.encode(junkData, writer)
	body := writer.Body.String()
	assert.NotEqual(t, math.Inf(1), body)
}

func TestConfigureAndPingRoute(t *testing.T) {

	sp := &mocks.SecretProvider{}
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest(http.MethodGet, clients.ApiPingRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()
	assert.Equal(t, "pong", body)

}

func TestConfigureAndVersionRoute(t *testing.T) {

	sp := &mocks.SecretProvider{}
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest(http.MethodGet, clients.ApiVersionRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()
	assert.Equal(t, "{\"version\":\"0.0.0\",\"sdk_version\":\"0.0.0\"}\n", body)

}
func TestConfigureAndConfigRoute(t *testing.T) {

	sp := &mocks.SecretProvider{}
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest(http.MethodGet, clients.ApiConfigRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	expected := `{"Writable":{"LogLevel":"","Pipeline":{"ExecutionOrder":"","UseTargetTypeOfByteArray":false,"Functions":null},"StoreAndForward":{"Enabled":false,"RetryInterval":"","MaxRetryCount":0},"InsecureSecrets":null},"Registry":{"Host":"","Port":0,"Type":""},"Service":{"BootTimeout":"","CheckInterval":"","Host":"","HTTPSCert":"","HTTPSKey":"","ServerBindAddr":"","Port":0,"Protocol":"","StartupMsg":"","ReadMaxLimit":0,"Timeout":""},"MessageBus":{"PublishHost":{"Host":"","Port":0,"Protocol":""},"SubscribeHost":{"Host":"","Port":0,"Protocol":""},"Type":"","Optional":null},"MqttBroker":{"Url":"","ClientId":"","ConnectTimeout":"","AutoReconnect":false,"KeepAlive":0,"QoS":0,"Retain":false,"SkipCertVerify":false,"SecretPath":"","AuthMode":""},"Binding":{"Type":"","SubscribeTopics":"","PublishTopic":""},"ApplicationSettings":null,"Clients":null,"Database":{"Type":"","Host":"","Port":0,"Timeout":"","MaxIdle":0,"BatchSize":0},"SecretStore":{"Host":"","Port":0,"Path":"","Protocol":"","Namespace":"","RootCaCertPath":"","ServerName":"","Authentication":{"AuthType":"","AuthToken":""},"AdditionalRetryAttempts":0,"RetryWaitPeriod":"","TokenFile":""}}` + "\n"

	body := rr.Body.String()
	assert.Equal(t, expected, body)
}

func TestConfigureAndMetricsRoute(t *testing.T) {
	sp := &mocks.SecretProvider{}
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest(http.MethodGet, clients.ApiMetricsRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()
	metrics := telemetry.SystemUsage{}
	json.Unmarshal([]byte(body), &metrics)
	assert.NotNil(t, body, "Metrics not populated")
	assert.NotZero(t, metrics.Memory.Alloc, "Expected Alloc value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.Frees, "Expected Frees value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.LiveObjects, "Expected LiveObjects value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.Mallocs, "Expected Mallocs value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.Sys, "Expected Sys value of metrics to be non-zero")
	assert.NotZero(t, metrics.Memory.TotalAlloc, "Expected TotalAlloc value of metrics to be non-zero")
	assert.NotNil(t, metrics.CpuBusyAvg, "Expected CpuBusyAvg value of metrics to be not nil")
}

func TestSetupTriggerRoute(t *testing.T) {
	sp := &mocks.SecretProvider{}
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())

	handlerFunctionNotCalled := true
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
		handlerFunctionNotCalled = false
	}

	webserver.SetupTriggerRoute(internal.ApiTriggerRoute, handler)

	req, _ := http.NewRequest(http.MethodGet, internal.ApiTriggerRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()

	assert.Equal(t, "test", body)
	assert.False(t, handlerFunctionNotCalled, "expected handler function to be called")
}

func TestPostSecretRoute(t *testing.T) {
	sp := &mocks.SecretProvider{}
	sp.On("StoreSecrets", "/MyPath", map[string]string{"MySecretKey": "MySecretValue"}).Return(nil)
	sp.On("StoreSecrets", "/MyPath", map[string]string{"MySecretKey1": "MySecretValue1", "MySecretKey2": "MySecretValue2"}).Return(nil)
	webserver := NewWebServer(config, sp, logClient, mux.NewRouter())
	webserver.ConfigureStandardRoutes()

	tests := []struct {
		name           string
		payload        []byte
		expectedStatus int
	}{
		{
			name:           "PostSecretRoute: Good case with one secret",
			payload:        []byte(`{"path":"MyPath","secrets":[{"key":"MySecretKey","value":"MySecretValue"}]}`),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PostSecretRoute: Good case with two secrets",
			payload:        []byte(`{"path":"MyPath","secrets":[{"key":"MySecretKey1","value":"MySecretValue1"}, {"key":"MySecretKey2","value":"MySecretValue2"}]}`),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PostSecretRoute: missing path",
			payload:        []byte(`{"secrets":[{"key":"MySecretKey1","value":"MySecretValue1"}, {"key":"MySecretKey2","value":"MySecretValue2"}]}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PostSecretRoute: missing secrets",
			payload:        []byte(`{"path":"MyPath"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PostSecretRoute: malformed payload",
			payload:        []byte(`<"path"="MyPath">`),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		currentTest := test
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, internal.ApiSecretsRoute, bytes.NewReader(currentTest.payload))
			rr := httptest.NewRecorder()
			webserver.router.ServeHTTP(rr, req)
			assert.Equal(t, currentTest.expectedStatus, rr.Result().StatusCode, "Expected secret doesn't match postSecret")
		})
	}
}

func TestValidateSecretRoute(t *testing.T) {
	secretDataBadPath := SecretData{Path: "/!$%&/foo", Secrets: []KeyValue{KeyValue{Key: "key", Value: "val"}}}
	assert.Error(t, secretDataBadPath.validateSecretData())

	secretDataEmptyKey := SecretData{Path: "/foo/bar", Secrets: []KeyValue{KeyValue{Key: "", Value: "val"}}}
	assert.Error(t, secretDataEmptyKey.validateSecretData())

	secretDataGoodPath := SecretData{Path: "/foo/bar", Secrets: []KeyValue{KeyValue{Key: "key", Value: "val"}}}
	assert.NoError(t, secretDataGoodPath.validateSecretData())
}
