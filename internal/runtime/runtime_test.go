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
package runtime

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceKey = "AppService-UnitTest"
)

var testAddEventRequest = createAddEventRequest()
var testV2Event = testAddEventRequest.Event

func createAddEventRequest() requests.AddEventRequest {
	event := dtos.NewEvent("Thermostat", "FamilyRoomThermostat", "Temperature")
	_ = event.AddSimpleReading("Temperature", v2.ValueTypeInt64, int64(72))
	request := requests.NewAddEventRequest(event)
	return request
}

func TestProcessMessageBusRequest(t *testing.T) {
	expected := http.StatusBadRequest

	badRequest := testAddEventRequest
	badRequest.Event.ProfileName = ""
	badRequest.Event.DeviceName = ""
	payload, err := json.Marshal(badRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}
	context := appfunction.NewContext("testId", dic, "")

	dummyTransform := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return true, "Hello"
	}

	runtime := GolangRuntime{}
	runtime.Initialize(nil)
	runtime.SetTransforms([]interfaces.AppFunction{dummyTransform})
	result := runtime.ProcessMessage(context, envelope)
	require.NotNil(t, result)
	assert.Equal(t, expected, result.ErrorCode)
}

func TestProcessMessageNoTransforms(t *testing.T) {
	expected := http.StatusInternalServerError

	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)
	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}
	context := appfunction.NewContext("testId", dic, "")

	runtime := GolangRuntime{}
	runtime.Initialize(nil)

	result := runtime.ProcessMessage(context, envelope)
	require.NotNil(t, result)
	assert.Equal(t, expected, result.ErrorCode)
}

func TestProcessMessageOneCustomTransform(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
		ReceivedTopic: uuid.NewString(),
	}
	context := appfunction.NewContext("testId", dic, "")

	transform1WasCalled := false
	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		require.NotNil(t, data, "should have been passed the first event from CoreData")
		if result, ok := data.(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			require.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event")
		}
		transform1WasCalled = true
		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.Initialize(nil)
	runtime.SetTransforms([]interfaces.AppFunction{transform1})
	result := runtime.ProcessMessage(context, envelope)
	require.Nil(t, result)
	require.True(t, transform1WasCalled, "transform1 should have been called")

	assertEventMetadataSet(t, context, envelope)
}

func TestProcessMessageTwoCustomTransforms(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
		ReceivedTopic: uuid.NewString(),
	}
	context := appfunction.NewContext("testId", dic, "")
	transform1WasCalled := false
	transform2WasCalled := false

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform1WasCalled = true
		require.NotNil(t, data, "should have been passed the first event from CoreData")
		if result, ok := data.(dtos.Event); ok {
			require.True(t, ok, "Should have received Event")
			assert.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected Event")
		}

		return true, "Transform1Result"
	}
	transform2 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform2WasCalled = true

		require.Equal(t, "Transform1Result", data, "Did not receive result from previous transform")

		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.Initialize(nil)
	runtime.SetTransforms([]interfaces.AppFunction{transform1, transform2})

	result := runtime.ProcessMessage(context, envelope)
	require.Nil(t, result)
	assert.True(t, transform1WasCalled, "transform1 should have been called")
	assert.True(t, transform2WasCalled, "transform2 should have been called")

	assertEventMetadataSet(t, context, envelope)
}

func TestProcessMessageThreeCustomTransformsOneFail(t *testing.T) {
	payload, err := json.Marshal(testAddEventRequest)
	require.NoError(t, err)

	envelope := types.MessageEnvelope{
		CorrelationID: "123-234-345-456",
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
		ReceivedTopic: uuid.NewString(),
	}
	context := appfunction.NewContext("testId", dic, "")

	transform1WasCalled := false
	transform2WasCalled := false
	transform3WasCalled := false

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform1WasCalled = true
		require.NotNil(t, data, "should have been passed the first event from CoreData")

		if result, ok := data.(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			require.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event")
		}

		return false, "Transform1Result"
	}
	transform2 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform2WasCalled = true
		require.Equal(t, "Transform1Result", data, "Did not receive result from previous transform")
		return true, "Hello"
	}
	transform3 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform3WasCalled = true
		require.Equal(t, "Transform1Result", data, "Did not receive result from previous transform")
		return true, "Hello"
	}
	runtime := GolangRuntime{}
	runtime.Initialize(nil)
	runtime.SetTransforms([]interfaces.AppFunction{transform1, transform2, transform3})

	result := runtime.ProcessMessage(context, envelope)
	require.Nil(t, result)
	assert.True(t, transform1WasCalled, "transform1 should have been called")
	assert.False(t, transform2WasCalled, "transform2 should NOT have been called")
	assert.False(t, transform3WasCalled, "transform3 should NOT have been called")

	assertEventMetadataSet(t, context, envelope)
}

func TestProcessMessageTransformError(t *testing.T) {
	// Error expected from FilterByDeviceName
	expectedError := "FilterByDeviceName: type received is not an Event"
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
		ReceivedTopic: uuid.NewString(),
	}
	context := appfunction.NewContext("testId", dic, "")

	// Let the Runtime know we are sending a RegistryInfo so it passes it to the first function
	runtime := GolangRuntime{TargetType: &config.RegistryInfo{}}
	runtime.Initialize(nil)
	// FilterByDeviceName with return an error if it doesn't receive and Event
	runtime.SetTransforms([]interfaces.AppFunction{transforms.NewFilterFor([]string{"SomeDevice"}).FilterByDeviceName})
	err := runtime.ProcessMessage(context, envelope)

	require.NotNil(t, err, "Expected an error")
	require.Error(t, err.Err, "Expected an error")
	assert.Equal(t, expectedError, err.Err.Error())
	assert.Equal(t, expectedErrorCode, err.ErrorCode)

	assertReceivedTopicSet(t, context, envelope)
}

func assertEventMetadataSet(t *testing.T, context *appfunction.Context, envelope types.MessageEnvelope) {
	assertReceivedTopicSet(t, context, envelope)

	v, f := context.GetValue(interfaces.DEVICENAME)
	require.True(t, f)
	assert.Equal(t, testAddEventRequest.Event.DeviceName, v)

	v, f = context.GetValue(interfaces.PROFILENAME)
	require.True(t, f)
	assert.Equal(t, testAddEventRequest.Event.ProfileName, v)

	v, f = context.GetValue(interfaces.SOURCENAME)
	require.True(t, f)
	assert.Equal(t, testAddEventRequest.Event.SourceName, v)
}

func assertReceivedTopicSet(t *testing.T, context *appfunction.Context, envelope types.MessageEnvelope) {
	v, f := context.GetValue(interfaces.RECEIVEDTOPIC)
	require.True(t, f)
	assert.Equal(t, envelope.ReceivedTopic, v)
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

	context := appfunction.NewContext("testing", dic, "")

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform1WasCalled = true

		require.Equal(t, expectedCorrelationID, appContext.CorrelationID(), "Context doesn't contain expected CorrelationID")

		if result, ok := data.(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			assert.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event, wrong device")
			assert.Equal(t, testV2Event.Id, result.Id, "Did not receive expected EdgeX event, wrong ID")
		}

		return false, nil
	}

	runtime := GolangRuntime{}
	runtime.Initialize(nil)
	runtime.SetTransforms([]interfaces.AppFunction{transform1})

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

	context := appfunction.NewContext("testing", dic, "")

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		transform1WasCalled = true

		require.Equal(t, expectedCorrelationID, appContext.CorrelationID(), "Context doesn't contain expected CorrelationID")

		if result, ok := data.(*dtos.Event); ok {
			require.True(t, ok, "Should have received EdgeX event")
			assert.Equal(t, testV2Event.DeviceName, result.DeviceName, "Did not receive expected EdgeX event, wrong device")
			assert.Equal(t, testV2Event.Id, result.Id, "Did not receive expected EdgeX event, wrong ID")
		}

		return false, nil
	}

	runtime := GolangRuntime{}
	runtime.Initialize(nil)
	runtime.SetTransforms([]interfaces.AppFunction{transform1})

	result := runtime.ProcessMessage(context, envelope)
	assert.Nil(t, result, "result should be null")
	assert.True(t, transform1WasCalled, "transform1 should have been called")
}

type CustomType struct {
	ID string `json:"id"`
}

// Must implement the Marshaller interface so SetResponseData will marshal it to JSON
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

			context := appfunction.NewContext("testing", dic, "")

			runtime := GolangRuntime{TargetType: currentTest.TargetType}
			runtime.Initialize(nil)
			runtime.SetTransforms([]interfaces.AppFunction{transforms.NewResponseData().SetResponseData})

			err := runtime.ProcessMessage(context, envelope)
			if currentTest.ErrorExpected {
				assert.NotNil(t, err, fmt.Sprintf("expected an error for test '%s'", currentTest.Name))
				assert.Error(t, err.Err, fmt.Sprintf("expected an error for test '%s'", currentTest.Name))
			} else {
				assert.Nil(t, err, fmt.Sprintf("unexpected error for test '%s'", currentTest.Name))
			}

			// ResponseData will be nil if an error occurred in the pipeline processing the data
			assert.Equal(t, currentTest.ExpectedOutputData, context.ResponseData(), fmt.Sprintf("'%s' test failed", currentTest.Name))

			switch currentTest.TargetType.(type) {
			case nil:
				assertEventMetadataSet(t, context, envelope)
			case *dtos.Event:
				assertEventMetadataSet(t, context, envelope)
			default:
				assertReceivedTopicSet(t, context, envelope)
			}
		})
	}
}

func TestExecutePipelinePersist(t *testing.T) {
	expectedItemCount := 1

	context := appfunction.NewContext("testing", dic, "")
	transformPassthru := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return true, data
	}

	runtime := GolangRuntime{ServiceKey: serviceKey}
	runtime.Initialize(updateDicWithMockStoreClient())

	httpPost := transforms.NewHTTPSender("http://nowhere", "", true).HTTPPost
	runtime.SetTransforms([]interfaces.AppFunction{transformPassthru, httpPost})
	payload := []byte("My Payload")

	// Target of this test
	actual := runtime.ExecutePipeline(payload, "", context, runtime.transforms, 0, false)

	require.NotNil(t, actual)
	require.Error(t, actual.Err, "Error expected from export function")
	storedObjects := mockRetrieveObjects(serviceKey)
	require.Equal(t, expectedItemCount, len(storedObjects), "unexpected item count")
	assert.Equal(t, serviceKey, storedObjects[0].AppServiceKey, "AppServiceKey not as expected")
	assert.Equal(t, context.CorrelationID(), storedObjects[0].CorrelationID, "CorrelationID not as expected")
}

func TestGolangRuntime_processEventPayload(t *testing.T) {
	jsonV2AddEventPayload, _ := json.Marshal(testAddEventRequest)
	cborV2AddEventPayload, _ := cbor.Marshal(testAddEventRequest)
	jsonV2EventPayload, _ := json.Marshal(testAddEventRequest.Event)
	cborV2EventPayload, _ := cbor.Marshal(testAddEventRequest.Event)

	notAnEvent := dtos.DeviceResource{
		Description: "Not An Event",
		Name:        "SomeResource",
	}
	jsonInvalidPayload, _ := json.Marshal(notAnEvent)
	cborInvalidPayload, _ := cbor.Marshal(notAnEvent)

	expectedV2Event := testV2Event

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
		{"invalid JSON", jsonInvalidPayload, clients.ContentTypeJSON, nil, true},
		{"invalid CBOR", cborInvalidPayload, clients.ContentTypeCBOR, nil, true},
	}

	target := GolangRuntime{}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			envelope := types.MessageEnvelope{}
			envelope.Payload = testCase.Payload
			envelope.ContentType = testCase.ContentType

			actual, err := target.processEventPayload(envelope, logger.NewMockClient())
			if testCase.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.Expected, actual)
		})
	}
}
