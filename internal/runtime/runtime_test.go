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
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ugorji/go/codec"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

var lc logger.LoggingClient

const (
	devID1     = "id1"
	serviceKey = "AppService-UnitTest"
)

func init() {
	lc = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
}
func TestProcessMessageNoTransforms(t *testing.T) {
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
	runtime.Initialize(nil)

	result := runtime.ProcessMessage(context, envelope)
	if result != nil {
		t.Fatal("result should be nil since no transforms have been passed")
	}
}
func TestProcessMessageOneCustomTransform(t *testing.T) {
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
		if len(params) < 1 {
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
	runtime.Initialize(nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})
	result := runtime.ProcessMessage(context, envelope)
	if result != nil {
		t.Fatal("result should be null")
	}
	if transform1WasCalled == false {
		t.Fatal("transform1 should have been called")
	}
}

func TestProcessMessageTwoCustomTransforms(t *testing.T) {
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
		if len(params) < 1 {
			t.Fatal("should have been passed the first event from CoreData")
		}
		if result, ok := params[0].(models.Event); ok {
			if ok == false {
				t.Fatal("Should have received Event")
			}

			if result.Device != devID1 {
				t.Fatal("Did not receive expected Event")
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
	runtime.Initialize(nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1, transform2})

	result := runtime.ProcessMessage(context, envelope)
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

func TestProcessMessageThreeCustomTransformsOneFail(t *testing.T) {
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
		if len(params) < 1 {
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

		return false, "Transform1Result"
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
	runtime.Initialize(nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1, transform2, transform3})

	result := runtime.ProcessMessage(context, envelope)
	if !assert.Nil(t, result) {
		t.Fatal("unexpected error")
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

func TestProcessMessageTransformError(t *testing.T) {
	// Error expected from FilterByDeviceName
	expectedError := "type received is not an Event"
	expectedErrorCode := http.StatusUnprocessableEntity

	// Send a RegistryInfo to the pipeline, instead of an Event
	registryInfo := common.RegistryInfo{
		Host: devID1,
	}
	payload, _ := json.Marshal(registryInfo)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}
	context := &appcontext.Context{
		LoggingClient: lc,
	}
	// Let the Runtime know we are sending a RegistryInfo so it passes it to the first function
	runtime := GolangRuntime{TargetType: &common.RegistryInfo{}}
	runtime.Initialize(nil)
	// FilterByDeviceName with return an error if it doesn't receive and Event
	runtime.SetTransforms([]appcontext.AppFunction{transforms.NewFilter([]string{"SomeDevice"}).FilterByDeviceName})
	err := runtime.ProcessMessage(context, envelope)

	if !assert.NotNil(t, err, "Expected an error") {
		t.Fatal()
	}
	if !assert.Error(t, err.Err, "Expected an error") {
		t.Fatal()
	}

	assert.Equal(t, err.Err.Error(), expectedError)
	assert.Equal(t, err.ErrorCode, expectedErrorCode)
}

func TestProcessMessageJSON(t *testing.T) {
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
		CorrelationID: expectedCorrelationID,
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
	runtime.Initialize(nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})

	result := runtime.ProcessMessage(context, envelope)
	if !assert.Nil(t, result, "result should be null") {
		t.Fatal()
	}

	if !assert.True(t, transform1WasCalled, "transform1 should have been called") {
		t.Fatal()
	}
}

func TestProcessMessageCBOR(t *testing.T) {
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
		CorrelationID: expectedCorrelationID,
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
	runtime.Initialize(nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})

	result := runtime.ProcessMessage(context, envelope)
	if !assert.Nil(t, result, "result should be null") {
		t.Fatal()
	}

	if !assert.True(t, transform1WasCalled, "transform1 should have been called") {
		t.Fatal()
	}
}

type CustomType struct {
	ID string `json:"id"`
}

// Must implement the Marshaler interface so SetOutputData will marshal it to JSON
func (custom CustomType) MarshalJSON() ([]byte, error) {
	test := struct {
		ID string `json:"id"`
	}{
		ID: custom.ID,
	}

	return json.Marshal(test)
}

func TestProcessMessageTargetType(t *testing.T) {
	eventIn := models.Event{
		Device: devID1,
	}
	eventJson, _ := json.Marshal(eventIn)

	eventCbor := &bytes.Buffer{}
	handle := &codec.CborHandle{}
	encoder := codec.NewEncoder(eventCbor, handle)
	encoder.Encode(eventIn)

	expected := CustomType{
		ID: "Id1",
	}
	customJson, _ := expected.MarshalJSON()
	byteData := []byte("This is my bytes")

	targetTypeTests := []struct {
		Name               string
		TargetType         interface{}
		Payload            []byte
		ContentType        string
		ExpectedOutputData []byte
		ErrorExpected      bool
	}{
		{"Default Nil Target Type", nil, eventJson, clients.ContentTypeJSON, eventJson, false},
		{"Event as Json", &models.Event{}, eventJson, clients.ContentTypeJSON, eventJson, false},
		{"Event as Cbor", &models.Event{}, eventCbor.Bytes(), clients.ContentTypeCBOR, eventJson, false}, // Not re-encoding as CBOR
		{"Custom Type Json", &CustomType{}, customJson, clients.ContentTypeJSON, customJson, false},
		{"Byte Slice", &[]byte{}, byteData, "application/binary", byteData, false},
		{"Target Type Not a pointer", models.Event{}, nil, "", nil, true},
	}

	for _, currentTest := range targetTypeTests {
		envelope := types.MessageEnvelope{
			CorrelationID: "123-234-345-456",
			Payload:       currentTest.Payload,
			ContentType:   currentTest.ContentType,
		}

		context := &appcontext.Context{
			LoggingClient: lc,
		}

		runtime := GolangRuntime{TargetType: currentTest.TargetType}
		runtime.Initialize(nil)
		runtime.SetTransforms([]appcontext.AppFunction{transforms.NewOutputData().SetOutputData})

		err := runtime.ProcessMessage(context, envelope)
		if currentTest.ErrorExpected {
			assert.NotNil(t, err, fmt.Sprintf("expected an error for test '%s'", currentTest.Name))
			assert.Error(t, err.Err, fmt.Sprintf("expected an error for test '%s'", currentTest.Name))
		} else {
			assert.Nil(t, err, fmt.Sprintf("unexpected error for test '%s'", currentTest.Name))
		}

		// OutputData will be nil if an error occurred in the pipeline processing the data
		assert.Equal(t, currentTest.ExpectedOutputData, context.OutputData, fmt.Sprintf("'%s' test failed", currentTest.Name))
	}
}

func TestExecutePipelinePersist(t *testing.T) {
	expectedItemCount := 1
	config := common.ConfigurationStruct{
		Writable: common.WritableInfo{
			LogLevel: "DEBUG",
			StoreAndForward: common.StoreAndForwardInfo{
				Enabled:       true,
				MaxRetryCount: 10},
		},
	}

	ctx := appcontext.Context{
		Configuration: config,
		LoggingClient: lc,
		CorrelationID: "CorrelationID",
		EventChecksum: "EventChecksum",
		EventID:       "EventID",
	}

	transformPassthru := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		return true, params[0]
	}

	runtime := GolangRuntime{ServiceKey: serviceKey}
	runtime.Initialize(creatMockStoreClient())

	httpPost := transforms.NewHTTPSender("http://nowhere", "", true).HTTPPost
	runtime.SetTransforms([]appcontext.AppFunction{transformPassthru, httpPost})
	payload := []byte("My Payload")

	// Target of this test
	err := runtime.executePipeline(payload, "", &ctx, runtime.transforms, 0, false)

	if assert.NotNil(t, err, "Error expected from export function") {
		storedObjects := mockRetrieveObjects(serviceKey)
		if assert.Equal(t, expectedItemCount, len(storedObjects), "unexpected item count") {
			assert.Equal(t, serviceKey, storedObjects[0].AppServiceKey, "AppServiceKey not as expected")
			assert.Equal(t, ctx.CorrelationID, storedObjects[0].CorrelationID, "CorrelationID not as expected")
			assert.Equal(t, ctx.EventID, storedObjects[0].EventID, "EventID not as expected")
			assert.Equal(t, ctx.EventChecksum, storedObjects[0].EventChecksum, "EventChecksum not as expected")
		}
	}
}
