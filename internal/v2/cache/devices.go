// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

var (
	dc *deviceCache
)

type DeviceCache interface {
	ForName(name string) (contract.Device, bool)
	ForId(id string) (contract.Device, bool)
	All() []contract.Device
	Add(device contract.Device) edgexErr.EdgeX
	Update(device contract.Device) edgexErr.EdgeX
	Remove(id string) edgexErr.EdgeX
	RemoveByName(name string) edgexErr.EdgeX
	UpdateAdminState(id string, state contract.AdminState) edgexErr.EdgeX
}

type deviceCache struct {
	deviceMap map[string]*contract.Device // key is Device name
	nameMap   map[string]string           // key is id, and value is Device name
	mutex     sync.Mutex
}

// ForName returns a Device with the given device name.
func (d *deviceCache) ForName(name string) (contract.Device, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	device, ok := d.deviceMap[name]
	return *device, ok
}

// ForId returns a device with the given device id.
func (d *deviceCache) ForId(id string) (contract.Device, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	name, ok := d.nameMap[id]
	if !ok {
		return contract.Device{}, ok
	}

	device, ok := d.deviceMap[name]
	return *device, ok
}

// All returns the current list of devices in the cache.
func (d *deviceCache) All() []contract.Device {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	i := 0
	devices := make([]contract.Device, len(d.deviceMap))
	for _, device := range d.deviceMap {
		devices[i] = *device
		i++
	}
	return devices
}

// Add adds a new device to the cache. This method is used to populate the
// device cache with pre-existing or recently-added devices from Core Metadata.
func (d *deviceCache) Add(device contract.Device) edgexErr.EdgeX {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.add(device)
}

func (d *deviceCache) add(device contract.Device) edgexErr.EdgeX {
	if _, ok := d.deviceMap[device.Name]; ok {
		errMsg := fmt.Sprintf("device %s has already existed in cache", device.Name)
		return edgexErr.NewCommonEdgeX(edgexErr.KindDuplicateName, errMsg, nil)
	}

	d.deviceMap[device.Name] = &device
	d.nameMap[device.Id] = device.Name
	return nil
}

// Update updates the device in the cache
func (d *deviceCache) Update(device contract.Device) edgexErr.EdgeX {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.remove(device.Id); err != nil {
		return err
	}

	return d.add(device)
}

// Remove removes the specified device by id from the cache.
func (d *deviceCache) Remove(id string) edgexErr.EdgeX {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.remove(id)
}

func (d *deviceCache) remove(id string) edgexErr.EdgeX {
	name, ok := d.nameMap[id]
	if !ok {
		errMsg := fmt.Sprintf("failed to find device with given id %s in cache", id)
		return edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	return d.removeByName(name)
}

// RemoveByName removes the specified device by name from the cache.
func (d *deviceCache) RemoveByName(name string) edgexErr.EdgeX {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.removeByName(name)
}

func (d *deviceCache) removeByName(name string) edgexErr.EdgeX {
	device, ok := d.deviceMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find device %s in cache", name)
		return edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	delete(d.nameMap, device.Id)
	delete(d.deviceMap, name)
	return nil
}

// UpdateAdminState updates the device admin state in cache by id. This method
// is used by the UpdateHandler to trigger update device admin state that's been
// updated directly to Core Metadata.
func (d *deviceCache) UpdateAdminState(id string, state contract.AdminState) edgexErr.EdgeX {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	name, ok := d.nameMap[id]
	if !ok {
		errMsg := fmt.Sprintf("failed to find device with given id %s in cache", id)
		return edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	d.deviceMap[name].AdminState = state
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
	dc = &deviceCache{deviceMap: dMap, nameMap: nameMap}
	return dc
}

func Devices() DeviceCache {
	return dc
}
