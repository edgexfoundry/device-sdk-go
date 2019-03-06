// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides management of device service related
// objects that may be distributed across one or more EdgeX
// core microservices.
//
package device

import (
	"context"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

// AddDevice adds a new Device to the device service and Core Metadata
// Returns new Device id or non-nil error.
func (s *Service) AddDevice(device models.Device) (id string, err error) {
	if d, ok := cache.Devices().ForName(device.Name); ok {
		return d.Id, fmt.Errorf("name conflicted, Device %s exists", device.Name)
	}

	common.LoggingClient.Debug(fmt.Sprintf("Adding managed device: : %v\n", device))

	prf, ok := cache.Profiles().ForName(device.Profile.Name)
	if !ok {
		errMsg := fmt.Sprintf("Device Profile %s doesn't exist for Device %v", device.Profile.Name, device)
		common.LoggingClient.Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	device.Origin = millis
	device.Service = common.CurrentDeviceService
	device.Profile = prf
	common.LoggingClient.Debug(fmt.Sprintf("Adding Device: %v", device))

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err = common.DeviceClient.Add(&device, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Add Device failed %v, error: %v", device, err))
		return "", err
	}
	if err = common.VerifyIdFormat(id, "Device"); err != nil {
		return "", err
	}
	device.Id = id
	cache.Devices().Add(device)

	return id, nil
}

// Devices return all managed Devices from cache
func (s *Service) Devices() []models.Device {
	return cache.Devices().All()
}

// GetDeviceByName returns device if it exists in EdgeX registration cache.
func (s *Service) GetDeviceByName(name string) (models.Device, error) {
	device, ok := cache.Devices().ForName(name)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", name)
		common.LoggingClient.Info(msg)
		return models.Device{}, fmt.Errorf(msg)
	}
	return device, nil
}

// RemoveDevice removes the specified Device by id from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *Service) RemoveDevice(id string) error {
	device, ok := cache.Devices().ForId(id)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", id)
		common.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LoggingClient.Debug(fmt.Sprintf("Removing managed Device: : %v\n", device))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := common.DeviceClient.Delete(id, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Delete Device %s from Core Metadata failed", id))
		return err
	}

	err = cache.Devices().Remove(id)
	return err
}

// RemoveDevice removes the specified Device by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *Service) RemoveDeviceByName(name string) error {
	device, ok := cache.Devices().ForName(name)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", name)
		common.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LoggingClient.Debug(fmt.Sprintf("Removing managed Device: : %v\n", device))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := common.DeviceClient.DeleteByName(name, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Delete Device %s from Core Metadata failed", name))
		return err
	}

	err = cache.Devices().RemoveByName(name)
	return err
}

// UpdateDevice updates the Device in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *Service) UpdateDevice(device models.Device) error {
	_, ok := cache.Devices().ForId(device.Id)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", device.Id)
		common.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	common.LoggingClient.Debug(fmt.Sprintf("Updating managed Device: : %v\n", device))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := common.DeviceClient.Update(device, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Update Device %s from Core Metadata failed: %v", device.Name, err))
		return err
	}

	err = cache.Devices().Update(device)
	return err
}
