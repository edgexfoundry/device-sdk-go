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
	"fmt"
	"net/http"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	contracts "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/controller/rest"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/gorilla/mux"
)

// WebServer handles the webserver configuration
type WebServer struct {
	Config         *common.ConfigurationStruct
	lc             logger.LoggingClient
	router         *mux.Router
	secretProvider interfaces.SecretProvider
	controller     *rest.Controller
}

// swagger:model
type Version struct {
	Version    string `json:"version"`
	SDKVersion string `json:"sdk_version"`
}

// NewWebserver returns a new instance of *WebServer
func NewWebServer(config *common.ConfigurationStruct, secretProvider interfaces.SecretProvider, lc logger.LoggingClient, router *mux.Router) *WebServer {
	ws := &WebServer{
		Config:         config,
		lc:             lc,
		router:         router,
		secretProvider: secretProvider,
		controller:     rest.NewController(router, lc, config, secretProvider),
	}

	return ws
}

// AddRoute enables support to leverage the existing webserver to add routes.
func (webserver *WebServer) AddRoute(routePath string, handler func(http.ResponseWriter, *http.Request), methods ...string) error {
	route := webserver.router.HandleFunc(routePath, handler).Methods(methods...)
	if routeErr := route.GetError(); routeErr != nil {
		return routeErr
	}
	return nil
}

// ConfigureStandardRoutes loads up the default routes
func (webserver *WebServer) ConfigureStandardRoutes() {
	router := webserver.router
	controller := webserver.controller

	webserver.lc.Info("Registering standard routes...")

	router.HandleFunc(contracts.ApiPingRoute, controller.Ping).Methods(http.MethodGet)
	router.HandleFunc(contracts.ApiVersionRoute, controller.Version).Methods(http.MethodGet)
	router.HandleFunc(contracts.ApiMetricsRoute, controller.Metrics).Methods(http.MethodGet)
	router.HandleFunc(contracts.ApiConfigRoute, controller.Config).Methods(http.MethodGet)
	router.HandleFunc(internal.ApiAddSecretRoute, controller.AddSecret).Methods(http.MethodPost)

	/// Trigger is not considered a standard route. Trigger route (when configured) is setup by the HTTP Trigger
	//  in internal/trigger/http/rest.go
}

// SetupTriggerRoute adds a route to handle trigger pipeline from REST request
func (webserver *WebServer) SetupTriggerRoute(path string, handlerForTrigger func(http.ResponseWriter, *http.Request)) {
	webserver.router.HandleFunc(path, handlerForTrigger)
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

	// this allows env overrides to explicitly set the value used
	// for ListenAndServe, as needed for different deployments
	addr := fmt.Sprintf("%v:%d", webserver.Config.Service.ServerBindAddr, webserver.Config.Service.Port)

	if webserver.Config.Service.Protocol == "https" {
		webserver.lc.Infof("Starting HTTPS Web Server on address %v", addr)
		errChannel <- http.ListenAndServeTLS(addr, webserver.Config.Service.HTTPSCert, webserver.Config.Service.HTTPSKey, http.TimeoutHandler(webserver.router, serviceTimeout, "Request timed out"))
	} else {
		webserver.lc.Infof("Starting HTTP Web Server on address %v", addr)
		errChannel <- http.ListenAndServe(addr, http.TimeoutHandler(webserver.router, serviceTimeout, "Request timed out"))
	}
}
