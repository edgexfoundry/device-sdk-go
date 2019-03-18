//
// Copyright (c) 2019 Intel Corporation
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
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"encoding/json"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var logClient logger.LoggingClient

func init() {
	logClient = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
}

func TestConfigureAndPingRoute(t *testing.T) {

	webserver := WebServer{
		LoggingClient: logClient,
	}
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest("GET", clients.ApiPingRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()
	assert.Equal(t, "pong", body)

}

func TestConfigureAndConfigRoute(t *testing.T) {

	webserver := WebServer{
		LoggingClient: logClient,
		Config: &common.ConfigurationStruct{},
	}
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest("GET", clients.ApiConfigRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	expected := `{"Writable":{"SubscribeTopic":"","PublishTopic":"","LogLevel":""},"Logging":{"EnableRemote":false,"File":""},"Registry":{"Host":"","Port":0,"Type":""},"Service":{"BootTimeout":0,"CheckInterval":"","ClientMonitor":0,"Host":"","Port":0,"Protocol":"","StartupMsg":"","ReadMaxLimit":0,"Timeout":0},"MessageBus":{"Host":"","Port":0,"Type":""},"Binding":{"Type":"","Name":"","Topic":""}}` + "\n"
	body := rr.Body.String()
	assert.Equal(t, expected, body)
}

func TestConfigureAndMetricsRoute(t *testing.T) {

	type Memory struct {
		Alloc      int
		TotalAlloc int
		Sys        int
		Mallocs    int
		//more members exist in Memory, but this should suffice for this test
	}
	type Metrics struct {
		Memory Memory
	}

	webserver := WebServer{
		LoggingClient: logClient,
	}
	webserver.ConfigureStandardRoutes()

	req, _ := http.NewRequest("GET", clients.ApiMetricsRoute, nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()
	metrics := Metrics{}
	json.Unmarshal([]byte(body), &metrics)
	assert.NotNil(t, body, "Metrics not populated")
	assert.NotZero(t, metrics.Memory.Alloc, "Expected value of metrics to be non-zero")

}

func TestSetupTriggerRoute(t *testing.T) {
	myRouter := mux.NewRouter()
	webserver := WebServer{
		LoggingClient: logClient,
		router:        myRouter,
	}

	handlerFunctionNotCalled := true
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
		handlerFunctionNotCalled = false
	}

	webserver.SetupTriggerRoute(handler)

	req, _ := http.NewRequest("GET", "/trigger", nil)
	rr := httptest.NewRecorder()
	webserver.router.ServeHTTP(rr, req)

	body := rr.Body.String()

	assert.Equal(t, "test", body)
	assert.False(t, handlerFunctionNotCalled, "expected handler function to be called")

}
