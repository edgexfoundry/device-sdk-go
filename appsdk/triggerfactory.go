//
// Copyright (c) 2020 Technotects
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
	"errors"
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/mqtt"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
	"strings"
)

func (sdk *AppFunctionsSDK) defaultTriggerMessageProcessor(edgexcontext *appcontext.Context, envelope types.MessageEnvelope) error {
	messageError := sdk.runtime.ProcessMessage(edgexcontext, envelope)

	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		return messageError.Err
	}

	return nil
}

func (sdk *AppFunctionsSDK) defaultTriggerContextBuilder(env types.MessageEnvelope) *appcontext.Context {
	return &appcontext.Context{
		CorrelationID:         env.CorrelationID,
		Configuration:         sdk.config,
		LoggingClient:         sdk.LoggingClient,
		EventClient:           sdk.EdgexClients.EventClient,
		ValueDescriptorClient: sdk.EdgexClients.ValueDescriptorClient,
		CommandClient:         sdk.EdgexClients.CommandClient,
		NotificationsClient:   sdk.EdgexClients.NotificationsClient,
		SecretProvider:        sdk.secretProvider,
	}
}

// RegisterCustomTriggerFactory allows users to register builders for custom trigger types
func (s *AppFunctionsSDK) RegisterCustomTriggerFactory(name string,
	factory func(TriggerConfig) (Trigger, error)) error {
	nu := strings.ToUpper(name)

	if nu == bindingTypeEdgeXMessageBus ||
		nu == bindingTypeMessageBus ||
		nu == bindingTypeHTTP ||
		nu == bindingTypeMQTT {
		return errors.New(fmt.Sprintf("cannot register custom trigger for builtin type (%s)", name))
	}

	if s.customTriggerFactories == nil {
		s.customTriggerFactories = make(map[string]func(sdk *AppFunctionsSDK) (Trigger, error), 1)
	}

	s.customTriggerFactories[nu] = func(sdk *AppFunctionsSDK) (Trigger, error) {
		return factory(TriggerConfig{
			Config:           s.config,
			Logger:           s.LoggingClient,
			ContextBuilder:   sdk.defaultTriggerContextBuilder,
			MessageProcessor: sdk.defaultTriggerMessageProcessor,
		})
	}

	return nil
}

// setupTrigger configures the appropriate trigger as specified by configuration.
func (sdk *AppFunctionsSDK) setupTrigger(configuration *common.ConfigurationStruct, runtime *runtime.GolangRuntime) Trigger {
	var t Trigger
	// Need to make dynamic, search for the binding that is input

	switch triggerType := strings.ToUpper(configuration.Binding.Type); triggerType {
	case bindingTypeHTTP:
		sdk.LoggingClient.Info("HTTP trigger selected")
		t = &http.Trigger{Configuration: configuration, Runtime: runtime, Webserver: sdk.webserver, EdgeXClients: sdk.EdgexClients}

	case bindingTypeMessageBus,
		bindingTypeEdgeXMessageBus: // Allows for more explicit name now that we have plain MQTT option also
		sdk.LoggingClient.Info("EdgeX MessageBus trigger selected")
		t = &messagebus.Trigger{Configuration: configuration, Runtime: runtime, EdgeXClients: sdk.EdgexClients}

	case bindingTypeMQTT:
		sdk.LoggingClient.Info("External MQTT trigger selected")
		t = mqtt.NewTrigger(configuration, runtime, sdk.EdgexClients, sdk.secretProvider)

	default:
		if factory, found := sdk.customTriggerFactories[triggerType]; found {
			var err error
			t, err = factory(sdk)
			if err != nil {
				sdk.LoggingClient.Error(fmt.Sprintf("failed to initialize custom trigger [%s]: %s", triggerType, err.Error()))
				panic(err)
			}
		} else {
			sdk.LoggingClient.Error(fmt.Sprintf("Invalid Trigger type of '%s' specified", configuration.Binding.Type))
		}
	}

	return t
}
