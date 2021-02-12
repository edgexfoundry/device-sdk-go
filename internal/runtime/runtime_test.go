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
package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var lc logger.LoggingClient

const (
	serviceKey = "AppService-UnitTest"
)

var testAddEventRequest = createAddEventRequest()
var testV2Event = testAddEventRequest.Event

func createAddEventRequest() requests.AddEventRequest {
	request := requests.NewAddRequest("Thermostat", "FamilyRoomThermostat")
	_ = request.Event.AddSimpleReading("Temperature", v2.ValueTypeInt64, int64(72))
	return request
}

var testV1Event = models.Event{
	ID:      testV2Event.Id,
	Device:  "FamilyRoomThermostat",
	Created: testV2Event.Created,
	Origin:  testV2Event.Origin,
	Readings: []models.Reading{
		{
			Id:        testV2Event.Readings[0].Id,
			Created:   testV2Event.Readings[0].Created,
			Origin:    testV2Event.Readings[0].Origin,
			Device:    "FamilyRoomThermostat",
			Name:      "Temperature",
			ValueType: v2.ValueTypeInt64,
			Value:     "72",
		},
	},
	Tags: nil,
}

func init() {
	lc = logger.NewMockClient()
}

func TestProcessMessageNoTransforms(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}
	context := &appcontext.Context{
		LoggingClient: lc,
	}
	runtime := GolangRuntime{}
	runtime.Initialize(nil, nil)

	result := runtime.ProcessMessage(context, envelope)
	require.Nil(t, result, "result should be nil since no transforms have been passed")
}

func TestProcessMessageOneCustomTransform(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}
	context := &appcontext.Context{
		LoggingClient: lc,
	}
	transform1WasCalled := false
	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		require.True(t, len(params) > 0, "should have been passed the first event from CoreData")
		if result, ok := params[0].(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			require.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event")
		}
		transform1WasCalled = true
		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.Initialize(nil, nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})
	result := runtime.ProcessMessage(context, envelope)
	require.Nil(t, result)
	require.True(t, transform1WasCalled, "transform1 should have been called")
}

func TestProcessMessageTwoCustomTransforms(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}
	context := &appcontext.Context{
		LoggingClient: lc,
	}
	transform1WasCalled := false
	transform2WasCalled := false

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true
		require.True(t, len(params) > 0, "should have been passed the first event from CoreData")
		if result, ok := params[0].(dtos.Event); ok {
			require.True(t, ok, "Should have received Event")
			assert.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected Event")
		}

		return true, "Transform1Result"
	}
	transform2 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform2WasCalled = true

		require.Equal(t, "Transform1Result", params[0], "Did not receive result from previous transform")

		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.Initialize(nil, nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1, transform2})

	result := runtime.ProcessMessage(context, envelope)
	require.Nil(t, result)
	assert.True(t, transform1WasCalled, "transform1 should have been called")
	assert.True(t, transform2WasCalled, "transform2 should have been called")
}

func TestProcessMessageThreeCustomTransformsOneFail(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
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
		require.True(t, len(params) > 0, "should have been passed the first event from CoreData")

		if result, ok := params[0].(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			require.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event")
		}

		return false, "Transform1Result"
	}
	transform2 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform2WasCalled = true
		require.Equal(t, "Transform1Result", params[0], "Did not receive result from previous transform")
		return true, "Hello"
	}
	transform3 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform3WasCalled = true
		require.Equal(t, "Transform1Result", params[0], "Did not receive result from previous transform")
		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.Initialize(nil, nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1, transform2, transform3})

	result := runtime.ProcessMessage(context, envelope)
	require.Nil(t, result)
	assert.True(t, transform1WasCalled, "transform1 should have been called")
	assert.False(t, transform2WasCalled, "transform2 should NOT have been called")
	assert.False(t, transform3WasCalled, "transform3 should NOT have been called")
}

func TestProcessMessageTransformError(t *testing.T) {
	// Error expected from FilterByDeviceName
	expectedError := "type received is not an Event"
	expectedErrorCode := http.StatusUnprocessableEntity

	// Send a RegistryInfo to the pipeline, instead of an Event
	registryInfo := config.RegistryInfo{
		Host: testV2Event.DeviceName,
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
	runtime := GolangRuntime{TargetType: &config.RegistryInfo{}}
	runtime.Initialize(nil, nil)
	// FilterByDeviceName with return an error if it doesn't receive and Event
	runtime.SetTransforms([]appcontext.AppFunction{transforms.NewFilterFor([]string{"SomeDevice"}).FilterByDeviceName})
	err := runtime.ProcessMessage(context, envelope)

	require.NotNil(t, err, "Expected an error")
	require.Error(t, err.Err, "Expected an error")
	assert.Equal(t, expectedError, err.Err.Error())
	assert.Equal(t, expectedErrorCode, err.ErrorCode)
}

func TestProcessMessageJSON(t *testing.T) {
	expectedCorrelationID := "123-234-345-456"

	transform1WasCalled := false

	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}

	context := &appcontext.Context{
		LoggingClient: lc,
		CorrelationID: expectedCorrelationID,
	}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true

		require.Equal(t, expectedCorrelationID, edgexcontext.CorrelationID, "Context doesn't contain expected CorrelationID")

		if result, ok := params[0].(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			assert.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event, wrong device")
			assert.Equal(t, testV2Event.Id, result.Id, "Did not receive expected EdgeX event, wrong ID")
		}

		return false, nil
	}

	runtime := GolangRuntime{}
	runtime.Initialize(nil, nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})

	result := runtime.ProcessMessage(context, envelope)
	assert.Nilf(t, result, "result should be null. Got %v", result)
	assert.True(t, transform1WasCalled, "transform1 should have been called")
}

func TestProcessMessageCBOR(t *testing.T) {
	expectedCorrelationID := "123-234-345-456"

	transform1WasCalled := false

	payload, err := cbor.Marshal(testAddEventRequest)
	assert.NoError(t, err, "expected no error when marshalling data")

	envelope := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   clients.ContentTypeCBOR,
	}

	context := &appcontext.Context{
		LoggingClient: lc,
		CorrelationID: expectedCorrelationID,
	}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transform1WasCalled = true

		require.Equal(t, expectedCorrelationID, edgexcontext.CorrelationID, "Context doesn't contain expected CorrelationID")

		if result, ok := params[0].(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			assert.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event, wrong device")
			assert.Equal(t, testV2Event.Id, result.Id, "Did not receive expected EdgeX event, wrong ID")
		}

		return false, nil
	}

	runtime := GolangRuntime{}
	runtime.Initialize(nil, nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})

	result := runtime.ProcessMessage(context, envelope)
	assert.Nil(t, result, "result should be null")
	assert.True(t, transform1WasCalled, "transform1 should have been called")
}

type CustomType struct {
	ID string `json:"id"`
}

// Must implement the Marshaller interface so SetOutputData will marshal it to JSON
func (custom CustomType) MarshalJSON() ([]byte, error) {
	test := struct {
		ID string `json:"id"`
	}{
		ID: custom.ID,
	}

	return json.Marshal(test)
}

func TestProcessMessageTargetType(t *testing.T) {
	jsonPayload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	eventJsonPayload, err := json.Marshal(testV2Event)
	require.NoError(t, err)

	cborPayload, err := cbor.Marshal(testAddEventRequest)
	assert.NoError(t, err)

	eventCborPayload, err := cbor.Marshal(testV2Event)
	require.NoError(t, err)

	expected := CustomType{
		ID: "Id1",
	}
	customJsonPayload, _ := expected.MarshalJSON()
	byteData := []byte("This is my bytes")

	targetTypeTests := []struct {
		Name               string
		TargetType         interface{}
		Payload            []byte
		ContentType        string
		ExpectedOutputData []byte
		ErrorExpected      bool
	}{
		{"JSON default Target Type", nil, jsonPayload, clients.ContentTypeJSON, eventJsonPayload, false},
		{"CBOR default Target Type", nil, cborPayload, clients.ContentTypeCBOR, eventJsonPayload, false},
		{"JSON Event Event DTO", &dtos.Event{}, eventJsonPayload, clients.ContentTypeJSON, eventJsonPayload, false},
		{"CBOR Event Event DTO", &dtos.Event{}, eventCborPayload, clients.ContentTypeCBOR, eventJsonPayload, false}, // Not re-encoding as CBOR
		{"Custom Type Json", &CustomType{}, customJsonPayload, clients.ContentTypeJSON, customJsonPayload, false},
		{"Byte Slice", &[]byte{}, byteData, "application/binary", byteData, false},
		{"Target Type Not a pointer", dtos.Event{}, nil, "", nil, true},
	}

	for _, currentTest := range targetTypeTests {
		t.Run(currentTest.Name, func(t *testing.T) {
			envelope := types.MessageEnvelope{
				CorrelationID: "123-234-345-456",
				Payload:       currentTest.Payload,
				ContentType:   currentTest.ContentType,
			}

			context := &appcontext.Context{
				LoggingClient: lc,
			}

			runtime := GolangRuntime{TargetType: currentTest.TargetType}
			runtime.Initialize(nil, nil)
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
		})
	}
}

func TestExecutePipelinePersist(t *testing.T) {
	expectedItemCount := 1
	configuration := common.ConfigurationStruct{
		Writable: common.WritableInfo{
			LogLevel: "DEBUG",
			StoreAndForward: common.StoreAndForwardInfo{
				Enabled:       true,
				MaxRetryCount: 10},
		},
	}

	ctx := appcontext.Context{
		Configuration: &configuration,
		LoggingClient: lc,
		CorrelationID: "CorrelationID",
	}

	transformPassthru := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		return true, params[0]
	}

	runtime := GolangRuntime{ServiceKey: serviceKey}
	runtime.Initialize(creatMockStoreClient(), nil)

	httpPost := transforms.NewHTTPSender("http://nowhere", "", true).HTTPPost
	runtime.SetTransforms([]appcontext.AppFunction{transformPassthru, httpPost})
	payload := []byte("My Payload")

	// Target of this test
	actual := runtime.ExecutePipeline(payload, "", &ctx, runtime.transforms, 0, false)

	require.NotNil(t, actual)
	require.Error(t, actual.Err, "Error expected from export function")
	storedObjects := mockRetrieveObjects(serviceKey)
	require.Equal(t, expectedItemCount, len(storedObjects), "unexpected item count")
	assert.Equal(t, serviceKey, storedObjects[0].AppServiceKey, "AppServiceKey not as expected")
	assert.Equal(t, ctx.CorrelationID, storedObjects[0].CorrelationID, "CorrelationID not as expected")
}

// TODO: Remove once switch completely to V2 Event DTOs
func TestProcessMessageJSONWithV1Event(t *testing.T) {
	expectedCorrelationID := "123-234-345-456"

	transform1WasCalled := false

	eventInBytes, _ := json.Marshal(testV1Event)
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

		require.Equal(t, expectedCorrelationID, edgexcontext.CorrelationID, "Context doesn't contain expected CorrelationID")

		if result, ok := params[0].(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			assert.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event, wrong device")
			assert.Equal(t, testV2Event.Id, result.Id, "Did not receive expected EdgeX event, wrong ID")
		}

		return false, nil
	}

	runtime := GolangRuntime{}
	runtime.Initialize(nil, nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})

	result := runtime.ProcessMessage(context, envelope)
	assert.Nil(t, result, "result should be null")
	assert.True(t, transform1WasCalled, "transform1 should have been called")
}

func TestGolangRuntime_processEventPayload(t *testing.T) {
	jsonV2AddEventPayload, _ := json.Marshal(testAddEventRequest)
	cborV2AddEventPayload, _ := cbor.Marshal(testAddEventRequest)
	jsonV2EventPayload, _ := json.Marshal(testAddEventRequest.Event)
	cborV2EventPayload, _ := cbor.Marshal(testAddEventRequest.Event)
	jsonV1EventPayload, _ := json.Marshal(testV1Event)
	cborV1EventPayload, _ := cbor.Marshal(testV1Event)

	notAnEvent := dtos.DeviceResource{
		Description: "Not An Event",
		Name:        "SomeResource",
	}
	jsonInvalidPayload, _ := json.Marshal(notAnEvent)
	cborInvalidPayload, _ := cbor.Marshal(notAnEvent)

	expectedV2Event := testV2Event
	expectedV2EventFromV1Event := testV2Event
	expectedV2EventFromV1Event.ProfileName = "Unknown"
	expectedV2EventFromV1Event.Readings = []dtos.BaseReading{testV2Event.Readings[0]}
	expectedV2EventFromV1Event.Readings[0].ProfileName = "Unknown"

	tests := []struct {
		Name        string
		Payload     []byte
		ContentType string
		Expected    *dtos.Event
		ExpectError bool
	}{
		{"JSON V2 Add Event DTO", jsonV2AddEventPayload, clients.ContentTypeJSON, &expectedV2Event, false},
		{"CBOR V2 Add Event DTO", cborV2AddEventPayload, clients.ContentTypeCBOR, &expectedV2Event, false},
		{"JSON V2 Event DTO", jsonV2EventPayload, clients.ContentTypeJSON, &expectedV2Event, false},
		{"CBOR V2 Event DTO", cborV2EventPayload, clients.ContentTypeCBOR, &expectedV2Event, false},
		{"JSON V1 Event", jsonV1EventPayload, clients.ContentTypeJSON, &expectedV2EventFromV1Event, false},
		{"CBOR V1 Event", cborV1EventPayload, clients.ContentTypeCBOR, &expectedV2EventFromV1Event, false},
		{"invalid JSON", jsonInvalidPayload, clients.ContentTypeJSON, nil, true},
		{"invalid CBOR", cborInvalidPayload, clients.ContentTypeCBOR, nil, true},
	}

	target := GolangRuntime{}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			envelope := types.MessageEnvelope{}
			envelope.Payload = testCase.Payload
			envelope.ContentType = testCase.ContentType

			actual, err := target.processEventPayload(envelope, lc)
			if testCase.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.Expected, actual)
		})
	}
}

func TestGolangRuntime_unmarshalV1EventToV2Event(t *testing.T) {
	target := GolangRuntime{}

	jsonV1Payload, _ := json.Marshal(testV1Event)
	cborV1Payload, _ := cbor.Marshal(testV1Event)

	expectedEvent := testV2Event
	expectedEvent.ProfileName = "Unknown"
	expectedEvent.Readings[0].ProfileName = "Unknown"

	tests := []struct {
		Name        string
		Payload     []byte
		ContentType string
	}{
		{"JSON V1 Event", jsonV1Payload, clients.ContentTypeJSON},
		{"CBOR V1 Event", cborV1Payload, clients.ContentTypeCBOR},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			envelope := types.MessageEnvelope{}
			envelope.Payload = testCase.Payload
			envelope.ContentType = testCase.ContentType

			actual, err := target.unmarshalV1EventToV2Event(envelope, lc)
			require.NoError(t, err)
			require.Equal(t, expectedEvent, *actual)
		})
	}
}
