//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/stretchr/testify/assert"
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

func Test_deviceCache_ForName(t *testing.T) {
	newDeviceCache([]models.Device{testDevice})

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
	newDeviceCache([]models.Device{testDevice})

	res := dc.All()
	require.Equal(t, len(res), len(dc.deviceMap))
}

func Test_deviceCache_Add(t *testing.T) {
	newDeviceCache([]models.Device{testDevice})

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
	newDeviceCache([]models.Device{testDevice})

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
	newDeviceCache([]models.Device{testDevice})

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
