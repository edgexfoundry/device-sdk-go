//
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

package mqtt

import (
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	sdkCommon "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	triggerMocks "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mocks"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mqtt/mocks"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	interfaceMocks "github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewTrigger(t *testing.T) {
	bnd := &triggerMocks.ServiceBinding{}
	mp := &triggerMocks.MessageProcessor{}

	got := NewTrigger(bnd, mp)

	require.NotNil(t, got)

	assert.Equal(t, bnd, got.serviceBinding)
	assert.Equal(t, mp, got.messageProcessor)
}

func TestTrigger_responseHandler(t *testing.T) {
	const topicWithPlaceholder = "/topic/with/{ph}/placeholder"
	const formattedTopic = "topic/with/ph-value/placeholder"
	const setContentType = "content-type"
	const correlationId = "corrid-1233523"
	var payload = []byte("not-empty")
	const retain = true
	var qos = byte(8)

	type fields struct {
		publishTopic string
		qos          byte
		retain       bool
	}
	type args struct {
		pipeline *interfaces.FunctionPipeline
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		setup   func(*triggerMocks.ServiceBinding, *interfaceMocks.AppFunctionContext, *mocks.Client, *mocks.Token)
	}{
		{name: "no response data", wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, _ *mocks.Client, _ *mocks.Token) {
			functionContext.On("ResponseData").Return(nil)
		}},
		{name: "topic format failed", fields: fields{qos: qos, retain: retain, publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: true, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, _ *mocks.Client, _ *mocks.Token) {
			functionContext.On("ResponseData").Return(payload)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return("", fmt.Errorf("apply values failed"))
		}},
		{name: "publish failed", fields: fields{qos: qos, retain: retain, publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: true, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.Client, token *mocks.Token) {
			functionContext.On("ResponseData").Return(payload)
			functionContext.On("ResponseContentType").Return(setContentType)
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, qos, retain, mock.Anything).Return(token)
			token.On("Wait").Return(true)
			token.On("Error").Return(fmt.Errorf("publish error"))
		}},
		{name: "happy", fields: fields{qos: qos, retain: retain, publishTopic: topicWithPlaceholder}, args: args{pipeline: &interfaces.FunctionPipeline{}}, wantErr: false, setup: func(processor *triggerMocks.ServiceBinding, functionContext *interfaceMocks.AppFunctionContext, client *mocks.Client, token *mocks.Token) {
			functionContext.On("ResponseData").Return(payload)
			functionContext.On("ResponseContentType").Return(setContentType)
			functionContext.On("CorrelationID").Return(correlationId)
			functionContext.On("ApplyValues", topicWithPlaceholder).Return(formattedTopic, nil)
			client.On("Publish", mock.Anything, qos, retain, mock.Anything).Return(token)
			token.On("Wait").Return(true)
			token.On("Error").Return(nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceBinding := &triggerMocks.ServiceBinding{}

			serviceBinding.On("Config").Return(&sdkCommon.ConfigurationStruct{Trigger: sdkCommon.TriggerInfo{EdgexMessageBus: sdkCommon.MessageBusConfig{PublishHost: sdkCommon.PublishHostInfo{PublishTopic: tt.fields.publishTopic}}}})
			serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

			ctx := &interfaceMocks.AppFunctionContext{}
			client := &mocks.Client{}
			token := &mocks.Token{}

			if tt.setup != nil {
				tt.setup(serviceBinding, ctx, client, token)
			}

			trigger := &Trigger{
				serviceBinding: serviceBinding,
				qos:            qos,
				retain:         retain,
				publishTopic:   tt.fields.publishTopic,
				mqttClient:     client,
			}
			if err := trigger.responseHandler(ctx, tt.args.pipeline); (err != nil) != tt.wantErr {
				t.Errorf("responseHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTrigger_messageHandler(t *testing.T) {
	tests := []struct {
		name                string
		data                []byte
		expectedContentType string
	}{
		{name: "assume CBOR", data: []byte("not json"), expectedContentType: common.ContentTypeCBOR},
		{name: "json array", data: []byte("[ json array"), expectedContentType: common.ContentTypeJSON},
		{name: "json object", data: []byte("{ json object"), expectedContentType: common.ContentTypeJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic := uuid.NewString()
			ctx := &appfunction.Context{}
			message := &mocks.Message{}
			message.On("Payload").Return(tt.data)
			message.On("Topic").Return(topic)

			serviceBinding := &triggerMocks.ServiceBinding{}
			serviceBinding.On("BuildContext", mock.Anything).Return(func(envelope types.MessageEnvelope) interfaces.AppFunctionContext {
				assert.Equal(t, tt.expectedContentType, envelope.ContentType)
				assert.Equal(t, tt.data, envelope.Payload)
				assert.Equal(t, topic, envelope.ReceivedTopic)
				assert.NotEmpty(t, envelope.CorrelationID)

				return ctx
			})
			serviceBinding.On("LoggingClient").Return(logger.NewMockClient())

			messageProcessor := &triggerMocks.MessageProcessor{}
			messageProcessor.On("MessageReceived", ctx, mock.Anything, mock.Anything).Return(func(inctx interfaces.AppFunctionContext, _ types.MessageEnvelope, _ interfaces.PipelineResponseHandler) error {
				assert.Equal(t, ctx, inctx)
				return nil
			})

			trigger := &Trigger{
				serviceBinding:   serviceBinding,
				messageProcessor: messageProcessor,
			}

			trigger.messageHandler(nil, message)
		})
	}
}
