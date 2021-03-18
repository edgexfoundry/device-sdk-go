//
// Copyright (c) 2020 Technotects
// Copyright (c) 2021 Intel Corporation
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
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

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

func (*mockCustomTrigger) Initialize(_ *sync.WaitGroup, _ context.Context, _ <-chan types.MessageEnvelope) (bootstrap.Deferred, error) {
	return nil, nil
}

func TestSetupTrigger_CustomType(t *testing.T) {
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

func TestSetupTrigger_CustomType_Error(t *testing.T) {
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

func TestSetupTrigger_CustomType_NotFound(t *testing.T) {
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
