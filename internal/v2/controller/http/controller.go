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
)

// V2Controller controller for V2 REST APIs
type V2Controller struct {
	lc logger.LoggingClient
}

// NewV2Controller creates and initializes an V2Controller
func NewV2Controller(lc logger.LoggingClient) *V2Controller {
	return &V2Controller{
		lc: lc,
	}
}

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (v2c *V2Controller) Ping(w http.ResponseWriter, r *http.Request) {
	pingResponse := common.NewPingResponse()

	v2c.sendResponse(w, contractsV2.ApiPingRoute, pingResponse, uuid.New().String())

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
