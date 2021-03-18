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

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/telemetry"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contracts "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"

	"github.com/gorilla/mux"
)

// Controller controller for V2 REST APIs
type Controller struct {
	router         *mux.Router
	secretProvider interfaces.SecretProvider
	lc             logger.LoggingClient
	config         *sdkCommon.ConfigurationStruct
}

// NewController creates and initializes an Controller
func NewController(router *mux.Router, dic *di.Container) *Controller {
	return &Controller{
		router:         router,
		secretProvider: bootstrapContainer.SecretProviderFrom(dic.Get),
		lc:             bootstrapContainer.LoggingClientFrom(dic.Get),
		config:         container.ConfigurationFrom(dic.Get),
	}
}

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *Controller) Ping(writer http.ResponseWriter, request *http.Request) {
	response := common.NewPingResponse()
	v2c.sendResponse(writer, request, contracts.ApiPingRoute, response, http.StatusOK)
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *Controller) Version(writer http.ResponseWriter, request *http.Request) {
	response := common.NewVersionSdkResponse(internal.ApplicationVersion, internal.SDKVersion)
	v2c.sendResponse(writer, request, contracts.ApiVersionRoute, response, http.StatusOK)
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *Controller) Config(writer http.ResponseWriter, request *http.Request) {
	response := common.NewConfigResponse(*v2c.config)
	v2c.sendResponse(writer, request, contracts.ApiVersionRoute, response, http.StatusOK)
}

// Metrics handles the request to the /metrics endpoint, memory and cpu utilization stats
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *Controller) Metrics(writer http.ResponseWriter, request *http.Request) {
	t := telemetry.NewSystemUsage()
	metrics := common.Metrics{
		MemAlloc:       t.Memory.Alloc,
		MemFrees:       t.Memory.Frees,
		MemLiveObjects: t.Memory.LiveObjects,
		MemMallocs:     t.Memory.Mallocs,
		MemSys:         t.Memory.Sys,
		MemTotalAlloc:  t.Memory.TotalAlloc,
		CpuBusyAvg:     uint8(t.CpuBusyAvg),
	}

	response := common.NewMetricsResponse(metrics)
	v2c.sendResponse(writer, request, contracts.ApiMetricsRoute, response, http.StatusOK)
}

// AddSecret handles the request to add App Service exclusive secret to the Secret Store
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *Controller) AddSecret(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		_ = request.Body.Close()
	}()

	secretRequest := common.SecretRequest{}
	err := json.NewDecoder(request.Body).Decode(&secretRequest)
	if err != nil {
		v2c.sendError(writer, request, errors.KindContractInvalid, "JSON decode failed", err, "")
		return
	}

	path, secret := v2c.prepareSecret(secretRequest)

	if err := v2c.secretProvider.StoreSecrets(path, secret); err != nil {
		v2c.sendError(writer, request, errors.KindServerError, "Storing secret failed", err, secretRequest.RequestId)
		return
	}

	response := common.NewBaseResponse(secretRequest.RequestId, "", http.StatusCreated)
	v2c.sendResponse(writer, request, internal.ApiAddSecretRoute, response, http.StatusCreated)
}

func (v2c *Controller) sendError(
	writer http.ResponseWriter,
	request *http.Request,
	errKind errors.ErrKind,
	message string,
	err error,
	requestID string) {
	edgexErr := errors.NewCommonEdgeX(errKind, message, err)
	v2c.lc.Error(edgexErr.Error())
	v2c.lc.Debug(edgexErr.DebugMessages())
	response := common.NewBaseResponse(requestID, edgexErr.Message(), edgexErr.Code())
	v2c.sendResponse(writer, request, internal.ApiAddSecretRoute, response, edgexErr.Code())
}

// sendResponse puts together the response packet for the V2 API
func (v2c *Controller) sendResponse(
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

func (v2c *Controller) prepareSecret(request common.SecretRequest) (string, map[string]string) {
	var secretsKV = make(map[string]string)
	for _, secret := range request.SecretData {
		secretsKV[secret.Key] = secret.Value
	}

	path := strings.TrimSpace(request.Path)

	// add '/' in the full URL path if it's not already at the end of the base path or sub path
	if !strings.HasSuffix(v2c.config.SecretStore.Path, "/") && !strings.HasPrefix(path, "/") {
		path = "/" + path
	} else if strings.HasSuffix(v2c.config.SecretStore.Path, "/") && strings.HasPrefix(path, "/") {
		// remove extra '/' in the full URL path because secret store's (Vault) APIs don't handle extra '/'.
		path = path[1:]
	}

	return path, secretsKV
}
