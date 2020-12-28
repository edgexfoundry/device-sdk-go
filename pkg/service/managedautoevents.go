// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 VMware
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// AddDeviceAutoEvent adds a new AutoEvent to the Device with given name
func (s *DeviceService) AddDeviceAutoEvent(deviceName string, event contract.AutoEvent) error {
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

	autoevent.GetManager().RestartForDevice(deviceName, nil)

	return nil
}

// RemoveDeviceAutoEvent removes an AutoEvent from the Device with given name
func (s *DeviceService) RemoveDeviceAutoEvent(deviceName string, event contract.AutoEvent) error {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		msg := fmt.Sprintf("Device %s cannot be found in cache", deviceName)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	autoevent.GetManager().StopForDevice(deviceName)
	for i, e := range device.AutoEvents {
		if e.Resource == event.Resource {
			s.LoggingClient.Debug(fmt.Sprintf("Removing auto event %s for device %s\n", e.Resource, deviceName))
			device.AutoEvents = append(device.AutoEvents[:i], device.AutoEvents[i+1:]...)
			break
		}
	}
	cache.Devices().Update(device)
	autoevent.GetManager().RestartForDevice(deviceName, nil)

	return nil
}
