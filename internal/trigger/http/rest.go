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

package http

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/webserver"
)

// Trigger implements Trigger to support Triggers
type Trigger struct {
	Configuration common.ConfigurationStruct
	Runtime       runtime.GolangRuntime
	outputData    []byte
	logging       logger.LoggingClient
	Webserver     *webserver.WebServer
}

// Initialize ...
func (trigger *Trigger) Initialize(logger logger.LoggingClient) error {
	trigger.logging = logger
	trigger.logging.Info("Initializing HTTP Trigger")
	trigger.Webserver.SetupTriggerRoute(trigger.requestHandler)
	trigger.logging.Info("HTTP Trigger Initialized")

	return nil
}
func (trigger *Trigger) requestHandler(writer http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var event models.Event
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&event)
	if err != nil {
		trigger.logging.Debug("HTTP Body not an Edgex Event")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	correlationID := r.Header.Get("X-Correlation-ID")
	edgexContext := &appcontext.Context{Configuration: trigger.Configuration,
		Trigger:       trigger,
		LoggingClient: trigger.logging,
		CorrelationID: correlationID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		trigger.logging.Error("Error marshaling data to []byte")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	envelope := &types.MessageEnvelope{
		CorrelationID: edgexContext.CorrelationID,
		Payload:       data,
	}

	trigger.Runtime.ProcessEvent(edgexContext, envelope)
	writer.Write(edgexContext.OutputData)

	trigger.outputData = nil
}
func getBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
