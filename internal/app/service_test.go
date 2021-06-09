//
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
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	triggerHttp "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/webserver"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	v2clients "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/http"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var lc logger.LoggingClient
var dic *di.Container
var target *Service
var baseUrl = "http://localhost:"

func TestMain(m *testing.M) {
	// No remote and no file results in STDOUT logging only
	lc = logger.NewMockClient()
	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &common.ConfigurationStruct{}
		},
	})

	target = NewService("unitTest", nil, "")
	target.dic = dic
	target.lc = lc

	m.Run()
}

func IsInstanceOf(objectPtr, typePtr interface{}) bool {
	return reflect.TypeOf(objectPtr) == reflect.TypeOf(typePtr)
}

func TestAddRoute(t *testing.T) {
	router := mux.NewRouter()

	ws := webserver.NewWebServer(dic, router)

	sdk := Service{
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
	sdk := Service{}
	pub, ok := sdk.AddBackgroundPublisher(1).(*backgroundPublisher)

	if !ok {
		assert.Fail(t, fmt.Sprintf("Unexpected BackgroundPublisher implementation encountered: %T", pub))
	}

	require.NotNil(t, pub.output, "publisher should have an output channel set")
	require.NotNil(t, sdk.backgroundPublishChannel, "svc should have a background channel set for passing to trigger initialization")

	// compare addresses since types will not match
	assert.Equal(t, fmt.Sprintf("%p", sdk.backgroundPublishChannel), fmt.Sprintf("%p", pub.output),
		"same channel should be referenced by the BackgroundPublisher and the SDK.")
}

func TestSetupHTTPTrigger(t *testing.T) {
	sdk := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: "htTp",
			},
		},
	}
	testRuntime := &runtime.GolangRuntime{}
	testRuntime.Initialize(dic)
	testRuntime.SetTransforms(sdk.transforms)
	trigger := sdk.setupTrigger(sdk.config, testRuntime)
	result := IsInstanceOf(trigger, (*triggerHttp.Trigger)(nil))
	assert.True(t, result, "Expected Instance of HTTP Trigger")
}

func TestSetupMessageBusTrigger(t *testing.T) {
	sdk := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMessageBus,
			},
		},
	}
	testRuntime := &runtime.GolangRuntime{}
	testRuntime.Initialize(dic)
	testRuntime.SetTransforms(sdk.transforms)
	trigger := sdk.setupTrigger(sdk.config, testRuntime)
	result := IsInstanceOf(trigger, (*messagebus.Trigger)(nil))
	assert.True(t, result, "Expected Instance of Message Bus Trigger")
}

func TestSetFunctionsPipelineNoTransforms(t *testing.T) {
	sdk := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMessageBus,
			},
		},
	}
	err := sdk.SetFunctionsPipeline()
	require.Error(t, err, "There should be an error")
	assert.Equal(t, "no transforms provided to pipeline", err.Error())
}

func TestSetFunctionsPipelineOneTransform(t *testing.T) {
	sdk := Service{
		lc:      lc,
		runtime: &runtime.GolangRuntime{},
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMessageBus,
			},
		},
	}
	function := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return true, nil
	}

	sdk.runtime.Initialize(dic)
	err := sdk.SetFunctionsPipeline(function)
	require.NoError(t, err)
	assert.Equal(t, 1, len(sdk.transforms))
}

func TestApplicationSettings(t *testing.T) {
	expectedSettingKey := "ApplicationName"
	expectedSettingValue := "simple-filter-xml"

	sdk := Service{
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
	sdk := Service{
		config: &common.ConfigurationStruct{},
	}

	appSettings := sdk.ApplicationSettings()
	require.Nil(t, appSettings, "returned application settings expected to be nil")
}

func TestGetAppSetting(t *testing.T) {
	goodSettingName := "ExportUrl"
	expectedGoodValue := "http:/somewhere.com"
	badSettingName := "DeviceName"

	svc := Service{
		config: &common.ConfigurationStruct{
			ApplicationSettings: map[string]string{
				goodSettingName: expectedGoodValue,
			},
		},
	}

	actual, err := svc.GetAppSetting(goodSettingName)
	require.NoError(t, err)
	assert.EqualValues(t, expectedGoodValue, actual, "actual application setting values not as expected")

	_, err = svc.GetAppSetting(badSettingName)
	require.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("%s setting not found in ApplicationSettings section", badSettingName))

	svc.config.ApplicationSettings = nil
	_, err = svc.GetAppSetting(goodSettingName)
	require.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("%s setting not found: ApplicationSettings section is missing", goodSettingName))
}

func TestGetAppSettingStrings(t *testing.T) {
	setting := "DeviceNames"
	expected := []string{"dev1", "dev2"}

	sdk := Service{
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

	sdk := Service{
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

	sdk := Service{
		config: &common.ConfigurationStruct{},
	}

	_, err := sdk.GetAppSettingStrings(setting)
	require.Error(t, err, "Expected an error")
	assert.Contains(t, err.Error(), expected, "Error not as expected")
}

func TestLoadConfigurablePipelineFunctionNotFound(t *testing.T) {
	sdk := Service{
		lc: lc,
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
	assert.Equal(t, "function 'Bogus' configuration not found in Pipeline.Functions section", err.Error())
	assert.Nil(t, appFunctions, "expected app functions list to be nil")
}

func TestLoadConfigurablePipelineNotABuiltInSdkFunction(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["Bogus"] = common.PipelineFunction{}

	sdk := Service{
		lc: lc,
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

func TestLoadConfigurablePipelineNumFunctions(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["FilterByDeviceName"] = common.PipelineFunction{
		Parameters: map[string]string{"DeviceNames": "Random-Float-Device, Random-Integer-Device"},
	}
	functions["Transform"] = common.PipelineFunction{
		Parameters: map[string]string{TransformType: TransformXml},
	}
	functions["SetResponseData"] = common.PipelineFunction{}

	sdk := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: "FilterByDeviceName, Transform, SetResponseData",
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
	functions["Compress"] = common.PipelineFunction{
		Parameters: map[string]string{Algorithm: CompressGZIP},
	}
	functions["SetResponseData"] = common.PipelineFunction{}

	sdk := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder:           "Compress, SetResponseData",
					UseTargetTypeOfByteArray: true,
					Functions:                functions,
				},
			},
		},
	}

	_, err := sdk.LoadConfigurablePipeline()
	require.NoError(t, err)
	require.NotNil(t, sdk.targetType)
	assert.Equal(t, reflect.Ptr, reflect.TypeOf(sdk.targetType).Kind())
	assert.Equal(t, reflect.TypeOf([]byte{}).Kind(), reflect.TypeOf(sdk.targetType).Elem().Kind())
}

func TestUseTargetTypeOfByteArrayFalse(t *testing.T) {
	functions := make(map[string]common.PipelineFunction)
	functions["Compress"] = common.PipelineFunction{
		Parameters: map[string]string{Algorithm: CompressGZIP},
	}
	functions["SetResponseData"] = common.PipelineFunction{}

	sdk := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder:           "Compress, SetResponseData",
					UseTargetTypeOfByteArray: false,
					Functions:                functions,
				},
			},
		},
	}

	_, err := sdk.LoadConfigurablePipeline()
	require.NoError(t, err)
	assert.Nil(t, sdk.targetType)
}

func TestSetServiceKey(t *testing.T) {
	sdk := Service{
		lc:                       lc,
		serviceKey:               "MyAppService",
		profileSuffixPlaceholder: interfaces.ProfileSuffixPlaceholder,
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
			originalServiceKey: "MyAppService" + interfaces.ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService",
		},
		{
			name:               "Profile specified, no override",
			profile:            "mqtt-export",
			originalServiceKey: "MyAppService-" + interfaces.ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService-mqtt-export",
		},
		{
			name:               "Profile specified with V2 override",
			profile:            "rules-engine",
			profileEnvVar:      envProfile,
			profileEnvValue:    "rules-engine-redis",
			originalServiceKey: "MyAppService-" + interfaces.ProfileSuffixPlaceholder,
			expectedServiceKey: "MyAppService-rules-engine-redis",
		},
		{
			name:               "No profile specified with V2 override",
			profileEnvVar:      envProfile,
			profileEnvValue:    "http-export",
			originalServiceKey: "MyAppService-" + interfaces.ProfileSuffixPlaceholder,
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
				err := os.Setenv(test.profileEnvVar, test.profileEnvValue)
				require.NoError(t, err)
			}
			if len(test.serviceKeyEnvValue) > 0 {
				err := os.Setenv(envServiceKey, test.serviceKeyEnvValue)
				require.NoError(t, err)
			}
			defer os.Clearenv()

			if len(test.serviceKeyCommandLineOverride) > 0 {
				sdk.commandLine.serviceKeyOverride = test.serviceKeyCommandLineOverride
			}

			sdk.serviceKey = test.originalServiceKey
			sdk.setServiceKey(test.profile)

			assert.Equal(t, test.expectedServiceKey, sdk.serviceKey)
		})
	}
}

func TestMakeItStop(t *testing.T) {
	stopCalled := false

	sdk := Service{
		ctx: contextGroup{
			stop: func() {
				stopCalled = true
			},
		},
		lc: logger.NewMockClient(),
	}

	sdk.MakeItStop()
	require.True(t, stopCalled, "Cancel function set at svc.stop should be called if set")

	sdk.ctx.stop = nil
	sdk.MakeItStop() //should avoid nil pointer
}

func TestFindMatchingFunction(t *testing.T) {
	svc := Service{
		lc:                       lc,
		serviceKey:               "MyAppService",
		profileSuffixPlaceholder: interfaces.ProfileSuffixPlaceholder,
	}

	configurable := reflect.ValueOf(NewConfigurable(svc.lc))

	tests := []struct {
		Name         string
		FunctionName string
		ExpectError  bool
	}{
		{"valid exact match AddTags", "AddTags", false},
		{"valid exact match HTTPExport", "HTTPExport", false},
		{"valid starts with match AddTags", "AddTagsExtra", false},
		{"valid starts with match HTTPExport", "HTTPExport2", false},
		{"invalid no match", "Bogus", true},
		{"invalid doesn't start with", "NextHTTPExport", true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_, _, err := svc.findMatchingFunction(configurable, test.FunctionName)
			if test.ExpectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_EventClient(t *testing.T) {
	actual := target.EventClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.EventClientName: func(get di.Get) interface{} {
			return v2clients.NewEventClient(baseUrl + "59880")
		},
	})

	actual = target.EventClient()
	assert.NotNil(t, actual)
}

func TestService_CommandClient(t *testing.T) {
	actual := target.CommandClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.CommandClientName: func(get di.Get) interface{} {
			return v2clients.NewCommandClient(baseUrl + "59882")
		},
	})

	actual = target.CommandClient()
	assert.NotNil(t, actual)
}

func TestService_DeviceServiceClient(t *testing.T) {
	actual := target.DeviceServiceClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.DeviceServiceClientName: func(get di.Get) interface{} {
			return v2clients.NewDeviceServiceClient(baseUrl + "59881")
		},
	})

	actual = target.DeviceServiceClient()
	assert.NotNil(t, actual)

}

func TestService_DeviceProfileClient(t *testing.T) {
	actual := target.DeviceProfileClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.DeviceProfileClientName: func(get di.Get) interface{} {
			return v2clients.NewDeviceProfileClient(baseUrl + "59881")
		},
	})

	actual = target.DeviceProfileClient()
	assert.NotNil(t, actual)
}

func TestService_DeviceClient(t *testing.T) {
	actual := target.DeviceClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.DeviceClientName: func(get di.Get) interface{} {
			return v2clients.NewDeviceClient(baseUrl + "59881")
		},
	})

	actual = target.DeviceClient()
	assert.NotNil(t, actual)

}

func TestService_NotificationClient(t *testing.T) {
	actual := target.NotificationClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.NotificationClientName: func(get di.Get) interface{} {
			return v2clients.NewNotificationClient(baseUrl + "59860")
		},
	})

	actual = target.NotificationClient()
	assert.NotNil(t, actual)

}

func TestService_SubscriptionClient(t *testing.T) {
	actual := target.SubscriptionClient()
	assert.Nil(t, actual)

	dic.Update(di.ServiceConstructorMap{
		container.SubscriptionClientName: func(get di.Get) interface{} {
			return v2clients.NewSubscriptionClient(baseUrl + "59860")
		},
	})

	actual = target.SubscriptionClient()
	assert.NotNil(t, actual)
}

func TestService_LoggingClient(t *testing.T) {
	actual := target.LoggingClient()
	assert.NotNil(t, actual)
}
