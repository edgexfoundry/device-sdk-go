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

	"bitbucket.org/tonyespy/gxds"
	"bitbucket.org/tonyespy/gxds/cache"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/gorilla/mux"
)

const deviceCommandTest = "device-command-test"
const testCmd = "TestCmd"

// Test Command REST call when service is locked.
func TestCommandServiceLocked(t *testing.T) {

	// TODO: add dummy Config
	lc := logger.NewClient("command_test", false, "./command_test.log")

	// Setup dummy service with logger, and 'locked=true'
	s := &Service{lc: lc, locked: true}
	s.Config = &gxds.Config{ServiceName: deviceCommandTest}

	ch := &commandHandler{fn: commandFunc, s: s}
	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", deviceV1, "nil", "nil"), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": "nil", "cmd": "nil"})

	rr := httptest.NewRecorder()
	ch.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusLocked {
		t.Errorf("ServiceLocked: handler returned wrong status code: got %v want %v",
			status, http.StatusLocked)
	}

	body := strings.TrimSpace(rr.Body.String())
	expected := deviceCommandTest + " is locked; GET " + deviceV1 + "/nil/nil"

	if body != expected {
		t.Errorf("ServiceLocked: handler returned wrong body:\nexpected: %s\ngot:      %s", expected, body)
	}
}

// TestCommandNoDevice tests the command REST call when the given deviceId doesn't
// specify an existing device.
func TestCommandNoDevice(t *testing.T) {
	badDeviceId := "5abae51de23bf81c9ef0f390"

	// TODO: add dummy Config
	lc := logger.NewClient("command_test", false, "./command_test.log")

	// Setup dummy service with logger, and mocked devices cache
	// Empty cache will by default have no devices.
	s := &Service{lc: lc}
	s.cd = cache.NewDevices(s.Config, nil)

	ch := &commandHandler{fn: commandFunc, s: s}
	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", deviceV1, badDeviceId, testCmd), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": badDeviceId, "cmd": testCmd})

	rr := httptest.NewRecorder()
	ch.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("NoDevice: handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

	body := strings.TrimSpace(rr.Body.String())
	expected := "device: " + badDeviceId + " not found; GET " + deviceV1 + "/" + badDeviceId + "/" + testCmd

	if body != expected {
		t.Errorf("ServiceLocked: handler returned wrong body:\nexpected: %s\ngot:      %s", expected, body)
	}
}
