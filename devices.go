// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package provides management of device service related
// objects that may be distributed across one or more EdgeX
// core microservices.
//
package device

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

// Inteface for device
type deviceCacheInterface interface {
	InitDeviceCache()
	Device(name string) *models.Device
	Devices() map[string]*models.Device
	Add(dev *models.Device) error
	AddById(id string) error
	Update(dev *models.Device) error
	UpdateAdminState(id string) error
	DeviceById(id string) *models.Device
	Remove(dev *models.Device) error
	IsDeviceLocked(id string) (exists, locked bool)
	SetDeviceOpState(name string, os models.OperatingState) error
	SetDeviceByIdOpState(id string, os models.OperatingState) error
}

// deviceCache is a local cache of devices seeded from Core Metadata.
type deviceCache struct {
	devices map[string]*models.Device
	names   map[string]string
}

var (
	dcOnce sync.Once
	dc     deviceCacheInterface
)

// Creates a singleton deviceCache instance.
func newDeviceCache(serviceId string) error {
	var retval error

	dcOnce.Do(func() {
		dc = &deviceCache{}

		mDevs, err := svc.dc.DevicesForService(serviceId)
		if err != nil {
			svc.lc.Error(fmt.Sprintf("DevicesForService error: %v\n", err))
			retval = err
		}

		svc.lc.Debug(fmt.Sprintf("returned devices %v\n", mDevs))

		dc.InitDeviceCache()

		for index, _ := range mDevs {
			dc.Add(&mDevs[index])
		}

		// TODO: call Protocol.initialize
		svc.lc.Debug(fmt.Sprintf("dstore: INITIALIZATION DONE! err=%v\n", err))
	})

	return retval
}

// Init basic state for deviceCache
func (d *deviceCache) InitDeviceCache() {
	d.devices = make(map[string]*models.Device)
	d.names = make(map[string]string)
}

// Adds a new device to the cache. This method is used to populate the
// devices cache with pre-existing devices from Core Metadata, as well
// as create new devices returned in a ScanList during discovery.
func (d *deviceCache) Add(dev *models.Device) error {

	// if device already exists in devices, delete & re-add
	if _, ok := d.devices[dev.Name]; ok {
		delete(d.devices, dev.Name)

		// TODO: remove from profiles
	}

	svc.lc.Debug(fmt.Sprintf("Adding managed device: : %v\n", dev))

	// TODO: per effective go, should these two stmts be collapsed?
	// check if this is commonly used in Go src & snapd.
	err := d.addDeviceToMetadata(dev)
	if err != nil {
		return err
	}

	// This is only the case for brand new devices
	if dev.OperatingState == models.OperatingState("ENABLED") {
		svc.lc.Debug(fmt.Sprintf("Initializing device: : %v\n", dev))
		// TODO: ${Protocol name}.initializeDevice(metaDevice);
	}

	return nil
}

// AddById adds a new device to the cache by id. This
// method is used by the UpdateHandler to trigger addition of a
// device that's been added directly to Core Metadata.
func (d *deviceCache) AddById(id string) error {
	return nil
}

// Device returns a device with the given name.
func (d *deviceCache) Device(name string) *models.Device {
	return nil
}

// DeviceById returns a device with the given device id.
func (d *deviceCache) DeviceById(id string) *models.Device {
	name, ok := d.names[id]
	if !ok {
		return nil
	}

	dev := d.devices[name]
	return dev
}

// Devices returns the current list of devices in the cache.
// TODO: based on the java code; we need to check how this function
// is used, as it's bad form to return an internal data struct to
// callers, especially when the result is a map, which can then be
// modified externally to this package.
func (d *deviceCache) Devices() map[string]*models.Device {
	return d.devices
}

// IsDeviceLocked returns a bool which indicates if the specified
// device exists, and a book which indicates whether the device is locked
func (d *deviceCache) IsDeviceLocked(id string) (exists, locked bool) {
	return false, false
}

// Remove removes the specified device from the cache.
func (d *deviceCache) Remove(dev *models.Device) error {
	err := svc.dc.Delete(dev.Id.Hex())
	if err != nil {
		return err
	}

	delete(d.devices, dev.Name)
	delete(d.devices, dev.Id.Hex())

	return nil
}

// SetDeviceOpState sets the operatingState of the device specified by name.
func (d *deviceCache) SetDeviceOpState(name string, os models.OperatingState) error {
	return nil
}

// SetDeviceByIdOpState sets the operatingState of the device specified by name.
func (d *deviceCache) SetDeviceByIdOpState(id string, os models.OperatingState) error {
	return nil
}

// Update updates the device in the cache and ensures that the
// copy in Core Metadata is also updated.
func (d *deviceCache) Update(dev *models.Device) error {
	err := svc.dc.Update(*dev)
	if err != nil {
		return err
	}

	// consider device name can be modified, so remove the old one and put new one
	if _, ok := d.names[dev.Id.Hex()]; ok {
		delete(d.devices, d.names[dev.Id.Hex()])
	}
	d.devices[dev.Name] = dev
	d.names[dev.Id.Hex()] = dev.Name

	return nil
}

// UpdateAdminState updates the device admin state in cache by id. This method
// is used by the UpdateHandler to trigger update device admin state that's been
// updated directly to Core Metadata.
func (d *deviceCache) UpdateAdminState(id string) error {
	name, ok := d.names[id]
	if !ok {
		return errors.New("Device not found")
	}
	dev, err := svc.dc.Device(id)
	if err != nil {
		return err
	}

	d.devices[name].AdminState = dev.AdminState
	return nil
}

// TODO: this should method should be broken into two separate
// functions, one which validates an existing device and adds
// it to the local cache, and one that adds a brand new device.
// The current method is an almost direct translation of the Java
// DeviceStore implementation.
func (d *deviceCache) addDeviceToMetadata(dev *models.Device) error {
	// TODO: fix metadata to indicate !found, vs. returned zeroed struct!
	svc.lc.Debug(fmt.Sprintf("Trying to find addressable for: %s\n", dev.Addressable.Name))
	addr, err := svc.ac.AddressableForName(dev.Addressable.Name)
	if err != nil {
		svc.lc.Error(fmt.Sprintf("AddressClient.AddressableForName: %s; failed: %v\n", dev.Addressable.Name, err))

		// If device exists in metadata, and lacks an Addressable, don't try to fix; skip instead
		if dev.Id.Valid() {
			return fmt.Errorf("Existing metadata dev has no addressable: %s", dev.Addressable.Name)
		}
	}

	// TODO: this is the best test for not-found for now...
	if addr.Name != dev.Addressable.Name {
		addr = dev.Addressable
		addr.BaseObject.Origin = time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)
		svc.lc.Debug(fmt.Sprintf("Creating new Addressable Object with name: %v", addr))

		id, err := svc.ac.Add(&addr)
		if err != nil {
			svc.lc.Error(fmt.Sprintf("AddressClient.Add: %s; failed: %v\n", dev.Addressable.Name, err))
			return err
		}

		// TODO: add back length check in from non-public metadata-clients logic
		//
		// if len(bodyBytes) != 24 || !bson.IsObjectIdHex(bodyString) {
		//
		if !bson.IsObjectIdHex(id) {
			return fmt.Errorf("Add addressable returned invalid Id: %s\n", id)
		} else {
			addr.Id = bson.ObjectIdHex(id)
			svc.lc.Debug(fmt.Sprintf("New addressable Id: %s\n", addr.Id.Hex()))
		}
	}

	// A device without a valid Id is new
	if dev.Id.Valid() == false {
		svc.lc.Debug(fmt.Sprintf("Trying to find device for: %s\n", dev.Name))
		mDev, err := svc.dc.DeviceForName(dev.Name)
		if err != nil {
			svc.lc.Error(fmt.Sprintf("DeviceClient.DeviceForName: %s; failed: %v\n", dev.Name, err))
		}

		// TODO: this is the best test for not-found for now...
		if mDev.Name != dev.Name {
			svc.lc.Debug(fmt.Sprintf("Adding Device to Metadata: %s\n", dev.Name))

			id, err := svc.dc.Add(dev)
			if err != nil {
				svc.lc.Error(fmt.Sprintf("DeviceClient.Add for %s failed: %v", dev.Name, err))
				return err
			}

			// TODO: add back length check in from non-public metadata-clients logic
			//
			// if len(bodyBytes) != 24 || !bson.IsObjectIdHex(bodyString) {
			//
			if !bson.IsObjectIdHex(id) {
				return fmt.Errorf("DeviceClient Add returned invalid id: %s\n", id)
			} else {
				dev.Id = bson.ObjectIdHex(id)
				svc.lc.Debug(fmt.Sprintf("New dev id: %s\n", dev.Id.Hex()))
			}
		} else {
			dev.Id = mDev.Id

			if dev.OperatingState != mDev.OperatingState {
				err := svc.dc.UpdateOpState(dev.Id.Hex(), string(dev.OperatingState))
				if err != nil {
					svc.lc.Error(fmt.Sprintf("DeviceClient.UpdateOpState: %s; failed: %v\n", dev.Name, err))
				}
			}
			// TODO: Java service doesn't check result, if UpdateOpState fails,
			// should device add fail too?
		}
	}

	err = pc.addDevice(dev)
	if err != nil {
		return err
	}

	d.devices[dev.Name] = dev
	d.names[dev.Id.Hex()] = dev.Name

	return nil
}

// FIXME: !threadsafe - none of the compare methods are threadsafe
// as other code can access the struct instances and potentially
// modify them while they're being compared.
func compareCommands(a []models.Command, b []models.Command) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func compareDevices(a models.Device, b models.Device) bool {
	labelsOk := compareStrings(a.Labels, b.Labels)
	profileOk := compareDeviceProfiles(a.Profile, b.Profile)
	serviceOk := compareDeviceServices(a.Service, b.Service)

	return a.Addressable == b.Addressable &&
		a.AdminState == b.AdminState &&
		a.Description == b.Description &&
		a.Id == b.Id &&
		a.Location == b.Location &&
		a.Name == b.Name &&
		a.OperatingState == b.OperatingState &&
		labelsOk &&
		profileOk &&
		serviceOk
}

func compareDeviceProfiles(a models.DeviceProfile, b models.DeviceProfile) bool {
	labelsOk := compareStrings(a.Labels, b.Labels)
	cmdsOk := compareCommands(a.Commands, b.Commands)
	devResourcesOk := compareDeviceResources(a.DeviceResources, b.DeviceResources)
	resourcesOk := compareResources(a.Resources, b.Resources)

	// TODO: Objects fields aren't compared as to do properly
	// requires introspection as Obects is a slice of interface{}

	return a.DescribedObject == b.DescribedObject &&
		a.Id == b.Id &&
		a.Name == b.Name &&
		a.Manufacturer == b.Manufacturer &&
		a.Model == b.Model &&
		labelsOk &&
		cmdsOk &&
		devResourcesOk &&
		resourcesOk

	return true
}

func compareDeviceResources(a []models.DeviceObject, b []models.DeviceObject) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		// TODO: Attributes aren't compared, as to do properly
		// requires introspection as Attributes is an interface{}

		if a[i].Description != b[i].Description ||
			a[i].Name != b[i].Name ||
			a[i].Tag != b[i].Tag ||
			a[i].Properties != b[i].Properties {
			return false
		}
	}

	return true
}

func compareDeviceServices(a models.DeviceService, b models.DeviceService) bool {

	serviceOk := compareServices(a.Service, b.Service)

	return a.AdminState == b.AdminState && serviceOk
}

func compareResources(a []models.ProfileResource, b []models.ProfileResource) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		getOk := compareResourceOperations(a[i].Get, b[i].Set)
		setOk := compareResourceOperations(a[i].Get, b[i].Set)

		if a[i].Name != b[i].Name && !getOk && !setOk {
			return false
		}
	}

	return true
}

func compareResourceOperations(a []models.ResourceOperation, b []models.ResourceOperation) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		secondaryOk := compareStrings(a[i].Secondary, b[i].Secondary)
		mappingsOk := compareStrStrMap(a[i].Mappings, b[i].Mappings)

		if a[i].Index != b[i].Index ||
			a[i].Operation != b[i].Operation ||
			a[i].Object != b[i].Object ||
			a[i].Property != b[i].Property ||
			a[i].Parameter != b[i].Parameter ||
			a[i].Resource != b[i].Resource ||
			!secondaryOk ||
			!mappingsOk {
			return false
		}
	}

	return true
}

func compareServices(a models.Service, b models.Service) bool {

	labelsOk := compareStrings(a.Labels, b.Labels)

	return a.DescribedObject == b.DescribedObject &&
		a.Id == b.Id &&
		a.Name == b.Name &&
		a.LastConnected == b.LastConnected &&
		a.LastReported == b.LastReported &&
		a.OperatingState == b.OperatingState &&
		a.Addressable == b.Addressable &&
		labelsOk
}

func compareStrings(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func compareStrStrMap(a map[string]string, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, av := range a {
		if bv, ok := b[k]; !ok || av != bv {
			return false
		}
	}

	return true
}
