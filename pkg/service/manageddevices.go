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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

// AddDevice adds a new Device to the Device Service and Core Metadata
// Returns new Device id or non-nil error.
func (s *DeviceServiceSDK) AddDevice(device contract.Device) (id string, err error) {
	if d, ok := cache.Devices().ForName(device.Name); ok {
		return d.Id, fmt.Errorf("name conflicted, Device %s exists", device.Name)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Adding managed device: : %s\n", device.Name))

	var prf contract.DeviceProfile
	prf, cacheExist := cache.Profiles().ForName(device.Profile.Name)
	if !cacheExist {
		dpc := s.edgexClients.DeviceProfileClient
		prf, err = dpc.DeviceProfileForName(context.Background(), device.Profile.Name)
		if err != nil {
			errMsg := fmt.Sprintf("Device Profile %s doesn't exist for Device %s", device.Profile.Name, device.Name)
			s.LoggingClient.Error(errMsg)
			return "", fmt.Errorf(errMsg)
		}
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	device.Origin = millis
	device.Service = s.deviceService
	device.Profile = prf

	dc := s.edgexClients.DeviceClient
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err = dc.Add(ctx, &device)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Add Device failed %s, error: %v", device.Name, err))
		return "", err
	}
	if err = common.VerifyIdFormat(id, "Device", s.LoggingClient); err != nil {
		return "", err
	}
	device.Id = id

	return id, nil
}

// Devices return all managed Devices from cache
func (s *DeviceServiceSDK) Devices() []contract.Device {
	return cache.Devices().All()
}

// GetDeviceByName returns the Device by its name if it exists in the cache, or returns an error.
func (s *DeviceServiceSDK) GetDeviceByName(name string) (contract.Device, error) {
	device, ok := cache.Devices().ForName(name)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", name)
		s.LoggingClient.Info(msg)
		return contract.Device{}, fmt.Errorf(msg)
	}
	return device, nil
}

// RemoveDevice removes the specified Device by id from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceServiceSDK) RemoveDevice(id string) error {
	device, ok := cache.Devices().ForId(id)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", id)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Removing managed Device: : %s\n", device.Name))
	dc := s.edgexClients.DeviceClient
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := dc.Delete(ctx, id)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Delete Device %s from Core Metadata failed", id))
	}

	return err
}

// RemoveDevice removes the specified Device by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceServiceSDK) RemoveDeviceByName(name string) error {
	device, ok := cache.Devices().ForName(name)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", name)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Removing managed Device: : %s\n", device.Name))
	dc := s.edgexClients.DeviceClient
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := dc.DeleteByName(ctx, name)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Delete Device %s from Core Metadata failed", name))
	}

	return err
}

// UpdateDevice updates the Device in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *DeviceServiceSDK) UpdateDevice(device contract.Device) error {
	_, ok := cache.Devices().ForId(device.Id)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", device.Id)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Updating managed Device: : %s\n", device.Name))
	dc := s.edgexClients.DeviceClient
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := dc.Update(ctx, device)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Update Device %s from Core Metadata failed: %v", device.Name, err))
	}

	return err
}

// UpdateDeviceOperatingState updates the Device's OperatingState with given name
// in Core Metadata and device service cache.
func (s *DeviceServiceSDK) UpdateDeviceOperatingState(deviceName string, state string) error {
	d, ok := cache.Devices().ForName(deviceName)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", deviceName)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Updating managed Device OperatingState: : %s\n", d.Name))
	dc := s.edgexClients.DeviceClient
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := dc.UpdateOpState(ctx, d.Id, state)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Update Device %s OperatingState from Core Metadata failed: %v", d.Name, err))
	}

	return err
}
