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

package trigger

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

type ServiceBinding interface {
	// ProcessMessage provides access to the runtime's ProcessMessage function
	ProcessMessage(appContext *appfunction.Context, envelope types.MessageEnvelope, pipeline *interfaces.FunctionPipeline) *runtime.MessageError
	// GetMatchingPipelines provides access to the runtime's GetMatchingPipelines function
	GetMatchingPipelines(incomingTopic string) []*interfaces.FunctionPipeline
	// BuildContext creates a context for a given message envelope
	BuildContext(env types.MessageEnvelope) interfaces.AppFunctionContext
	// SecretProvider provides access to this service's secret provider for the trigger
	SecretProvider() messaging.SecretDataProvider
	// Config provides access to this service's configuration for the trigger
	Config() *common.ConfigurationStruct
	// LoggingClient provides access to this service's logging clietn for the trigger
	LoggingClient() logger.LoggingClient
	// LoadCustomConfig provides access to the service's LoadCustomConfig function
	LoadCustomConfig(config interfaces.UpdatableConfig, sectionName string) error
}
