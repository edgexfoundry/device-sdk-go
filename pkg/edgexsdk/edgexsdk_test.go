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

package edgexsdk

import (
	"reflect"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	"github.com/stretchr/testify/assert"
)

var lc logger.LoggingClient

func init() {
	lc = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
}
func TestSetPipelineNoTransforms(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	err := sdk.SetPipeline()
	if err == nil {
		t.Fatal("Should return error")
	}
	assert.Equal(t, err.Error(), "No transforms provided to pipeline", "Incorrect error message received")
}
func TestSetPipelineNoTransformsNil(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	transform1 := func(edgexcontext excontext.Context, params ...interface{}) (bool, interface{}) {
		return false, nil
	}
	err := sdk.SetPipeline(transform1)
	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, len(sdk.transforms), 1, "sdk.Transforms should have 1 transform")
}

func TestFilterByDeviceID(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	deviceIDs := []string{"GS1-AC-Drive01"}

	trx := sdk.FilterByDeviceID(deviceIDs)
	assert.NotNil(t, trx, "return result from FilterByDeviceID should not be nil")
}

func TestFilterByValueDescriptor(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	valueDescriptors := []string{"GS1-AC-Drive01"}

	trx := sdk.FilterByValueDescriptor(valueDescriptors)
	assert.NotNil(t, trx, "return result from FilterByValueDescriptor should not be nil")
}

func TestTransformToXML(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	trx := sdk.TransformToXML()
	assert.NotNil(t, trx, "return result from TransformToXML should not be nil")
}

func TestTransformToJSON(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	trx := sdk.TransformToJSON()
	assert.NotNil(t, trx, "return result from TransformToJSON should not be nil")
}

func TestHTTPPost(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	trx := sdk.HTTPPost("http://url")
	assert.NotNil(t, trx, "return result from HTTPPost should not be nil")
}

// func TestMakeItRun(t *testing.T) {

// 	sdk := AppFunctionsSDK{
// 		config: common.ConfigurationStruct{
// 			Bindings: []common.Binding{
// 				common.Binding{
// 					Type: "http",
// 				},
// 			},
// 		},
// 	}

// 	sdk.MakeItRun()

// }
func IsInstanceOf(objectPtr, typePtr interface{}) bool {
	return reflect.TypeOf(objectPtr) == reflect.TypeOf(typePtr)
}
func TestSetupHTTPTrigger(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Binding: common.Binding{
				Type: "htTp",
			},
		},
	}
	runtime := runtime.GolangRuntime{Transforms: sdk.transforms}
	trigger := sdk.setupTrigger(sdk.config, runtime)
	result := IsInstanceOf(trigger, (*http.Trigger)(nil))
	if !result {
		t.Error("Expected HTTP Trigger")
	}
}
func TestSetupMessageBusTrigger(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		config: common.ConfigurationStruct{
			Binding: common.Binding{
				Type: "meSsaGebus",
			},
		},
	}
	runtime := runtime.GolangRuntime{Transforms: sdk.transforms}
	trigger := sdk.setupTrigger(sdk.config, runtime)
	result := IsInstanceOf(trigger, (*messagebus.Trigger)(nil))
	if !result {
		t.Error("Expected Message Bus Trigger")
	}
}
