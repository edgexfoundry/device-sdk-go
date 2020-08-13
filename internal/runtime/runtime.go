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

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"sync"

	"github.com/fxamacker/cbor/v2"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

const unmarshalErrorMessage = "Unable to unmarshal message payload as %s"

// GolangRuntime represents the golang runtime environment
type GolangRuntime struct {
	TargetType     interface{}
	ServiceKey     string
	transforms     []appcontext.AppFunction
	isBusyCopying  sync.Mutex
	storeForward   storeForwardInfo
	secretProvider security.SecretProvider
}

type MessageError struct {
	Err       error
	ErrorCode int
}

// ProcessMessage sends the contents of the message thru the functions pipeline
func (gr *GolangRuntime) ProcessMessage(edgexcontext *appcontext.Context, envelope types.MessageEnvelope) *MessageError {

	edgexcontext.LoggingClient.Debug("Processing message: " + strconv.Itoa(len(gr.transforms)) + " Transforms")

	if gr.TargetType == nil {
		gr.TargetType = &models.Event{}
	}

	if reflect.TypeOf(gr.TargetType).Kind() != reflect.Ptr {
		err := fmt.Errorf("TargetType must be a pointer, not a value of the target type.")
		edgexcontext.LoggingClient.Error(err.Error())
		return &MessageError{Err: err, ErrorCode: http.StatusInternalServerError}
	}

	// Must make a copy of the type so that data isn't retained between calls.
	target := reflect.New(reflect.ValueOf(gr.TargetType).Elem().Type()).Interface()

	// Only set when the data is binary so function receiving it knows how to deal with it.
	var contentType string

	switch target.(type) {
	case *[]byte:
		target = &envelope.Payload
		contentType = envelope.ContentType

	default:
		switch envelope.ContentType {
		case clients.ContentTypeJSON:

			if err := json.Unmarshal([]byte(envelope.Payload), target); err != nil {
				message := fmt.Sprintf(unmarshalErrorMessage, "JSON")
				edgexcontext.LoggingClient.Error(
					message, "error", err.Error(),
					clients.CorrelationHeader, envelope.CorrelationID)
				err = fmt.Errorf("%s : %s", message, err.Error())
				return &MessageError{Err: err, ErrorCode: http.StatusBadRequest}
			}

			event, ok := target.(*models.Event)
			if ok {
				// Needed for Marking event as handled
				edgexcontext.EventID = event.ID
			}

		case clients.ContentTypeCBOR:
			err := cbor.Unmarshal([]byte(envelope.Payload), target)
			if err != nil {
				message := fmt.Sprintf(unmarshalErrorMessage, "CBOR")
				edgexcontext.LoggingClient.Error(
					message, "error", err.Error(),
					clients.CorrelationHeader, envelope.CorrelationID)
				err = fmt.Errorf("%s : %s", message, err.Error())
				return &MessageError{Err: err, ErrorCode: http.StatusBadRequest}
			}

			// Needed for Marking event as handled
			edgexcontext.EventChecksum = envelope.Checksum

		default:
			message := "content type for input data not supported"
			edgexcontext.LoggingClient.Error(message,
				clients.ContentType, envelope.ContentType,
				clients.CorrelationHeader, envelope.CorrelationID)
			err := fmt.Errorf("'%s' %s", envelope.ContentType, message)
			return &MessageError{Err: err, ErrorCode: http.StatusBadRequest}
		}
	}

	edgexcontext.CorrelationID = envelope.CorrelationID

	// All functions expect an object, not a pointer to an object, so must use reflection to
	// dereference to pointer to the object
	target = reflect.ValueOf(target).Elem().Interface()

	// Make copy of transform functions to avoid disruption of pipeline when updating the pipeline from registry
	gr.isBusyCopying.Lock()
	transforms := make([]appcontext.AppFunction, len(gr.transforms))
	copy(transforms, gr.transforms)
	gr.isBusyCopying.Unlock()

	return gr.ExecutePipeline(target, contentType, edgexcontext, transforms, 0, false)

}

// Initialize sets the internal reference to the StoreClient for use when Store and Forward is enabled
func (gr *GolangRuntime) Initialize(storeClient interfaces.StoreClient, secretProvider security.SecretProvider) {
	gr.storeForward.storeClient = storeClient
	gr.storeForward.runtime = gr
	gr.secretProvider = secretProvider
}

// SetTransforms is thread safe to set transforms
func (gr *GolangRuntime) SetTransforms(transforms []appcontext.AppFunction) {
	gr.isBusyCopying.Lock()
	gr.transforms = transforms
	gr.storeForward.pipelineHash = gr.storeForward.calculatePipelineHash() // Only need to calculate hash when the pipeline changes.
	gr.isBusyCopying.Unlock()
}

func (gr *GolangRuntime) ExecutePipeline(target interface{}, contentType string, edgexcontext *appcontext.Context,
	transforms []appcontext.AppFunction, startPosition int, isRetry bool) *MessageError {

	var result interface{}
	var continuePipeline = true

	edgexcontext.SecretProvider = gr.secretProvider

	for functionIndex, trxFunc := range transforms {
		if functionIndex < startPosition {
			continue
		}

		edgexcontext.RetryData = nil

		if result == nil {
			continuePipeline, result = trxFunc(edgexcontext, target, contentType)
		} else {
			continuePipeline, result = trxFunc(edgexcontext, result)
		}

		if continuePipeline != true {
			if result != nil {
				if err, ok := result.(error); ok {
					edgexcontext.LoggingClient.Error(
						fmt.Sprintf("Pipeline function #%d resulted in error", functionIndex),
						"error", err.Error(), clients.CorrelationHeader, edgexcontext.CorrelationID)
					if edgexcontext.RetryData != nil && !isRetry {
						gr.storeForward.storeForLaterRetry(edgexcontext.RetryData, edgexcontext, functionIndex)
					}

					return &MessageError{Err: err, ErrorCode: http.StatusUnprocessableEntity}
				}
			}
			break
		}
	}

	return nil
}

func (gr *GolangRuntime) StartStoreAndForward(
	appWg *sync.WaitGroup,
	appCtx context.Context,
	enabledWg *sync.WaitGroup,
	enabledCtx context.Context,
	serviceKey string,
	config *common.ConfigurationStruct,
	edgeXClients common.EdgeXClients) {

	gr.storeForward.startStoreAndForwardRetryLoop(appWg, appCtx, enabledWg, enabledCtx, serviceKey, config, edgeXClients)
}
