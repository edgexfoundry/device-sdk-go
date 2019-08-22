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
	"encoding/json"
	"reflect"
	"strconv"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
	"github.com/ugorji/go/codec"
)

// GolangRuntime represents the golang runtime environment
type GolangRuntime struct {
	TargetType    interface{}
	transforms    []appcontext.AppFunction
	isBusyCopying sync.Mutex
}

// ProcessEvent handles processing the event
func (gr *GolangRuntime) ProcessEvent(edgexcontext *appcontext.Context, envelope types.MessageEnvelope) error {

	edgexcontext.LoggingClient.Debug("Processing Event: " + strconv.Itoa(len(gr.transforms)) + " Transforms")

	if gr.TargetType == nil {
		gr.TargetType = &models.Event{}
	}
	target := gr.TargetType

	if reflect.TypeOf(target).Kind() != reflect.Ptr {
		edgexcontext.LoggingClient.Error("pipeline seed target type must be a pointer to an object of the target type, not a value of the target type.")
		return nil
	}

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
				edgexcontext.LoggingClient.Error("Unable to JSON unmarshal EdgeX Event: "+err.Error(), clients.CorrelationHeader, envelope.CorrelationID)
				return nil
			}

			event, ok := target.(*models.Event)
			if ok {
				// Needed for Marking event as handled
				edgexcontext.EventID = event.ID
			}

		case clients.ContentTypeCBOR:
			x := codec.CborHandle{}
			err := codec.NewDecoderBytes([]byte(envelope.Payload), &x).Decode(&target)
			if err != nil {
				edgexcontext.LoggingClient.Error("Unable to CBOR unmarshal EdgeX Event: "+err.Error(), clients.CorrelationHeader, envelope.CorrelationID)
				return nil
			}

			// Needed for Marking event as handled
			edgexcontext.EventChecksum = envelope.Checksum

		default:
			edgexcontext.LoggingClient.Error("'"+envelope.ContentType+"' content type for EdgeX Event not supported: ", clients.CorrelationHeader, envelope.CorrelationID)
			return nil
		}
	}

	edgexcontext.CorrelationID = envelope.CorrelationID

	// All functions expect an object, not a pointer to an object, so must use reflection to
	// dereference to pointer to the object
	target = reflect.ValueOf(target).Elem().Interface()

	var result interface{}
	var continuePipeline = true

	// Make copy of transform functions to avoid disruption of pipeline when updating the pipeline from registry
	gr.isBusyCopying.Lock()
	transforms := make([]appcontext.AppFunction, len(gr.transforms))
	copy(transforms, gr.transforms)
	gr.isBusyCopying.Unlock()

	for _, trxFunc := range transforms {
		if result != nil {
			continuePipeline, result = trxFunc(edgexcontext, result)
		} else {
			continuePipeline, result = trxFunc(edgexcontext, target, contentType)
		}
		if continuePipeline != true {
			if result != nil {
				if result, ok := result.(error); ok {
					edgexcontext.LoggingClient.Error((result).(error).Error())
				}
			}
			break
		}
	}

	return nil
}

// SetTransforms is thread safe to set transforms
func (gr *GolangRuntime) SetTransforms(transforms []appcontext.AppFunction) {
	gr.isBusyCopying.Lock()
	gr.transforms = transforms
	gr.isBusyCopying.Unlock()
}
