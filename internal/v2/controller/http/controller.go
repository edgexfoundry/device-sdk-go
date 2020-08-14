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
	"fmt"
	"net/http"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/telemetry"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/v2/dtos/requests"
)

// V2HttpController controller for V2 REST APIs
type V2HttpController struct {
	router         *mux.Router
	secretProvider security.SecretProvider
	lc             logger.LoggingClient
	config         *sdkCommon.ConfigurationStruct
}

// NewV2HttpController creates and initializes an V2HttpController
func NewV2HttpController(
	router *mux.Router,
	lc logger.LoggingClient,
	config *sdkCommon.ConfigurationStruct,
	secretProvider security.SecretProvider) *V2HttpController {
	return &V2HttpController{
		router:         router,
		secretProvider: secretProvider,
		lc:             lc,
		config:         config,
	}
}

// ConfigureStandardRoutes loads standard V2 routes
func (v2c *V2HttpController) ConfigureStandardRoutes() {
	v2c.lc.Info("Registering standard V2 routes...")
	v2c.router.HandleFunc(contractsV2.ApiPingRoute, v2c.Ping).Methods(http.MethodGet)
	v2c.router.HandleFunc(contractsV2.ApiVersionRoute, v2c.Version).Methods(http.MethodGet)
	v2c.router.HandleFunc(contractsV2.ApiMetricsRoute, v2c.Metrics).Methods(http.MethodGet)
	v2c.router.HandleFunc(contractsV2.ApiConfigRoute, v2c.Config).Methods(http.MethodGet)
	v2c.router.HandleFunc(internal.ApiV2SecretsRoute, v2c.Secrets).Methods(http.MethodPost)

	/// V2 Trigger is not considered a standard route. Trigger route (when configured) is setup by the HTTP Trigger
	//  in internal/trigger/http/rest.go
}

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2HttpController) Ping(writer http.ResponseWriter, request *http.Request) {
	response := common.NewPingResponse()
	v2c.sendResponse(writer, request, contractsV2.ApiPingRoute, response, http.StatusOK)
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2HttpController) Version(writer http.ResponseWriter, request *http.Request) {
	response := common.NewVersionSdkResponse(internal.ApplicationVersion, internal.SDKVersion)
	v2c.sendResponse(writer, request, contractsV2.ApiVersionRoute, response, http.StatusOK)
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2HttpController) Config(writer http.ResponseWriter, request *http.Request) {
	response := common.NewConfigResponse(*v2c.config)
	v2c.sendResponse(writer, request, contractsV2.ApiVersionRoute, response, http.StatusOK)
}

// Metrics handles the request to the /metrics endpoint, memory and cpu utilization stats
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2HttpController) Metrics(writer http.ResponseWriter, request *http.Request) {
	telem := telemetry.NewSystemUsage()
	metrics := common.Metrics{
		MemAlloc:       telem.Memory.Alloc,
		MemFrees:       telem.Memory.Frees,
		MemLiveObjects: telem.Memory.LiveObjects,
		MemMallocs:     telem.Memory.Mallocs,
		MemSys:         telem.Memory.Sys,
		MemTotalAlloc:  telem.Memory.TotalAlloc,
		CpuBusyAvg:     uint8(telem.CpuBusyAvg),
	}

	response := common.NewMetricsResponse(metrics)
	v2c.sendResponse(writer, request, contractsV2.ApiMetricsRoute, response, http.StatusOK)
}

// Secrets handles the request to add App Service exclusive secrets to the Secret Store
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2HttpController) Secrets(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		_ = request.Body.Close()
	}()

	secretRequest := requests.SecretsRequest{}
	err := json.NewDecoder(request.Body).Decode(&secretRequest)
	if err != nil {
		response := common.NewBaseResponse("unknown", err.Error(), http.StatusBadRequest)
		v2c.sendResponse(writer, request, internal.ApiV2SecretsRoute, response, http.StatusBadRequest)
		return
	}

	path, secrets := v2c.prepareSecrets(secretRequest)

	if err := v2c.secretProvider.StoreSecrets(path, secrets); err != nil {
		msg := fmt.Sprintf("Storing secrets failed: %v", err)
		response := common.NewBaseResponse(secretRequest.RequestID, msg, http.StatusInternalServerError)
		v2c.sendResponse(writer, request, internal.ApiV2SecretsRoute, response, http.StatusInternalServerError)
		return
	}

	response := common.NewBaseResponseNoMessage(secretRequest.RequestID, http.StatusCreated)
	v2c.sendResponse(writer, request, internal.ApiV2SecretsRoute, response, http.StatusCreated)
}

// sendResponse puts together the response packet for the V2 API
func (v2c *V2HttpController) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	api string,
	response interface{},
	statusCode int) {

	correlationID := request.Header.Get(internal.CorrelationHeaderKey)

	writer.WriteHeader(statusCode)
	writer.Header().Set(internal.CorrelationHeaderKey, correlationID)
	writer.Header().Set(clients.ContentType, clients.ContentTypeJSON)

	data, err := json.Marshal(response)
	if err != nil {
		v2c.lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = writer.Write(data)
	if err != nil {
		v2c.lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (v2c *V2HttpController) prepareSecrets(request requests.SecretsRequest) (string, map[string]string) {
	var secretsKV = make(map[string]string)
	for _, secret := range request.Secrets {
		secretsKV[secret.Key] = secret.Value
	}

	path := strings.TrimSpace(request.Path)

	// add '/' in the full URL path if it's not already at the end of the basepath or subpath
	if !strings.HasSuffix(v2c.config.SecretStoreExclusive.Path, "/") && !strings.HasPrefix(path, "/") {
		path = "/" + path
	} else if strings.HasSuffix(v2c.config.SecretStoreExclusive.Path, "/") && strings.HasPrefix(path, "/") {
		// remove extra '/' in the full URL path because secret store's (Vault) APIs don't handle extra '/'.
		path = path[1:]
	}

	return path, secretsKV
}
