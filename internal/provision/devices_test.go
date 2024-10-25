// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"path"
	"testing"
)

func Test_processDevices(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		secretProvider     interfaces.SecretProvider
		expectedNumDevices int
	}{
		{"valid load device from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "devices", "simple-device.yml"), nil, 2},
		{"valid load devices from uri", "https://raw.githubusercontent.com/edgexfoundry/device-virtual-go/main/cmd/res/devices/devices.yaml", nil, 5},
		{"invalid load device empty path", "", nil, 0},
		{"invalid load device from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "devices", "bogus.yml"), nil, 0},
		{"invalid load device invalid uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/devices/bogus.yml", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := logger.MockLogger{}
			dic, _ := NewMockDIC()
			err := cache.InitCache(TestDeviceService, TestDeviceService, dic)
			require.NoError(t, err)
			addDeviceRequests := processDevices(tt.path, tt.path, TestDeviceService, tt.secretProvider, lc)
			assert.Equal(t, tt.expectedNumDevices, len(addDeviceRequests))
		})
	}
}

func Test_loadDevicesFromURI(t *testing.T) {
	tests := []struct {
		name                string
		path                string
		serviceName         string
		secretProvider      interfaces.SecretProvider
		expectedNumDevices  int
		expectedEdgexErrMsg string
	}{
		{"valid load from uri",
			"https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/internal/provision/uri-test-files/devices/index.json",
			"TestDevice",
			nil,
			2, ""},
		{"invalid load from uri",
			"https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/internal/provision/uri-test-files/devices/bogus.json",
			"TestDevice",
			nil,
			0, "failed to load Devices list from URI"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addDeviceReq []requests.AddDeviceRequest
			lc := logger.MockLogger{}
			dic, _ := NewMockDIC()
			edgexErr := cache.InitCache(TestDeviceService, TestDeviceService, dic)
			require.NoError(t, edgexErr)
			parsedURI, err := url.Parse(tt.path)
			require.NoError(t, err)
			addDeviceReq, edgexErr = loadDevicesFromURI(tt.path, parsedURI, tt.serviceName, tt.secretProvider, lc)
			assert.Equal(t, tt.expectedNumDevices, len(addDeviceReq))
			if edgexErr != nil {
				assert.Contains(t, edgexErr.Error(), tt.expectedEdgexErrMsg)
			}
		})
	}
}
