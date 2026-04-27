//
// Copyright (C) 2020-2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	contractsCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	internalCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

var d = sdkModels.DiscoveredDevice{
	Name: "device-sdk-test",
}

func newDeviceService() *deviceService {
	return &deviceService{
		serviceKey: "test-service",
		lc:         logger.NewMockClient(),
	}
}

func Test_processAsyncFilterAndAdd_bypassValidation(t *testing.T) {
	const testServiceKey = "test-service"

	pw := models.ProvisionWatcher{
		AdminState: models.Unlocked,
	}

	discovered := sdkModels.DiscoveredDevice{
		Name: "new-device",
		Protocols: map[string]models.ProtocolProperties{
			"http": {"host": "localhost"},
		},
	}

	addCalled := make(chan struct{}, 1)
	dcMock := &clientMocks.DeviceClient{}
	dcMock.On("DevicesByServiceName", mock.Anything, testServiceKey, 0, -1).Return(
		responses.MultiDevicesResponse{}, nil)
	dcMock.On("AddWithQueryParams",
		mock.Anything,
		mock.Anything,
		map[string]string{internalCommon.BypassValidationQueryParam: contractsCommon.ValueTrue},
	).Return(nil, nil).Run(func(mock.Arguments) { addCalled <- struct{}{} })

	pwcMock := &clientMocks.ProvisionWatcherClient{}
	pwcMock.On("ProvisionWatchersByServiceName", mock.Anything, testServiceKey, 0, -1).Return(
		responses.MultiProvisionWatchersResponse{}, nil)

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return dcMock
		},
		bootstrapContainer.ProvisionWatcherClientName: func(get di.Get) interface{} {
			return pwcMock
		},
	})

	err := cache.InitCache(testServiceKey, testServiceKey, dic)
	require.NoError(t, err)
	err = cache.ProvisionWatchers().Add(pw)
	require.NoError(t, err)

	deviceCh := make(chan []sdkModels.DiscoveredDevice, 1)
	ds := &deviceService{
		serviceKey: testServiceKey,
		lc:         logger.NewMockClient(),
		dic:        dic,
		deviceCh:   deviceCh,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go ds.processAsyncFilterAndAdd(ctx)

	deviceCh <- []sdkModels.DiscoveredDevice{discovered}

	select {
	case <-addCalled:
	case <-time.After(time.Second):
		t.Fatal("AddWithQueryParams was not called within timeout")
	}
}

func Test_checkAllowList(t *testing.T) {
	ds := newDeviceService()
	pw := models.ProvisionWatcher{
		Name: "test-watcher",
		Identifiers: map[string]string{
			"host": "localhost",
			"port": "3[0-9]{2}",
		},
	}

	onlyOneMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "localhost",
			"port": "301",
		},
	}
	oneOfProtocolsMatch := map[string]models.ProtocolProperties{
		"tcp": {
			"host": "localhost",
			"port": "80",
		},
		"http": {
			"host": "localhost",
			"port": "301",
		},
	}
	noIdentifiersMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "192.168.0.1",
			"port": "400",
		},
	}
	someIdentifiersMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "127.0.0.1",
			"port": "301",
		},
		"tcp": {
			"host": "localhost",
			"port": "80",
		},
	}
	noMatchInSingleIdentifier := map[string]models.ProtocolProperties{
		"http": {
			"port": "301",
		},
		"tcp": {
			"host": "localhost",
		},
	}

	tests := []struct {
		name      string
		protocols map[string]models.ProtocolProperties
		expected  bool
	}{
		{"pass - match found", onlyOneMatch, true},
		{"pass - one match found in multiple protocol", oneOfProtocolsMatch, true},
		{"fail - none of identifier match in one protocol", noIdentifiersMatch, false},
		{"fail - only partial of identifiers match in one protocol", someIdentifiersMatch, false},
		{"fail - all of the identifiers match but across different protocol", noMatchInSingleIdentifier, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			d.Protocols = testCase.protocols
			result := ds.checkAllowList(d, pw)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func Test_checkBlockList(t *testing.T) {
	ds := newDeviceService()
	pw := models.ProvisionWatcher{
		Name: "test-watcher",
		BlockingIdentifiers: map[string][]string{
			"port": []string{"399", "398", "397"},
		},
	}

	noBlockingIdentifierFound := map[string]models.ProtocolProperties{
		"http": {
			"host": "localhost",
		},
		"tcp": {
			"host": "127.0.0.1",
		},
	}
	noBlockingIdentifierMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "localhost",
			"port": "400",
		},
		"tcp": {
			"host": "localhost",
			"port": "80",
		},
	}
	blockingIdentifierMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "localhost",
			"port": "399",
		},
		"tcp": {
			"host": "localhost",
			"port": "80",
		},
	}

	tests := []struct {
		name      string
		protocols map[string]models.ProtocolProperties
		expected  bool
	}{
		{"pass - no blocking identifier found", noBlockingIdentifierFound, true},
		{"pass - blocking identifier found but not match", noBlockingIdentifierMatch, true},
		{"fail - blocking identifier match", blockingIdentifierMatch, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			d.Protocols = testCase.protocols
			result := ds.checkBlockList(d, pw)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
