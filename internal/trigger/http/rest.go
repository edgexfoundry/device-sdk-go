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
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/webserver"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

// Trigger implements Trigger to support Triggers
type Trigger struct {
	Configuration common.ConfigurationStruct
	Runtime       runtime.GolangRuntime
	outputData    []byte
	logging       logger.LoggingClient
	Webserver     *webserver.WebServer
	EventClient   coredata.EventClient
}

// Initialize initializes the Trigger for logging and REST route
func (trigger *Trigger) Initialize(logger logger.LoggingClient) error {
	trigger.logging = logger
	trigger.logging.Info("Initializing HTTP Trigger")
	trigger.Webserver.SetupTriggerRoute(trigger.requestHandler)
	trigger.logging.Info("HTTP Trigger Initialized")

	return nil
}

func (trigger *Trigger) requestHandler(writer http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	contentType := r.Header.Get(clients.ContentType)

	if contentType != clients.ContentTypeJSON && contentType != clients.ContentTypeCBOR {
		trigger.logging.Debug("HTTP content type not supported", clients.ContentType, contentType)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		trigger.logging.Debug("Error reading HTTP Body", "error", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	trigger.logging.Debug("Request Body read", "byte count", len(data))

	correlationID := r.Header.Get("X-Correlation-ID")
	edgexContext := &appcontext.Context{
		Configuration: trigger.Configuration,
		Trigger:       trigger,
		LoggingClient: trigger.logging,
		CorrelationID: correlationID,
		EventClient:   trigger.EventClient,
	}

	trigger.logging.Trace("Received message from http", clients.CorrelationHeader, correlationID)
	trigger.logging.Debug("Received message from http", clients.ContentType, contentType)

	envelope := types.MessageEnvelope{
		CorrelationID: correlationID,
		ContentType:   contentType,
		Payload:       data,
	}

	trigger.Runtime.ProcessEvent(edgexContext, envelope)
	writer.Write(edgexContext.OutputData)

	if edgexContext.OutputData != nil {
		trigger.logging.Trace("Sent http response message", clients.CorrelationHeader, correlationID)
	}

	trigger.outputData = nil
}
