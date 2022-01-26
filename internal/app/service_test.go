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

	"github.com/google/uuid"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	builtin "github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	triggerHttp "github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/webserver"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http"
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

	os.Exit(m.Run())
}

func IsInstanceOf(objectPtr, typePtr interface{}) bool {
	return reflect.TypeOf(objectPtr) == reflect.TypeOf(typePtr)
}

func TestAddRoute(t *testing.T) {
	router := mux.NewRouter()

	ws := webserver.NewWebServer(dic, router, uuid.NewString())

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

func TestAddBackgroundPublisherNoTopic(t *testing.T) {
	sdk := Service{
		config: &common.ConfigurationStruct{},
	}

	p, err := sdk.AddBackgroundPublisher(1)

	require.Error(t, err)
	require.Nil(t, p)
}

func TestAddBackgroundPublisherMessageBus(t *testing.T) {
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMessageBus,
				EdgexMessageBus: common.MessageBusConfig{
					PublishHost: common.PublishHostInfo{
						PublishTopic: "topic",
					},
				},
			},
		}}

	p, err := sdk.AddBackgroundPublisher(1)

	require.NoError(t, err)

	pub, ok := p.(*backgroundPublisher)

	if !ok {
		assert.Fail(t, fmt.Sprintf("Unexpected BackgroundPublisher implementation encountered: %T", pub))
	}

	require.NotNil(t, pub.output, "publisher should have an output channel set")
	require.NotNil(t, sdk.backgroundPublishChannel, "svc should have a background channel set for passing to trigger initialization")
	require.Equal(t, sdk.config.Trigger.EdgexMessageBus.PublishHost.PublishTopic, pub.topic)

	// compare addresses since types will not match
	assert.Equal(t, fmt.Sprintf("%p", sdk.backgroundPublishChannel), fmt.Sprintf("%p", pub.output),
		"same channel should be referenced by the BackgroundPublisher and the SDK.")
}

func TestAddBackgroundPublisher_Arbitrary(t *testing.T) {
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: "NOT MQTT OR HTTP",
				EdgexMessageBus: common.MessageBusConfig{
					PublishHost: common.PublishHostInfo{
						PublishTopic: "topic",
					},
				},
			},
		}}

	p, err := sdk.AddBackgroundPublisher(1)

	require.NoError(t, err)

	pub, ok := p.(*backgroundPublisher)

	if !ok {
		assert.Fail(t, fmt.Sprintf("Unexpected BackgroundPublisher implementation encountered: %T", pub))
	}

	require.NotNil(t, pub.output, "publisher should have an output channel set")
	require.NotNil(t, sdk.backgroundPublishChannel, "svc should have a background channel set for passing to trigger initialization")
	require.Equal(t, sdk.config.Trigger.EdgexMessageBus.PublishHost.PublishTopic, pub.topic)

	// compare addresses since types will not match
	assert.Equal(t, fmt.Sprintf("%p", sdk.backgroundPublishChannel), fmt.Sprintf("%p", pub.output),
		"same channel should be referenced by the BackgroundPublisher and the SDK.")
}

func TestAddBackgroundPublisher_Custom_Topic(t *testing.T) {
	sdk := Service{config: &common.ConfigurationStruct{}}

	topic := uuid.NewString()

	p, err := sdk.AddBackgroundPublisherWithTopic(1, topic)

	require.NoError(t, err)

	pub, ok := p.(*backgroundPublisher)

	if !ok {
		assert.Fail(t, fmt.Sprintf("Unexpected BackgroundPublisher implementation encountered: %T", pub))
	}

	require.NotNil(t, pub.output, "publisher should have an output channel set")
	require.NotNil(t, sdk.backgroundPublishChannel, "svc should have a background channel set for passing to trigger initialization")
	require.Equal(t, topic, pub.topic)

	// compare addresses since types will not match
	assert.Equal(t, fmt.Sprintf("%p", sdk.backgroundPublishChannel), fmt.Sprintf("%p", pub.output),
		"same channel should be referenced by the BackgroundPublisher and the SDK.")
}

func TestAddBackgroundPublisher_MQTT(t *testing.T) {
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMQTT,
				EdgexMessageBus: common.MessageBusConfig{
					PublishHost: common.PublishHostInfo{
						PublishTopic: "topic",
					},
				},
			},
		}}

	pub, err := sdk.AddBackgroundPublisher(1)

	require.Error(t, err)
	require.Nil(t, pub)
}

func TestAddBackgroundPublisher_HTTP(t *testing.T) {
	sdk := Service{
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeHTTP,
				EdgexMessageBus: common.MessageBusConfig{
					PublishHost: common.PublishHostInfo{
						PublishTopic: "topic",
					},
				},
			},
		}}

	pub, err := sdk.AddBackgroundPublisher(1)

	require.Error(t, err)
	require.Nil(t, pub)
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

	testRuntime := runtime.NewGolangRuntime("", nil, dic)
	testRuntime.SetDefaultFunctionsPipeline(nil)

	sdk.runtime = testRuntime

	trigger := sdk.setupTrigger(sdk.config)
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
	testRuntime := runtime.NewGolangRuntime("", nil, dic)
	testRuntime.SetDefaultFunctionsPipeline(nil)

	sdk.runtime = testRuntime

	trigger := sdk.setupTrigger(sdk.config)
	result := IsInstanceOf(trigger, (*messagebus.Trigger)(nil))
	assert.True(t, result, "Expected Instance of Message Bus Trigger")
}

func TestSetDefaultFunctionsPipelineNoTransforms(t *testing.T) {
	sdk := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMessageBus,
			},
		},
	}
	err := sdk.SetDefaultFunctionsPipeline()
	require.Error(t, err, "There should be an error")
	assert.Equal(t, "no transforms provided to pipeline", err.Error())
}

func TestSetDefaultFunctionsPipelineOneTransform(t *testing.T) {
	service := Service{
		lc:      lc,
		runtime: runtime.NewGolangRuntime("", nil, dic),
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMessageBus,
			},
		},
	}
	function := func(appContext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
		return true, nil
	}

	err := service.SetDefaultFunctionsPipeline(function)
	require.NoError(t, err)
}

func TestService_AddFunctionsPipelineForTopics(t *testing.T) {
	service := Service{
		lc:      lc,
		dic:     dic,
		runtime: runtime.NewGolangRuntime("", nil, dic),
		config: &common.ConfigurationStruct{
			Trigger: common.TriggerInfo{
				Type: TriggerTypeMessageBus,
			},
		},
	}

	tags := builtin.NewTags(nil)

	transforms := []interfaces.AppFunction{tags.AddTags}

	err := service.SetDefaultFunctionsPipeline(transforms...)

	require.NoError(t, err)

	defaultTopics := []string{"#"}

	tests := []struct {
		name        string
		id          string
		trigger     string
		topics      []string
		transforms  []interfaces.AppFunction
		expectError bool
	}{
		{"Happy Path", "123", TriggerTypeMessageBus, defaultTopics, transforms, false},
		{"Duplicate Id", interfaces.DefaultPipelineId, TriggerTypeMessageBus, defaultTopics, transforms, true},
		{"Happy Path Custom", "124", "CUSTOM TRIGGER", defaultTopics, transforms, false},
		{"Empty Topic", "125", TriggerTypeMessageBus, []string{" "}, transforms, true},
		{"Empty Topics", "126", TriggerTypeMessageBus, nil, transforms, true},
		{"No Transforms", "127", TriggerTypeMessageBus, defaultTopics, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service.config.Trigger.Type = test.trigger

			err := service.AddFunctionsPipelineForTopics(test.id, test.topics, test.transforms...)
			if test.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			actual := service.runtime.GetPipelineById(test.id)
			assert.Equal(t, transforms, actual.Transforms)
		})
	}
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

func TestLoadConfigurableFunctionPipelinesDefaultNotFound(t *testing.T) {
	service := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder:    "Bogus",
					PerTopicPipelines: make(map[string]common.TopicPipeline),
					Functions:         make(map[string]common.PipelineFunction),
				},
			},
		},
	}

	tests := []struct {
		name                   string
		defaultExecutionOrder  string
		perTopicExecutionOrder string
	}{
		{"Default Not Found", "Bogus", ""},
		{"PerTopicNotFound", "", "Bogus"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service.config.Writable.Pipeline.ExecutionOrder = test.defaultExecutionOrder
			if len(test.perTopicExecutionOrder) > 0 {
				service.config.Writable.Pipeline.PerTopicPipelines["bogus"] = common.TopicPipeline{
					Id:             "bogus",
					Topics:         "#",
					ExecutionOrder: test.perTopicExecutionOrder,
				}
			}

			appFunctions, err := service.LoadConfigurableFunctionPipelines()
			require.Error(t, err, "expected error for function not found in config")
			assert.Contains(t, err.Error(), "function 'Bogus' configuration not found in Pipeline.Functions section")
			assert.Nil(t, appFunctions, "expected app functions list to be nil")

		})
	}
}

func TestLoadConfigurableFunctionPipelinesNotABuiltInSdkFunction(t *testing.T) {
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

	appFunctions, err := sdk.LoadConfigurableFunctionPipelines()
	require.Error(t, err, "expected error")
	assert.Contains(t, err.Error(), "function Bogus is not a built in SDK function")
	assert.Nil(t, appFunctions, "expected app functions list to be nil")
}

func TestLoadConfigurableFunctionPipelinesNumFunctions(t *testing.T) {
	expectedPipelinesCount := 2
	expectedTransformsCount := 3
	perTopicPipelineId := "pre-topic"

	transforms := make(map[string]common.PipelineFunction)
	transforms["FilterByDeviceName"] = common.PipelineFunction{
		Parameters: map[string]string{"DeviceNames": "Random-Float-Device, Random-Integer-Device"},
	}
	transforms["Transform"] = common.PipelineFunction{
		Parameters: map[string]string{TransformType: TransformXml},
	}
	transforms["SetResponseData"] = common.PipelineFunction{}

	sdk := Service{
		lc: lc,
		config: &common.ConfigurationStruct{
			Writable: common.WritableInfo{
				Pipeline: common.PipelineInfo{
					ExecutionOrder: "FilterByDeviceName, Transform, SetResponseData",
					PerTopicPipelines: map[string]common.TopicPipeline{
						perTopicPipelineId: {
							Id:             perTopicPipelineId,
							Topics:         "#",
							ExecutionOrder: "FilterByDeviceName, Transform, SetResponseData",
						},
					},
					Functions: transforms,
				},
			},
		},
	}

	pipelines, err := sdk.LoadConfigurableFunctionPipelines()
	require.NoError(t, err)
	require.NotNil(t, pipelines, "expected app pipelines list to be set")
	assert.Equal(t, expectedPipelinesCount, len(pipelines))

	pipeline, found := pipelines[interfaces.DefaultPipelineId]
	require.True(t, found)
	assert.Equal(t, expectedTransformsCount, len(pipeline.Transforms))

	pipeline, found = pipelines[perTopicPipelineId]
	require.True(t, found)
	assert.Equal(t, expectedTransformsCount, len(pipeline.Transforms))
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
			return clients.NewEventClient(baseUrl + "59880")
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
			return clients.NewCommandClient(baseUrl + "59882")
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
			return clients.NewDeviceServiceClient(baseUrl + "59881")
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
			return clients.NewDeviceProfileClient(baseUrl + "59881")
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
			return clients.NewDeviceClient(baseUrl + "59881")
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
			return clients.NewNotificationClient(baseUrl + "59860")
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
			return clients.NewSubscriptionClient(baseUrl + "59860")
		},
	})

	actual = target.SubscriptionClient()
	assert.NotNil(t, actual)
}

func TestService_LoggingClient(t *testing.T) {
	actual := target.LoggingClient()
	assert.NotNil(t, actual)
}

func TestService_BuildContext(t *testing.T) {
	sdk := Service{
		dic: dic,
	}

	correlationId := uuid.NewString()

	contentType := uuid.NewString()

	appCtx := sdk.BuildContext(correlationId, contentType)

	require.NotNil(t, appCtx)

	require.Equal(t, correlationId, appCtx.CorrelationID())
	require.Equal(t, contentType, appCtx.InputContentType())

	castCtx := appCtx.(*appfunction.Context)

	require.NotNil(t, castCtx)
	require.Equal(t, dic, castCtx.Dic)
}
