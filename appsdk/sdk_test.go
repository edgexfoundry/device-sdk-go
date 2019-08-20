//
// Copyright (c) 2019 Intel Corporation
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
	"reflect"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	triggerHttp "github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/assert"
)

var lc logger.LoggingClient
var context *appcontext.Context

func init() {
	lc = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")

}

func IsInstanceOf(objectPtr, typePtr interface{}) bool {
	return reflect.TypeOf(objectPtr) == reflect.TypeOf(typePtr)
}
func TestSetupHTTPTrigger(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "htTp",
			},
		},
	}
	runtime := &runtime.GolangRuntime{}
	runtime.SetTransforms(sdk.transforms)
	trigger := sdk.setupTrigger(sdk.config, runtime)
	result := IsInstanceOf(trigger, (*triggerHttp.Trigger)(nil))
	assert.True(t, result, "Expected Instance of HTTP Trigger")
}
func TestSetupMessageBusTrigger(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "meSsaGebus",
			},
		},
	}
	runtime := &runtime.GolangRuntime{}
	runtime.SetTransforms(sdk.transforms)
	trigger := sdk.setupTrigger(sdk.config, runtime)
	result := IsInstanceOf(trigger, (*messagebus.Trigger)(nil))
	assert.True(t, result, "Expected Instance of Message Bus Trigger")
}

func TestApplicationSettings(t *testing.T) {
	expectedSettingKey := "ApplicationName"
	expectedSettingValue := "simple-filter-xml"

	sdk := AppFunctionsSDK{}

	sdk.configDir = "../examples/simple-filter-xml/res"
	err := sdk.initializeConfiguration()

	assert.NoError(t, err, "failed to initialize configuration")

	appSettings := sdk.ApplicationSettings()
	if !assert.NotNil(t, appSettings, "returned application settings is nil") {
		t.Fatal()
	}

	actual, ok := appSettings[expectedSettingKey]
	if !assert.True(t, ok, "expected application setting key not fond") {
		t.Fatal()
	}

	assert.Equal(t, expectedSettingValue, actual, "actual application setting value not as expected")

}

func TestApplicationSettingsNil(t *testing.T) {
	sdk := AppFunctionsSDK{}

	sdk.configDir = "../examples/simple-filter-xml-post/res"
	err := sdk.initializeConfiguration()
	assert.NoError(t, err, "failed to initialize configuration")

	appSettings := sdk.ApplicationSettings()
	if !assert.Nil(t, appSettings, "returned application settings expected to be nil") {
		t.Fatal()
	}
}

func TestLoadConfigurablePipelineFunctionNotFound(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: "Bogus",
					Functions:      make(map[string]common.PipelineFunction),
				},
			},
		},
	}

	appFunctions, err := sdk.LoadConfigurablePipeline()
	assert.Error(t, err, "expected error for function not found in config")
	assert.Equal(t, err.Error(), "Function Bogus configuration not found in Pipeline.Functions section")
	assert.Nil(t, appFunctions, "expected app functions list to be nil")
}

func TestLoadConfigurablePipelineNotABuiltInSdkFunction(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["Bogus"] = common.PipelineFunction{}

	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: "Bogus",
					Functions:      functions,
				},
			},
		},
	}

	appFunctions, err := sdk.LoadConfigurablePipeline()
	assert.Error(t, err, "expected error")
	assert.Equal(t, err.Error(), "Function Bogus is not a built in SDK function")
	assert.Nil(t, appFunctions, "expected app functions list to be nil")
}

func TestLoadConfigurablePipelineAddressableConfig(t *testing.T) {
	functionName := "MQTTSend"
	functions := make(map[string]common.PipelineFunction)
	functions[functionName] = common.PipelineFunction{
		Parameters: map[string]string{"qos": "0", "autoreconnect": "false"},
		Addressable: models.Addressable{
			Address:   "localhost",
			Port:      1883,
			Protocol:  "tcp",
			Publisher: "MyApp",
			Topic:     "sampleTopic",
		},
	}

	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: functionName,
					Functions:      functions,
				},
			},
		},
	}

	appFunctions, err := sdk.LoadConfigurablePipeline()
	assert.NoError(t, err, "")
	assert.NotNil(t, appFunctions, "expected app functions list to be set")
}

func TestLoadConfigurablePipelineNumFunctions(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["FilterByDeviceName"] = common.PipelineFunction{
		Parameters: map[string]string{"FilterValues": "Random-Float-Device, Random-Integer-Device"},
	}
	functions["TransformToXML"] = common.PipelineFunction{}
	functions["SetOutputData"] = common.PipelineFunction{}

	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: "FilterByDeviceName, TransformToXML, SetOutputData",
					Functions:      functions,
				},
			},
		},
	}

	appFunctions, err := sdk.LoadConfigurablePipeline()
	assert.NoError(t, err, "")
	assert.NotNil(t, appFunctions, "expected app functions list to be set")
	assert.Equal(t, 3, len(appFunctions))
}

func TestUseTargetTypeOfByteArrayTrue(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["CompressWithGZIP"] = common.PipelineFunction{}
	functions["SetOutputData"] = common.PipelineFunction{}

	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder:           "CompressWithGZIP, SetOutputData",
					UseTargetTypeOfByteArray: true,
					Functions:                functions,
				},
			},
		},
	}

	_, err := sdk.LoadConfigurablePipeline()
	assert.NoError(t, err, "")
	assert.NotNil(t, sdk.TargetType)
	assert.Equal(t, reflect.Ptr, reflect.TypeOf(sdk.TargetType).Kind())
	assert.Equal(t, reflect.TypeOf([]byte{}).Kind(), reflect.TypeOf(sdk.TargetType).Elem().Kind())
}

func TestUseTargetTypeOfByteArrayFalse(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["CompressWithGZIP"] = common.PipelineFunction{}
	functions["SetOutputData"] = common.PipelineFunction{}

	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder:           "CompressWithGZIP, SetOutputData",
					UseTargetTypeOfByteArray: false,
					Functions:                functions,
				},
			},
		},
	}

	_, err := sdk.LoadConfigurablePipeline()
	assert.NoError(t, err, "")
	assert.Nil(t, sdk.TargetType)
}
