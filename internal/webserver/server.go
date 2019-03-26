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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/telemetry"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/gorilla/mux"
)

// WebServer handles the webserver configuration
type WebServer struct {
	Config        *common.ConfigurationStruct
	LoggingClient logger.LoggingClient
	router        *mux.Router
}

// Test if the service is working
func (webserver *WebServer) pingHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.Write([]byte("pong"))
}

func (webserver *WebServer) configHandler(writer http.ResponseWriter, _ *http.Request) {
	webserver.encode(webserver.Config, writer)
}

// Helper function for encoding things for returning from REST calls
func (webserver *WebServer) encode(data interface{}, writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(writer)
	err := enc.Encode(data)
	// Problems encoding
	if err != nil {
		webserver.LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (webserver *WebServer) metricsHandler(writer http.ResponseWriter, _ *http.Request) {
	telem := telemetry.NewSystemUsage()

	webserver.encode(telem, writer)

	return
}

// ConfigureStandardRoutes loads up some default routes
func (webserver *WebServer) ConfigureStandardRoutes() {
	webserver.LoggingClient.Info("Registering standard routes...")
	webserver.router = mux.NewRouter()

	// Ping Resource
	webserver.router.HandleFunc(clients.ApiPingRoute, webserver.pingHandler).Methods(http.MethodGet)

	// Configuration
	webserver.router.HandleFunc(clients.ApiConfigRoute, webserver.configHandler).Methods(http.MethodGet)

	// Metrics
	webserver.router.HandleFunc(clients.ApiMetricsRoute, webserver.metricsHandler).Methods(http.MethodGet)

}

// SetupTriggerRoute adds a route to handle trigger pipeline from HTTP request
func (webserver *WebServer) SetupTriggerRoute(handlerForTrigger func(http.ResponseWriter, *http.Request)) {

	webserver.router.HandleFunc("/trigger", handlerForTrigger)
}

// StartHTTPServer starts the http server
func (webserver *WebServer) StartHTTPServer(errChannel chan error) {
	webserver.LoggingClient.Info(fmt.Sprintf("Starting HTTP Server on port :%d", webserver.Config.Service.Port))
	go func() {
		p := fmt.Sprintf(":%d", webserver.Config.Service.Port)
		errChannel <- http.ListenAndServe(p, webserver.router)
	}()
}
