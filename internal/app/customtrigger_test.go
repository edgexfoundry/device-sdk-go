//
// Copyright (c) 2020 Technotects
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

package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/app/mocks"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"sync"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mqtt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRegisterCustomTriggerFactory_HTTP(t *testing.T) {
	name := strings.ToTitle(TriggerTypeHTTP)

	sdk := Service{}
	err := sdk.RegisterCustomTriggerFactory(name, nil)

	require.Error(t, err, "should throw error")
	require.Zero(t, len(sdk.customTriggerFactories), "nothing should be registered")
}

func TestRegisterCustomTriggerFactory_EdgeXMessageBus(t *testing.T) {
	name := strings.ToTitle(TriggerTypeMessageBus)

	sdk := Service{}
	err := sdk.RegisterCustomTriggerFactory(name, nil)

	require.Error(t, err, "should throw error")
	require.Zero(t, len(sdk.customTriggerFactories), "nothing should be registered")
}

func TestRegisterCustomTriggerFactory_MQTT(t *testing.T) {
	name := strings.ToTitle(TriggerTypeMQTT)

	sdk := Service{}
	err := sdk.RegisterCustomTriggerFactory(name, nil)

	require.Error(t, err, "should throw error")
	require.Zero(t, len(sdk.customTriggerFactories), "nothing should be registered")
}

func TestRegisterCustomTrigger(t *testing.T) {
	name := "cUsToM tRiGgEr"
	trig := mockCustomTrigger{}

	builder := func(c interfaces.TriggerConfig) (interfaces.Trigger, error) {
		return &trig, nil
	}
	sdk := Service{config: &common.ConfigurationStruct{}}

	err := sdk.RegisterCustomTriggerFactory(name, builder)

	require.Nil(t, err, "should not throw error")
	require.Equal(t, len(sdk.customTriggerFactories), 1, "provided function should be registered")

	registeredBuilder := sdk.customTriggerFactories[strings.ToUpper(name)]
	require.NotNil(t, registeredBuilder, "provided function should be registered with uppercase name")

	res, err := registeredBuilder(&sdk)
	require.NoError(t, err)
	require.Equal(t, res, &trig, "will be wrapped but should ultimately return result of builder")
}

func TestSetupTrigger_HTTP(t *testing.T) {
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeHTTP,
			},
		},
		lc: logger.MockLogger{},
	}

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &http.Trigger{}, trigger, "should be an http trigger")
}

func TestSetupTrigger_EdgeXMessageBus(t *testing.T) {
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMessageBus,
			},
		},
		lc: logger.MockLogger{},
	}

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &messagebus.Trigger{}, trigger, "should be an edgex-messagebus trigger")
}

func TestSetupTrigger_MQTT(t *testing.T) {
	config := &common.ConfigurationStruct{
		Trigger: common.TriggerInfo{
			Type: TriggerTypeMQTT,
		},
	}

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return config
		},
	})

	sdk := Service{
		dic:    dic,
		config: config,
		lc:     lc,
	}

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &mqtt.Trigger{}, trigger, "should be an external-MQTT trigger")
}

type mockCustomTrigger struct {
}

func (*mockCustomTrigger) Initialize(_ *sync.WaitGroup, _ context.Context, _ <-chan interfaces.BackgroundMessage) (bootstrap.Deferred, error) {
	return nil, nil
}

func Test_Service_setupTrigger_CustomType(t *testing.T) {
	triggerName := uuid.New().String()

	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: triggerName,
			},
		},
		lc: logger.MockLogger{},
	}

	err := sdk.RegisterCustomTriggerFactory(triggerName, func(c interfaces.TriggerConfig) (interfaces.Trigger, error) {
		return &mockCustomTrigger{}, nil
	})
	require.NoError(t, err)

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &mockCustomTrigger{}, trigger, "should be a custom trigger")
}

func Test_Service_SetupTrigger_CustomTypeError(t *testing.T) {
	triggerName := uuid.New().String()

	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: triggerName,
			},
		},
		lc: logger.MockLogger{},
	}

	err := sdk.RegisterCustomTriggerFactory(triggerName, func(c interfaces.TriggerConfig) (interfaces.Trigger, error) {
		return &mockCustomTrigger{}, errors.New("this should force returning nil even though we'll have a value")
	})
	require.NoError(t, err)

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.Nil(t, trigger, "should be nil")
}

func Test_Service_SetupTrigger_CustomTypeNotFound(t *testing.T) {
	triggerName := uuid.New().String()

	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: triggerName,
			},
		},
		lc: logger.MockLogger{},
	}

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.Nil(t, trigger, "should be nil")
}

func Test_customTriggerBinding_buildContext(t *testing.T) {
	container := &di.Container{}
	correlationId := uuid.NewString()
	contentType := uuid.NewString()

	bnd := &customTriggerBinding{
		dic: container,
	}

	got := bnd.buildContext(types.MessageEnvelope{CorrelationID: correlationId, ContentType: contentType})

	require.NotNil(t, got)

	assert.Equal(t, correlationId, got.CorrelationID())
	assert.Equal(t, contentType, got.InputContentType())

	ctx, ok := got.(*appfunction.Context)
	require.True(t, ok)
	assert.Equal(t, container, ctx.Dic)
}

func Test_customTriggerBinding_processMessage(t *testing.T) {
	type returns struct {
		runtimeProcessor interface{}
		pipelineMatcher  interface{}
	}
	type args struct {
		ctx      interfaces.AppFunctionContext
		envelope types.MessageEnvelope
	}
	tests := []struct {
		name    string
		setup   returns
		args    args
		wantErr int
	}{
		{
			name:    "no pipelines",
			setup:   returns{},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 0,
		},
		{
			name: "single pipeline",
			setup: returns{
				pipelineMatcher: []*interfaces.FunctionPipeline{{}},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 0,
		},
		{
			name: "single pipeline error",
			setup: returns{
				pipelineMatcher:  []*interfaces.FunctionPipeline{{}},
				runtimeProcessor: &runtime.MessageError{Err: fmt.Errorf("some error")},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 1,
		},
		{
			name: "multi pipeline",
			setup: returns{
				pipelineMatcher: []*interfaces.FunctionPipeline{{}, {}, {}},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 0,
		},
		{
			name: "multi pipeline single err",
			setup: returns{
				pipelineMatcher: []*interfaces.FunctionPipeline{{}, {Id: "errorid"}, {}},
				runtimeProcessor: func(appContext *appfunction.Context, envelope types.MessageEnvelope, pipeline *interfaces.FunctionPipeline) *runtime.MessageError {
					if pipeline.Id == "errorid" {
						return &runtime.MessageError{Err: fmt.Errorf("new error")}
					}
					return nil
				},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 1,
		},
		{
			name: "multi pipeline multi err",
			setup: returns{
				pipelineMatcher:  []*interfaces.FunctionPipeline{{}, {}, {}},
				runtimeProcessor: &runtime.MessageError{Err: fmt.Errorf("new error")},
			},
			args:    args{envelope: types.MessageEnvelope{CorrelationID: uuid.NewString(), ContentType: uuid.NewString(), ReceivedTopic: uuid.NewString()}, ctx: &appfunction.Context{}},
			wantErr: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tsb := mocks.TriggerServiceBinding{}

			tsb.On("ProcessMessage", mock.Anything, mock.Anything, mock.Anything).Return(tt.setup.runtimeProcessor)
			tsb.On("GetMatchingPipelines", tt.args.envelope.ReceivedTopic).Return(tt.setup.pipelineMatcher)

			bnd := &customTriggerBinding{
				TriggerServiceBinding: &tsb,
				log:                   logger.NewMockClient(),
			}

			err := bnd.processMessage(tt.args.ctx, tt.args.envelope)

			require.Equal(t, err == nil, tt.wantErr == 0)

			if err != nil {
				merr, ok := err.(*multierror.Error)

				require.True(t, ok)

				assert.Equal(t, tt.wantErr, merr.Len())
			}
		})
	}
}
