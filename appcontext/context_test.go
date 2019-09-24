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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/assert"
)

func TestComplete(t *testing.T) {
	ctx := Context{}
	testData := "output data"
	ctx.Complete([]byte(testData))
	assert.Equal(t, []byte(testData), ctx.OutputData)
}

var eventClient coredata.EventClient
var params types.EndpointParams
var lc logger.LoggingClient

func init() {
	params = types.EndpointParams{
		ServiceKey:  clients.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: false,
		Url:         "http://test" + clients.ApiEventRoute,
		Interval:    1000,
	}
	eventClient = coredata.NewEventClient(params, startup.Endpoint{RegistryClient: nil})
	lc = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
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
	eventClient = coredata.NewEventClient(params, mockEventEndpoint{})
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
	params.Url = ts.URL + clients.ApiEventRoute
	eventClient = coredata.NewEventClient(params, mockEventEndpoint{})

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
		w.Write([]byte("newId"))
		if r.Method != http.MethodPost {
			t.Errorf("expected http method is POST, active http method is : %s", r.Method)
		}
		url := clients.ApiEventRoute
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}

	}))

	defer ts.Close()
	params.Url = ts.URL + clients.ApiEventRoute
	eventClient = coredata.NewEventClient(params, mockEventEndpoint{})
	ctx := Context{
		EventClient:   eventClient,
		LoggingClient: lc,
	}
	newEvent := &models.Event{
		ID:     "newId",
		Device: "device-name",
		Origin: 1567802840199266000,
		Readings: []models.Reading{
			models.Reading{
				Device:      "device-name",
				Name:        "device-resource",
				Value:       "value",
				BinaryValue: []uint8(nil),
			},
		},
	}
	result, err := ctx.PushToCoreData("device-name", "device-resource", "value")
	assert.NotNil(t, result)
	assert.Equal(t, newEvent.ID, result.ID)
	assert.Equal(t, newEvent.Device, result.Device)
	assert.Equal(t, newEvent.Readings[0].Name, result.Readings[0].Name)
	assert.Equal(t, newEvent.Readings[0].Value, result.Readings[0].Value)

	assert.Nil(t, err)

}

func TestMarkAsPushedEventId(t *testing.T) {
	testID := "eventId"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPut {
			t.Errorf("expected http method is PUT, active http method is : %s", r.Method)
		}
		//"http://test/api/v1/event/id/eventId"
		url := clients.ApiEventRoute + "/id/" + testID
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}
	}))

	defer ts.Close()
	params.Url = ts.URL + clients.ApiEventRoute
	eventClient = coredata.NewEventClient(params, mockEventEndpoint{})

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

type mockEventEndpoint struct {
}

func (e mockEventEndpoint) Monitor(params types.EndpointParams, ch chan string) {
	switch params.ServiceKey {
	case clients.CoreDataServiceKey:
		url := e.Fetch(params)
		ch <- url
		break
	default:
		ch <- ""
	}
}

func (e mockEventEndpoint) Fetch(params types.EndpointParams) string {
	return fmt.Sprintf("http://%s:%v%s", "localhost", 48080, params.Path)
}
