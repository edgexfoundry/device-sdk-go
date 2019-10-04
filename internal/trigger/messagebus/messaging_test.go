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

package messagebus

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

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
	logClient = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
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

	trigger := Trigger{Configuration: config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	trigger.Initialize(&sync.WaitGroup{}, context.Background())
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

	trigger := Trigger{Configuration: config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	err := trigger.Initialize(&sync.WaitGroup{}, context.Background())
	assert.NotNil(t, err)
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

	transformWasCalled := false

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transformWasCalled = true
		assert.Equal(t, expectedEvent, params[0])
		return false, nil
	}

	runtime := &runtime.GolangRuntime{}
	runtime.Initialize(nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}
	trigger.Initialize(&sync.WaitGroup{}, context.Background())

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
	if !assert.NoError(t, err, "Unable to create to publisher") {
		t.Fatal()
	}

	assert.False(t, transformWasCalled)
	err = testClient.Publish(message, "") //transform1 should be called after this executes
	if !assert.NoError(t, err, "Failed to publish message") {
		t.Fatal()
	}

	time.Sleep(3 * time.Second)
	assert.True(t, transformWasCalled, "Transform never called")

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

	transformWasCalled := false

	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		transformWasCalled = true
		assert.Equal(t, expectedEvent, params[0])
		edgexcontext.Complete([]byte("Transformed")) //transformed message published to message bus
		return false, nil

	}

	runtime := &runtime.GolangRuntime{}
	runtime.Initialize(nil)
	runtime.SetTransforms([]appcontext.AppFunction{transform1})
	trigger := Trigger{Configuration: config, Runtime: runtime, EdgeXClients: common.EdgeXClients{LoggingClient: logClient}}

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
	if !assert.NoError(t, err, "Failed to create test client") {
		t.Fatal()
	}

	testTopics := []types.TopicChannel{{Topic: trigger.Configuration.Binding.PublishTopic, Messages: make(chan types.MessageEnvelope)}}
	testMessageErrors := make(chan error)

	testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus

	trigger.Initialize(&sync.WaitGroup{}, context.Background())

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       expectedPayload,
		ContentType:   clients.ContentTypeJSON,
	}

	assert.False(t, transformWasCalled)
	err = testClient.Publish(message, "SubscribeTopic")
	if !assert.NoError(t, err, "Failed to publish message") {
		t.Fatal()
	}
	time.Sleep(3 * time.Second)
	if !assert.True(t, transformWasCalled, "Transform never called") {
		t.Fatal()
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

		}
	}
}
