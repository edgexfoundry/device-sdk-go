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

package appsdk

import (
	"context"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// Trigger interface provides an abstract means to pass messages through the function pipeline
type Trigger interface {
	// Initialize performs post creation initializations
	Initialize(wg *sync.WaitGroup, ctx context.Context, background <-chan types.MessageEnvelope) (bootstrap.Deferred, error)
}

// TriggerMessageProcessor provides an interface that can be used by custom triggers to invoke the runtime
type TriggerMessageProcessor func(ctx *appcontext.Context, envelope types.MessageEnvelope) error

// TriggerContextBuilder provides an interface to construct an appcontext.Context for message
type TriggerContextBuilder func(env types.MessageEnvelope) *appcontext.Context

// TriggerConfig provides a container to pass context needed to user defined triggers
type TriggerConfig struct {
	Config           *common.ConfigurationStruct
	Logger           logger.LoggingClient
	ContextBuilder   TriggerContextBuilder
	MessageProcessor TriggerMessageProcessor
}
