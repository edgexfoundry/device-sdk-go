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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/webserver"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// Trigger implements Trigger to support Triggers
type Trigger struct {
	Configuration *common.ConfigurationStruct
	Runtime       *runtime.GolangRuntime
	outputData    []byte
	Webserver     *webserver.WebServer
	EdgeXClients  common.EdgeXClients
}

// Initialize initializes the Trigger for logging and REST route
func (trigger *Trigger) Initialize(appWg *sync.WaitGroup, appCtx context.Context, background <-chan types.MessageEnvelope) (bootstrap.Deferred, error) {
	logger := trigger.EdgeXClients.LoggingClient

	if background != nil {
		return nil, errors.New("background publishing not supported for services using HTTP trigger")
	}

	logger.Info("Initializing HTTP Trigger")
	trigger.Webserver.SetupTriggerRoute(internal.ApiTriggerRoute, trigger.requestHandler)
	logger.Info("HTTP Trigger Initialized")

	return nil, nil
}

func (trigger *Trigger) requestHandler(writer http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	logger := trigger.EdgeXClients.LoggingClient
	contentType := r.Header.Get(clients.ContentType)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("Error reading HTTP Body", "error", err)
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf("Error reading HTTP Body: %s", err.Error())))
		return
	}

	logger.Debug("Request Body read", "byte count", len(data))

	correlationID := r.Header.Get(internal.CorrelationHeaderKey)
	edgexContext := &appcontext.Context{
		CorrelationID:         correlationID,
		Configuration:         trigger.Configuration,
		LoggingClient:         trigger.EdgeXClients.LoggingClient,
		EventClient:           trigger.EdgeXClients.EventClient,
		ValueDescriptorClient: trigger.EdgeXClients.ValueDescriptorClient,
		CommandClient:         trigger.EdgeXClients.CommandClient,
		NotificationsClient:   trigger.EdgeXClients.NotificationsClient,
	}

	logger.Trace("Received message from http", clients.CorrelationHeader, correlationID)
	logger.Debug("Received message from http", clients.ContentType, contentType)

	envelope := types.MessageEnvelope{
		CorrelationID: correlationID,
		ContentType:   contentType,
		Payload:       data,
	}

	messageError := trigger.Runtime.ProcessMessage(edgexContext, envelope)
	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		writer.WriteHeader(messageError.ErrorCode)
		writer.Write([]byte(messageError.Err.Error()))
		return
	}

	if len(edgexContext.ResponseContentType) > 0 {
		writer.Header().Set(clients.ContentType, edgexContext.ResponseContentType)
	}
	writer.Write(edgexContext.OutputData)

	if edgexContext.OutputData != nil {
		logger.Trace("Sent http response message", clients.CorrelationHeader, correlationID)
	}

	trigger.outputData = nil
}
