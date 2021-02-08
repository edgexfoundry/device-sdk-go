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

package appsdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigurableFilterByProfileName(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameters", map[string]string{"": ""}, true},
		{"Empty Parameters", map[string]string{ProfileNames: ""}, false},
		{"Valid Parameters", map[string]string{ProfileNames: "GS1-AC-Drive, GS0-DC-Drive, GSX-ACDC-Drive"}, false},
		{"Empty FilterOut Parameters", map[string]string{ProfileNames: "GS1-AC-Drive, GS0-DC-Drive, GSX-ACDC-Drive", FilterOut: ""}, true},
		{"Valid FilterOut Parameters", map[string]string{ProfileNames: "GS1-AC-Drive, GS0-DC-Drive, GSX-ACDC-Drive", FilterOut: "true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.FilterByProfileName(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from FilterByProfileName should be nil")
			} else {
				assert.NotNil(t, trx, "return result from FilterByProfileName should not be nil")
			}
		})
	}
}

func TestConfigurableFilterByDeviceName(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameters", map[string]string{"": ""}, true},
		{"Empty Parameters", map[string]string{DeviceNames: ""}, false},
		{"Valid Parameters", map[string]string{DeviceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03"}, false},
		{"Empty FilterOut Parameters", map[string]string{DeviceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: ""}, true},
		{"Valid FilterOut Parameters", map[string]string{DeviceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: "true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.FilterByDeviceName(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from FilterByDeviceName should be nil")
			} else {
				assert.NotNil(t, trx, "return result from FilterByDeviceName should not be nil")
			}
		})
	}
}

func TestConfigurableFilterByResourceName(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameters", map[string]string{"": ""}, true},
		{"Empty Parameters", map[string]string{ResourceNames: ""}, false},
		{"Valid Parameters", map[string]string{ResourceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03"}, false},
		{"Empty FilterOut Parameters", map[string]string{ResourceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: ""}, true},
		{"Valid FilterOut Parameters", map[string]string{ResourceNames: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: "true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.FilterByResourceName(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from FilterByResourceName should be nil")
			} else {
				assert.NotNil(t, trx, "return result from FilterByResourceName should not be nil")
			}
		})
	}
}

func TestConfigurableTransformToXML(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{}

	trx := configurable.TransformToXML()
	assert.NotNil(t, trx, "return result from TransformToXML should not be nil")
}

func TestConfigurableTransformToJSON(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{}

	trx := configurable.TransformToJSON()
	assert.NotNil(t, trx, "return result from TransformToJSON should not be nil")
}

func TestConfigurableHTTPPost(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	params := make(map[string]string)

	// no url in params
	params[""] = "fake"
	trx := configurable.HTTPPost(params)
	assert.Nil(t, trx, "return result from HTTPPost should be nil")

	// no mime type
	params[Url] = "http://url"
	trx = configurable.HTTPPost(params)
	assert.Nil(t, trx, "return result from HTTPPost should be nil")

	params[MimeType] = ""
	trx = configurable.HTTPPost(params)
	assert.NotNil(t, trx, "return result from HTTPPost should not be nil")

	params[PersistOnError] = "true"
	trx = configurable.HTTPPost(params)
	assert.NotNil(t, trx, "return result from HTTPPost should not be nil")
}

func TestConfigurableHTTPPostJSON(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	params := make(map[string]string)

	// no url in params
	params[""] = ""
	trx := configurable.HTTPPostJSON(params)
	assert.Nil(t, trx, "return result from HTTPPostJSON should be nil")

	params[Url] = "http://url"
	params[PersistOnError] = "true"
	trx = configurable.HTTPPostJSON(params)
	assert.NotNil(t, trx, "return result from HTTPPostJSON should not be nil")
}

func TestConfigurableHTTPPostXML(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	params := make(map[string]string)

	// no url in params
	params[""] = ""
	trx := configurable.HTTPPostXML(params)
	assert.Nil(t, trx, "return result from HTTPPostXML should be nil")

	params[Url] = "http://url"
	params[PersistOnError] = "true"
	trx = configurable.HTTPPostXML(params)
	assert.NotNil(t, trx, "return result from HTTPPostXML should not be nil")
}

func TestConfigurableSetOutputData(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	tests := []struct {
		name      string
		params    map[string]string
		expectNil bool
	}{
		{"Non Existent Parameter", map[string]string{}, false},
		{"Valid Parameter With Value", map[string]string{ResponseContentType: "application/json"}, false},
		{"Valid Parameter Without Value", map[string]string{ResponseContentType: ""}, false},
		{"Unknown Parameter", map[string]string{"Unknown": "scary/text"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.SetOutputData(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from SetOutputData should be nil")
			} else {
				assert.NotNil(t, trx, "return result from SetOutputData should not be nil")
			}
		})
	}
}

func TestBatchByCount(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	params := make(map[string]string)
	params[BatchThreshold] = "30"
	trx := configurable.BatchByCount(params)
	assert.NotNil(t, trx, "return result from BatchByCount should not be nil")
}
func TestBatchByTime(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	params := make(map[string]string)
	params[TimeInterval] = "10"
	trx := configurable.BatchByTime(params)
	assert.NotNil(t, trx, "return result from BatchByTime should not be nil")
}
func TestBatchByTimeAndCount(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	params := make(map[string]string)
	params[BatchThreshold] = "30"
	params[TimeInterval] = "10"

	trx := configurable.BatchByTimeAndCount(params)
	assert.NotNil(t, trx, "return result from BatchByTimeAndCount should not be nil")
}

func TestJSONLogic(t *testing.T) {
	params := make(map[string]string)
	params[Rule] = "{}"

	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}
	trx := configurable.JSONLogic(params)
	assert.NotNil(t, trx, "return result from JSONLogic should not be nil")

}
func TestConfigurableMQTTSecretSend(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	params := make(map[string]string)
	params[BrokerAddress] = "mqtt://broker:8883"
	params[Topic] = "topic"
	params[SecretPath] = "/path"
	params[ClientID] = "clientid"
	params[Qos] = "0"
	params[Retain] = "true"
	params[AutoReconnect] = "true"
	params[SkipVerify] = "true"
	params[PersistOnError] = "false"
	params[AuthMode] = "none"

	trx := configurable.MQTTSecretSend(params)
	assert.NotNil(t, trx, "return result from MQTTSecretSend should not be nil")
}

func TestAppFunctionsSDKConfigurable_AddTags(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	tests := []struct {
		Name      string
		ParamName string
		TagsSpec  string
		ExpectNil bool
	}{
		{"Good - non-empty list", Tags, "GatewayId:HoustonStore000123,Latitude:29.630771,Longitude:-95.377603", false},
		{"Good - empty list", Tags, "", false},
		{"Bad - No : separator", Tags, "GatewayId HoustonStore000123, Latitude:29.630771,Longitude:-95.377603", true},
		{"Bad - Missing value", Tags, "GatewayId:,Latitude:29.630771,Longitude:-95.377603", true},
		{"Bad - Missing key", Tags, "GatewayId:HoustonStore000123,:29.630771,Longitude:-95.377603", true},
		{"Bad - Missing key & value", Tags, ":,:,:", true},
		{"Bad - No Tags parameter", "NotTags", ":,:,:", true},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			params := make(map[string]string)
			params[testCase.ParamName] = testCase.TagsSpec

			transform := configurable.AddTags(params)
			assert.Equal(t, testCase.ExpectNil, transform == nil)
		})
	}
}
