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

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/assert"
)

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

func TestConfigurableFilterByValueDescriptor(t *testing.T) {
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
		{"Empty Parameters", map[string]string{ValueDescriptors: ""}, false},
		{"Valid Parameters", map[string]string{ValueDescriptors: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03"}, false},
		{"Empty FilterOut Parameters", map[string]string{ValueDescriptors: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: ""}, true},
		{"Valid FilterOut Parameters", map[string]string{ValueDescriptors: "GS1-AC-Drive01, GS1-AC-Drive02, GS1-AC-Drive03", FilterOut: "true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := configurable.FilterByValueDescriptor(tt.params)
			if tt.expectNil {
				assert.Nil(t, trx, "return result from FilterByValueDescriptor should be nil")
			} else {
				assert.NotNil(t, trx, "return result from FilterByValueDescriptor should not be nil")
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

func TestConfigurableMQTTSend(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{
		Sdk: &AppFunctionsSDK{
			LoggingClient: lc,
		},
	}

	params := make(map[string]string)
	addr := models.Addressable{}
	params[PersistOnError] = "true"
	trx := configurable.MQTTSend(params, addr)
	assert.NotNil(t, trx, "return result from MQTTSend should not be nil")
}

func TestConfigurableSetOutputData(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{}

	trx := configurable.SetOutputData()
	assert.NotNil(t, trx, "return result from SetOutputData should not be nil")
}

func TestConfigurableMarkAsPushed(t *testing.T) {
	configurable := AppFunctionsSDKConfigurable{}

	trx := configurable.MarkAsPushed()
	assert.NotNil(t, trx, "return result from MarkAsPushed should not be nil")
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
	assert.NotNil(t, trx, "return result from MQTTSend should not be nil")
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
	assert.NotNil(t, trx, "return result from MQTTSend should not be nil")
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
	assert.NotNil(t, trx, "return result from MQTTSend should not be nil")
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
	assert.NotNil(t, trx, "return result from MQTTSend should not be nil")
}
