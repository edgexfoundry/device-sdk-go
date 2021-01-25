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

package appcontext

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	localURL "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/urlclient/local"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplete(t *testing.T) {
	ctx := Context{}
	testData := "output data"
	ctx.Complete([]byte(testData))
	assert.Equal(t, []byte(testData), ctx.OutputData)
}

var eventClient coredata.EventClient
var lc logger.LoggingClient

func init() {
	eventClient = coredata.NewEventClient(localURL.New("http://test" + clients.ApiEventRoute))
	lc = logger.NewMockClient()
}

func TestPushToCore(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("newId"))
		if r.Method != http.MethodPost {
			t.Errorf("expected http method is POST, active http method is : %s", r.Method)
		}
		url := clients.ApiEventRoute
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}

	}))

	defer ts.Close()
	eventClient = coredata.NewEventClient(localURL.New(ts.URL + clients.ApiEventRoute))
	ctx := Context{
		EventClient:   eventClient,
		LoggingClient: lc,
	}
	expectedEvent := &dtos.Event{
		Versionable: common.NewVersionable(),
		DeviceName:  "device-name",
		Readings: []dtos.BaseReading{
			{
				Versionable:  common.NewVersionable(),
				DeviceName:   "device-name",
				ResourceName: "device-resource",
				ValueType:    v2.ValueTypeString,
				SimpleReading: dtos.SimpleReading{
					Value: "value",
				},
			},
		},
	}
	actualEvent, err := ctx.PushToCoreData("device-name", "device-resource", "value")
	require.NoError(t, err)

	assert.NotNil(t, actualEvent)
	assert.Equal(t, expectedEvent.ApiVersion, actualEvent.ApiVersion)
	assert.Equal(t, expectedEvent.DeviceName, actualEvent.DeviceName)
	assert.True(t, len(expectedEvent.Readings) == 1)
	assert.Equal(t, expectedEvent.Readings[0].DeviceName, actualEvent.Readings[0].DeviceName)
	assert.Equal(t, expectedEvent.Readings[0].ResourceName, actualEvent.Readings[0].ResourceName)
	assert.Equal(t, expectedEvent.Readings[0].Value, actualEvent.Readings[0].Value)
	assert.Equal(t, expectedEvent.Readings[0].ValueType, actualEvent.Readings[0].ValueType)
	assert.Equal(t, expectedEvent.Readings[0].ApiVersion, actualEvent.Readings[0].ApiVersion)
}

func TestSetRetryData(t *testing.T) {
	ctx := Context{}
	testData := "output data"
	ctx.SetRetryData([]byte(testData))
	assert.Equal(t, []byte(testData), ctx.RetryData)
}
