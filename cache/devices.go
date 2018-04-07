// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package provides management of device service related
// objects that may be distributed across one or more EdgeX
// core microservices.
//
package cache

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/core/clients/metadataclients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/tonyespy/gxds"
	"gopkg.in/mgo.v2/bson"
)

// Devices is a local cache of devices seeded from Core Metadata.
// TODO: review Go review comments, to see if this (and other code
// in this package) should use singular names (e.g. Device).
type Devices struct {
	config  *gxds.Config
	proto   gxds.ProtocolDriver
	devices map[string]*models.Device
	names   map[string]string
	ac      metadataclients.AddressableClient
	dc      metadataclients.DeviceClient
	lc      logger.LoggingClient
}

var (
	dcOnce  sync.Once
	devices *Devices
)

// Creates a singleton Devices cache instance.
func NewDevices(c *gxds.Config, lc logger.LoggingClient, proto gxds.ProtocolDriver) *Devices {

	dcOnce.Do(func() {
		devices = &Devices{config: c, proto: proto, lc: lc}
	})

	return devices
}

// Add a new device to the cache. This method is used to populate the
// devices cache with pre-existing devices from Core Metadata, as well
// as create new devices returned in a ScanList during discovery.
func (d *Devices) Add(dev *models.Device) error {

	// if device already exists in devices, delete & re-add
	if _, ok := d.devices[dev.Name]; ok {
		delete(d.devices, dev.Name)

		// TODO: remove from profiles
	}

	d.lc.Debug(fmt.Sprintf("Adding managed device: : %v\n", dev))

	// TODO: per effective go, should these two stmts be collapsed?
	// check if this is commonly used in Go src & snapd.
	err := d.addDeviceToMetadata(dev)
	if err != nil {
		return err
	}

	// This is only the case for brand new devices
	if dev.OperatingState == models.OperatingState("ENABLED") {
		d.lc.Debug(fmt.Sprintf("Initializing device: : %v\n", dev))
		// TODO: ${Protocol name}.initializeDevice(metaDevice);
	}

	return nil
}

// AddById adds a new device to the cache by id. This
// method is used by the UpdateHandler to trigger addition of a
// device that's been added directly to Core Metadata.
func (d *Devices) AddById(id string) error {
	return nil
}

// Device returns a device with the given name.
func (d *Devices) Device(name string) *models.Device {
	return nil
}

// DeviceById returns a device with the given device id.
func (d *Devices) DeviceById(id string) *models.Device {
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
func (d *Devices) Devices() map[string]*models.Device {
	return d.devices
}

// Init initializes the device cache.
func (d *Devices) Init(serviceId string) error {

	metaPort := strconv.Itoa(d.config.MetadataPort)
	d.ac = metadataclients.NewAddressableClient("http://" + d.config.MetadataHost +
		":" + metaPort + "/api/v1/addressable")

	d.dc = metadataclients.NewDeviceClient("http://" + d.config.MetadataHost +
		":" + metaPort + "/api/v1/device")

	mDevs, err := d.dc.DevicesForService(serviceId)
	if err != nil {
		d.lc.Error(fmt.Sprintf("DevicesForService error: %v\n", err))
		return err
	}

	d.lc.Debug(fmt.Sprintf("returned devices %v\n", mDevs))

	d.devices = make(map[string]*models.Device)
	d.names = make(map[string]string)

	// TODO: initialize watchers.initialize

	// TODO: consider removing this logic, as the use case for
	// re-adding devices that exist in the core-metadata service
	// isn't clear (crash recovery)?
	for _, md := range mDevs {
		err = d.dc.UpdateOpState(md.Id.Hex(), "DISABLED")
		if err != nil {
			d.lc.Error(fmt.Sprintf("Update metadata DeviceOpState failed: %s; error: %v", md.Name, err))
		}

		md.OperatingState = models.OperatingState("DISABLED")
		d.Add(&md)
	}

	// TODO: call Protocol.initialize
	d.lc.Debug(fmt.Sprintf("dstore: INITIALIZATION DONE! err=%v\n", err))

	return err
}

// IsDeviceLocked returns a bool which indicates if the specified
// device exists, and a book which indicates whether the device is locked
func (d *Devices) IsDeviceLocked(id string) (exists, locked bool) {
	return false, false
}

// RemoveById removes the specified (by id) device from the cache.
func (d *Devices) RemoveById(id string) error {
	return nil
}

// SetDeviceOpState sets the operatingState of the device specified by name.
func (d *Devices) SetDeviceOpState(name string, os models.OperatingState) error {
	return nil
}

// SetDeviceByIdOpState sets the operatingState of the device specified by name.
func (d *Devices) SetDeviceByIdOpState(id string, os models.OperatingState) error {
	return nil
}

// Update updates the device in the cache and ensures that the
// copy in Core Metadata is also updated.
func (d *Devices) Update(id string) error {
	return nil
}

// TODO: this should method should  be broken into two separate
// functions, one which validates an existing device and adds
// it to the local cache, and one that adds a brand new device.
// The current method is an almost direct translation of the Java
// DeviceStore implementation.
func (d *Devices) addDeviceToMetadata(dev *models.Device) error {
	// TODO: fix metadataclients to indicate !found, vs. returned zeroed struct!
	d.lc.Debug(fmt.Sprintf("Trying to find addressable for: %s\n", dev.Addressable.Name))
	addr, err := d.ac.AddressableForName(dev.Addressable.Name)
	if err != nil {
		d.lc.Error(fmt.Sprintf("AddressClient.AddressableForName: %s; failed: %v\n", dev.Addressable.Name, err))

		// If device exists in metadata, and lacks an Addressable, don't try to fix; skip instead
		if dev.Id.Valid() {
			return fmt.Errorf("Existing metadata dev has no addressable: %s", dev.Addressable.Name)
		}
	}

	// TODO: this is the best test for not-found for now...
	if addr.Name != dev.Addressable.Name {
		addr = dev.Addressable
		addr.BaseObject.Origin = time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)
		d.lc.Debug(fmt.Sprintf("Creating new Addressable Object with name: %v", addr))

		id, err := d.ac.Add(&addr)
		if err != nil {
			d.lc.Error(fmt.Sprintf("AddressClient.Add: %s; failed: %v\n", dev.Addressable.Name, err))
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
			d.lc.Debug(fmt.Sprintf("New addressable Id: %s\n", addr.Id.Hex()))
		}
	}

	// A device without a valid Id is new
	if dev.Id.Valid() == false {
		d.lc.Debug(fmt.Sprintf("Trying to find device for: %s\n", dev.Name))
		mDev, err := d.dc.DeviceForName(dev.Name)
		if err != nil {
			d.lc.Error(fmt.Sprintf("DeviceClient.DeviceForName: %s; failed: %v\n", dev.Name, err))
		}

		// TODO: this is the best test for not-found for now...
		if mDev.Name != dev.Name {
			d.lc.Debug(fmt.Sprintf("Adding Device to Metadata: %s\n", dev.Name))

			id, err := d.dc.Add(dev)
			if err != nil {
				d.lc.Error(fmt.Sprintf("DeviceClient.Add for %s failed: %v", dev.Name, err))
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
				d.lc.Debug(fmt.Sprintf("New dev id: %s\n", dev.Id.Hex()))
			}
		} else {
			dev.Id = mDev.Id

			if dev.OperatingState != mDev.OperatingState {
				err := d.dc.UpdateOpState(dev.Id.Hex(), string(dev.OperatingState))
				if err != nil {
					d.lc.Error(fmt.Sprintf("DeviceClient.UpdateOpState: %s; failed: %v\n", dev.Name, err))
				}
			}
			// TODO: Java service doesn't check result, if UpdateOpState fails,
			// should device add fail too?
		}
	}

	err = profiles.addDevice(dev)
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

	// TODO: Objects fields aren't compared

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
		attributesOk := compareStrStrMap(a[i].Attributes, b[i].Attributes)

		if a[i].Description != b[i].Description ||
			a[i].Name != b[i].Name ||
			a[i].Tag != b[i].Tag ||
			a[i].Properties != b[i].Properties &&
				!attributesOk {
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
