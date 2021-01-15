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
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/mqtt"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRegisterCustomTriggerFactory_HTTP(t *testing.T) {
	name := strings.ToTitle(bindingTypeHTTP)

	sdk := AppFunctionsSDK{}
	err := sdk.RegisterCustomTriggerFactory(name, nil)

	require.Error(t, err, "should throw error")
	require.Zero(t, len(sdk.customTriggerFactories), "nothing should be registered")
}

func TestRegisterCustomTriggerFactory_MessageBus(t *testing.T) {
	name := strings.ToTitle(bindingTypeMessageBus)

	sdk := AppFunctionsSDK{}
	err := sdk.RegisterCustomTriggerFactory(name, nil)

	require.Error(t, err, "should throw error")
	require.Zero(t, len(sdk.customTriggerFactories), "nothing should be registered")
}

func TestRegisterCustomTriggerFactory_EdgeXMessageBus(t *testing.T) {
	name := strings.ToTitle(bindingTypeEdgeXMessageBus)

	sdk := AppFunctionsSDK{}
	err := sdk.RegisterCustomTriggerFactory(name, nil)

	require.Error(t, err, "should throw error")
	require.Zero(t, len(sdk.customTriggerFactories), "nothing should be registered")
}

func TestRegisterCustomTriggerFactory_MQTT(t *testing.T) {
	name := strings.ToTitle(bindingTypeMQTT)

	sdk := AppFunctionsSDK{}
	err := sdk.RegisterCustomTriggerFactory(name, nil)

	require.Error(t, err, "should throw error")
	require.Zero(t, len(sdk.customTriggerFactories), "nothing should be registered")
}

func TestRegisterCustomTrigger(t *testing.T) {
	name := "cUsToM tRiGgEr"
	trig := mockCustomTrigger{}

	builder := func(c TriggerConfig) (Trigger, error) {
		return &trig, nil
	}
	sdk := AppFunctionsSDK{}
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
	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "http",
			},
		},
		LoggingClient: logger.MockLogger{},
	}

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &http.Trigger{}, trigger, "should be an http trigger")
}

func TestSetupTrigger_MessageBus(t *testing.T) {
	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "messagebus",
			},
		},
		LoggingClient: logger.MockLogger{},
	}

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &messagebus.Trigger{}, trigger, "should be a messagebus trigger")
}

func TestSetupTrigger_EdgeXMessageBus(t *testing.T) {
	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "edgex-messagebus",
			},
		},
		LoggingClient: logger.MockLogger{},
	}

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &messagebus.Trigger{}, trigger, "should be an edgex-messagebus trigger")
}

func TestSetupTrigger_MQTT(t *testing.T) {
	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "external-mqtt",
			},
		},
		LoggingClient: logger.MockLogger{},
	}

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &mqtt.Trigger{}, trigger, "should be an external-MQTT trigger")
}

type mockCustomTrigger struct {
}

func (*mockCustomTrigger) Initialize(wg *sync.WaitGroup, ctx context.Context, background <-chan types.MessageEnvelope) (bootstrap.Deferred, error) {
	return nil, nil
}

func TestSetupTrigger_CustomType(t *testing.T) {
	triggerName := uuid.New().String()

	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: triggerName,
			},
		},
		LoggingClient: logger.MockLogger{},
	}

	sdk.RegisterCustomTriggerFactory(triggerName, func(c TriggerConfig) (Trigger, error) {
		return &mockCustomTrigger{}, nil
	})

	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	require.NotNil(t, trigger, "should be defined")
	require.IsType(t, &mockCustomTrigger{}, trigger, "should be a custom trigger")
}
