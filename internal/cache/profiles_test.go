// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/mock"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var dps []contract.DeviceProfile

func init() {
	common.DeviceProfileClient = &mock.DeviceProfileClientMock{}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	dps, _ = common.DeviceProfileClient.DeviceProfiles(ctx)
}

func TestNewProfileCache(t *testing.T) {
	dpc := newProfileCache([]contract.DeviceProfile{})
	if _, ok := dpc.(ProfileCache); !ok {
		t.Error("the newProfileCache function supposed to return a value which holds the ProfileCache type")
	}
}

func TestProfileCache_ForName(t *testing.T) {
	dpc := newProfileCache(dps)
	if dp0, ok := dpc.ForName(mock.DeviceProfileRandomBoolGenerator.Name); !ok {
		t.Error("supposed to find a matching device profile in cache by a valid device name")
	} else {
		assert.Equal(t, mock.DeviceProfileRandomBoolGenerator, dp0)
	}
}

func TestProfileCache_ForId(t *testing.T) {
	dpc := newProfileCache(dps)
	if dp0, ok := dpc.ForId(mock.DeviceProfileRandomBoolGenerator.Id); !ok {
		t.Error("supposed to find a matching device profile in cache by a valid device profile id")
	} else {
		assert.Equal(t, mock.DeviceProfileRandomBoolGenerator, dp0)
	}
}

func TestProfileCache_All(t *testing.T) {
	dpc := newProfileCache(dps)
	dpsFromCache := dpc.All()

	for _, dpFromCache := range dpsFromCache {
		for _, dp := range dps {
			if dpFromCache.Id == dp.Id {
				assert.Equal(t, dp, dpFromCache)
				continue
			}
		}
	}
}

func TestProfileCache_Add(t *testing.T) {
	dpc := newProfileCache(dps)
	if err := dpc.Add(mock.NewDeviceProfile); err != nil {
		t.Error("failed to add a new device profile to cache")
	}
	if dp3, found := dpc.ForId(mock.NewDeviceProfile.Id); !found {
		t.Error("unable to find the device profile which just be added to cache")
	} else {
		assert.Equal(t, mock.NewDeviceProfile, dp3)
	}
	if err := dpc.Add(mock.DuplicateDeviceProfileRandomFloatGenerator); err == nil {
		t.Error("supposed to get an error when adding a duplicate device profile to cache")
	}
}

func TestProfileCache_RemoveByName(t *testing.T) {
	dpc := newProfileCache(dps)

	if err := dpc.RemoveByName(mock.NewDeviceProfile.Name); err == nil {
		t.Error("supposed to get an error when removing a device profile which doesn't exist in cache")
	}

	if err := dpc.RemoveByName(mock.DeviceProfileRandomBoolGenerator.Name); err != nil {
		t.Error("failed to remove device profile from cache by name")
	}

	if _, found := dpc.ForName(mock.DeviceProfileRandomBoolGenerator.Name); found {
		t.Error("unable to remove device profile from cache by name")
	}
}

func TestProfileCache_Remove(t *testing.T) {
	dpc := newProfileCache(dps)

	if err := dpc.Remove(mock.NewDeviceProfile.Id); err == nil {
		t.Error("supposed to get an error when removing a device profile which doesn't exist in cache")
	}

	if err := dpc.Remove(mock.DeviceProfileRandomBoolGenerator.Id); err != nil {
		t.Error("failed to remove device profile from cache by id")
	}

	if _, found := dpc.ForId(mock.DeviceProfileRandomBoolGenerator.Id); found {
		t.Error("unable to remove device profile from cache by id")
	}
}

func TestProfileCache_Update(t *testing.T) {
	dpc := newProfileCache(dps)

	if err := dpc.Update(mock.NewDeviceProfile); err == nil {
		t.Error("supposed to get an error when updating a device profile which doesn't exist in cache")
	}

	mock.DeviceProfileRandomBoolGenerator.Description = "TestProfileCache_Update"
	if err := dpc.Update(mock.DeviceProfileRandomBoolGenerator); err != nil {
		t.Error("failed to update device profile in cache")
	}

	if udp0, found := dpc.ForId(mock.DeviceProfileRandomBoolGenerator.Id); !found {
		t.Error("unable to find the device profile in cache after updating it")
	} else {
		assert.Equal(t, mock.DeviceProfileRandomBoolGenerator, udp0)
	}
}

func TestProfileCache_DeviceResource(t *testing.T) {
	dpc := newProfileCache(dps)

	if _, found := dpc.DeviceResource(mock.NewDeviceProfile.Name, mock.NewDeviceProfile.DeviceResources[0].Name); found {
		t.Error("not supposed to find a matching device resource by a profile name not in cache")
	}

	if dr, found := dpc.DeviceResource(mock.DeviceProfileRandomBoolGenerator.Name, mock.DeviceProfileRandomBoolGenerator.DeviceResources[0].Name); !found {
		t.Error("supposed to find a matching device resource")
	} else {
		assert.Equal(t, mock.DeviceProfileRandomBoolGenerator.DeviceResources[0], dr)
	}
}

func TestProfileCache_CommandExists(t *testing.T) {
	dpc := newProfileCache(dps)

	if _, err := dpc.CommandExists(mock.NewDeviceProfile.Name, mock.NewDeviceProfile.CoreCommands[0].Name); err == nil {
		t.Error("DeviceProfileRandomFloatGenerator is not in cache, supposed to get an error")
	}
	if exists, err := dpc.CommandExists(mock.DeviceProfileRandomBoolGenerator.Name, mock.DeviceProfileRandomBoolGenerator.CoreCommands[0].Name); err != nil {
		t.Error("DeviceProfileRandomBoolGenerator exists in cache, not supposed to get an error")
	} else if !exists {
		t.Error("DeviceProfileRandomBoolGenerator.Commands[0] exists in cache, the returned value should be true")
	}

	if exists, _ := dpc.CommandExists(mock.DeviceProfileRandomBoolGenerator.Name, "arbitaryNameXXX"); exists {
		t.Error("arbitaryNameXXX doesn't belong to any command in DeviceProfileRandomBoolGenerator.Commands, the returned value should be false")
	}
}

func TestProfileCache_ResourceOperations(t *testing.T) {
	dpc := newProfileCache(dps)

	if _, err := dpc.ResourceOperations(mock.NewDeviceProfile.Name, mock.NewDeviceProfile.CoreCommands[0].Name, common.GetCmdMethod); err == nil {
		t.Error("DeviceProfileRandomFloatGenerator is not in cache, supposed to get an error")
	}
	if _, err := dpc.ResourceOperations(mock.NewDeviceProfile.Name, mock.NewDeviceProfile.CoreCommands[0].Name, common.SetCmdMethod); err == nil {
		t.Error("DeviceProfileRandomFloatGenerator is not in cache, supposed to get an error")
	}

	if ros, err := dpc.ResourceOperations(mock.DeviceProfileRandomBoolGenerator.Name, mock.DeviceProfileRandomBoolGenerator.CoreCommands[0].Name, common.GetCmdMethod); err != nil {
		t.Error("DeviceProfileRandomBoolGenerator exists in cache, not supposed to get an error")
	} else {
		assert.Equal(t, mock.DeviceProfileRandomBoolGenerator.DeviceCommands[0].Get, ros)
	}
	if ros, err := dpc.ResourceOperations(mock.DeviceProfileRandomBoolGenerator.Name, mock.DeviceProfileRandomBoolGenerator.CoreCommands[0].Name, common.SetCmdMethod); err != nil {
		t.Error("DeviceProfileRandomBoolGenerator exists in cache, not supposed to get an error")
	} else {
		assert.Equal(t, mock.DeviceProfileRandomBoolGenerator.DeviceCommands[0].Set, ros)
	}

	if _, err := dpc.ResourceOperations(mock.DeviceProfileRandomBoolGenerator.Name, "arbitaryNameXXX", common.GetCmdMethod); err == nil {
		t.Error("the input cmd name is not belong to DeviceProfileRandomBoolGenerator, supposed to get an error")
	}
}

func TestProfileCache_ResourceOperation(t *testing.T) {
	dpc := newProfileCache(dps)

	if _, err := dpc.ResourceOperation(mock.NewDeviceProfile.Name, mock.NewDeviceProfile.DeviceCommands[0].Get[0].Object, common.GetCmdMethod); err == nil {
		t.Error("DeviceProfileRandomFloatGenerator is not in cache, supposed to get an error")
	}
	if _, err := dpc.ResourceOperation(mock.NewDeviceProfile.Name, mock.NewDeviceProfile.DeviceCommands[0].Get[0].Object, common.SetCmdMethod); err == nil {
		t.Error("DeviceProfileRandomFloatGenerator is not in cache, supposed to get an error")
	}

	if ro, err := dpc.ResourceOperation(mock.DeviceProfileRandomBoolGenerator.Name, mock.DeviceProfileRandomBoolGenerator.DeviceCommands[0].Get[0].Object, common.GetCmdMethod); err != nil {
		t.Error("DeviceProfileRandomBoolGenerator exists in cache, not supposed to get an error")
	} else {
		assert.Equal(t, mock.DeviceProfileRandomBoolGenerator.DeviceCommands[0].Get[0], ro)
	}
	if ro, err := dpc.ResourceOperation(mock.DeviceProfileRandomBoolGenerator.Name, mock.DeviceProfileRandomBoolGenerator.DeviceCommands[0].Get[0].Object, common.GetCmdMethod); err != nil {
		t.Error("DeviceProfileRandomBoolGenerator exists in cache, not supposed to get an error")
	} else {
		assert.Equal(t, mock.DeviceProfileRandomBoolGenerator.DeviceCommands[0].Get[0], ro)
	}

	if _, err := dpc.ResourceOperation(mock.DeviceProfileRandomBoolGenerator.Name, "arbitaryNameXXX", common.GetCmdMethod); err == nil {
		t.Error("the input object name of resource operation is not belong to DeviceProfileRandomBoolGenerator, supposed to get an error")
	}
}
