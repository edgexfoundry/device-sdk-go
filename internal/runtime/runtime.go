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
	"strconv"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
	"github.com/ugorji/go/codec"
)

// GolangRuntime represents the golang runtime environment
type GolangRuntime struct {
	Transforms []func(*appcontext.Context, ...interface{}) (bool, interface{})
}

// ProcessEvent handles processing the event
func (gr GolangRuntime) ProcessEvent(edgexcontext *appcontext.Context, envelope types.MessageEnvelope) error {

	edgexcontext.LoggingClient.Debug("Processing Event: " + strconv.Itoa(len(gr.Transforms)) + " Transforms")
	var event models.Event

	switch envelope.ContentType {
	case clients.ContentTypeJSON:
		if err := json.Unmarshal([]byte(envelope.Payload), &event); err != nil {
			edgexcontext.LoggingClient.Error("Unable to JSON unmarshal EdgeX Event: "+err.Error(), clients.CorrelationHeader, envelope.CorrelationID)
			return nil
		}

		// Needed for Marking event as handled
		edgexcontext.EventID = event.ID

	case clients.ContentTypeCBOR:
		x := codec.CborHandle{}
		err := codec.NewDecoderBytes([]byte(envelope.Payload), &x).Decode(&event)
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

	edgexcontext.CorrelationID = envelope.CorrelationID
	edgexcontext.EventID = event.ID
	var result interface{}
	var continuePipeline = true
	for _, trxFunc := range gr.Transforms {
		if result != nil {
			continuePipeline, result = trxFunc(edgexcontext, result)
		} else {
			continuePipeline, result = trxFunc(edgexcontext, event)
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
