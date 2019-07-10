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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/assert"
)

func TestSetFunctionsPipelineNoTransforms(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	err := sdk.SetFunctionsPipeline()
	assert.NotNil(t, err, "Should return error")
	assert.Equal(t, err.Error(), "No transforms provided to pipeline", "Incorrect error message received")
}

func TestSetFunctionsPipelineOneTransformNil(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
		runtime:       &runtime.GolangRuntime{},
	}
	transform1 := func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
		return false, nil
	}
	err := sdk.SetFunctionsPipeline(transform1)
	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, len(sdk.transforms), 1, "sdk.Transforms should have 1 transform")
}

func TestDeviceNameFilter(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	deviceIDs := []string{"GS1-AC-Drive01"}

	trx := sdk.DeviceNameFilter(deviceIDs)
	assert.NotNil(t, trx, "return result from DeviceNameFilter should not be nil")
}

func TestValueDescriptorFilter(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	valueDescriptors := []string{"GS1-AC-Drive01"}

	trx := sdk.ValueDescriptorFilter(valueDescriptors)
	assert.NotNil(t, trx, "return result from ValueDescriptorFilter should not be nil")
}

func TestAESTransform(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}

	encryptionDeets := models.EncryptionDetails{
		Key:        "key string",
		InitVector: "init vector string",
	}

	trx := sdk.AESTransform(encryptionDeets)
	assert.NotNil(t, trx, "return result from AESTransform should not be nil")
}

func TestXMLTransform(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	trx := sdk.XMLTransform()
	assert.NotNil(t, trx, "return result from XMLTransform should not be nil")
}

func TestJSONTransform(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	trx := sdk.JSONTransform()
	assert.NotNil(t, trx, "return result from JSONTransform should not be nil")
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
	pushedHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Contains(t, r.URL.EscapedPath(), "234")
	}
	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	ts2 := httptest.NewServer(http.HandlerFunc(pushedHandler))
	defer ts.Close()
	defer ts2.Close()
	//Setup eventClient
	params := types.EndpointParams{
		ServiceKey:  clients.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: false,
		Url:         ts2.URL,
		Interval:    1000,
	}
	eventClient := coredata.NewEventClient(params, startup.Endpoint{RegistryClient: nil})

	context = &appcontext.Context{
		EventID:       "234",
		LoggingClient: lc,
		EventClient:   eventClient,
	}
	// used to ensure MarkAsPushed is called
	trx := sdk.HTTPPostJSON(ts.URL)
	assert.NotNil(t, trx, "return result from HTTPPostJSON should not be nil")

	result, data := trx(context, msgStr)
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
	// used to ensure MarkAsPushed is called
	pushedHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Contains(t, r.URL.EscapedPath(), "123")
	}
	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	ts2 := httptest.NewServer(http.HandlerFunc(pushedHandler))
	defer ts.Close()
	defer ts2.Close()
	//Setup eventClient
	params := types.EndpointParams{
		ServiceKey:  clients.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: false,
		Url:         ts2.URL,
		Interval:    1000,
	}
	eventClient := coredata.NewEventClient(params, startup.Endpoint{RegistryClient: nil})

	context = &appcontext.Context{
		EventID:       "123",
		LoggingClient: lc,
		EventClient:   eventClient,
	}

	trx := sdk.HTTPPostXML(ts.URL)
	assert.NotNil(t, trx, "return result from HTTPPostXML should not be nil")

	result, data := trx(context, msgStr)
	assert.True(t, result, "continuePipeline should be true")
	assert.Equal(t, "RESPONSE", (string)((data).([]byte)), "response should be \"RESPONSE\"")
}

func TestGZIPTransform(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	trx := sdk.GZIPTransform()
	assert.NotNil(t, trx, "return result from GZIPTransform should not be nil")
}

func TestZLIBTransform(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}
	trx := sdk.ZLIBTransform()
	assert.NotNil(t, trx, "return result from ZLIBTransform should not be nil")
}

func TestMQTTSend(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}

	addr := models.Addressable{}
	trx := sdk.MQTTSend(addr, "cert", "key", byte(0), false, false)
	assert.NotNil(t, trx, "return result from MQTTSend should not be nil")
}

func TestSetOutputData(t *testing.T) {
	sdk := AppFunctionsSDK{
		LoggingClient: lc,
	}

	trx := sdk.SetOutputData()
	assert.NotNil(t, trx, "return result from SetOutputData should not be nil")
}
