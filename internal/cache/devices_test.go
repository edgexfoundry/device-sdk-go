//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"testing"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testDevice = models.Device{
	Name:           TestDevice,
	AdminState:     models.Unlocked,
	OperatingState: models.Up,
}

var newDevice = models.Device{
	Name:           "newDevice",
	AdminState:     models.Unlocked,
	OperatingState: models.Unlocked,
}

func mockDic() *di.Container {
	mockMetricsManager := &mocks.MetricsManager{}
	mockMetricsManager.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMetricsManager.On("Unregister", mock.Anything)
	return di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return mockMetricsManager
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "INFO",
				},
				Service: bootstrapConfig.ServiceInfo{
					EnableNameFieldEscape: true,
				},
			}
		},
	})
}

func Test_deviceCache_ForName(t *testing.T) {
	dic := mockDic()
	newDeviceCache([]models.Device{testDevice}, dic)

	tests := []struct {
		name       string
		deviceName string
		device     models.Device
		expected   bool
	}{
		{"Invalid - empty name", "", models.Device{}, false},
		{"Invalid - nonexistent Device name", "nil", models.Device{}, false},
		{"Valid", TestDevice, testDevice, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := dc.ForName(tt.deviceName)
			assert.Equal(t, res, tt.device, "ForName returns wrong Device")
			assert.Equal(t, ok, tt.expected, "ForName returns opposite result")
		})
	}
}

func Test_deviceCache_All(t *testing.T) {
	dic := mockDic()
	newDeviceCache([]models.Device{testDevice}, dic)

	res := dc.All()
	require.Equal(t, len(res), len(dc.deviceMap))
}

func Test_deviceCache_Add(t *testing.T) {
	dic := mockDic()
	newDeviceCache([]models.Device{testDevice}, dic)

	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid", false},
		{"Invalid - duplicate Device name", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dc.Add(newDevice)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_deviceCache_RemoveByName(t *testing.T) {
	dic := mockDic()
	newDeviceCache([]models.Device{testDevice}, dic)

	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid", false},
		{"Invalid - nonexistent Device name", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dc.RemoveByName(TestDevice)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_deviceCache_UpdateAdminState(t *testing.T) {
	dic := mockDic()
	newDeviceCache([]models.Device{testDevice}, dic)

	tests := []struct {
		name          string
		deviceName    string
		state         models.AdminState
		expectedError bool
	}{
		{"Invalid - nonexistent Device name", "nil", models.Locked, true},
		{"Invalid - invalid AdminState", TestDevice, "INVALID", true},
		{"Valid", TestDevice, models.Locked, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dc.UpdateAdminState(tt.deviceName, tt.state)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_deviceCache_SetLastConnectedByName(t *testing.T) {
	dic := mockDic()
	newDeviceCache([]models.Device{testDevice}, dic)

	// Make currentTimestamp return currentTimeInstant constant in unit test
	currentTimeInstant := time.Now().UnixNano()
	currentTimestamp = func() int64 {
		return currentTimeInstant
	}

	dc.SetLastConnectedByName(TestDevice)
	lastConnectedTime := dc.GetLastConnectedByName(TestDevice)
	require.Equal(t, currentTimeInstant, lastConnectedTime)
}

func Test_deviceCache_GetLastConnectedByName(t *testing.T) {
	dic := mockDic()
	newDeviceCache([]models.Device{testDevice}, dic)

	lastConnectedTime := dc.GetLastConnectedByName(TestDevice)
	require.Equal(t, int64(0), lastConnectedTime)
}
