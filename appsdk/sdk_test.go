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
	"io/ioutil"
	http "net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	triggerHttp "github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
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
	assert.NotNil(t, err, "Should return error")
	assert.Equal(t, err.Error(), "No transforms provided to pipeline", "Incorrect error message received")
}
func TestSetPipelineNoTransformsNil(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
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
	trx := sdk.HTTPPost("http://url", "")
	assert.NotNil(t, trx, "return result from HTTPPost should not be nil")
}
func TestHTTPPostJSON(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	msgStr := "POST ME"
	path := "/"
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(([]byte)("RESPONSE"))
		readMsg, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		assert.Equal(t, msgStr, (string)(readMsg), "Invalid msg received %v, expected %v", (string)(readMsg), msgStr)
		assert.Equal(t, "application/json", r.Header.Get("Content-type"), "Unexpected content-type received %s, expected %s", r.Header.Get("Content-type"), "application/xml")
		assert.Equal(t, path, r.URL.EscapedPath(), "Invalid path received %s, expected %s", r.URL.EscapedPath(), path)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	trx := sdk.HTTPPostJSON(ts.URL)
	assert.NotNil(t, trx, "return result from HTTPPostJSON should not be nil")

	ctx := &appcontext.Context{
		LoggingClient: lc,
	}
	result, data := trx(ctx, msgStr)
	assert.True(t, result, "continuePipeline should be true")
	assert.Equal(t, "RESPONSE", (string)((data).([]byte)), "response should be \"RESPONSE\"")
}
func TestHTTPPostXML(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	msgStr := "POST ME"
	path := "/"
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(([]byte)("RESPONSE"))
		readMsg, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		assert.Equal(t, msgStr, (string)(readMsg), "Invalid msg received %v, expected %v", (string)(readMsg), msgStr)
		assert.Equal(t, "application/xml", r.Header.Get("Content-type"), "Unexpected content-type received %s, expected %s", r.Header.Get("Content-type"), "application/xml")
		assert.Equal(t, path, r.URL.EscapedPath(), "Invalid path received %s, expected %s", r.URL.EscapedPath(), path)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	trx := sdk.HTTPPostXML(ts.URL)
	assert.NotNil(t, trx, "return result from HTTPPostXML should not be nil")

	ctx := &appcontext.Context{
		LoggingClient: lc,
	}
	result, data := trx(ctx, msgStr)
	assert.True(t, result, "continuePipeline should be true")
	assert.Equal(t, "RESPONSE", (string)((data).([]byte)), "response should be \"RESPONSE\"")
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
	runtime := runtime.GolangRuntime{Transforms: sdk.transforms}
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
	runtime := runtime.GolangRuntime{Transforms: sdk.transforms}
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
