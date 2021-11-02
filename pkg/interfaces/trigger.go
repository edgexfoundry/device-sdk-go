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

package interfaces

import (
	"context"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

type TriggerConfig struct {
	// Logger exposes the logging client passed from the service
	Logger logger.LoggingClient
	// ContextBuilder contructs a context the trigger can specify for processing the received message
	ContextBuilder TriggerContextBuilder
	// MessageProcessor processes a message on the services default pipeline
	// Deprecated: use MessageReceived for multi-pipeline support
	MessageProcessor TriggerMessageProcessor
	// MessageReceived sends a message to the runtime for processing.
	MessageReceived TriggerMessageHandler
	// ConfigLoader is a function of type TriggerConfigLoader that can be used to load custom configuration sections for the trigger.s
	ConfigLoader TriggerConfigLoader
}

// Trigger provides an abstract means to pass messages to the function pipeline
type Trigger interface {
	// Initialize performs post creation initializations
	Initialize(wg *sync.WaitGroup, ctx context.Context, background <-chan BackgroundMessage) (bootstrap.Deferred, error)
}

// TriggerMessageProcessor provides an interface that can be used by custom triggers to invoke the runtime
type TriggerMessageProcessor func(ctx AppFunctionContext, envelope types.MessageEnvelope) error

// TriggerMessageHandler provides an interface that can be used by custom triggers to invoke the runtime
type TriggerMessageHandler func(ctx AppFunctionContext, envelope types.MessageEnvelope, responseHandler PipelineResponseHandler) error

// TriggerContextBuilder provides an interface to construct an AppFunctionContext for message
type TriggerContextBuilder func(env types.MessageEnvelope) AppFunctionContext

// TriggerConfigLoader provides an interface that can be used by custom triggers to load custom configuration elements
type TriggerConfigLoader func(config UpdatableConfig, sectionName string) error

// PipelineResponseHandler provides a function signature that can be passed to MessageProcessor to handle pipeline output(s)
type PipelineResponseHandler func(ctx AppFunctionContext, pipeline *FunctionPipeline) error
