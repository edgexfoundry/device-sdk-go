// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	dc *deviceCache
)

type DeviceCache interface {
	ForName(name string) (contract.Device, bool)
	ForId(id string) (contract.Device, bool)
	All() []contract.Device
	Add(device contract.Device) error
	Update(device contract.Device) error
	Remove(id string) error
	RemoveByName(name string) error
	UpdateAdminState(id string, state contract.AdminState) error
}

type deviceCache struct {
	dMap    map[string]*contract.Device // key is Device name
	nameMap map[string]string           // key is id, and value is Device name
}

// ForName returns a Device with the given name.
func (d *deviceCache) ForName(name string) (contract.Device, bool) {
	if device, ok := d.dMap[name]; ok {
		return *device, ok
	} else {
		return contract.Device{}, ok
	}
}

// ForId returns a device with the given device id.
func (d *deviceCache) ForId(id string) (contract.Device, bool) {
	name, ok := d.nameMap[id]
	if !ok {
		return contract.Device{}, ok
	}

	if device, ok := d.dMap[name]; ok {
		return *device, ok
	} else {
		return contract.Device{}, ok
	}
}

// All() returns the current list of devices in the cache.
func (d *deviceCache) All() []contract.Device {
	devices := make([]contract.Device, len(d.dMap))
	i := 0
	for _, device := range d.dMap {
		devices[i] = *device
		i++
	}
	return devices
}

// Adds a new device to the cache. This method is used to populate the
// devices cache with pre-existing devices from Core Metadata, as well
// as create new devices returned in a ScanList during discovery.
func (d *deviceCache) Add(device contract.Device) error {
	if _, ok := d.dMap[device.Name]; ok {
		return fmt.Errorf("device %s has already existed in cache", device.Name)
	}
	d.dMap[device.Name] = &device
	d.nameMap[device.Id] = device.Name
	return nil
}

// Update updates the device in the cache
func (d *deviceCache) Update(device contract.Device) error {
	if err := d.Remove(device.Id); err != nil {
		return err
	}
	return d.Add(device)
}

// Remove removes the specified device by id from the cache.
func (d *deviceCache) Remove(id string) error {
	name, ok := d.nameMap[id]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", id)
	}

	return d.RemoveByName(name)
}

// RemoveByName removes the specified device by name from the cache.
func (d *deviceCache) RemoveByName(name string) error {
	device, ok := d.dMap[name]
	if !ok {
		return fmt.Errorf("device %s does not exist in cache", name)
	}

	delete(d.nameMap, device.Id)
	delete(d.dMap, name)
	return nil
}

// UpdateAdminState updates the device admin state in cache by id. This method
// is used by the UpdateHandler to trigger update device admin state that's been
// updated directly to Core Metadata.
func (d *deviceCache) UpdateAdminState(id string, state contract.AdminState) error {
	name, ok := d.nameMap[id]
	if !ok {
		return fmt.Errorf("device %s cannot be found in cache", id)
	}

	d.dMap[name].AdminState = state
	return nil
}

func newDeviceCache(devices []contract.Device) DeviceCache {
	defaultSize := len(devices) * 2
	dMap := make(map[string]*contract.Device, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	for i, d := range devices {
		dMap[d.Name] = &devices[i]
		nameMap[d.Id] = d.Name
	}
	dc = &deviceCache{dMap: dMap, nameMap: nameMap}
	return dc
}

func Devices() DeviceCache {
	if dc == nil {
		InitCache()
	}
	return dc
}
