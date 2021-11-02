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

package app

import (
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	triggerMocks "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mocks"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_simpleTriggerServiceBinding_BuildContext(t *testing.T) {
	container := &di.Container{}
	correlationId := uuid.NewString()
	contentType := uuid.NewString()

	bnd := &simpleTriggerServiceBinding{&Service{dic: container}, nil}

	got := bnd.BuildContext(types.MessageEnvelope{CorrelationID: correlationId, ContentType: contentType})

	require.NotNil(t, got)

	assert.Equal(t, correlationId, got.CorrelationID())
	assert.Equal(t, contentType, got.InputContentType())

	ctx, ok := got.(*appfunction.Context)
	require.True(t, ok)
	assert.Equal(t, container, ctx.Dic)
}

func Test_triggerMessageProcessor_MessageReceived(t *testing.T) {
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
		nilRh   bool
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
				pipelineMatcher:  []*interfaces.FunctionPipeline{{}},
				runtimeProcessor: nil,
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
			tsb := triggerMocks.ServiceBinding{}

			tsb.On("ProcessMessage", mock.Anything, mock.Anything, mock.Anything).Return(tt.setup.runtimeProcessor)
			tsb.On("GetMatchingPipelines", tt.args.envelope.ReceivedTopic).Return(tt.setup.pipelineMatcher)
			tsb.On("LoggingClient").Return(lc)

			bnd := &triggerMessageProcessor{
				&tsb,
			}

			var rh interfaces.PipelineResponseHandler

			if !tt.nilRh {
				rh = func(ctx interfaces.AppFunctionContext, pipeline *interfaces.FunctionPipeline) error {
					assert.Equal(t, tt.args.ctx, ctx)
					return nil
				}
			}

			err := bnd.MessageReceived(tt.args.ctx, tt.args.envelope, rh)

			require.Equal(t, err == nil, tt.wantErr == 0)

			if err != nil {
				if merr, ok := err.(*multierror.Error); ok {
					assert.Equal(t, tt.wantErr, merr.Len())
				}
			}
		})
	}
}
