// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 VMware
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/cache"
)

// AddDeviceAutoEvent adds a new AutoEvent to the Device with given name
func (s *DeviceService) AddDeviceAutoEvent(deviceName string, event models.AutoEvent) error {
	found := false
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", deviceName)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	for _, e := range device.AutoEvents {
		if e.Resource == event.Resource {
			s.LoggingClient.Debug(fmt.Sprintf("Updating existing auto event %s for device %s\n", e.Resource, deviceName))
			e.Frequency = event.Frequency
			e.OnChange = event.OnChange
			found = true
			break
		}
	}

	if !found {
		s.LoggingClient.Debug(fmt.Sprintf("Adding new auto event to device %s: %v\n", deviceName, event))
		device.AutoEvents = append(device.AutoEvents, event)
		cache.Devices().Update(device)
	}

	s.manager.RestartForDevice(deviceName)

	return nil
}

// RemoveDeviceAutoEvent removes an AutoEvent from the Device with given name
func (s *DeviceService) RemoveDeviceAutoEvent(deviceName string, event models.AutoEvent) error {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", deviceName)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.manager.StopForDevice(deviceName)
	for i, e := range device.AutoEvents {
		if e.Resource == event.Resource {
			s.LoggingClient.Debug(fmt.Sprintf("Removing auto event %s for device %s\n", e.Resource, deviceName))
			device.AutoEvents = append(device.AutoEvents[:i], device.AutoEvents[i+1:]...)
			break
		}
	}
	cache.Devices().Update(device)
	s.manager.RestartForDevice(deviceName)

	return nil
}
