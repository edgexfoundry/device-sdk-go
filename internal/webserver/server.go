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
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/telemetry"
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

// swagger:model
type Version struct {
	Version    string `json:"version"`
	SDKVersion string `json:"sdk_version"`
}

// NewWebserver returns a new instance of *WebServer
func NewWebServer(config *common.ConfigurationStruct, lc logger.LoggingClient, router *mux.Router) *WebServer {
	ws := &WebServer{
		Config:        config,
		LoggingClient: lc,
		router:        router,
	}

	return ws
}

//
// swagger:operation GET /ping System_Management_Agent Ping
//
// Ping
//
// Test if the service is working
//
// ---
// produces:
// - application/text
//
// Schemes:
//  - http
//
// Responses:
//  '200':
//    description: \"pong\" response
//    schema:
//      type: string
//
func (webserver *WebServer) pingHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.Write([]byte("pong"))
}

// swagger:operation GET /config System_Management_Agent Config
//
// Config
//
// Gets the currently defined configuration
//
// ---
// produces:
// - application/json
//
// Schemes:
//  - http
//
// Responses:
//  '200':
//    description: Get configuration
//    schema:
//      "$ref": "#/definitions/ConfigurationStruct"
//
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

// swagger:operation GET /metrics System_Management_Agent Metrics
//
// Metrics
//
// Gets the current metrics
//
// ---
// produces:
// - application/json
//
// Schemes:
//  - http
//
// Responses:
//  '200':
//    description: Get metrics
//    schema:
//      "$ref": "#/definitions/SystemUsage"
//
func (webserver *WebServer) metricsHandler(writer http.ResponseWriter, _ *http.Request) {
	telem := telemetry.NewSystemUsage()

	webserver.encode(telem, writer)

	return
}

// swagger:operation GET /version System_Management_Agent Version
//
// Version
//
// Gets the current version of both the SDK and the version of this application that uses the SDK.
//
// ---
// produces:
// - application/json
//
// Schemes:
//  - http
//
// Responses:
//  '200':
//    description: Get current version
//    schema:
//      "$ref": "#/definitions/Version"
//
func (webserver *WebServer) versionHandler(writer http.ResponseWriter, _ *http.Request) {
	version := Version{
		Version:    internal.ApplicationVersion,
		SDKVersion: internal.SDKVersion,
	}
	webserver.encode(version, writer)

	return
}

// AddRoute enables support to leverage the existing webserver to add routes.
func (webserver *WebServer) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) {
	webserver.router.HandleFunc(route, handler).Methods(methods...)
}

// ConfigureStandardRoutes loads up some default routes
func (webserver *WebServer) ConfigureStandardRoutes() {
	webserver.LoggingClient.Info("Registering standard routes...")

	// Ping Resource
	webserver.router.HandleFunc(clients.ApiPingRoute, webserver.pingHandler).Methods(http.MethodGet)

	// Configuration
	webserver.router.HandleFunc(clients.ApiConfigRoute, webserver.configHandler).Methods(http.MethodGet)

	// Metrics
	webserver.router.HandleFunc(clients.ApiMetricsRoute, webserver.metricsHandler).Methods(http.MethodGet)

	// Version
	webserver.router.HandleFunc(clients.ApiVersionRoute, webserver.versionHandler).Methods(http.MethodGet)
}

// SetupTriggerRoute adds a route to handle trigger pipeline from HTTP request
// swagger:operation POST /trigger Trigger Trigger
//
// Trigger
//
// Available when HTTPTrigger is specified as the binding in configuration. This API
// provides a way to initiate and start processing the defined pipeline using the data submitted.
//
// ---
// produces:
// - application/json
// consumes:
// - application/json
// parameters:
//   - in: body
//     name: Data Event
//     description: |
//       This is the data that will processed the configured pipeline. Typically this is an EdgeX event as described below, however, it can
//       ingest other forms of data if a custom Target Type (https://github.com/edgexfoundry/app-functions-sdk-go/blob/master/README.md#target-type) is being used.
//     required: true
//     schema:
//       "$ref": "#/definitions/Event"
// Responses:
//  '200':
//    description: Get current version
//    schema:
//      "$ref": "#/definitions/Version"
//
func (webserver *WebServer) SetupTriggerRoute(handlerForTrigger func(http.ResponseWriter, *http.Request)) {
	webserver.router.HandleFunc(internal.ApiTriggerRoute, handlerForTrigger)
}

// StartWebServer starts the web server
func (webserver *WebServer) StartWebServer(errChannel chan error) {
	go func() {
		if serviceTimeout, err := time.ParseDuration(webserver.Config.Service.Timeout); err != nil {
			errChannel <- fmt.Errorf("failed to parse Service.Timeout: %v", err)
		} else {
			listenAndServe(webserver, serviceTimeout, errChannel)
		}
	}()
}

// Helper function to handle HTTPs or HTTP connection based on the configured protocol
func listenAndServe(webserver *WebServer, serviceTimeout time.Duration, errChannel chan error) {

	p := fmt.Sprintf(":%d", webserver.Config.Service.Port)

	if webserver.Config.Service.Protocol == "https" {
		webserver.LoggingClient.Info(fmt.Sprintf("Starting HTTPS Web Server on port :%d", webserver.Config.Service.Port))
		errChannel <- http.ListenAndServeTLS(p, webserver.Config.Service.HTTPSCert, webserver.Config.Service.HTTPSKey, http.TimeoutHandler(webserver.router, serviceTimeout, "Request timed out"))
	} else {
		webserver.LoggingClient.Info(fmt.Sprintf("Starting HTTP Web Server on port :%d", webserver.Config.Service.Port))
		errChannel <- http.ListenAndServe(p, http.TimeoutHandler(webserver.router, serviceTimeout, "Request timed out"))
	}
}
