//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
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
			Name:      TestDeviceCommand,
			ReadWrite: v2.ReadWrite_R,
			ResourceOperations: []models.ResourceOperation{
				models.ResourceOperation{DeviceResource: TestDeviceResource},
			},
		},
	},
}

var newProfile = models.DeviceProfile{
	Name: "newProfile",
	DeviceResources: []models.DeviceResource{
		models.DeviceResource{Name: "newResource"},
	},
	DeviceCommands: []models.DeviceCommand{
		models.DeviceCommand{
			Name:      "newCommand",
			ReadWrite: v2.ReadWrite_R,
			ResourceOperations: []models.ResourceOperation{
				models.ResourceOperation{DeviceResource: "newResource"},
			},
		},
	},
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

func Test_profileCache_DeviceCommand(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name          string
		profileName   string
		commandName   string
		deviceCommand models.DeviceCommand
		expected      bool
	}{
		{"Invalid - nonexistent Profile name", "nil", TestDeviceCommand, models.DeviceCommand{}, false},
		{"Invalid - nonexistent Command name", TestProfile, "nil", models.DeviceCommand{}, false},
		{"Valid", TestProfile, TestDeviceCommand, testProfile.DeviceCommands[0], true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := pc.DeviceCommand(tt.profileName, tt.commandName)
			assert.Equal(t, res, tt.deviceCommand, "DeviceResource returns wrong deviceResource")
			assert.Equal(t, ok, tt.expected, "DeviceResource returns opposite result")
		})
	}
}

func Test_profileCache_ResourceOperation(t *testing.T) {
	newProfileCache([]models.DeviceProfile{testProfile})

	tests := []struct {
		name          string
		profile       string
		resource      string
		res           models.ResourceOperation
		expectedError bool
	}{
		{"Invalid - nonexistent Profile name", "nil", TestDeviceResource, models.ResourceOperation{}, true},
		{"Invalid - nonexistent DeviceResource name", TestProfile, "nil", models.ResourceOperation{}, true},
		{"Valid", TestProfile, TestDeviceResource, testProfile.DeviceCommands[0].ResourceOperations[0], false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro, err := pc.ResourceOperation(tt.profile, tt.resource)
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
