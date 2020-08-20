// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

// AddDeviceProfile adds a new DeviceProfile to the Device Service and Core Metadata
// Returns new DeviceProfile id or non-nil error.
func (s *DeviceService) AddDeviceProfile(profile contract.DeviceProfile) (id string, err error) {
	if p, ok := cache.Profiles().ForName(profile.Name); ok {
		return p.Id, fmt.Errorf("name conflicted, Profile %s exists", profile.Name)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Adding managed Profile: : %s", profile.Name))
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	profile.Origin = millis

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err = s.edgexClients.DeviceProfileClient.Add(ctx, &profile)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Add Profile failed %s, error: %v", profile.Name, err))
		return "", err
	}
	if err = common.VerifyIdFormat(id, "Device Profile"); err != nil {
		return "", err
	}
	profile.Id = id
	_ = cache.Profiles().Add(profile)

	provision.CreateDescriptorsFromProfile(
		&profile,
		s.LoggingClient,
		s.edgexClients.GeneralClient,
		s.edgexClients.ValueDescriptorClient)

	return id, nil
}

// DeviceProfiles return all managed DeviceProfiles from cache
func (s *DeviceService) DeviceProfiles() []contract.DeviceProfile {
	return cache.Profiles().All()
}

// RemoveDeviceProfile removes the specified DeviceProfile by id from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceService) RemoveDeviceProfile(id string) error {
	profile, ok := cache.Profiles().ForId(id)
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", id)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Removing managed DeviceProfile: : %s", profile.Name))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := s.edgexClients.DeviceProfileClient.Delete(ctx, id)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Delete DeviceProfile %s from Core Metadata failed", id))
		return err
	}

	err = cache.Profiles().Remove(id)
	return err
}

// RemoveDeviceProfileByName removes the specified DeviceProfile by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceService) RemoveDeviceProfileByName(name string) error {
	profile, ok := cache.Profiles().ForName(name)
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", name)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Removing managed DeviceProfile: : %s", profile.Name))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := s.edgexClients.DeviceProfileClient.DeleteByName(ctx, name)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Delete DeviceProfile %s from Core Metadata failed", name))
		return err
	}

	err = cache.Profiles().RemoveByName(profile.Name)
	return err
}

// UpdateDeviceProfile updates the DeviceProfile in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *DeviceService) UpdateDeviceProfile(profile contract.DeviceProfile) error {
	_, ok := cache.Profiles().ForId(profile.Id)
	if !ok {
		msg := fmt.Sprintf("DeviceProfile %s cannot be found in cache", profile.Id)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Updating managed DeviceProfile: : %s", profile.Name))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := s.edgexClients.DeviceProfileClient.Update(ctx, profile)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Update DeviceProfile %s from Core Metadata failed: %v", profile.Name, err))
		return err
	}

	err = cache.Profiles().Update(profile)
	provision.CreateDescriptorsFromProfile(
		&profile,
		s.LoggingClient,
		s.edgexClients.GeneralClient,
		s.edgexClients.ValueDescriptorClient)

	return err
}

// ResourceOperation retrieves the first matched ResourceOpereation instance from cache according to
// the Device name, Device Resource name, and the method (get or set).
func (s *DeviceService) ResourceOperation(deviceName string, deviceResource string, method string) (contract.ResourceOperation, bool) {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		s.LoggingClient.Error(fmt.Sprintf("retrieving ResourceOperation - Device %s not found", deviceName))
	}

	ro, err := cache.Profiles().ResourceOperation(device.Profile.Name, deviceResource, method)
	if err != nil {
		s.LoggingClient.Error(err.Error())
		return ro, false
	}
	return ro, true
}

// DeviceResource retrieves the specific DeviceResource instance from cache according to
// the Device name and Device Resource name
func (s *DeviceService) DeviceResource(deviceName string, deviceResource string, _ string) (contract.DeviceResource, bool) {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		s.LoggingClient.Error(fmt.Sprintf("retrieving DeviceResource - Device %s not found", deviceName))
	}

	dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, deviceResource)
	if !ok {
		return dr, false
	}
	return dr, true
}
