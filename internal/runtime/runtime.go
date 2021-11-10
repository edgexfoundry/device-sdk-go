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

package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/fxamacker/cbor/v2"
)

const (
	TopicWildCard       = "#"
	TopicLevelSeparator = "/"
)

func NewFunctionPipeline(id string, topics []string, transforms []interfaces.AppFunction) interfaces.FunctionPipeline {
	pipeline := interfaces.FunctionPipeline{
		Id:         id,
		Transforms: transforms,
		Topics:     topics,
		Hash:       calculatePipelineHash(transforms),
	}

	return pipeline
}

// GolangRuntime represents the golang runtime environment
type GolangRuntime struct {
	TargetType    interface{}
	ServiceKey    string
	pipelines     map[string]*interfaces.FunctionPipeline
	isBusyCopying sync.Mutex
	storeForward  storeForwardInfo
	dic           *di.Container
}

type MessageError struct {
	Err       error
	ErrorCode int
}

// NewGolangRuntime creates and initializes the GolangRuntime instance
func NewGolangRuntime(serviceKey string, targetType interface{}, dic *di.Container) *GolangRuntime {
	gr := &GolangRuntime{
		ServiceKey: serviceKey,
		TargetType: targetType,
		dic:        dic,
		pipelines:  make(map[string]*interfaces.FunctionPipeline),
	}

	gr.storeForward.dic = dic
	gr.storeForward.runtime = gr

	return gr
}

// SetDefaultFunctionsPipeline sets the default function pipeline
func (gr *GolangRuntime) SetDefaultFunctionsPipeline(transforms []interfaces.AppFunction) {
	pipeline := gr.GetDefaultPipeline() // ensures the default pipeline exists
	gr.SetFunctionsPipelineTransforms(pipeline.Id, transforms)
}

// SetFunctionsPipelineTransforms sets the transforms for an existing function pipeline.
// Non-existent pipelines are ignored
func (gr *GolangRuntime) SetFunctionsPipelineTransforms(id string, transforms []interfaces.AppFunction) {
	lc := bootstrapContainer.LoggingClientFrom(gr.dic.Get)

	pipeline := gr.pipelines[id]
	if pipeline != nil {
		gr.isBusyCopying.Lock()
		pipeline.Transforms = transforms
		pipeline.Hash = calculatePipelineHash(transforms)
		gr.isBusyCopying.Unlock()
		lc.Infof("Transforms set for `%s` pipeline", id)
	} else {
		lc.Warnf("Unable to set transforms for `%s` pipeline: Pipeline not found", id)
	}
}

// ClearAllFunctionsPipelineTransforms clears the transforms for all existing function pipelines.
func (gr *GolangRuntime) ClearAllFunctionsPipelineTransforms() {
	gr.isBusyCopying.Lock()
	for index := range gr.pipelines {
		gr.pipelines[index].Transforms = nil
		gr.pipelines[index].Hash = ""
	}
	gr.isBusyCopying.Unlock()
}

// AddFunctionsPipeline is thread safe to set transforms
func (gr *GolangRuntime) AddFunctionsPipeline(id string, topics []string, transforms []interfaces.AppFunction) error {
	_, exists := gr.pipelines[id]
	if exists {
		return fmt.Errorf("pipeline with Id='%s' already exists", id)
	}

	gr.addFunctionsPipeline(id, topics, transforms)

	return nil
}

func (gr *GolangRuntime) addFunctionsPipeline(id string, topics []string, transforms []interfaces.AppFunction) *interfaces.FunctionPipeline {
	pipeline := NewFunctionPipeline(id, topics, transforms)
	gr.isBusyCopying.Lock()
	gr.pipelines[id] = &pipeline
	gr.isBusyCopying.Unlock()

	return &pipeline
}

// ProcessMessage sends the contents of the message through the functions pipeline
func (gr *GolangRuntime) ProcessMessage(
	appContext *appfunction.Context,
	envelope types.MessageEnvelope,
	pipeline *interfaces.FunctionPipeline) *MessageError {
	lc := appContext.LoggingClient()

	if len(pipeline.Transforms) == 0 {
		err := fmt.Errorf("no transforms configured for pipleline Id='%s'. Please check log for earlier errors loading pipeline", pipeline.Id)
		logError(lc, err, envelope.CorrelationID)
		return &MessageError{Err: err, ErrorCode: http.StatusInternalServerError}
	}

	appContext.AddValue(interfaces.RECEIVEDTOPIC, envelope.ReceivedTopic)
	appContext.AddValue(interfaces.PIPELINEID, pipeline.Id)

	lc.Debugf("Pipeline '%s' processing message %d Transforms", pipeline.Id, len(pipeline.Transforms))

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
		lc.Debug("Expecting raw byte data")
		target = &envelope.Payload

	case *dtos.Event:
		lc.Debug("Expecting an AddEventRequest or Event DTO")

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

		appContext.AddValue(interfaces.DEVICENAME, event.DeviceName)
		appContext.AddValue(interfaces.PROFILENAME, event.ProfileName)
		appContext.AddValue(interfaces.SOURCENAME, event.SourceName)

		target = event

	default:
		customTypeName := di.TypeInstanceToName(target)
		lc.Debugf("Expecting a custom type of %s", customTypeName)

		// Expecting a custom type so just unmarshal into the target type.
		if err := gr.unmarshalPayload(envelope, target); err != nil {
			err = fmt.Errorf("unable to process custom object received of type '%s': %s", customTypeName, err.Error())
			logError(lc, err, envelope.CorrelationID)
			return &MessageError{Err: err, ErrorCode: http.StatusBadRequest}
		}
	}

	appContext.SetCorrelationID(envelope.CorrelationID)

	// All functions expect an object, not a pointer to an object, so must use reflection to
	// dereference to pointer to the object
	target = reflect.ValueOf(target).Elem().Interface()

	// Make copy of transform functions to avoid disruption of pipeline when updating the pipeline from registry
	gr.isBusyCopying.Lock()
	execPipeline := &interfaces.FunctionPipeline{
		Id:         pipeline.Id,
		Transforms: make([]interfaces.AppFunction, len(pipeline.Transforms)),
		Topics:     pipeline.Topics,
		Hash:       pipeline.Hash,
	}
	copy(execPipeline.Transforms, pipeline.Transforms)
	gr.isBusyCopying.Unlock()

	return gr.ExecutePipeline(target, envelope.ContentType, appContext, execPipeline, 0, false)
}

func (gr *GolangRuntime) ExecutePipeline(
	target interface{},
	contentType string,
	appContext *appfunction.Context,
	pipeline *interfaces.FunctionPipeline,
	startPosition int,
	isRetry bool) *MessageError {

	var result interface{}
	var continuePipeline bool

	for functionIndex, trxFunc := range pipeline.Transforms {
		if functionIndex < startPosition {
			continue
		}

		appContext.SetRetryData(nil)

		if result == nil {
			appContext.SetInputContentType(contentType)
			continuePipeline, result = trxFunc(appContext, target)
		} else {
			continuePipeline, result = trxFunc(appContext, result)
		}

		if !continuePipeline {
			if result != nil {
				if err, ok := result.(error); ok {
					appContext.LoggingClient().Errorf(
						"Pipeline (%s) function #%d resulted in error: %s (%s=%s)",
						pipeline.Id,
						functionIndex,
						err.Error(),
						common.CorrelationHeader,
						appContext.CorrelationID())
					if appContext.RetryData() != nil && !isRetry {
						gr.storeForward.storeForLaterRetry(appContext.RetryData(), appContext, pipeline, functionIndex)
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
	serviceKey string) {

	gr.storeForward.startStoreAndForwardRetryLoop(appWg, appCtx, enabledWg, enabledCtx, serviceKey)
}

func (gr *GolangRuntime) processEventPayload(envelope types.MessageEnvelope, lc logger.LoggingClient) (*dtos.Event, error) {

	lc.Debug("Attempting to process Payload as an AddEventRequest DTO")
	requestDto := requests.AddEventRequest{}

	// Note that DTO validation is called during the unmarshaling
	// which results in a KindContractInvalid error
	requestDtoErr := gr.unmarshalPayload(envelope, &requestDto)
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
		err = common.Validate(event)
		if err == nil {
			lc.Debug("Using Event DTO received")
			return event, nil
		}
	}

	// Check for validation error
	if edgexErrors.Kind(err) != edgexErrors.KindContractInvalid {
		return nil, err
	}

	// Still unable to process so assume have invalid AddEventRequest DTO
	return nil, requestDtoErr
}

func (gr *GolangRuntime) unmarshalPayload(envelope types.MessageEnvelope, target interface{}) error {
	var err error

	contentType := strings.Split(envelope.ContentType, ";")[0]

	switch contentType {
	case common.ContentTypeJSON:
		err = json.Unmarshal(envelope.Payload, target)

	case common.ContentTypeCBOR:
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
		case strings.ToLower(common.ValueTypeBinary):
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

func (gr *GolangRuntime) GetDefaultPipeline() *interfaces.FunctionPipeline {
	pipeline := gr.pipelines[interfaces.DefaultPipelineId]
	if pipeline == nil {
		pipeline = gr.addFunctionsPipeline(interfaces.DefaultPipelineId, []string{TopicWildCard}, nil)
	}
	return pipeline
}

func (gr *GolangRuntime) GetMatchingPipelines(incomingTopic string) []*interfaces.FunctionPipeline {
	var matches []*interfaces.FunctionPipeline

	if len(gr.pipelines) == 0 {
		return matches
	}

	for _, pipeline := range gr.pipelines {
		if topicMatches(incomingTopic, pipeline.Topics) {
			matches = append(matches, pipeline)
		}
	}

	return matches
}

func (gr *GolangRuntime) GetPipelineById(id string) *interfaces.FunctionPipeline {
	return gr.pipelines[id]
}

func topicMatches(incomingTopic string, pipelineTopics []string) bool {
	for _, pipelineTopic := range pipelineTopics {
		if pipelineTopic == TopicWildCard {
			return true
		}

		wildcardCount := strings.Count(pipelineTopic, TopicWildCard)
		switch wildcardCount {
		case 0:
			if incomingTopic == pipelineTopic {
				return true
			}
		default:
			pipelineLevels := strings.Split(pipelineTopic, TopicLevelSeparator)
			incomingLevels := strings.Split(incomingTopic, TopicLevelSeparator)

			if len(pipelineLevels) > len(incomingLevels) {
				continue
			}

			for index, level := range pipelineLevels {
				if level == TopicWildCard {
					incomingLevels[index] = TopicWildCard
				}
			}

			incomingWithWildCards := strings.Join(incomingLevels, "/")
			if strings.Index(incomingWithWildCards, pipelineTopic) == 0 {
				return true
			}
		}
	}
	return false
}

func calculatePipelineHash(transforms []interfaces.AppFunction) string {
	hash := "Pipeline-functions: "
	for _, item := range transforms {
		name := runtime.FuncForPC(reflect.ValueOf(item).Pointer()).Name()
		hash = hash + " " + name
	}

	return hash
}

func logError(lc logger.LoggingClient, err error, correlationID string) {
	lc.Errorf("%s. %s=%s", err.Error(), common.CorrelationHeader, correlationID)
}
