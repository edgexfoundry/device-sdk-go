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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
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
