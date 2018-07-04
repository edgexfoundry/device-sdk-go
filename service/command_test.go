// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// TODO: imports commented out till cache objects become interfaces
	//	"time"

	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	//	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/gorilla/mux"
	//	"gopkg.in/mgo.v2/bson"
)

const deviceCommandTest = "device-command-test"
const testCmd = "TestCmd"

// Test Command REST call when service is locked.
func TestCommandServiceLocked(t *testing.T) {
	ch := &commandHandler{fn: commandFunc}
	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", v1Device, "nil", "nil"), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": "nil", "cmd": "nil"})

	rr := httptest.NewRecorder()
	ch.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusLocked {
		t.Errorf("ServiceLocked: handler returned wrong status code: got %v want %v",
			status, http.StatusLocked)
	}

	body := strings.TrimSpace(rr.Body.String())
	expected := deviceCommandTest + " is locked; GET " + v1Device + "/nil/nil"

	if body != expected {
		t.Errorf("ServiceLocked: handler returned wrong body:\nexpected: %s\ngot:      %s", expected, body)
	}
}

// TestCommandNoDevice tests the command REST call when the given deviceId doesn't
// specify an existing device.
func TestCommandNoDevice(t *testing.T) {
	badDeviceId := "5abae51de23bf81c9ef0f390"

	lc := logger.NewClient("command_test", false, "./command_test.log")

	// Setup dummy service with logger, and mocked devices cache
	// Empty cache will by default have no devices.
	s := &Service{lc: lc}
	newDeviceCache("fakeID")

	ch := &commandHandler{fn: commandFunc, s: s}
	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", v1Device, badDeviceId, testCmd), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": badDeviceId, "cmd": testCmd})

	rr := httptest.NewRecorder()
	ch.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("NoDevice: handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

	body := strings.TrimSpace(rr.Body.String())
	expected := "device: " + badDeviceId + " not found; GET " + v1Device + "/" + badDeviceId + "/" + testCmd

	if body != expected {
		t.Errorf("ServiceLocked: handler returned wrong body:\nexpected: %s\ngot:      %s", expected, body)
	}
}

// TestCommandNoDevice tests the command REST call when the device specified
// by deviceId is locked.
func TestCommandDeviceLocked(t *testing.T) {
	// Empty cache will by default have no devices.
	newDeviceCache("fakeID")

	/* TODO: adding a device to the devices cache requires a live metadata instance. We need
	 * create interfaces for all of the caches, so that they can be mocked in unit tests.

	millis := time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)

	// TODO: does HTTPMethod need to be specified?
	addr = models.Addressable{
		BaseObject: models.BaseObject{
			Origin: millis,
		},
		Name:       s.Config.ServiceName,
		HTTPMethod: "POST",
		Protocol:   "HTTP",
		Address:    "localhost",
		Port:       "2112",
		Path:       "/api/v1/callback",
	}
	addr.Origin = millis

	// Create a locked Device
	d := &models.Device{Name: "DummyDevice", AdminState: "LOCKED", OperatingState: "ENABLED"}
	d.Id = bson.ObjectIdHex(testDeviceId)

	s.cd.Add(d)

	ch := &commandHandler{fn: commandFunc, s: s}
	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", v1Device, testDeviceId, testCmd), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": testDeviceId, "cmd": testCmd})

	rr := httptest.NewRecorder()
	ch.ServeHTTP(rr, req)

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
}
