// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"context"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

// AddDeviceProfile adds a new DeviceProfile to the device service and Core Metadata
// Returns new DeviceProfile id or non-nil error.
func (s *Service) AddDeviceProfile(profile models.DeviceProfile) (id string, err error) {
	if p, ok := cache.Profiles().ForName(profile.Name); ok {
		return p.Id, fmt.Errorf("name conflicted, Profile %s exists", profile.Name)
	}

	common.LoggingClient.Debug(fmt.Sprintf("Adding managed Profile: : %v\n", profile))
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	profile.Origin = millis
	common.LoggingClient.Debug(fmt.Sprintf("Adding Profile: %v", profile))

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err = common.DeviceProfileClient.Add(&profile, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Add Profile failed %v, error: %v", profile, err))
		return "", err
	}
	if err = common.VerifyIdFormat(id, "Device Profile"); err != nil {
		return "", err
	}
	profile.Id = id
	cache.Profiles().Add(profile)

	provision.CreateDescriptorsFromProfile(&profile)

	return id, nil
}

// DeviceProfiles return all managed DeviceProfiles from cache
func (s *Service) DeviceProfiles() []models.DeviceProfile {
	return cache.Profiles().All()
}

// RemoveDeviceProfile removes the specified DeviceProfile by id from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *Service) RemoveDeviceProfile(id string) error {
	profile, ok := cache.Profiles().ForId(id)
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", id)
		common.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LoggingClient.Debug(fmt.Sprintf("Removing managed DeviceProfile: : %v\n", profile))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := common.DeviceProfileClient.Delete(id, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Delete DeviceProfile %s from Core Metadata failed", id))
		return err
	}

	err = cache.Profiles().Remove(id)
	return err
}

// RemoveDeviceProfileByName removes the specified DeviceProfile by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (*Service) RemoveDeviceProfileByName(name string) error {
	profile, ok := cache.Profiles().ForName(name)
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", name)
		common.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LoggingClient.Debug(fmt.Sprintf("Removing managed DeviceProfile: : %v\n", profile))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := common.DeviceProfileClient.DeleteByName(name, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Delete DeviceProfile %s from Core Metadata failed", name))
		return err
	}

	err = cache.Profiles().RemoveByName(profile.Name)
	return err
}

// UpdateDeviceProfile updates the DeviceProfile in the cache and ensures that the
// copy in Core Metadata is also updated.
func (*Service) UpdateDeviceProfile(profile models.DeviceProfile) error {
	_, ok := cache.Profiles().ForId(profile.Id)
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", profile.Id)
		common.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LoggingClient.Debug(fmt.Sprintf("Updating managed DeviceProfile: : %v\n", profile))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := common.DeviceProfileClient.Update(profile, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Update DeviceProfile %s from Core Metadata failed: %v", profile.Name, err))
		return err
	}

	err = cache.Profiles().Update(profile)
	provision.CreateDescriptorsFromProfile(&profile)

	return err
}

// ResourceOperation retrieves the first matched ResourceOpereation instance from cache according to
// the Device name, Device resource (object) name, and the method (get or set)
func (*Service) ResourceOperation(deviceName string, object string, method string) (models.ResourceOperation, bool) {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		common.LoggingClient.Error(fmt.Sprintf("retrieving ResourceOperation - Device %s not found", deviceName))
	}

	ro, err := cache.Profiles().ResourceOperation(device.Profile.Name, object, method)
	if err != nil {
		common.LoggingClient.Error(err.Error())
		return ro, false
	}
	return ro, true
}

// DeviceResource retrieves the specific DeviceResource instance from cache according to
// the Device name and Device resource (object) name
func (*Service) DeviceResource(deviceName string, object string, method string) (models.DeviceResource, bool) {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		common.LoggingClient.Error(fmt.Sprintf("retrieving DeviceResource - Device %s not found", deviceName))
	}

	dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, object)
	if !ok {
		return dr, false
	}
	return dr, true
}
