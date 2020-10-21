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

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
	"github.com/stretchr/testify/assert"
)

var logClient logger.LoggingClient

func init() {
	logClient = logger.NewMockClient()
}

func TestInitialize(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:           "meSsaGebus",
			PublishTopic:   "publish",
			SubscribeTopic: "events",
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

	runtime := &runtime.GolangRuntime{}

	trigger := Trigger{Configuration: &config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	assert.NotNil(t, trigger.client, "Expected client to be set")
	assert.Equal(t, 1, len(trigger.topics))
	assert.Equal(t, "events", trigger.topics[0].Topic)
	assert.NotNil(t, trigger.topics[0].Messages)
}

func TestInitializeBadConfiguration(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:           "meSsaGebus",
			PublishTopic:   "publish",
			SubscribeTopic: "events",
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

	runtime := &runtime.GolangRuntime{}

	trigger := Trigger{Configuration: &config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	assert.Error(t, err)
}

func TestInitializeAndProcessEventWithNoOutput(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:           "meSsaGebus",
			PublishTopic:   "",
			SubscribeTopic: "",
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

	expectedPayload := []byte(`{"id":"5888dea1bd36573f4681d6f9","created":1485364897029,"modified":1485364897029,"origin":1471806386919,"pushed":0,"device":"livingroomthermostat","readings":[{"id":"5888dea0bd36573f4681d6f8","created":1485364896983,"modified":1485364896983,"origin":1471806386919,"pushed":0,"name":"temperature","value":"38","device":"livingroomthermostat"}]}`)
	var expectedEvent models.Event
	json.Unmarshal([]byte(expectedPayload), &expectedEvent)

	transformWasCalled := common.AtomicBool{}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transformWasCalled.Set(true)
		assert.Equal(t, expectedEvent, params[0])
		return false, nil
	}

	runtime := &runtime.GolangRuntime{}
	runtime.Initialize(nil, nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: &config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       expectedPayload,
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
			Type:           "meSsaGebus",
			PublishTopic:   "PublishTopic",
			SubscribeTopic: "SubscribeTopic",
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

	expectedCorrelationID := "123"

	expectedPayload := []byte(`{"id":"5888dea1bd36573f4681d6f9","created":1485364897029,"modified":1485364897029,"origin":1471806386919,"pushed":0,"device":"livingroomthermostat","readings":[{"id":"5888dea0bd36573f4681d6f8","created":1485364896983,"modified":1485364896983,"origin":1471806386919,"pushed":0,"name":"temperature","value":"38","device":"livingroomthermostat"}]}`)
	var expectedEvent models.Event
	json.Unmarshal([]byte(expectedPayload), &expectedEvent)

	transformWasCalled := common.AtomicBool{}

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transformWasCalled.Set(true)
		assert.Equal(t, expectedEvent, params[0])
		edgexcontext.Complete([]byte("Transformed")) //transformed message published to message bus
		return false, nil

	}

	runtime := &runtime.GolangRuntime{}
	runtime.Initialize(nil, nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: &config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}

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

	testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus

	trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       expectedPayload,
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

		}
	}
}

func TestInitializeAndProcessBackgroundMessage(t *testing.T) {

	config := common.ConfigurationStruct{
		Binding: common.BindingInfo{
			Type:           "meSsaGebus",
			PublishTopic:   "PublishTopic",
			SubscribeTopic: "SubscribeTopic",
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

	runtime := &runtime.GolangRuntime{}
	runtime.Initialize(nil, nil)
	trigger := Trigger{Configuration: &config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}

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

	testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus

	background := make(chan types.MessageEnvelope)

	trigger.Initialize(&sync.WaitGroup{}, context.Background(), background)

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
