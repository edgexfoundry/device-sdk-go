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

package webserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/handlers"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/controller/rest"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	"github.com/gorilla/mux"
)

// WebServer handles the webserver configuration
type WebServer struct {
	dic        *di.Container
	config     *sdkCommon.ConfigurationStruct
	lc         logger.LoggingClient
	router     *mux.Router
	controller *rest.Controller
}

// swagger:model
type Version struct {
	Version    string `json:"version"`
	SDKVersion string `json:"sdk_version"`
}

// NewWebServer returns a new instance of *WebServer
func NewWebServer(dic *di.Container, router *mux.Router) *WebServer {
	ws := &WebServer{
		lc:         bootstrapContainer.LoggingClientFrom(dic.Get),
		config:     container.ConfigurationFrom(dic.Get),
		router:     router,
		controller: rest.NewController(router, dic),
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

	router.HandleFunc(common.ApiPingRoute, controller.Ping).Methods(http.MethodGet)
	router.HandleFunc(common.ApiVersionRoute, controller.Version).Methods(http.MethodGet)
	router.HandleFunc(common.ApiMetricsRoute, controller.Metrics).Methods(http.MethodGet)
	router.HandleFunc(common.ApiConfigRoute, controller.Config).Methods(http.MethodGet)
	router.HandleFunc(internal.ApiAddSecretRoute, controller.AddSecret).Methods(http.MethodPost)

	router.Use(handlers.ProcessCORS(webserver.config.Service.CORSConfiguration))

	// Handle the CORS preflight request
	router.Methods(http.MethodOptions).MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.Header.Get(handlers.AccessControlRequestMethod) != ""
	}).HandlerFunc(handlers.HandlePreflight(webserver.config.Service.CORSConfiguration))

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
		if serviceTimeout, err := time.ParseDuration(webserver.config.Service.RequestTimeout); err != nil {
			errChannel <- fmt.Errorf("failed to parse Service.RequestTimeout: %v", err)
		} else {
			webserver.listenAndServe(serviceTimeout, errChannel)
		}
	}()
}

// Helper function to handle HTTPs or HTTP connection based on the configured protocol
func (webserver *WebServer) listenAndServe(serviceTimeout time.Duration, errChannel chan error) {
	config := webserver.config
	lc := webserver.lc

	// The Host value is the default bind address value if the ServerBindAddr value is not specified
	// this allows env overrides to explicitly set the value used for ListenAndServe,
	// as needed for different deployments
	bindAddress := config.Service.Host
	if len(config.Service.ServerBindAddr) != 0 {
		bindAddress = config.Service.ServerBindAddr
	}
	addr := fmt.Sprintf("%s:%d", bindAddress, config.Service.Port)

	if config.HttpServer.Protocol == "https" {
		provider := bootstrapContainer.SecretProviderFrom(webserver.dic.Get)
		httpsSecretData, err := provider.GetSecret(config.HttpServer.SecretName)
		if err != nil {
			lc.Errorf("unable to find HTTPS Secret %s in Secret Store: %w", config.HttpServer.SecretName, err)
			errChannel <- err
			return
		}

		httpsCert, ok := httpsSecretData[config.HttpServer.HTTPSCertName]
		if !ok {
			lc.Errorf("unable to find HTTPS Cert in Secret Data as %s. Check configuration", config.HttpServer.HTTPSCertName, err)
			errChannel <- err
			return
		}

		httpsKey, ok := httpsSecretData[config.HttpServer.HTTPSKeyName]
		if !ok {
			lc.Errorf("unable to find HTTPS Key in Secret Data as %s. Check configuration.", config.HttpServer.HTTPSKeyName, err)
			errChannel <- err
			return
		}

		lc.Infof("Starting HTTPS Web Server on address %s", addr)

		errChannel <- http.ListenAndServeTLS(addr, httpsCert, httpsKey, http.TimeoutHandler(webserver.router, serviceTimeout, "Request timed out"))
	} else {
		lc.Infof("Starting HTTP Web Server on address %s", addr)
		errChannel <- http.ListenAndServe(addr, http.TimeoutHandler(webserver.router, serviceTimeout, "Request timed out"))
	}
}
