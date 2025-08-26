// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
// # Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
	"net/url"
	"path"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_processDevices(t *testing.T) {
	tests := []struct {
		name                  string
		path                  string
		update                bool
		secretProvider        interfaces.SecretProvider
		expectedNumDevices    int
		expectedUpdateDevices int
	}{
		{"valid load device from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "devices", "simple-device.yml"), false, nil, 2, 0},
		{"valid load devices from uri", "https://raw.githubusercontent.com/edgexfoundry/device-virtual-go/main/cmd/res/devices/devices.yaml", false, nil, 6, 0},
		{"valid overwrite device from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "devices", "simple-device.yml"), true, nil, 1, 1},
		{"invalid load device empty path", "", false, nil, 0, 0},
		{"invalid load device from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "devices", "bogus.yml"), false, nil, 0, 0},
		{"invalid load device invalid uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/devices/bogus.yml", false, nil, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := logger.MockLogger{}
			dic, _ := NewMockDIC()
			err := cache.InitCache(TestDeviceService, TestDeviceService, dic)
			if tt.update {
				cache.Devices().Add(models.Device{
					Name: "Simple-Device01",
				})
			}

			require.NoError(t, err)
			addDeviceRequests, updateDeviceRequests := processDevices(tt.path, tt.path, TestDeviceService, tt.update, tt.secretProvider, lc)
			assert.Equal(t, tt.expectedNumDevices, len(addDeviceRequests))
			if tt.update {
				assert.Equal(t, tt.expectedUpdateDevices, len(updateDeviceRequests))
			}
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
			var updateDeviceReq []requests.UpdateDeviceRequest
			lc := logger.MockLogger{}
			dic, _ := NewMockDIC()
			edgexErr := cache.InitCache(TestDeviceService, TestDeviceService, dic)
			require.NoError(t, edgexErr)
			parsedURI, err := url.Parse(tt.path)
			require.NoError(t, err)
			addDeviceReq, updateDeviceReq, edgexErr = loadDevicesFromURI(tt.path, parsedURI, tt.serviceName, false, tt.secretProvider, lc)
			assert.Equal(t, tt.expectedNumDevices, len(addDeviceReq))
			assert.Equal(t, 0, len(updateDeviceReq))
			if edgexErr != nil {
				assert.Contains(t, edgexErr.Error(), tt.expectedEdgexErrMsg)
			}
		})
	}
}
