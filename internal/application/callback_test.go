// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
)

const (
	testPWBaseService = "testPWService"
	testPWInstance    = testPWBaseService + "_1"
	testPWOtherSvc    = "otherService"
	testPWCachedName  = "cachedWatcher"
	testPWNewName     = "newWatcher"
	testPWId          = "pw-id"
	testPWProfile     = "testPWProfile"
)

// initPWTestCache seeds the cache with testPWCachedName owned by testPWBaseService, under instance name testPWInstance.
func initPWTestCache(t *testing.T) *di.Container {
	t.Helper()

	dcMock := &clientMocks.DeviceClient{}
	dcMock.On("DevicesByServiceName", context.Background(), testPWInstance, 0, -1).
		Return(responses.MultiDevicesResponse{}, nil)
	dpcMock := &clientMocks.DeviceProfileClient{}
	dpcMock.On("DeviceProfileByName", context.Background(), testPWProfile).
		Return(responses.DeviceProfileResponse{Profile: dtos.DeviceProfile{
			DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{Name: testPWProfile},
		}}, nil)
	pwcMock := &clientMocks.ProvisionWatcherClient{}
	pwcMock.On("ProvisionWatchersByServiceName", context.Background(), testPWBaseService, 0, -1).
		Return(responses.MultiProvisionWatchersResponse{ProvisionWatchers: []dtos.ProvisionWatcher{
			{Id: testPWId, Name: testPWCachedName, ServiceName: testPWBaseService, AdminState: models.Unlocked,
				DiscoveredDevice: dtos.DiscoveredDevice{ProfileName: testPWProfile}},
		}}, nil)

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.DeviceServiceName: func(_ di.Get) any {
			return &models.DeviceService{Name: testPWInstance}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(_ di.Get) any { return logger.NewMockClient() },
		bootstrapContainer.DeviceClientName:           func(_ di.Get) any { return dcMock },
		bootstrapContainer.DeviceProfileClientName:    func(_ di.Get) any { return dpcMock },
		bootstrapContainer.ProvisionWatcherClientName: func(_ di.Get) any { return pwcMock },
	})
	require.NoError(t, cache.InitCache(testPWInstance, testPWBaseService, dic))
	return dic
}

func newUpdatePWRequest(name, serviceName string) requests.UpdateProvisionWatcherRequest {
	pw := models.ProvisionWatcher{Id: testPWId, Name: name, ServiceName: serviceName, AdminState: models.Locked,
		DiscoveredDevice: models.DiscoveredDevice{ProfileName: testPWProfile}}
	return requests.NewUpdateProvisionWatcherRequest(dtos.FromProvisionWatcherModelToUpdateDTO(pw))
}

func TestUpdateProvisionWatcher(t *testing.T) {
	tests := []struct {
		name          string
		pwName        string
		pwServiceName string
		expectCached  bool
		expectError   bool
	}{
		{
			name:          "update cached watcher owned by base name",
			pwName:        testPWCachedName,
			pwServiceName: testPWBaseService,
			expectCached:  true,
		},
		{
			name:          "add uncached watcher owned by base name",
			pwName:        testPWNewName,
			pwServiceName: testPWBaseService,
			expectCached:  true,
		},
		{
			name:          "remove cached watcher moved to instance name",
			pwName:        testPWCachedName,
			pwServiceName: testPWInstance,
		},
		{
			name:          "reject uncached watcher owned by instance name",
			pwName:        testPWNewName,
			pwServiceName: testPWInstance,
			expectError:   true,
		},
		{
			name:          "remove cached watcher moved to other service",
			pwName:        testPWCachedName,
			pwServiceName: testPWOtherSvc,
		},
		{
			name:          "reject uncached watcher owned by other service",
			pwName:        testPWNewName,
			pwServiceName: testPWOtherSvc,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dic := initPWTestCache(t)

			err := UpdateProvisionWatcher(newUpdatePWRequest(tt.pwName, tt.pwServiceName), testPWBaseService, dic)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, errors.KindEntityDoesNotExist, errors.Kind(err))
				assert.Contains(t, err.Error(), tt.pwName)
			} else {
				require.NoError(t, err)
			}

			pw, ok := cache.ProvisionWatchers().ForName(tt.pwName)
			require.Equal(t, tt.expectCached, ok)
			if tt.expectCached {
				assert.Equal(t, testPWId, pw.Id)
				assert.Equal(t, tt.pwServiceName, pw.ServiceName)
				assert.Equal(t, models.AdminState(models.Locked), pw.AdminState)
			}
		})
	}
}
