//
// Copyright (c) 2021 Intel Corporation
// Copyright (c) 2021 One Track Consulting
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
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"io"
	"net/http"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

type TriggerRouteManager interface {
	SetupTriggerRoute(path string, handlerForTrigger func(http.ResponseWriter, *http.Request))
}

type TriggerResponseWriter interface {
	http.ResponseWriter
}

type TriggerRequestReader interface {
	io.ReadCloser
}

// Trigger implements Trigger to support Triggers
type Trigger struct {
	serviceBinding   trigger.ServiceBinding
	messageProcessor trigger.MessageProcessor
	RouteManager     TriggerRouteManager
}

func NewTrigger(bnd trigger.ServiceBinding, mp trigger.MessageProcessor, trm TriggerRouteManager) *Trigger {
	return &Trigger{
		serviceBinding:   bnd,
		messageProcessor: mp,
		RouteManager:     trm,
	}
}

// Initialize initializes the Trigger for logging and REST route
func (trigger *Trigger) Initialize(_ *sync.WaitGroup, _ context.Context, background <-chan interfaces.BackgroundMessage) (bootstrap.Deferred, error) {
	lc := trigger.serviceBinding.LoggingClient()

	if background != nil {
		return nil, errors.New("background publishing not supported for services using HTTP trigger")
	}

	lc.Info("Initializing HTTP Trigger")
	trigger.RouteManager.SetupTriggerRoute(internal.ApiTriggerRoute, trigger.requestHandler)
	lc.Info("HTTP Trigger Initialized")

	return nil, nil
}

func (trigger *Trigger) requestHandler(writer http.ResponseWriter, r *http.Request) {
	lc := trigger.serviceBinding.LoggingClient()
	defer func() {
		if r.Body != nil {
			_ = r.Body.Close()
		}
	}()

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

	lc.Trace("Received message from http", common.CorrelationHeader, correlationID)
	lc.Debug("Received message from http", common.ContentType, contentType)

	envelope := types.MessageEnvelope{
		CorrelationID: correlationID,
		ContentType:   contentType,
		Payload:       data,
	}

	appContext := trigger.serviceBinding.BuildContext(envelope)

	defaultPipeline := trigger.serviceBinding.GetDefaultPipeline()
	messageError := trigger.serviceBinding.ProcessMessage(appContext.(*appfunction.Context), envelope, defaultPipeline)

	if messageError != nil {
		writer.WriteHeader(messageError.ErrorCode)
		_, _ = writer.Write([]byte(messageError.Err.Error()))
	}
}

func getResponseHandler(writer http.ResponseWriter, lc logger.LoggingClient) interfaces.PipelineResponseHandler {
	return func(ctx interfaces.AppFunctionContext, pipeline *interfaces.FunctionPipeline) error {
		if len(ctx.ResponseContentType()) > 0 {
			writer.Header().Set(common.ContentType, ctx.ResponseContentType())
		}

		_, err := writer.Write(ctx.ResponseData())
		if err != nil {
			lc.Errorf("unable to write ResponseData as HTTP response: %s", err.Error())
			return err
		}

		if ctx.ResponseData() != nil {
			lc.Trace("Sent http response message", common.CorrelationHeader, ctx.CorrelationID())
		}
		return nil
	}
}
