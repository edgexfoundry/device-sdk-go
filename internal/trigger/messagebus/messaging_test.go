//
// Copyright (c) 2021 Intel Corporation
// Copyright (c) 2021 One Track Consulting
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
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus/mocks"
	interfaceMocks "github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces/mocks"
	bootstrapMessaging "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"

	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	triggerMocks "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mocks"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	bootstrapMocks "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note the constant TriggerTypeMessageBus can not be used due to cyclic imports
const TriggerTypeMessageBus = "EDGEX-MESSAGEBUS"

var addEventRequest = createTestEventRequest()

func createTestEventRequest() requests.AddEventRequest {
	event := dtos.NewEvent("thermostat", "LivingRoomThermostat", "temperature")
	_ = event.AddSimpleReading("temperature", common.ValueTypeInt64, int64(38))
	request := requests.NewAddEventRequest(event)
	return request
}

func TestInitializeNotSecure(t *testing.T) {

	config := sdkCommon.ConfigurationStruct{
		Trigger: sdkCommon.TriggerInfo{
			Type: TriggerTypeMessageBus,
			EdgexMessageBus: sdkCommon.MessageBusConfig{
				Type: "zero",

				PublishHost: sdkCommon.PublishHostInfo{
					Host:         "*",
					Port:         5563,
					Protocol:     "tcp",
					PublishTopic: "publish",
				},
				SubscribeHost: sdkCommon.SubscribeHostInfo{
					Host:            "localhost",
					Port:            5563,
					Protocol:        "tcp",
					SubscribeTopics: "events",
				},
			},
		},
	}

	serviceBinding := &triggerMocks.ServiceBinding{}
	serviceBinding.On("Config").Return(&config)
	serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

	trigger := NewTrigger(serviceBinding, nil)

	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)
	assert.NotNil(t, trigger.client, "Expected client to be set")
	assert.Equal(t, 1, len(trigger.topics))
	assert.Equal(t, "events", trigger.topics[0].Topic)
	assert.NotNil(t, trigger.topics[0].Messages)
}

func TestInitializeSecure(t *testing.T) {
	secretName := "redisdb"

	config := sdkCommon.ConfigurationStruct{
		Trigger: sdkCommon.TriggerInfo{
			Type: TriggerTypeMessageBus,
			EdgexMessageBus: sdkCommon.MessageBusConfig{
				Type: "zero",

				PublishHost: sdkCommon.PublishHostInfo{
					Host:         "*",
					Port:         5563,
					Protocol:     "tcp",
					PublishTopic: "publish",
				},
				SubscribeHost: sdkCommon.SubscribeHostInfo{
					Host:            "localhost",
					Port:            5563,
					Protocol:        "tcp",
					SubscribeTopics: "events",
				},
				Optional: map[string]string{
					bootstrapMessaging.AuthModeKey:   bootstrapMessaging.AuthModeUsernamePassword,
					bootstrapMessaging.SecretNameKey: secretName,
				},
			},
		},
	}

	mock := bootstrapMocks.SecretProvider{}
	mock.On("GetSecret", secretName).Return(map[string]string{
		bootstrapMessaging.SecretUsernameKey: "user",
		bootstrapMessaging.SecretPasswordKey: "password",
	}, nil)

	serviceBinding := &triggerMocks.ServiceBinding{}
	serviceBinding.On("Config").Return(&config)
	serviceBinding.On("LoggingClient").Return(logger.NewMockClient())
	serviceBinding.On("SecretProvider").Return(&mock)

	trigger := NewTrigger(serviceBinding, nil)

	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)
	assert.NotNil(t, trigger.client, "Expected client to be set")
	assert.Equal(t, 1, len(trigger.topics))
	assert.Equal(t, "events", trigger.topics[0].Topic)
	assert.NotNil(t, trigger.topics[0].Messages)
}

func TestInitializeBadConfiguration(t *testing.T) {

	config := sdkCommon.ConfigurationStruct{
		Trigger: sdkCommon.TriggerInfo{
			Type: TriggerTypeMessageBus,

			EdgexMessageBus: sdkCommon.MessageBusConfig{
				Type: "aaaa", //as type is not "zero", should return an error on client initialization
				PublishHost: sdkCommon.PublishHostInfo{
					Host:         "*",
					Port:         5568,
					Protocol:     "tcp",
					PublishTopic: "publish",
				},
				SubscribeHost: sdkCommon.SubscribeHostInfo{
					Host:            "localhost",
					Port:            5568,
					Protocol:        "tcp",
					SubscribeTopics: "events",
				},
			},
		},
	}

	serviceBinding := &triggerMocks.ServiceBinding{}
	serviceBinding.On("Config").Return(&config)
	serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

	trigger := NewTrigger(serviceBinding, nil)

	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	assert.Error(t, err)
}

func TestInitializeAndProcessEvent(t *testing.T) {

	config := sdkCommon.ConfigurationStruct{
		Trigger: sdkCommon.TriggerInfo{
			Type: TriggerTypeMessageBus,
			EdgexMessageBus: sdkCommon.MessageBusConfig{
				Type: "zero",
				PublishHost: sdkCommon.PublishHostInfo{
					Host:         "*",
					Port:         5566,
					Protocol:     "tcp",
					PublishTopic: "",
				},
				SubscribeHost: sdkCommon.SubscribeHostInfo{
					Host:            "localhost",
					Port:            5564,
					Protocol:        "tcp",
					SubscribeTopics: "",
				},
			},
		},
	}

	expectedCorrelationID := "123"

	messageProcessed := make(chan bool, 1)

	expectedContext := appfunction.NewContext(uuid.NewString(), nil, "")

	serviceBinding := &triggerMocks.ServiceBinding{}
	serviceBinding.On("Config").Return(&config)
	serviceBinding.On("LoggingClient").Return(logger.NewMockClient())
	serviceBinding.On("BuildContext", mock.Anything).Return(expectedContext)

	messageProcessor := &triggerMocks.MessageProcessor{}
	messageProcessor.On("MessageReceived", expectedContext, mock.Anything, mock.AnythingOfType("interfaces.PipelineResponseHandler")).Return(func(interfaces.AppFunctionContext, types.MessageEnvelope, interfaces.PipelineResponseHandler) error {
		messageProcessed <- true
		return nil
	})

	trigger := NewTrigger(serviceBinding, messageProcessor)

	_, err := trigger.Initialize(&sync.WaitGroup{}, context.Background(), nil)
	require.NoError(t, err)

	payload, err := json.Marshal(addEventRequest)
	require.NoError(t, err)

	message := types.MessageEnvelope{
		CorrelationID: expectedCorrelationID,
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
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
	case <-messageProcessed:
		// do nothing, just need to fall out.
	case <-time.After(5 * time.Second):
		require.Fail(t, "Message never processed")
	}
}

func TestInitializeAndProcessBackgroundMessage(t *testing.T) {

	config := sdkCommon.ConfigurationStruct{
		Trigger: sdkCommon.TriggerInfo{
			Type: TriggerTypeMessageBus,
			EdgexMessageBus: sdkCommon.MessageBusConfig{
				Type: "zero",
				PublishHost: sdkCommon.PublishHostInfo{
					Host:         "*",
					Port:         5588,
					Protocol:     "tcp",
					PublishTopic: "PublishTopic",
				},
				SubscribeHost: sdkCommon.SubscribeHostInfo{
					Host:            "localhost",
					Port:            5590,
					Protocol:        "tcp",
					SubscribeTopics: "SubscribeTopic",
				},
			},
		},
	}

	expectedCorrelationID := "123"

	expectedPayload := []byte(`{"id":"5888dea1bd36573f4681d6f9","origin":1471806386919,"pushed":0,"device":"livingroomthermostat","readings":[{"id":"5888dea0bd36573f4681d6f8","created":1485364896983,"modified":1485364896983,"origin":1471806386919,"pushed":0,"name":"temperature","value":"38","device":"livingroomthermostat"}]}`)

	serviceBinding := &triggerMocks.ServiceBinding{}
	serviceBinding.On("Config").Return(&config)
	serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

	trigger := NewTrigger(serviceBinding, nil)

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

	backgroundTopic := uuid.NewString()

	testTopics := []types.TopicChannel{{Topic: backgroundTopic, Messages: make(chan types.MessageEnvelope)}}
	testMessageErrors := make(chan error)

	err = testClient.Subscribe(testTopics, testMessageErrors) //subscribe in order to receive transformed output to the bus
	require.NoError(t, err)

	background := make(chan interfaces.BackgroundMessage)

	_, err = trigger.Initialize(&sync.WaitGroup{}, context.Background(), background)
	require.NoError(t, err)

	background <- mockBackgroundMessage{
		Payload: types.MessageEnvelope{
			CorrelationID: expectedCorrelationID,
			Payload:       expectedPayload,
			ContentType:   common.ContentTypeJSON,
		},
		DeliverToTopic: backgroundTopic,
	}

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

type mockBackgroundMessage struct {
	DeliverToTopic string
	Payload        types.MessageEnvelope
}

func (bg mockBackgroundMessage) Topic() string {
	return bg.DeliverToTopic
}

func (bg mockBackgroundMessage) Message() types.MessageEnvelope {
	return bg.Payload
}

func TestTrigger_responseHandler(t *testing.T) {
	const topicWithPlaceholder = "/topic/with/{ph}/placeholder"
	const formattedTopic = "topic/with/ph-value/placeholder"
	const setContentType = "content-type"
	const correlationId = "corrid-1233523"
	var setContentTypePayload = []byte("not-empty")
	var inferJsonPayload = []byte("{not-empty")
	var inferJsonArrayPayload = []byte("[not-empty")

	type fields struct {
		publishTopic string
	}
	type args struct {
		pipeline *interfaces.FunctionPipeline
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		setup   func(*triggerMocks.ServiceBinding, *interfaceMocks.AppFunctionContext, *mocks.MessageClient)
	}{
		{name: "no response data", wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, _ *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(nil)
		}},
		{name: "topic format failed", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: true, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, _ *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(setContentTypePayload)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return("", fmt.Errorf("apply values failed"))
		}},
		{name: "publish failed", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: true, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(setContentTypePayload)
			functionContext.On("ResponseContentType").Return(setContentType)
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				return fmt.Errorf("publish failed")
			})
		}},
		{name: "happy", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(setContentTypePayload)
			functionContext.On("ResponseContentType").Return(setContentType)
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				assert.Equal(t, correlationId, envelope.CorrelationID)
				assert.Equal(t, setContentType, envelope.ContentType)
				assert.Equal(t, setContentTypePayload, envelope.Payload)
				return nil
			})
		}},
		{name: "happy assume CBOR", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(setContentTypePayload)
			functionContext.On("ResponseContentType").Return("")
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				assert.Equal(t, correlationId, envelope.CorrelationID)
				assert.Equal(t, common.ContentTypeCBOR, envelope.ContentType)
				assert.Equal(t, setContentTypePayload, envelope.Payload)
				return nil
			})
		}},
		{name: "happy infer JSON", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(inferJsonPayload)
			functionContext.On("ResponseContentType").Return("")
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				assert.Equal(t, correlationId, envelope.CorrelationID)
				assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
				assert.Equal(t, inferJsonPayload, envelope.Payload)
				return nil
			})
		}},
		{name: "happy infer JSON array", fields: fields{publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.MessageClient) {
			functionContext.On("ResponseData").Return(inferJsonArrayPayload)
			functionContext.On("ResponseContentType").Return("")
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, mock.Anything).Return(func(envelope types.MessageEnvelope, s string) error {
				assert.Equal(t, correlationId, envelope.CorrelationID)
				assert.Equal(t, common.ContentTypeJSON, envelope.ContentType)
				assert.Equal(t, inferJsonArrayPayload, envelope.Payload)
				return nil
			})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceBinding := &triggerMocks.ServiceBinding{}

			serviceBinding.On("Config").Return(&sdkCommon.ConfigurationStruct{Trigger: sdkCommon.TriggerInfo{EdgexMessageBus: sdkCommon.MessageBusConfig{PublishHost: sdkCommon.PublishHostInfo{PublishTopic: tt.fields.publishTopic}}}})
			serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

			ctx := &interfaceMocks.AppFunctionContext{}
			client := &mocks.MessageClient{}

			if tt.setup != nil {
				tt.setup(serviceBinding, ctx, client)
			}

			trigger := &Trigger{
				serviceBinding: serviceBinding,
				client:         client,
			}
			if err := trigger.responseHandler(ctx, tt.args.pipeline); (err != nil) != tt.wantErr {
				t.Errorf("responseHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
