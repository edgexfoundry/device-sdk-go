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

func Test_processProvisionWatcherFile(t *testing.T) {

	tests := []struct {
		name                         string
		path                         string
		secretProvider               interfaces.SecretProvider
		expectedNumProvisionWatchers int
	}{
		{"valid load provision watcher from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "provisionwatchers", "Simple-Provision-Watcher.yml"), nil, 1},
		{"valid load provision watcher invalid uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/provisionwatchers/Simple-Provision-Watcher.yml", nil, 1},
		{"invalid load provision watcher empty path", "", nil, 0},
		{"invalid load provision watcher from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "provisionwatchers", "bogus.yml"), nil, 0},
		{"invalid load provision watcher invalid uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/provisionwatchers/bogus.yml", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addProvisionWatchersReq []requests.AddProvisionWatcherRequest
			lc := logger.MockLogger{}
			dic, _ := NewMockDIC()
			cache.InitCache(TestDeviceService, TestDeviceService, dic)
			addProvisionWatchersReq = processProvisonWatcherFile(tt.path, tt.secretProvider, lc, addProvisionWatchersReq)
			assert.Equal(t, tt.expectedNumProvisionWatchers, len(addProvisionWatchersReq))
		})
	}
}
