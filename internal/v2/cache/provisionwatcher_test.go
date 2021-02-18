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

var testProvisionWatcher = models.ProvisionWatcher{
	Name:       TestProvisionWatcher,
	AdminState: models.Unlocked,
}

var newProvisionWatcher = models.ProvisionWatcher{
	Name:       "newProvisionWatcher",
	AdminState: models.Unlocked,
}

func Test_provisionWatcherCache_ForName(t *testing.T) {
	newProvisionWatcherCache([]models.ProvisionWatcher{testProvisionWatcher})

	tests := []struct {
		name             string
		pwName           string
		provisionWatcher models.ProvisionWatcher
		expected         bool
	}{
		{"Invalid - empty name", "", models.ProvisionWatcher{}, false},
		{"Invalid - nonexistent ProvisionWatcher name", "nil", models.ProvisionWatcher{}, false},
		{"Valid", TestProvisionWatcher, testProvisionWatcher, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := pwc.ForName(tt.pwName)
			assert.Equal(t, res, tt.provisionWatcher, "ForName returns wrong ProvisionWatcher")
			assert.Equal(t, ok, tt.expected, "ForName returns opposite result")
		})
	}
}

func Test_provisionWatcherCache_All(t *testing.T) {
	newProvisionWatcherCache([]models.ProvisionWatcher{testProvisionWatcher})

	res := pwc.All()
	require.Equal(t, len(res), len(pwc.pwMap))
}

func Test_provisionWatcherCache_Add(t *testing.T) {
	newProvisionWatcherCache([]models.ProvisionWatcher{testProvisionWatcher})

	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid", false},
		{"Invalid - duplicate ProvisionWatcher name", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pwc.Add(newProvisionWatcher)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_provisionWatcherCache_RemoveByName(t *testing.T) {
	newProvisionWatcherCache([]models.ProvisionWatcher{testProvisionWatcher})

	tests := []struct {
		name          string
		expectedError bool
	}{
		{"Valid", false},
		{"Invalid - nonexistent ProvisionWatcher name", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pwc.RemoveByName(TestProvisionWatcher)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_provisionWatcherCache_UpdateAdminState(t *testing.T) {
	newProvisionWatcherCache([]models.ProvisionWatcher{testProvisionWatcher})

	tests := []struct {
		name          string
		pwName        string
		state         models.AdminState
		expectedError bool
	}{
		{"Invalid - nonexistent ProvisionWatcher name", "nil", models.Locked, true},
		{"Invalid - invalid AdminState", TestProvisionWatcher, "INVALID", true},
		{"Valid", TestProvisionWatcher, models.Locked, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pwc.UpdateAdminState(tt.pwName, tt.state)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
