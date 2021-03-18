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

package messagebus

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note the constant TriggerTypeMessageBus can not be used due to cyclic imports
const TriggerTypeMessageBus = "EDGEX-MESSAGEBUS"

var addEventRequest = createTestEventRequest()
var expectedEvent = addEventRequest.Event

func createTestEventRequest() requests.AddEventRequest {
	event := dtos.NewEvent("thermostat", "LivingRoomThermostat", "temperature")
	_ = event.AddSimpleReading("temperature", v2.ValueTypeInt64, int64(38))
	request := requests.NewAddEventRequest(event)
	return request
}

var dic *di.Container

func TestMain(m *testing.M) {
	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
	m.Run()
}

func TestInitialize(t *testing.T) {

	config := common.ConfigurationStruct{
		Trigger: common.TriggerInfo{
			Type:            TriggerTypeMessageBus,
			PublishTopic:    "publish",
			SubscribeTopics: "events",
			EdgexMessageBus: types.MessageBusConfig{
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
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
	})

	goRuntime := &runtime.GolangRuntime{}

	trigger := NewTrigger(dic, goRuntime)

	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)
	assert.NotNil(t, trigger.client, "Expected client to be set")
	assert.Equal(t, 1, len(trigger.topics))
	assert.Equal(t, "events", trigger.topics[0].Topic)
	assert.NotNil(t, trigger.topics[0].Messages)
}

func TestInitializeBadConfiguration(t *testing.T) {

	config := common.ConfigurationStruct{
		Trigger: common.TriggerInfo{
			Type:            TriggerTypeMessageBus,
			PublishTopic:    "publish",
			SubscribeTopics: "events",

			EdgexMessageBus: types.MessageBusConfig{
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
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
	})

	goRuntime := &runtime.GolangRuntime{}

	trigger := NewTrigger(dic, goRuntime)
	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	assert.Error(t, err)
}

func TestInitializeAndProcessEventWithNoOutput(t *testing.T) {

	config := common.ConfigurationStruct{
		Trigger: common.TriggerInfo{
			Type:            TriggerTypeMessageBus,
			PublishTopic:    "",
			SubscribeTopics: "",
			EdgexMessageBus: types.MessageBusConfig{
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
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
	})

	expectedCorrelationID := "123"

	transformWasCalled := make(chan bool, 1)

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		assert.Equal(t, expectedEvent, data)
		transformWasCalled <- true
		return false, nil
	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(dic)
	goRuntime.SetTransforms([]interfaces.AppFunction{transform1})
	trigger := NewTrigger(dic, goRuntime)
	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)

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

	err = testClient.Publish(message, "") //transform1 should be called after this executes
	require.NoError(t, err, "Failed to publish message")

	select {
	case <-transformWasCalled:
		// do nothing, just need to fall out.
	case <-time.After(3 * time.Second):
		require.Fail(t, "Transform never called")
	}
}

func TestInitializeAndProcessEventWithOutput(t *testing.T) {

	config := common.ConfigurationStruct{
		Trigger: common.TriggerInfo{
			Type:            TriggerTypeMessageBus,
			PublishTopic:    "PublishTopic",
			SubscribeTopics: "SubscribeTopic",
			EdgexMessageBus: types.MessageBusConfig{
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
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
	})

	responseContentType := uuid.New().String()

	expectedCorrelationID := "123"

	transformWasCalled := make(chan bool, 1)

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		assert.Equal(t, expectedEvent, data)
		appContext.SetResponseContentType(responseContentType)
		appContext.SetResponseData([]byte("Transformed")) //transformed message published to message bus
		transformWasCalled <- true
		return false, nil

	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(dic)
	goRuntime.SetTransforms([]interfaces.AppFunction{transform1})
	trigger := NewTrigger(dic, goRuntime)

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

	testTopics := []types.TopicChannel{{Topic: config.Trigger.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
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

	err = testClient.Publish(message, "SubscribeTopic")
	require.NoError(t, err, "Failed to publish message")

	select {
	case <-transformWasCalled:
		// do nothing, just need to fall out.
	case <-time.After(3 * time.Second):
		require.Fail(t, "Transform never called")
	}
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
		Trigger: common.TriggerInfo{
			Type:            TriggerTypeMessageBus,
			PublishTopic:    "PublishTopic",
			SubscribeTopics: "SubscribeTopic",
			EdgexMessageBus: types.MessageBusConfig{
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
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
	})

	expectedCorrelationID := "123"

	transformWasCalled := make(chan bool, 1)

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		assert.Equal(t, expectedEvent, data)
		appContext.SetResponseData([]byte("{;)Transformed")) //transformed message published to message bus
		transformWasCalled <- true
		return false, nil

	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(dic)
	goRuntime.SetTransforms([]interfaces.AppFunction{transform1})
	trigger := NewTrigger(dic, goRuntime)

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

	testTopics := []types.TopicChannel{{Topic: config.Trigger.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
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

	err = testClient.Publish(message, "SubscribeTopic")
	require.NoError(t, err, "Failed to publish message")

	select {
	case <-transformWasCalled:
		// do nothing, just need to fall out.
	case <-time.After(3 * time.Second):
		require.Fail(t, "Transform never called")
	}

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
		Trigger: common.TriggerInfo{
			Type:            TriggerTypeMessageBus,
			PublishTopic:    "PublishTopic",
			SubscribeTopics: "SubscribeTopic",
			EdgexMessageBus: types.MessageBusConfig{
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
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
	})

	expectedCorrelationID := "123"

	transformWasCalled := make(chan bool, 1)

	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		assert.Equal(t, expectedEvent, data)
		appContext.SetResponseData([]byte("Transformed")) //transformed message published to message bus
		transformWasCalled <- true
		return false, nil
	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(dic)
	goRuntime.SetTransforms([]interfaces.AppFunction{transform1})
	trigger := NewTrigger(dic, goRuntime)
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

	testTopics := []types.TopicChannel{{Topic: config.Trigger.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
	testMessageErrors := make(chan error)

	err = testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus
	require.NoError(t, err)
	_, err = trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)

	payload, _ := json.Marshal(addEventRequest)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   clients.ContentTypeJSON,
	}

	err = testClient.Publish(message, "SubscribeTopic")
	require.NoError(t, err, "Failed to publish message")

	select {
	case <-transformWasCalled:
		// do nothing, just need to fall out.
	case <-time.After(3 * time.Second):
		require.Fail(t, "Transform never called")
	}

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
		Trigger: common.TriggerInfo{
			Type:            TriggerTypeMessageBus,
			PublishTopic:    "PublishTopic",
			SubscribeTopics: "SubscribeTopic",
			EdgexMessageBus: types.MessageBusConfig{
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
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
	})

	expectedCorrelationID := "123"

	expectedPayload := []byte(`{"id":"5888dea1bd36573f4681d6f9","created":1485364897029,"modified":1485364897029,"origin":1471806386919,"pushed":0,"device":"livingroomthermostat","readings":[{"id":"5888dea0bd36573f4681d6f8","created":1485364896983,"modified":1485364896983,"origin":1471806386919,"pushed":0,"name":"temperature","value":"38","device":"livingroomthermostat"}]}`)

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(dic)
	trigger := NewTrigger(dic, goRuntime)

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

	testTopics := []types.TopicChannel{{Topic: config.Trigger.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
	testMessageErrors := make(chan error)

	err = testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus
	require.NoError(t, err)

	background := make(chan types.MessageEnvelope)

	_, err = trigger.Initialize(&sync.WaitGroup{}, context.Background(), background)
	require.NoError(t, err)

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
		Trigger: common.TriggerInfo{
			Type:            TriggerTypeMessageBus,
			PublishTopic:    "",
			SubscribeTopics: "t1,t2",
			EdgexMessageBus: types.MessageBusConfig{
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
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config
		},
	})

	expectedCorrelationID := "123"

	transformWasCalled := make(chan bool, 1)
	transform1 := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		require.Equal(t, expectedEvent, data)
		transformWasCalled <- true
		return false, nil
	}

	goRuntime := &runtime.GolangRuntime{}
	goRuntime.Initialize(dic)
	goRuntime.SetTransforms([]interfaces.AppFunction{transform1})
	trigger := NewTrigger(dic, goRuntime)
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
	case <-transformWasCalled:
		// do nothing, just need to fall out.
	case <-time.After(3 * time.Second):
		require.Fail(t, "Transform never called for t1")
	}

	err = testClient.Publish(message, "t2") //transform1 should be called after this executes
	require.NoError(t, err, "Failed to publish message")

	select {
	case <-transformWasCalled:
		// do nothing, just need to fall out.
	case <-time.After(3 * time.Second):
		require.Fail(t, "Transform never called t2")
	}
}
