// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func Test_processDevices(t *testing.T) {
	tests := []struct {
		name                string
		path                string
		secretProvider      interfaces.SecretProvider
		expectedNumProfiles int
	}{
		{"valid load device from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "devices", "simple-device.yml"), nil, 1},
		{"valid load device from uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/devices/simple-device.yml", nil, 1},
		{"invalid load device empty path", "", nil, 0},
		{"invalid load device from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "devices", "bogus.yml"), nil, 0},
		{"invalid load device invalid uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/devices/bogus.yml", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addDeviceRequests []requests.AddDeviceRequest
			lc := logger.MockLogger{}
			dic, _ := NewMockDIC()
			cache.InitCache(TestDeviceService, TestDeviceService, dic)
			addDeviceRequests = processDevices(tt.path, TestDeviceService, tt.secretProvider, lc, addDeviceRequests)
			assert.Equal(t, tt.expectedNumProfiles, len(addDeviceRequests))
		})
	}
}
