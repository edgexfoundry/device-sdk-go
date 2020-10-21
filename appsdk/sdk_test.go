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
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	triggerHttp "github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/webserver"
)

var lc logger.LoggingClient

func TestMain(m *testing.M) {
	// No remote and no file results in STDOUT logging only
	lc = logger.NewMockClient()
	m.Run()
}

func IsInstanceOf(objectPtr, typePtr interface{}) bool {
	return reflect.TypeOf(objectPtr) == reflect.TypeOf(typePtr)
}

func TestAddRoute(t *testing.T) {
	router := mux.NewRouter()
	ws := webserver.NewWebServer(&common.ConfigurationStruct{}, nil, lc, router)

	sdk := AppFunctionsSDK{
		webserver: ws,
	}
	_ = sdk.AddRoute("/test", func(http.ResponseWriter, *http.Request) {}, http.MethodGet)
	_ = router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		assert.Equal(t, "/test", path)
		return nil
	})

}

func TestAddBackgroundPublisher(t *testing.T) {
	sdk := AppFunctionsSDK{}
	pub, ok := sdk.AddBackgroundPublisher(1).(*backgroundPublisher)

	if !ok {
		assert.Fail(t, fmt.Sprintf("Unexpected BackgroundPublisher implementation encountered: %T", pub))
	}

	require.NotNil(t, pub.output, "publisher should have an output channel set")
	require.NotNil(t, sdk.backgroundChannel, "sdk should have a background channel set for passing to trigger intitialization")

	// compare addresses since types will not match
	assert.Equal(t, fmt.Sprintf("%p", sdk.backgroundChannel), fmt.Sprintf("%p", pub.output),
		"same channel should be referenced by the BackgroundPublisher and the SDK.")
}

func TestSetupHTTPTrigger(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "htTp",
			},
		},
	}
	testRuntime := &runtime.GolangRuntime{}
	testRuntime.Initialize(nil, nil)
	testRuntime.SetTransforms(sdk.transforms)
	trigger := sdk.setupTrigger(sdk.config, testRuntime)
	result := IsInstanceOf(trigger, (*triggerHttp.Trigger)(nil))
	assert.True(t, result, "Expected Instance of HTTP Trigger")
}

func TestSetupMessageBusTrigger(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "meSsaGebus",
			},
		},
	}
	testRuntime := &runtime.GolangRuntime{}
	testRuntime.Initialize(nil, nil)
	testRuntime.SetTransforms(sdk.transforms)
	trigger := sdk.setupTrigger(sdk.config, testRuntime)
	result := IsInstanceOf(trigger, (*messagebus.Trigger)(nil))
	assert.True(t, result, "Expected Instance of Message Bus Trigger")
}

func TestSetFunctionsPipelineNoTransforms(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "meSsaGebus",
			},
		},
	}
	err := sdk.SetFunctionsPipeline()
	require.Error(t, err, "There should be an error")
	assert.Equal(t, "no transforms provided to pipeline", err.Error())
}

func TestSetFunctionsPipelineOneTransform(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		runtime:       &runtime.GolangRuntime{},
		config: &common.ConfigurationStruct{
			Binding: common.BindingInfo{
				Type: "meSsaGebus",
			},
		},
	}
	function := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		return true, nil
	}

	sdk.runtime.Initialize(nil, nil)
	err := sdk.SetFunctionsPipeline(function)
	require.NoError(t, err)
	assert.Equal(t, 1, len(sdk.transforms))
}

func TestApplicationSettings(t *testing.T) {
	expectedSettingKey := "ApplicationName"
	expectedSettingValue := "simple-filter-xml"

	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{
			ApplicationSettings: map[string]string{
				"ApplicationName": "simple-filter-xml",
			},
		},
	}

	appSettings := sdk.ApplicationSettings()
	require.NotNil(t, appSettings, "returned application settings is nil")

	actual, ok := appSettings[expectedSettingKey]
	require.True(t, ok, "expected application setting key not found")
	assert.Equal(t, expectedSettingValue, actual, "actual application setting value not as expected")
}

func TestApplicationSettingsNil(t *testing.T) {
	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{},
	}

	appSettings := sdk.ApplicationSettings()
	require.Nil(t, appSettings, "returned application settings expected to be nil")
}

func TestGetAppSettingStrings(t *testing.T) {
	setting := "DeviceNames"
	expected := []string{"dev1", "dev2"}

	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{
			ApplicationSettings: map[string]string{
				"DeviceNames": "dev1,   dev2",
			},
		},
	}

	actual, err := sdk.GetAppSettingStrings(setting)
	require.NoError(t, err, "unexpected error")
	assert.EqualValues(t, expected, actual, "actual application setting values not as expected")
}

func TestGetAppSettingStringsSettingMissing(t *testing.T) {
	setting := "DeviceNames"
	expected := "setting not found in ApplicationSettings"

	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{
			ApplicationSettings: map[string]string{},
		},
	}

	_, err := sdk.GetAppSettingStrings(setting)
	require.Error(t, err, "Expected an error")
	assert.Contains(t, err.Error(), expected, "Error not as expected")
}

func TestGetAppSettingStringsNoAppSettings(t *testing.T) {
	setting := "DeviceNames"
	expected := "ApplicationSettings section is missing"

	sdk := AppFunctionsSDK{
		config: &common.ConfigurationStruct{},
	}

	_, err := sdk.GetAppSettingStrings(setting)
	require.Error(t, err, "Expected an error")
	assert.Contains(t, err.Error(), expected, "Error not as expected")
}

func TestLoadConfigurablePipelineFunctionNotFound(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: "Bogus",
					Functions:      make(map[string]common.PipelineFunction),
				},
			},
		},
	}

	appFunctions, err := sdk.LoadConfigurablePipeline()
	require.Error(t, err, "expected error for function not found in config")
	assert.Equal(t, "function Bogus configuration not found in Pipeline.Functions section", err.Error())
	assert.Nil(t, appFunctions, "expected app functions list to be nil")
}

func TestLoadConfigurablePipelineNotABuiltInSdkFunction(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["Bogus"] = common.PipelineFunction{}

	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: "Bogus",
					Functions:      functions,
				},
			},
		},
	}

	appFunctions, err := sdk.LoadConfigurablePipeline()
	require.Error(t, err, "expected error")
	assert.Equal(t, "function Bogus is not a built in SDK function", err.Error())
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
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: functionName,
					Functions:      functions,
				},
			},
		},
	}

	appFunctions, err := sdk.LoadConfigurablePipeline()
	require.NoError(t, err)
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
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: "FilterByDeviceName, TransformToXML, SetOutputData",
					Functions:      functions,
				},
			},
		},
	}

	appFunctions, err := sdk.LoadConfigurablePipeline()
	require.NoError(t, err)
	require.NotNil(t, appFunctions, "expected app functions list to be set")
	assert.Equal(t, 3, len(appFunctions))
}

func TestUseTargetTypeOfByteArrayTrue(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["CompressWithGZIP"] = common.PipelineFunction{}
	functions["SetOutputData"] = common.PipelineFunction{}

	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: &common.ConfigurationStruct{
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
	require.NoError(t, err)
	require.NotNil(t, sdk.TargetType)
	assert.Equal(t, reflect.Ptr, reflect.TypeOf(sdk.TargetType).Kind())
	assert.Equal(t, reflect.TypeOf([]byte{}).Kind(), reflect.TypeOf(sdk.TargetType).Elem().Kind())
}

func TestUseTargetTypeOfByteArrayFalse(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["CompressWithGZIP"] = common.PipelineFunction{}
	functions["SetOutputData"] = common.PipelineFunction{}

	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: &common.ConfigurationStruct{
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
	require.NoError(t, err)
	assert.Nil(t, sdk.TargetType)
}

func TestSetServiceKey(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		ServiceKey:    "MyAppService",
	}

	tests := []struct {
		name                          string
		profile                       string
		profileEnvVar                 string
		profileEnvValue               string
		serviceKeyEnvValue            string
		serviceKeyCommandLineOverride string
		originalServiceKey            string
		expectedServiceKey            string
	}{
		{
			name:               "No profile",
			originalServiceKey: "MyAppService" + ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService",
		},
		{
			name:               "Profile specified, no override",
			profile:            "mqtt-export",
			originalServiceKey: "MyAppService-" + ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService-mqtt-export",
		},
		{
			name:               "Profile specified with V1 override",
			profile:            "rules-engine",
			profileEnvVar:      envV1Profile,
			profileEnvValue:    "rules-engine-mqtt",
			originalServiceKey: "MyAppService-" + ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService-rules-engine-mqtt",
		},
		{
			name:               "Profile specified with V2 override",
			profile:            "rules-engine",
			profileEnvVar:      envProfile,
			profileEnvValue:    "rules-engine-redis",
			originalServiceKey: "MyAppService-" + ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService-rules-engine-redis",
		},
		{
			name:               "No profile specified with V1 override",
			profileEnvVar:      envV1Profile,
			profileEnvValue:    "sample",
			originalServiceKey: "MyAppService-" + ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService-sample",
		},
		{
			name:               "No profile specified with V2 override",
			profileEnvVar:      envProfile,
			profileEnvValue:    "http-export",
			originalServiceKey: "MyAppService-" + ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService-http-export",
		},
		{
			name:               "No ProfileSuffixPlaceholder with override",
			profileEnvVar:      envProfile,
			profileEnvValue:    "my-profile",
			originalServiceKey: "MyCustomAppService",
			expectedServiceKey: "MyCustomAppService",
		},
		{
			name:               "No ProfileSuffixPlaceholder with profile specified, no override",
			profile:            "my-profile",
			originalServiceKey: "MyCustomAppService",
			expectedServiceKey: "MyCustomAppService",
		},
		{
			name:                          "Service Key command-line override, no profile",
			serviceKeyCommandLineOverride: "MyCustomAppService",
			originalServiceKey:            "AppService",
			expectedServiceKey:            "MyCustomAppService",
		},
		{
			name:                          "Service Key command-line override, with profile",
			serviceKeyCommandLineOverride: "AppService-<profile>-MyCloud",
			profile:                       "http-export",
			originalServiceKey:            "AppService",
			expectedServiceKey:            "AppService-http-export-MyCloud",
		},
		{
			name:               "Service Key ENV override, no profile",
			serviceKeyEnvValue: "MyCustomAppService",
			originalServiceKey: "AppService",
			expectedServiceKey: "MyCustomAppService",
		},
		{
			name:               "Service Key ENV override, with profile",
			serviceKeyEnvValue: "AppService-<profile>-MyCloud",
			profile:            "http-export",
			originalServiceKey: "AppService",
			expectedServiceKey: "AppService-http-export-MyCloud",
		},
	}

	// Just in case...
	os.Clearenv()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if len(test.profileEnvVar) > 0 && len(test.profileEnvValue) > 0 {
				os.Setenv(test.profileEnvVar, test.profileEnvValue)
			}
			if len(test.serviceKeyEnvValue) > 0 {
				os.Setenv(envServiceKey, test.serviceKeyEnvValue)
			}
			defer os.Clearenv()

			if len(test.serviceKeyCommandLineOverride) > 0 {
				sdk.serviceKeyOverride = test.serviceKeyCommandLineOverride
			}

			sdk.ServiceKey = test.originalServiceKey
			sdk.setServiceKey(test.profile)

			assert.Equal(t, test.expectedServiceKey, sdk.ServiceKey)
		})
	}
}
