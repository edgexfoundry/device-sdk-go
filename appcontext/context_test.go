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

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	localURL "github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
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
func TestMarkAsPushedNoClient(t *testing.T) {
	ctx := Context{
		LoggingClient: lc,
	}
	err := ctx.MarkAsPushed()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'CoreData' is missing from Clients")
}

func TestMarkAsPushedNoEventIdOrChecksum(t *testing.T) {
	eventClient = coredata.NewEventClient(localURL.New("http://test" + clients.ApiEventRoute))
	ctx := Context{
		LoggingClient: lc,
		EventClient:   eventClient,
	}
	err := ctx.MarkAsPushed()
	assert.NotNil(t, err)
	assert.Equal(t, "No EventID or EventChecksum Provided", err.Error())
}

func TestMarkAsPushedNoChecksum(t *testing.T) {
	testChecksum := "checksumValue"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPut {
			t.Errorf("expected http method is PUT, active http method is : %s", r.Method)
		}
		url := clients.ApiEventRoute + "/checksum/" + testChecksum
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}
	}))

	defer ts.Close()
	eventClient = coredata.NewEventClient(localURL.New(ts.URL + clients.ApiEventRoute))

	ctx := Context{
		EventChecksum: testChecksum,
		CorrelationID: "correlationId",
		EventClient:   eventClient,
		LoggingClient: lc,
	}
	err := ctx.MarkAsPushed()
	assert.Nil(t, err)

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
	newEvent := &models.Event{
		ID:     "newId",
		Device: "device-name",
		Origin: 1567802840199266000,
		Readings: []models.Reading{
			{
				Device:      "device-name",
				Name:        "device-resource",
				Value:       "value",
				BinaryValue: []uint8(nil),
			},
		},
	}
	result, err := ctx.PushToCoreData("device-name", "device-resource", "value")
	require.NoError(t, err)

	assert.NotNil(t, result)
	assert.Equal(t, newEvent.ID, result.ID)
	assert.Equal(t, newEvent.Device, result.Device)
	assert.Equal(t, newEvent.Readings[0].Name, result.Readings[0].Name)
	assert.Equal(t, newEvent.Readings[0].Value, result.Readings[0].Value)
}

func TestMarkAsPushedEventId(t *testing.T) {
	testID := "eventId"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPut {
			t.Errorf("expected http method is PUT, active http method is : %s", r.Method)
		}
		url := clients.ApiEventRoute + "/id/" + testID
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}
	}))

	defer ts.Close()
	eventClient = coredata.NewEventClient(localURL.New(ts.URL + clients.ApiEventRoute))

	ctx := Context{
		EventID:       testID,
		CorrelationID: "correlationId",
		EventClient:   eventClient,
		LoggingClient: lc,
	}

	err := ctx.MarkAsPushed()

	assert.Nil(t, err)
}

func TestSetRetryData(t *testing.T) {
	ctx := Context{}
	testData := "output data"
	ctx.SetRetryData([]byte(testData))
	assert.Equal(t, []byte(testData), ctx.RetryData)
}
