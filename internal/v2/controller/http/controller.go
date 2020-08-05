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

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/google/uuid"

	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/telemetry"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
)

// V2Controller controller for V2 REST APIs
type V2Controller struct {
	lc     logger.LoggingClient
	config *sdkCommon.ConfigurationStruct
}

// NewV2Controller creates and initializes an V2Controller
func NewV2Controller(lc logger.LoggingClient, config *sdkCommon.ConfigurationStruct) *V2Controller {
	return &V2Controller{
		lc:     lc,
		config: config,
	}
}

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2Controller) Ping(w http.ResponseWriter, _ *http.Request) {
	response := common.NewPingResponse()
	v2c.sendResponse(w, contractsV2.ApiPingRoute, response, uuid.New().String())
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2Controller) Version(w http.ResponseWriter, _ *http.Request) {
	response := common.NewVersionSdkResponse(internal.ApplicationVersion, internal.SDKVersion)
	v2c.sendResponse(w, contractsV2.ApiVersionRoute, response, uuid.New().String())
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2Controller) Config(w http.ResponseWriter, _ *http.Request) {
	response := common.NewConfigResponse(*v2c.config)
	v2c.sendResponse(w, contractsV2.ApiVersionRoute, response, uuid.New().String())
}

// Metrics handles the request to the /metrics endpoint, memory and cpu utilization stats
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2Controller) Metrics(w http.ResponseWriter, r *http.Request) {
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
	v2c.sendResponse(w, contractsV2.ApiMetricsRoute, response, uuid.New().String())
}

// sendResponse puts together the response packet for the V2 API
// api is the V2 API path
// item is the object or data that is sent back as part of the response
// correlationID is a unique identifier correlating a request to its associated response
func (v2c *V2Controller) sendResponse(w http.ResponseWriter, api string, item interface{}, correlationID string) {
	data, err := json.Marshal(item)
	if err != nil {
		v2c.lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		v2c.lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(clients.CorrelationHeader, correlationID)
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}
