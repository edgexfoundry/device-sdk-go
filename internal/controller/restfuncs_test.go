// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"bytes"
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/mock"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	"github.com/gorilla/mux"
)

const (
	badDeviceId       = "e0fe7ac0-f7f3-4b76-b1b0-4b9bf4788d3e"
	deviceCommandTest = "device-command-test"
	testCmd           = "TestCmd"
)

// Test callback REST calls
func TestCallback(t *testing.T) {
	var tests = []struct {
		name   string
		method string
		body   string
		code   int
	}{
		{"Empty body", http.MethodPut, "", http.StatusBadRequest},
		{"Empty json", http.MethodPut, "{}", http.StatusBadRequest},
		{"Invalid type", http.MethodPut, `{"id":"1ef435eb-5060-49b0-8d55-8d4e43239800","type":"INVALID"}`, http.StatusBadRequest},
		{"Invalid method", http.MethodPost, `{"id":"1ef435eb-5060-49b0-8d55-8d4e43239800","type":"DEVICE"}`, http.StatusBadRequest},
		{"Invalid id", http.MethodPut, `{"id":"1ef435eb-5060-49b0-8d55-8d4e43239800","type":"DEVICE"}`, http.StatusBadRequest},
	}

	lc := logger.NewClient("update_test", false, "./device-simple.log", "DEBUG")
	common.LoggingClient = lc
	common.DeviceClient = &mock.DeviceClientMock{}
	r := InitRestRoutes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonStr = []byte(tt.body)
			req := httptest.NewRequest(tt.method, common.APICallbackRoute, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			fmt.Printf("rr.code = %v\n", rr.Code)
			if status := rr.Code; status != tt.code {
				t.Errorf("CallbackHandler: handler returned wrong status code: got %v want %v",
					status, http.StatusLocked)
			}
		})
	}
}

// Test Command REST call when service is locked.
func TestCommandServiceLocked(t *testing.T) {
	lc := logger.NewClient("command_test", false, "./command_test.log", "DEBUG")
	common.LoggingClient = lc
	common.ServiceLocked = true
	common.ServiceName = deviceCommandTest
	r := InitRestRoutes()

	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", clients.ApiDeviceRoute, "nil", "nil"), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": "nil", "cmd": "nil"})

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusLocked {
		t.Errorf("ServiceLocked: handler returned wrong status code: got %v want %v",
			status, http.StatusLocked)
	}

	body := strings.TrimSpace(rr.Body.String())
	expected := deviceCommandTest + " is locked; GET " + clients.ApiDeviceRoute + "/nil/nil"

	if body != expected {
		t.Errorf("ServiceLocked: handler returned wrong body:\nexpected: %s\ngot:      %s", expected, body)
	}
}

// TestCommandNoDevice tests the command REST call when the given deviceId doesn't
// specify an existing device.
func TestCommandNoDevice(t *testing.T) {
	lc := logger.NewClient("command_test", false, "./command_test.log", "DEBUG")
	common.LoggingClient = lc
	common.ServiceLocked = false
	common.ValueDescriptorClient = &mock.ValueDescriptorMock{}
	r := InitRestRoutes()

	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", clients.ApiDeviceRoute, badDeviceId, testCmd), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": badDeviceId, "cmd": testCmd})

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("NoDevice: handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

	body := strings.TrimSpace(rr.Body.String())
	expected := "Device: " + badDeviceId + " not found; GET " + clients.ApiDeviceRoute + "/" + badDeviceId + "/" + testCmd

	if body != expected {
		t.Errorf("No Device: handler returned wrong body:\nexpected: %s\ngot:      %s", expected, body)
	}
}

// TestCommandNoDevice tests the command REST call when the device specified
// by deviceId is locked.
//func TestCommandDeviceLocked(t *testing.T) {
//	lc := logger.NewClient("command_test", false, "./command_test.log")
//	common.LoggingClient = lc
//r := InitRestRoutes()
// Empty cache will by default have no devices.
//dc = &deviceCache{}

/* TODO: adding a device to the devices cache requires a live metadata instance. We need
 * create interfaces for all of the caches, so that they can be mocked in unit tests.

millis := time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)

// TODO: does HTTPMethod need to be specified?
addr = models.Addressable{
	BaseObject: models.BaseObject{
		Origin: millis,
	},
	name:       s.Config.ServiceName,
	HTTPMethod: "POST",
	Protocol:   "HTTP",
	Address:    "localhost",
	Port:       "2112",
	Path:       "/api/v1/callback",
}
addr.Origin = millis

// Create a locked Device
d := &models.Device{name: "DummyDevice", AdminState: "LOCKED", OperatingState: "ENABLED"}
d.Id = testDeviceId

s.cd.Add(d)

req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", v1Device, testDeviceId, testCmd), nil)
req = mux.SetURLVars(req, map[string]string{"deviceId": testDeviceId, "cmd": testCmd})

rr := httptest.NewRecorder()
svc.r.ServeHTTP(rr, req)

if status := rr.Code; status != http.StatusLocked {
	t.Errorf("NoDevice: handler returned wrong status code: got %v want %v",
		status, http.StatusLocked)
}

body := strings.TrimSpace(rr.Body.String())
expected := "device: " + testDeviceId + " locked; GET " + v1Device + "/" + testDeviceId + "/" + testCmd

if body != expected {
	t.Errorf("DeviceLocked: handler returned wrong body:\nexpected: %s\ngot:      %s", expected, body)
}
*/
//}
