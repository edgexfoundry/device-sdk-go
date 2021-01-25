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

package messagebus

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var logClient logger.LoggingClient

var expectedEvent = dtos.Event{
	Versionable: commonDTO.NewVersionable(),
	Id:          "7a1707f0-166f-4c4b-bc9d-1d54c74e0137",
	DeviceName:  "LivingRoomThermostat",
	ProfileName: "thermostat",
	Created:     1485364897029,
	Origin:      1471806386919,
	Readings: []dtos.BaseReading{
		{
			Versionable:   commonDTO.NewVersionable(),
			Id:            "7a1707f0-166f-4c4b-bc9d-1d54c74e0145",
			Created:       1485364896983,
			Origin:        1471806386919,
			DeviceName:    "LivingRoomThermostat",
			ResourceName:  "temperature",
			ProfileName:   "thermostat",
			ValueType:     v2.ValueTypeInt64,
			SimpleReading: dtos.SimpleReading{Value: "38"},
		},
	},
	Tags: nil,
}

var addEventRequest = requests.AddEventRequest{
	BaseRequest: commonDTO.BaseRequest{},
	Event:       expectedEvent,
}

func init() {
	logClient = logger.NewMockClient()
}

func TestInitialize(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:            "meSsaGebus",
			PublishTopic:    "publish",
			SubscribeTopics: "events",
		},
		MessageBus: types.MessageBusConfig{
			Type: "zero",
			PublishHost: types.HostInfo{
				Host:     "*",
				Port:     5563,
				Protocol: "tcp",
			},
			SubscribeHost: types.HostInfo{
				Host:     "localhost",
				Port:     5563,
				Protocol: "tcp",
			},
		},
	}

	goRuntime := &runtime.GolangRuntime{}

	trigger := Trigger{Configuration: &config, Runtime: goRuntime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	_, _ = trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	assert.NotNil(t, trigger.client, "Expected client to be set")
	assert.Equal(t, 1, len(trigger.topics))
	assert.Equal(t, "events", trigger.topics[0].Topic)
	assert.NotNil(t, trigger.topics[0].Messages)
}

func TestInitializeBadConfiguration(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:            "meSsaGebus",
			PublishTopic:    "publish",
			SubscribeTopics: "events",
		},
		MessageBus: types.MessageBusConfig{
			Type: "aaaa", //as type is not "zero", should return an error on client initialization
			PublishHost: types.HostInfo{
				Host:     "*",
				Port:     5568,
				Protocol: "tcp",
			},
			SubscribeHost: types.HostInfo{
				Host:     "localhost",
				Port:     5568,
				Protocol: "tcp",
			},
		},
	}

	goRuntime := &runtime.GolangRuntime{}

	trigger := Trigger{Configuration: &config, Runtime: goRuntime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	assert.Error(t, err)
}

func TestInitializeAndProcessEventWithNoOutput(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:            "meSsaGebus",
			PublishTopic:    "",
			SubscribeTopics: "",
		},
		MessageBus: types.MessageBusConfig{
			Type: "zero",
			PublishHost: types.HostInfo{
				Host:     "*",
				Port:     5566,
				Protocol: "tcp",
			},
			SubscribeHost: types.HostInfo{
				Host:     "localhost",
				Port:     5564,
				Protocol: "tcp",
			},
		},
	}

	expectedCorrelationID := "123"

	transformWasCalled := common.AtomicBool{}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transformWasCalled.Set(true)
		assert.Equal(t, expectedEvent, params[0])
		return false, nil
	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(nil, nil)
	goRuntime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: &config, Runtime: goRuntime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	_, _ = trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)

	payload, err := json.Marshal(addEventRequest)
	require.NoError(t, err)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}

	testClientConfig := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     5564,
			Protocol: "tcp",
		},
		Type: "zero",
	}

	testClient, err := messaging.NewMessageClient(testClientConfig)
	require.NoError(t, err, "Unable to create to publisher")
	assert.False(t, transformWasCalled.Value())

	err = testClient.Publish(message, "") //transform1 should be called after this executes
	require.NoError(t, err, "Failed to publish message")

	time.Sleep(3 * time.Second)
	assert.True(t, transformWasCalled.Value(), "Transform never called")
}

func TestInitializeAndProcessEventWithOutput(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:            "meSsaGebus",
			PublishTopic:    "PublishTopic",
			SubscribeTopics: "SubscribeTopic",
		},
		MessageBus: types.MessageBusConfig{
			Type: "zero",
			PublishHost: types.HostInfo{
				Host:     "*",
				Port:     5586,
				Protocol: "tcp",
			},
			SubscribeHost: types.HostInfo{
				Host:     "localhost",
				Port:     5584,
				Protocol: "tcp",
			},
		},
	}

	responseContentType := uuid.New().String()

	expectedCorrelationID := "123"

	transformWasCalled := common.AtomicBool{}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transformWasCalled.Set(true)
		assert.Equal(t, expectedEvent, params[0])
		edgexcontext.ResponseContentType = responseContentType
		edgexcontext.Complete([]byte("Transformed")) //transformed message published to message bus
		return false, nil

	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(nil, nil)
	goRuntime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: &config, Runtime: goRuntime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}

	testClientConfig := types.MessageBusConfig{
		SubscribeHost: types.HostInfo{
			Host:     "localhost",
			Port:     5586,
			Protocol: "tcp",
		},
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     5584,
			Protocol: "tcp",
		},
		Type: "zero",
	}
	testClient, err := messaging.NewMessageClient(testClientConfig) //new client to publish & subscribe
	require.NoError(t, err, "Failed to create test client")

	testTopics := []types.TopicChannel{{Topic: trigger.Configuration.Binding.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
	testMessageErrors := make(chan error)

	err = testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus
	require.NoError(t, err)
	_, err = trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)

	payload, err := json.Marshal(addEventRequest)
	require.NoError(t, err)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}

	assert.False(t, transformWasCalled.Value())
	err = testClient.Publish(message, "SubscribeTopic")
	require.NoError(t, err, "Failed to publish message")

	time.Sleep(3 * time.Second)
	require.True(t, transformWasCalled.Value(), "Transform never called")

	receiveMessage := true

	for receiveMessage {
		select {
		case msgErr := <-testMessageErrors:
			receiveMessage = false
			assert.Error(t, msgErr)
		case msgs := <-testTopics[0].Messages:
			receiveMessage = false
			assert.Equal(t, "Transformed", string(msgs.Payload))
			assert.Equal(t, responseContentType, msgs.ContentType)
		}
	}
}

func TestInitializeAndProcessEventWithOutput_InferJSON(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:            "meSsaGebus",
			PublishTopic:    "PublishTopic",
			SubscribeTopics: "SubscribeTopic",
		},
		MessageBus: types.MessageBusConfig{
			Type: "zero",
			PublishHost: types.HostInfo{
				Host:     "*",
				Port:     5701,
				Protocol: "tcp",
			},
			SubscribeHost: types.HostInfo{
				Host:     "localhost",
				Port:     5702,
				Protocol: "tcp",
			},
		},
	}

	expectedCorrelationID := "123"

	transformWasCalled := common.AtomicBool{}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transformWasCalled.Set(true)
		assert.Equal(t, expectedEvent, params[0])
		edgexcontext.Complete([]byte("{;)Transformed")) //transformed message published to message bus
		return false, nil

	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(nil, nil)
	goRuntime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: &config, Runtime: goRuntime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}

	testClientConfig := types.MessageBusConfig{
		SubscribeHost: types.HostInfo{
			Host:     "localhost",
			Port:     5701,
			Protocol: "tcp",
		},
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     5702,
			Protocol: "tcp",
		},
		Type: "zero",
	}
	testClient, err := messaging.NewMessageClient(testClientConfig) //new client to publish & subscribe
	require.NoError(t, err, "Failed to create test client")

	testTopics := []types.TopicChannel{{Topic: trigger.Configuration.Binding.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
	testMessageErrors := make(chan error)

	err = testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus
	require.NoError(t, err)
	_, err = trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)

	payload, err := json.Marshal(addEventRequest)
	require.NoError(t, err)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}

	assert.False(t, transformWasCalled.Value())
	err = testClient.Publish(message, "SubscribeTopic")
	require.NoError(t, err, "Failed to publish message")

	time.Sleep(3 * time.Second)
	require.True(t, transformWasCalled.Value(), "Transform never called")

	receiveMessage := true

	for receiveMessage {
		select {
		case msgErr := <-testMessageErrors:
			receiveMessage = false
			assert.Error(t, msgErr)
		case msgs := <-testTopics[0].Messages:
			receiveMessage = false
			assert.Equal(t, "{;)Transformed", string(msgs.Payload))
			assert.Equal(t, clients.ContentTypeJSON, msgs.ContentType)
		}
	}
}

func TestInitializeAndProcessEventWithOutput_AssumeCBOR(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:            "meSsaGebus",
			PublishTopic:    "PublishTopic",
			SubscribeTopics: "SubscribeTopic",
		},
		MessageBus: types.MessageBusConfig{
			Type: "zero",
			PublishHost: types.HostInfo{
				Host:     "*",
				Port:     5703,
				Protocol: "tcp",
			},
			SubscribeHost: types.HostInfo{
				Host:     "localhost",
				Port:     5704,
				Protocol: "tcp",
			},
		},
	}

	expectedCorrelationID := "123"

	transformWasCalled := common.AtomicBool{}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transformWasCalled.Set(true)
		assert.Equal(t, expectedEvent, params[0])
		edgexcontext.Complete([]byte("Transformed")) //transformed message published to message bus
		return false, nil
	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(nil, nil)
	goRuntime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: &config, Runtime: goRuntime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}

	testClientConfig := types.MessageBusConfig{
		SubscribeHost: types.HostInfo{
			Host:     "localhost",
			Port:     5703,
			Protocol: "tcp",
		},
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     5704,
			Protocol: "tcp",
		},
		Type: "zero",
	}
	testClient, err := messaging.NewMessageClient(testClientConfig) //new client to publish & subscribe
	require.NoError(t, err, "Failed to create test client")

	testTopics := []types.TopicChannel{{Topic: trigger.Configuration.Binding.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
	testMessageErrors := make(chan error)

	err = testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus
	require.NoError(t, err)
	_, _ = trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)

	payload, _ := json.Marshal(addEventRequest)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}

	assert.False(t, transformWasCalled.Value())
	err = testClient.Publish(message, "SubscribeTopic")
	require.NoError(t, err, "Failed to publish message")

	time.Sleep(3 * time.Second)
	require.True(t, transformWasCalled.Value(), "Transform never called")

	receiveMessage := true

	for receiveMessage {
		select {
		case msgErr := <-testMessageErrors:
			receiveMessage = false
			assert.Error(t, msgErr)
		case msgs := <-testTopics[0].Messages:
			receiveMessage = false
			assert.Equal(t, "Transformed", string(msgs.Payload))
			assert.Equal(t, clients.ContentTypeCBOR, msgs.ContentType)
		}
	}
}

func TestInitializeAndProcessBackgroundMessage(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:            "meSsaGebus",
			PublishTopic:    "PublishTopic",
			SubscribeTopics: "SubscribeTopic",
		},
		MessageBus: types.MessageBusConfig{
			Type: "zero",
			PublishHost: types.HostInfo{
				Host:     "*",
				Port:     5588,
				Protocol: "tcp",
			},
			SubscribeHost: types.HostInfo{
				Host:     "localhost",
				Port:     5590,
				Protocol: "tcp",
			},
		},
	}

	expectedCorrelationID := "123"

	expectedPayload := []byte(`{"id":"5888dea1bd36573f4681d6f9","created":1485364897029,"modified":1485364897029,"origin":1471806386919,"pushed":0,"device":"livingroomthermostat","readings":[{"id":"5888dea0bd36573f4681d6f8","created":1485364896983,"modified":1485364896983,"origin":1471806386919,"pushed":0,"name":"temperature","value":"38","device":"livingroomthermostat"}]}`)

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(nil, nil)
	trigger := Trigger{Configuration: &config, Runtime: goRuntime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}

	testClientConfig := types.MessageBusConfig{
		SubscribeHost: types.HostInfo{
			Host:     "localhost",
			Port:     5588,
			Protocol: "tcp",
		},
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     5590,
			Protocol: "tcp",
		},
		Type: "zero",
	}
	testClient, err := messaging.NewMessageClient(testClientConfig) //new client to publish & subscribe
	require.NoError(t, err, "Failed to create test client")

	testTopics := []types.TopicChannel{{Topic: trigger.Configuration.Binding.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
	testMessageErrors := make(chan error)

	err = testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus
	require.NoError(t, err)

	background := make(chan types.MessageEnvelope)

	_, _ = trigger.Initialize(&sync.WaitGroup{}, context.Background(), background)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       expectedPayload,
		ContentType:   clients.ContentTypeJSON,
	}

	background <- message

	receiveMessage := true

	for receiveMessage {
		select {
		case msgErr := <-testMessageErrors:
			receiveMessage = false
			assert.Error(t, msgErr)
		case msgs := <-testTopics[0].Messages:
			receiveMessage = false
			assert.Equal(t, expectedPayload, msgs.Payload)
		}
	}
}

func TestInitializeAndProcessEventMultipleTopics(t *testing.T) {
	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:            "edgeX-meSsaGebus",
			PublishTopic:    "",
			SubscribeTopics: "t1,t2",
		},
		MessageBus: types.MessageBusConfig{
			Type: "zero",
			PublishHost: types.HostInfo{
				Host:     "*",
				Port:     5592,
				Protocol: "tcp",
			},
			SubscribeHost: types.HostInfo{
				Host:     "localhost",
				Port:     5594,
				Protocol: "tcp",
			},
		},
	}

	expectedCorrelationID := "123"

	done := make(chan bool)
	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		require.Equal(t, expectedEvent, params[0])
		done <- true
		return false, nil
	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(nil, nil)
	goRuntime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: &config, Runtime: goRuntime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)

	payload, _ := json.Marshal(addEventRequest)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}

	testClientConfig := types.MessageBusConfig{
		PublishHost: types.HostInfo{
			Host:     "*",
			Port:     5594,
			Protocol: "tcp",
		},
		Type: "zero",
	}

	testClient, err := messaging.NewMessageClient(testClientConfig)
	require.NoError(t, err, "Unable to create to publisher")

	err = testClient.Publish(message, "t1") //transform1 should be called after this executes
	require.NoError(t, err, "Failed to publish message")

	select {
	case <-done:
		// do nothing, just need to fall out.
	case <-time.After(3 * time.Second):
		require.Fail(t, "Transform never called for t1")
	}

	err = testClient.Publish(message, "t2") //transform1 should be called after this executes
	require.NoError(t, err, "Failed to publish message")

	select {
	case <-done:
		// do nothing, just need to fall out.
	case <-time.After(3 * time.Second):
		require.Fail(t, "Transform never called t2")
	}
}
