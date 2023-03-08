// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 VMware
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
)

// AddDeviceAutoEvent adds a new AutoEvent to the Device with given name
func (s *deviceService) AddDeviceAutoEvent(deviceName string, event models.AutoEvent) error {
	found := false
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		msg := fmt.Sprintf("failed to find device %s in cache", deviceName)
		s.lc.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	for _, e := range device.AutoEvents {
		if e.SourceName == event.SourceName {
			s.lc.Debugf("Updating existing AutoEvent %s for device %s", e.SourceName, deviceName)
			e.Interval = event.Interval
			e.OnChange = event.OnChange
			found = true
			break
		}
	}

	if !found {
		s.lc.Debugf("Adding new AutoEvent %s to device %s", event.SourceName, deviceName)
		device.AutoEvents = append(device.AutoEvents, event)
		err := cache.Devices().Update(device)
		if err != nil {
			s.lc.Errorf("failed to update device %s with AutoEvent change", deviceName)
			return err
		}
	}
	s.autoEventManager.RestartForDevice(deviceName)

	return nil
}

// RemoveDeviceAutoEvent removes an AutoEvent from the Device with given name
func (s *deviceService) RemoveDeviceAutoEvent(deviceName string, event models.AutoEvent) error {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		msg := fmt.Sprintf("failed to find device %s cannot in cache", deviceName)
		s.lc.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	for i, e := range device.AutoEvents {
		if e.SourceName == event.SourceName {
			s.lc.Debugf("Removing AutoEvent %s for device %s", e.SourceName, deviceName)
			device.AutoEvents = append(device.AutoEvents[:i], device.AutoEvents[i+1:]...)
			break
		}
	}

	err := cache.Devices().Update(device)
	if err != nil {
		s.lc.Errorf("failed to update device %s with AutoEvent change", deviceName)
		return err
	}
	s.autoEventManager.RestartForDevice(deviceName)

	return nil
}
