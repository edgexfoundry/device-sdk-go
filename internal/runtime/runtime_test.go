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
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ugorji/go/codec"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

var lc logger.LoggingClient

const (
	devID1        = "id1"
	devID2        = "id2"
	readingName1  = "sensor1"
	readingValue1 = "123.45"
)

func init() {
	lc = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
}
func TestProcessEventNoTransforms(t *testing.T) {
	// Event from device 1

	eventIn := models.Event{
		Device: devID1,
	}
	eventInBytes, _ := json.Marshal(eventIn)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       eventInBytes,
		ContentType:   clients.ContentTypeJSON,
	}
	context := &appcontext.Context{
		LoggingClient: lc,
	}
	runtime := GolangRuntime{}

	result := runtime.ProcessEvent(context, envelope)
	if result != nil {
		t.Fatal("result should be nil since no transforms have been passed")
	}
}
func TestProcessEventOneCustomTransform(t *testing.T) {
	// Event from device 1

	eventIn := models.Event{
		Device: devID1,
	}
	eventInBytes, _ := json.Marshal(eventIn)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       eventInBytes,
		ContentType:   clients.ContentTypeJSON,
	}
	context := &appcontext.Context{
		LoggingClient: lc,
	}
	transform1WasCalled := false
	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		if len(params) != 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(*models.Event); ok {
			if ok == false {
				t.Fatal("Should have receieved CoreData event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected CoreData event")
			}
		}
		transform1WasCalled = true
		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.SetTransforms([]appcontext.AppFunction{transform1})
	result := runtime.ProcessEvent(context, envelope)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
}
func TestProcessEventTwoCustomTransforms(t *testing.T) {
	// Event from device 1

	eventIn := models.Event{
		Device: devID1,
	}
	eventInBytes, _ := json.Marshal(eventIn)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       eventInBytes,
		ContentType:   clients.ContentTypeJSON,
	}
	context := &appcontext.Context{
		LoggingClient: lc,
	}
	transform1WasCalled := false
	transform2WasCalled := false

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true
		if len(params) != 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(*models.Event); ok {
			if ok == false {
				t.Fatal("Should have receieved CoreData event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected CoreData event")
			}
		}

		return true, "Transform1Result"
	}
	transform2 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform2WasCalled = true

		if params[0] != "Transform1Result" {
			t.Fatal("Did not recieve result from previous transform")
		}
		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.SetTransforms([]appcontext.AppFunction{transform1, transform2})

	result := runtime.ProcessEvent(context, envelope)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
	if transform2WasCalled == false {
		t.Fatal("transform2 should have been called")
	}
}
func TestProcessEventThreeCustomTransformsOneFail(t *testing.T) {
	// Event from device 1

	eventIn := models.Event{
		Device: devID1,
	}
	eventInBytes, _ := json.Marshal(eventIn)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       eventInBytes,
		ContentType:   clients.ContentTypeJSON,
	}
	context := &appcontext.Context{
		LoggingClient: lc,
	}
	transform1WasCalled := false
	transform2WasCalled := false
	transform3WasCalled := false

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true
		if len(params) != 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(*models.Event); ok {
			if ok == false {
				t.Fatal("Should have receieved CoreData event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected CoreData event")
			}
		}

		return false, errors.New("Transform1Result")
	}
	transform2 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform2WasCalled = true

		if params[0] != "Transform1Result" {
			t.Fatal("Did not recieve result from previous transform")
		}
		return true, "Hello"
	}
	transform3 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform3WasCalled = true

		if params[0] != "Transform1Result" {
			t.Fatal("Did not recieve result from previous transform")
		}
		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.SetTransforms([]appcontext.AppFunction{transform1, transform2, transform3})

	result := runtime.ProcessEvent(context, envelope)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
	if transform2WasCalled == true {
		t.Fatal("transform2 should NOT have been called")
	}
	if transform3WasCalled == true {
		t.Fatal("transform3 should NOT have been called")
	}
}

func TestProcessEventJSON(t *testing.T) {
	// Event from device 1
	expectedEventID := "1234"
	expectedCorrelationID := "123-234-345-456"
	eventIn := models.Event{
		ID:     expectedEventID,
		Device: devID1,
	}

	transform1WasCalled := false

	eventInBytes, _ := json.Marshal(eventIn)
	envelope := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       eventInBytes,
		ContentType:   clients.ContentTypeJSON,
	}

	context := &appcontext.Context{
		LoggingClient: lc,
	}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true

		if !assert.Equal(t, expectedEventID, edgexcontext.EventID, "Context doesn't contain expected EventID") {
			t.Fatal()
		}

		if !assert.Equal(t, expectedCorrelationID, edgexcontext.CorrelationID, "Context doesn't contain expected CorrelationID") {
			t.Fatal()
		}

		if result, ok := params[0].(*models.Event); ok {
			if !assert.True(t, ok, "Should have received CoreData event") {
				t.Fatal()
			}

			assert.Equal(t, devID1, result.Device, "Did not receive expected CoreData event, wrong device")
			assert.Equal(t, expectedEventID, result.ID, "Did not receive expected CoreData event, wrong ID")
		}

		return false, nil
	}

	runtime := GolangRuntime{}
	runtime.SetTransforms([]appcontext.AppFunction{transform1})

	result := runtime.ProcessEvent(context, envelope)
	if !assert.Nil(t, result, "result should be null") {
		t.Fatal()
	}

	if !assert.True(t, transform1WasCalled, "transform1 should have been called") {
		t.Fatal()
	}
}

func TestProcessEventCBOR(t *testing.T) {
	// Event from device 1
	expectedEventID := "6789"
	expectedCorrelationID := "123-234-345-456"
	expectedChecksum := "1234567890"
	eventIn := models.Event{
		ID:     expectedEventID,
		Device: devID1,
	}

	transform1WasCalled := false

	buffer := &bytes.Buffer{}
	handle := &codec.CborHandle{}
	encoder := codec.NewEncoder(buffer, handle)
	encoder.Encode(eventIn)

	envelope := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       buffer.Bytes(),
		ContentType:   clients.ContentTypeCBOR,
		Checksum:      expectedChecksum,
	}

	context := &appcontext.Context{
		LoggingClient: lc,
	}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true

		if !assert.Equal(t, expectedChecksum, edgexcontext.EventChecksum, "Context doesn't contain expected EventChecksum") {
			t.Fatal()
		}

		if !assert.Equal(t, expectedCorrelationID, edgexcontext.CorrelationID, "Context doesn't contain expected CorrelationID") {
			t.Fatal()
		}

		if result, ok := params[0].(*models.Event); ok {
			if !assert.True(t, ok, "Should have received CoreData event") {
				t.Fatal()
			}

			assert.Equal(t, devID1, result.Device, "Did not receive expected CoreData event, wrong device")
			assert.Equal(t, expectedEventID, result.ID, "Did not receive expected CoreData event, wrong ID")
		}

		return false, nil
	}

	runtime := GolangRuntime{}
	runtime.SetTransforms([]appcontext.AppFunction{transform1})

	result := runtime.ProcessEvent(context, envelope)
	if !assert.Nil(t, result, "result should be null") {
		t.Fatal()
	}

	if !assert.True(t, transform1WasCalled, "transform1 should have been called") {
		t.Fatal()
	}
}
