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
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	dbInterfaces "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/interfaces"

	"github.com/fxamacker/cbor/v2"
)

// GolangRuntime represents the golang runtime environment
type GolangRuntime struct {
	TargetType     interface{}
	ServiceKey     string
	transforms     []appcontext.AppFunction
	isBusyCopying  sync.Mutex
	storeForward   storeForwardInfo
	secretProvider interfaces.SecretProvider
}

type MessageError struct {
	Err       error
	ErrorCode int
}

// Initialize sets the internal reference to the StoreClient for use when Store and Forward is enabled
func (gr *GolangRuntime) Initialize(storeClient dbInterfaces.StoreClient, secretProvider interfaces.SecretProvider) {
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

// ProcessMessage sends the contents of the message thru the functions pipeline
func (gr *GolangRuntime) ProcessMessage(edgexcontext *appcontext.Context, envelope types.MessageEnvelope) *MessageError {
	lc := edgexcontext.LoggingClient

	if len(gr.transforms) == 0 {
		err := errors.New("No transforms configured. Please check log for errors loading pipeline")
		logError(lc, err, envelope.CorrelationID)
		return &MessageError{Err: err, ErrorCode: http.StatusInternalServerError}
	}

	lc.Debugf("Processing message %d Transforms", len(gr.transforms))

	// Default Target Type for the function pipeline is an Event DTO.
	// The Event DTO can be wrapped in an AddEventRequest DTO or just be the un-wrapped Event DTO,
	// which is handled dynamically below.
	if gr.TargetType == nil {
		gr.TargetType = &dtos.Event{}
	}

	if reflect.TypeOf(gr.TargetType).Kind() != reflect.Ptr {
		err := errors.New("TargetType must be a pointer, not a value of the target type")
		logError(lc, err, envelope.CorrelationID)
		return &MessageError{Err: err, ErrorCode: http.StatusInternalServerError}
	}

	// Must make a copy of the type so that data isn't retained between calls for custom types
	target := reflect.New(reflect.ValueOf(gr.TargetType).Elem().Type()).Interface()

	switch target.(type) {
	case *[]byte:
		lc.Debug("Pipeline is expecting raw byte data")
		target = &envelope.Payload

	case *dtos.Event:
		lc.Debug("Pipeline is expecting an AddEventRequest or Event DTO")

		// Dynamically process either AddEventRequest or Event DTO
		event, err := gr.processEventPayload(envelope, lc)
		if err != nil {
			errorCode := http.StatusInternalServerError
			if edgexErrors.Kind(err) == edgexErrors.KindContractInvalid {
				errorCode = http.StatusBadRequest
			}

			err = fmt.Errorf("unable to process payload %s", err.Error())
			logError(lc, err, envelope.CorrelationID)

			return &MessageError{Err: err, ErrorCode: errorCode}
		}

		if lc.LogLevel() == models.DebugLog {
			gr.debugLogEvent(lc, event)
		}

		target = event

	default:
		customTypeName := di.TypeInstanceToName(target)
		lc.Debugf("Pipeline is expecting a custom type of %s", customTypeName)

		// Expecting a custom type so just unmarshal into the target type.
		if err := gr.unmarshalPayload(envelope, target); err != nil {
			err = fmt.Errorf("unable to process custom object received of type '%s': %s", customTypeName, err.Error())
			logError(lc, err, envelope.CorrelationID)
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

	return gr.ExecutePipeline(target, envelope.ContentType, edgexcontext, transforms, 0, false)
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

func (gr *GolangRuntime) processEventPayload(envelope types.MessageEnvelope, lc logger.LoggingClient) (*dtos.Event, error) {

	lc.Debug("Attempting to process Payload as an AddEventRequest DTO")
	requestDto := requests.AddEventRequest{}

	// Note that DTO validation is called during the unmarshaling
	// which results in a KindContractInvalid error
	// TODO: Remove Validate() call once the CBOR unmarshalling calls validation like JSON does
	requestDtoErr := gr.unmarshalPayload(envelope, &requestDto)
	if requestDtoErr == nil && envelope.ContentType == clients.ContentTypeCBOR {
		requestDtoErr = requestDto.Validate()
	}

	if requestDtoErr == nil {
		lc.Debug("Using Event DTO from AddEventRequest DTO")

		// Determine that we have an AddEventRequest DTO
		return &requestDto.Event, nil
	}

	// Check for validation error
	if edgexErrors.Kind(requestDtoErr) != edgexErrors.KindContractInvalid {
		return nil, requestDtoErr
	}

	// KindContractInvalid indicates that we likely don't have an AddEventRequest
	// so try to process as Event
	lc.Debug("Attempting to process Payload as an Event DTO")
	event := &dtos.Event{}
	err := gr.unmarshalPayload(envelope, event)
	if err == nil {
		err = v2.Validate(event)
		if err == nil {
			lc.Debug("Using Event DTO received")
			return event, nil
		}
	}

	// Check for validation error
	if edgexErrors.Kind(err) != edgexErrors.KindContractInvalid {
		return nil, err
	}

	// KindContractInvalid indicates that we likely don't have an Event DTO
	// so try to process as V1 Event
	// TODO: Remove this V1 detection once fully switched over to V2 DTOs.
	event, err = gr.unmarshalV1EventToV2Event(envelope, lc)
	if err == nil {
		return event, nil
	}

	// Still unable to process so assume have invalid AddEventRequest DTO
	return nil, requestDtoErr
}

// TODO: Remove when completely switched to V2 Event DTO
func (gr *GolangRuntime) unmarshalV1EventToV2Event(envelope types.MessageEnvelope, lc logger.LoggingClient) (*dtos.Event, error) {
	lc.Debug("Processing payload as V1 Event model")

	var err error
	v1Event := models.Event{}

	err = gr.unmarshalPayload(envelope, &v1Event)
	if err != nil {
		return nil, err
	}

	_, err = v1Event.Validate()
	if err != nil {
		return nil, err
	}

	v2Event := dtos.Event{
		Versionable: commonDTO.NewVersionable(),
		Id:          v1Event.ID,
		DeviceName:  v1Event.Device,
		ProfileName: "Unknown",
		Created:     v1Event.Created,
		Origin:      v1Event.Origin,
		Tags:        v1Event.Tags,
	}

	// V1 Event ID may not be set if Core Data persistence is turned off
	if len(v2Event.Id) == 0 {
		v2Event.Id = uuid.NewString()
	}

	for _, v1Reading := range v1Event.Readings {
		v2Reading := dtos.BaseReading{
			Versionable:  commonDTO.NewVersionable(),
			Id:           v1Reading.Id,
			Created:      v1Reading.Created,
			Origin:       v1Reading.Origin,
			DeviceName:   v1Reading.Device,
			ResourceName: v1Reading.Name,
			ProfileName:  "Unknown",
			ValueType:    v1Reading.ValueType,
		}

		// V1 Reading ID may not be set if Core Data persistence is turned off
		if len(v2Reading.Id) == 0 {
			v2Reading.Id = uuid.NewString()
		}

		if v1Reading.ValueType == v2.ValueTypeBinary {
			v2Reading.BinaryValue = v1Reading.BinaryValue
		} else {
			v2Reading.Value = v1Reading.Value
		}

		v2Event.Readings = append(v2Event.Readings, v2Reading)
	}

	lc.Debug("Using Event DTO created from V1 Event Model")

	return &v2Event, nil
}

func (gr *GolangRuntime) unmarshalPayload(envelope types.MessageEnvelope, target interface{}) error {
	var err error

	switch envelope.ContentType {
	case clients.ContentTypeJSON:
		err = json.Unmarshal(envelope.Payload, target)

	case clients.ContentTypeCBOR:
		err = cbor.Unmarshal(envelope.Payload, target)

	default:
		err = fmt.Errorf("unsupported content-type '%s' recieved", envelope.ContentType)
	}

	return err
}

func (gr *GolangRuntime) debugLogEvent(lc logger.LoggingClient, event *dtos.Event) {
	lc.Debugf("Event Received with ProfileName=%s, DeviceName=%s and ReadingCount=%d",
		event.ProfileName,
		event.DeviceName,
		len(event.Readings))
	if len(event.Tags) > 0 {
		lc.Debugf("Event tags are: [%v]", event.Tags)
	} else {
		lc.Debug("Event has no tags")
	}

	for index, reading := range event.Readings {
		switch strings.ToLower(reading.ValueType) {
		case strings.ToLower(v2.ValueTypeBinary):
			lc.Debugf("Reading #%d received with ResourceName=%s, ValueType=%s, MediaType=%s and BinaryValue of size=`%d`",
				index+1,
				reading.ResourceName,
				reading.ValueType,
				reading.MediaType,
				len(reading.BinaryValue))
		default:
			lc.Debugf("Reading #%d received with ResourceName=%s, ValueType=%s, Value=`%s`",
				index+1,
				reading.ResourceName,
				reading.ValueType,
				reading.Value)
		}
	}
}

func logError(lc logger.LoggingClient, err error, correlationID string) {
	lc.Errorf("%s. %s=%s", err.Error(), clients.CorrelationHeader, correlationID)
}
