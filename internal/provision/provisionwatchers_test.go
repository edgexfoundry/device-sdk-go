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
	"github.com/stretchr/testify/require"
	"net/url"
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
		{"valid load provision watcher from valid uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/provisionwatchers/Simple-Provision-Watcher.yml", nil, 1},
		{"invalid load provision watcher empty path", "", nil, 0},
		{"invalid load provision watcher from file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "provisionwatchers", "bogus.yml"), nil, 0},
		{"invalid load provision watcher from invalid uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/provisionwatchers/bogus.yml", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addProvisionWatchersReq []requests.AddProvisionWatcherRequest
			lc := logger.MockLogger{}
			dic, _ := NewMockDIC()
			err := cache.InitCache(TestDeviceService, TestDeviceService, dic)
			require.NoError(t, err)
			addProvisionWatchersReq = processProvisionWatcherFile(tt.path, tt.path, tt.secretProvider, lc)
			assert.Equal(t, tt.expectedNumProvisionWatchers, len(addProvisionWatchersReq))
		})
	}
}

func Test_loadProvisionWatchersFromURI(t *testing.T) {
	tests := []struct {
		name                         string
		path                         string
		secretProvider               interfaces.SecretProvider
		expectedNumProvisionWatchers int
		expectedEdgexErrMsg          string
	}{
		{"valid load from uri",
			"https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/internal/provision/uri-test-files/provisionwatchers/index.json",
			nil,
			2, ""},
		{"invalid load from uri",
			"https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/internal/provision/uri-test-files/provisionwatchers/bogus.json",
			nil,
			0, "failed to load Provision Watchers list from URI"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addProvisionWatchersReq []requests.AddProvisionWatcherRequest
			lc := logger.MockLogger{}
			dic, _ := NewMockDIC()
			edgexErr := cache.InitCache(TestDeviceService, TestDeviceService, dic)
			require.NoError(t, edgexErr)
			parsedURI, err := url.Parse(tt.path)
			require.NoError(t, err)
			addProvisionWatchersReq, edgexErr = loadProvisionWatchersFromURI(tt.path, parsedURI, tt.secretProvider, lc)
			assert.Equal(t, tt.expectedNumProvisionWatchers, len(addProvisionWatchersReq))
			if edgexErr != nil {
				assert.Contains(t, edgexErr.Error(), tt.expectedEdgexErrMsg)
			}
		})
	}
}
