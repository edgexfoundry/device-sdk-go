//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/stretchr/testify/assert"
)

var testProfile = models.DeviceProfile{
	Name: TestProfile,
	DeviceResources: []models.DeviceResource{
		models.DeviceResource{Name: TestDeviceResource},
	},
	DeviceCommands: []models.DeviceCommand{
		models.DeviceCommand{
			Name: TestDeviceCommand,
			Get: []models.ResourceOperation{
				models.ResourceOperation{DeviceResource: TestDeviceResource},
			},
		},
	},
	CoreCommands: nil,
}

var newProfile = models.DeviceProfile{
	Name: "newProfile",
	DeviceResources: []models.DeviceResource{
		models.DeviceResource{Name: "newResource"},
	},
	DeviceCommands: []models.DeviceCommand{
		models.DeviceCommand{
			Name: "newCommand",
			Get: []models.ResourceOperation{
				models.ResourceOperation{DeviceResource: "newResource"},
			},
		},
	},
	CoreCommands: nil,
}

func Test_profileCache_ForName(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name        string
		profileName string
		profile     models.DeviceProfile
		expected    bool
	}{
		{"Invalid - empty name", "", models.DeviceProfile{}, false},
		{"Invalid - nonexistent Profile name", "nil", models.DeviceProfile{}, false},
		{"Valid", TestProfile, testProfile, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := pc.ForName(tt.profileName)
			assert.Equal(t, res, tt.profile, "ForName returns wrong Profile")
			assert.Equal(t, ok, tt.expected, "ForName returns opposite result")
		})
	}
}

func Test_profileCache_All(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	res := pc.All()
	assert.Equal(t, len(res), len(pc.deviceProfileMap))
}

func Test_profileCache_Add(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid", false},
		{"Invalid - duplicate Profile", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pc.Add(newProfile)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_profileCache_RemoveByName(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid", false},
		{"Invalid - nonexistent Profile name", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pc.RemoveByName(TestProfile)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_profileCache_DeviceResource(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name           string
		profileName    string
		resourceName   string
		deviceResource models.DeviceResource
		expected       bool
	}{
		{"Invalid - nonexistent Profile name", "nil", TestDeviceResource, models.DeviceResource{}, false},
		{"Invalid - nonexistent Resource name", TestProfile, "nil", models.DeviceResource{}, false},
		{"Valid", TestProfile, TestDeviceResource, testProfile.DeviceResources[0], true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := pc.DeviceResource(tt.profileName, tt.resourceName)
			assert.Equal(t, res, tt.deviceResource, "DeviceResource returns wrong deviceResource")
			assert.Equal(t, ok, tt.expected, "DeviceResource returns opposite result")
		})
	}
}

func Test_profileCache_CommandExists(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name     string
		profile  string
		cmd      string
		method   string
		expected bool
	}{
		{"Invalid - nonexistent Profile name", "nil", TestDeviceCommand, "GET", false},
		{"Invalid - nonexistent Command name", TestProfile, "nil", "GET", false},
		{"Invalid - invalid method", TestProfile, TestDeviceCommand, "INVALID", false},
		{"Invalid - nonexistent method", TestProfile, TestDeviceCommand, "SET", false},
		{"Valid", TestProfile, TestDeviceCommand, "gEt", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := pc.CommandExists(tt.profile, tt.cmd, tt.method)
			assert.Equal(t, ok, tt.expected)
		})
	}
}

func Test_profileCache_ResourceOperations(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name          string
		profile       string
		cmd           string
		method        string
		res           []models.ResourceOperation
		expectedError bool
	}{
		{"Invalid - nonexistent Profile name", "nil", TestDeviceCommand, "GET", nil, true},
		{"Invalid - nonexistent Command name", TestProfile, "nil", "GET", nil, true},
		{"Invalid - invalid method", TestProfile, TestDeviceCommand, "INVALID", nil, true},
		{"Valid", TestProfile, TestDeviceCommand, "GET", testProfile.DeviceCommands[0].Get, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := pc.ResourceOperations(tt.profile, tt.cmd, tt.method)
			if tt.expectedError {
				assert.NotNil(t, err)
				assert.Nil(t, res)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, res, tt.res)
			}
		})
	}
}

func Test_profileCache_ResourceOperation(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name          string
		profile       string
		resource      string
		method        string
		res           models.ResourceOperation
		expectedError bool
	}{
		{"Invalid - nonexistent Profile name", "nil", TestDeviceResource, "GET", models.ResourceOperation{}, true},
		{"Invalid - nonexistent DeviceResource name", TestProfile, "nil", "GET", models.ResourceOperation{}, true},
		{"Invalid - invalid method", TestProfile, TestDeviceResource, "INVALID", models.ResourceOperation{}, true},
		{"Valid", TestProfile, TestDeviceResource, "Get", testProfile.DeviceCommands[0].Get[0], false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro, err := pc.ResourceOperation(tt.profile, tt.resource, tt.method)
			if tt.expectedError {
				assert.NotNil(t, err)
				assert.Equal(t, ro, tt.res)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, ro, tt.res)
			}
		})
	}
}
