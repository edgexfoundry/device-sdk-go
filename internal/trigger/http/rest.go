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

package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/webserver"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// Trigger implements Trigger to support Triggers
type Trigger struct {
	dic        *di.Container
	Runtime    *runtime.GolangRuntime
	Webserver  *webserver.WebServer
	outputData []byte
}

func NewTrigger(dic *di.Container, runtime *runtime.GolangRuntime, webserver *webserver.WebServer) *Trigger {
	return &Trigger{
		dic:       dic,
		Runtime:   runtime,
		Webserver: webserver,
	}
}

// Initialize initializes the Trigger for logging and REST route
func (trigger *Trigger) Initialize(_ *sync.WaitGroup, _ context.Context, background <-chan interfaces.BackgroundMessage) (bootstrap.Deferred, error) {
	lc := bootstrapContainer.LoggingClientFrom(trigger.dic.Get)

	if background != nil {
		return nil, errors.New("background publishing not supported for services using HTTP trigger")
	}

	lc.Info("Initializing HTTP Trigger")
	trigger.Webserver.SetupTriggerRoute(internal.ApiTriggerRoute, trigger.requestHandler)
	lc.Info("HTTP Trigger Initialized")

	return nil, nil
}

func (trigger *Trigger) requestHandler(writer http.ResponseWriter, r *http.Request) {
	lc := bootstrapContainer.LoggingClientFrom(trigger.dic.Get)
	defer func() { _ = r.Body.Close() }()

	contentType := r.Header.Get(common.ContentType)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		lc.Error("Error reading HTTP Body", "error", err)
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(fmt.Sprintf("Error reading HTTP Body: %s", err.Error())))
		return
	}

	lc.Debug("Request Body read", "byte count", len(data))

	correlationID := r.Header.Get(common.CorrelationHeader)

	appContext := appfunction.NewContext(correlationID, trigger.dic, contentType)

	lc.Trace("Received message from http", common.CorrelationHeader, correlationID)
	lc.Debug("Received message from http", common.ContentType, contentType)

	envelope := types.MessageEnvelope{
		CorrelationID: correlationID,
		ContentType:   contentType,
		Payload:       data,
	}

	messageError := trigger.Runtime.ProcessMessage(appContext, envelope, trigger.Runtime.GetDefaultPipeline())
	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		writer.WriteHeader(messageError.ErrorCode)
		_, _ = writer.Write([]byte(messageError.Err.Error()))
		return
	}

	if len(appContext.ResponseContentType()) > 0 {
		writer.Header().Set(common.ContentType, appContext.ResponseContentType())
	}

	_, err = writer.Write(appContext.ResponseData())
	if err != nil {
		lc.Errorf("unable to write ResponseData as HTTP response: %s", err.Error())
		return
	}

	if appContext.ResponseData() != nil {
		lc.Trace("Sent http response message", common.CorrelationHeader, correlationID)
	}

	trigger.outputData = nil
}
