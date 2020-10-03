// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/mock"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

const (
	badDeviceId = "e0fe7ac0-f7f3-4b76-b1b0-4b9bf4788d3e"
	testCmd     = "TestCmd"
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
		{"Invalid method", http.MethodGet, `{"id":"1ef435eb-5060-49b0-8d55-8d4e43239800","type":"DEVICE"}`, http.StatusBadRequest},
		{"Invalid id", http.MethodPut, `{"id":"","type":"DEVICE"}`, http.StatusBadRequest},
	}

	lc := logger.NewClientStdOut("device-sdk-test", false, "DEBUG")
	deviceClient := &mock.DeviceClientMock{}
	ds := contract.DeviceService{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		container.MetadataDeviceClientName: func(get di.Get) interface{} {
			return deviceClient
		},
		container.DeviceServiceName: func(get di.Get) interface{} {
			return ds
		},
	})

	r := mux.NewRouter()
	controller := NewRestController(r, dic)
	controller.InitRestRoutes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonStr = []byte(tt.body)
			req := httptest.NewRequest(tt.method, common.APICallbackRoute, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			controller.router.ServeHTTP(rr, req)
			fmt.Printf("rr.code = %v\n", rr.Code)
			if status := rr.Code; status != tt.code {
				t.Errorf("CallbackHandler: handler returned wrong status code: got %v want %v",
					status, tt.code)
			}
		})
	}
}

// Test Command REST call when service is locked.
func TestCommandServiceLocked(t *testing.T) {
	lc := logger.NewClientStdOut("device-sdk-test", false, "DEBUG")
	ds := contract.DeviceService{
		AdminState: contract.Locked,
	}
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		container.DeviceServiceName: func(get di.Get) interface{} {
			return ds
		},
	})
	r := mux.NewRouter()
	controller := NewRestController(r, dic)
	controller.InitRestRoutes()

	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", clients.ApiDeviceRoute, "nil", "nil"), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": "nil", "cmd": "nil"})

	rr := httptest.NewRecorder()
	controller.router.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusLocked {
		t.Errorf("ServiceLocked: handler returned wrong status code: got %v want %v",
			status, http.StatusLocked)
	}

	body := strings.TrimSpace(rr.Body.String())
	expected := "Service is locked; GET " + clients.ApiDeviceRoute + "/nil/nil"

	if body != expected {
		t.Errorf("ServiceLocked: handler returned wrong body:\nexpected: %s\ngot:      %s", expected, body)
	}
}

// TestCommandNoDevice tests the command REST call when the given deviceId doesn't
// specify an existing device.
func TestCommandNoDevice(t *testing.T) {
	lc := logger.NewClientStdOut("device-sdk-test", false, "DEBUG")
	ds := contract.DeviceService{}
	dc := &mock.DeviceClientMock{}
	vdc := &mock.ValueDescriptorMock{}
	pwc := &mock.ProvisionWatcherClientMock{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		container.DeviceServiceName: func(get di.Get) interface{} {
			return ds
		},
		container.MetadataDeviceClientName: func(get di.Get) interface{} {
			return dc
		},
		container.CoredataValueDescriptorClientName: func(get di.Get) interface{} {
			return vdc
		},
		container.MetadataProvisionWatcherClientName: func(get di.Get) interface{} {
			return pwc
		},
	})
	cache.InitCache("device-sdk-test", lc, vdc, dc, pwc)

	r := mux.NewRouter()
	controller := NewRestController(r, dic)
	controller.InitRestRoutes()

	req := httptest.NewRequest("GET", fmt.Sprintf("%s/%s/%s", clients.ApiDeviceRoute, badDeviceId, testCmd), nil)
	req = mux.SetURLVars(req, map[string]string{"deviceId": badDeviceId, "cmd": testCmd})

	rr := httptest.NewRecorder()
	controller.router.ServeHTTP(rr, req)

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
