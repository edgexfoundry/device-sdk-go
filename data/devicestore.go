// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
// TODO: one file in each package should contain a doc comment for the package.
// Alternatively, this doc can go in a file called doc.go.
package data

import (
	"fmt"
	"os"
	"sync"
	"time"

	"bitbucket.org/clientcto/go-core-clients/metadataclients"
	"bitbucket.org/clientcto/go-core-domain/models"
	"bitbucket.org/tonyespy/gxds"
)

type DeviceStore struct {
	proto    *gxds.ProtocolHandler
	devices  map[string]models.Device
	profiles *profileStore
	ac       metadataclients.AddressableClient
	dc       metadataclients.DeviceClient
}

var (
	dsOnce sync.Once
	deviceStore *DeviceStore
)

// TODO: used by Init() to populate the local cache
// with devices pre-existing in metadata service, and
// also by ScanList, to add newly detected devices.
func (ds *DeviceStore) Add(device models.Device) error {

	// if device already exists in devices, delete & re-add
	if _, ok := ds.devices[device.Name]; ok {
		delete(ds.devices, device.Name)

		// TODO: remove from profiles
	}

	fmt.Fprintf(os.Stdout, "Adding managed device: : %v\n", device)

	err := ds.addDeviceToMetadata(device)
	if err != nil {
		return err
	}

	// This is only the case for brand new devices
	if device.OperatingState == models.OperatingState("enabled") {
		fmt.Fprintf(os.Stdout, "Initializing device: : %v\n", device)
		// ${Protocol name}.initializeDevice(metaDevice);
	}

	return nil
}

// TODO: revisit the use case for this function.  Currently
// it's used by updatehandler to add a device with a known
// Id, which was added to metadata by an external service
// while the deviceservice is running.
func (ds *DeviceStore) AddById(deviceId string) error {
	return nil
}

func (ds *DeviceStore) GetDevice(deviceName string) *models.Device {
	return nil
}

func (ds *DeviceStore) GetDeviceById(deviceId string) *models.Device {
	return nil
}

// TODO: based on the java code; we need to check how this function
// is used, as it's bad form to return an internal data struct to
// callers, especially when the result is a map, which can then be
// modified externally to this package.
func (ds *DeviceStore) GetDevices() map[string]models.Device {
	return ds.devices
}

func (ds *DeviceStore) GetMetaDevice(deviceName string) *models.Device {
	return nil
}

func (ds *DeviceStore) GetMetaDeviceById(deviceId string) *models.Device {
	return nil
}

func (ds *DeviceStore) GetMetaDevices() []models.Device {
	return []models.Device{}
}

func (ds *DeviceStore) Init(serviceId string) error {
	ds.ac = metadataclients.NewAddressableClient()
	ds.dc = metadataclients.NewDeviceClient()

	metaDevices, err := ds.dc.DevicesForService(serviceId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DevicesForService error: %v\n", err)
		return err
	}

	fmt.Fprintf(os.Stderr, "returned devices %v\n", metaDevices)

	ds.devices = make(map[string]models.Device)

	// TODO: initialize watchers
	// TODO: call Protocol.initialize

	ds.profiles, err = newProfileStore()
	ds.profiles.Init()

	// TODO: consider removing this logic, as the use case for
	// re-adding devices that exist in the core-metadata service
	// isn't clear (crash recovery)?
	for _, device := range metaDevices {
		err = ds.dc.UpdateOpState(device.Id.Hex(), "disabled")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Update metadata DeviceOpState failed: %s; error: %v",
				device.Name, err)
		}

		device.OperatingState = models.OperatingState("disabled")
		ds.Add(device)
	}

	return err
}

func (ds *DeviceStore) IsDeviceLocked(deviceId string) bool {
	return false
}

// New DeviceStore
func NewDeviceStore(proto *gxds.ProtocolHandler) (*DeviceStore) {

	// config.once.Do(func() { config.init(filename) })
	dsOnce.Do(func() {
		deviceStore = &DeviceStore{proto: proto}
	})

	return deviceStore
}

func (ds *DeviceStore) Remove(device models.Device) error {
	// remove(device):
	//  - if devices(device):
	//    - remove from map
	//    - call protocol.disconnect
	//    - dc.updateOpState(deviceId, OperatingState.disabled)
	//    - profiles.remove

	return nil
}

func (ds *DeviceStore) RemoveById(deviceId string) error {
	return nil
}

func (ds *DeviceStore) SetDeviceOpState(deviceName string, state models.OperatingState) error {
	return nil
}

func (ds *DeviceStore) SetDeviceByIdOpState(deviceId string, state models.OperatingState) error {
	return nil
}

func (ds *DeviceStore) Update(deviceId string) error {
	return nil
}

func (ds *DeviceStore) UpdateProfile(profileId string) error {
	return nil
}

// TODO: this should probably be broken into two separate
// functions, one which validates an existing device and
// adds it to the local cache, and one that adds a brand
// new device.

func (ds *DeviceStore) addDeviceToMetadata(device models.Device) error {
	// TODO: fix metadataclients to indicate !found, vs. returned zeroed struct!
	fmt.Fprintf(os.Stderr, "Trying to find addressable for: %s\n", device.Addressable.Name)
	addr, err := ds.ac.AddressableForName(device.Addressable.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AddressClient.AddressableForName: %s; failed: %v\n", device.Addressable.Name, err)

		// If device exists in metadata, and lacks an Addressable, don't try to fix; skip instead
		if device.Id.Valid() {
			return fmt.Errorf("Existing metadata device has no addressable: %s", device.Addressable.Name)
		}
	}

	// TODO: this is the best test for not-found for now...
	if addr.Name != device.Addressable.Name {
		addr = device.Addressable
		addr.BaseObject.Origin = time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)
		fmt.Fprintf(os.Stdout, "Creating new Addressable Object with name: %v", addr)

		err := ds.ac.Add(&addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "AddressClient.Add: %s; failed: %v\n", device.Addressable.Name, err)
			return err
		}
	}

	// A device without a valid Id is new
	if device.Id.Valid() == false {
		fmt.Fprintf(os.Stdout, "Trying to find device for: %s\n", device.Name)
		metaDevice, err := ds.dc.DeviceForName(device.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "DeviceClient.DeviceForName: %s; failed: %v\n", device.Name, err)
		}

		// TODO: this is the best test for not-found for now...
		if metaDevice.Name != device.Name {
			fmt.Fprintf(os.Stdout, "Adding Device to Metadata: %s\n", device.Name)

			err := ds.dc.Add(&device)
			if err != nil {
				fmt.Fprintf(os.Stderr, "DeviceClient.Add for %s failed: %v", device.Name, err)
				return err
			}
		} else {
			device.Id = metaDevice.Id

			if device.OperatingState != metaDevice.OperatingState {
				err := ds.dc.UpdateOpState(device.Id.Hex(), string(device.OperatingState))
				if err != nil {
					fmt.Fprintf(os.Stderr, "DeviceClient.UpdateOpState: %s; failed: %v\n", device.Name, err)
				}
			}
			// TODO: Java service doesn't check result, if UpdateOpState fails,
			// should device add fail too?
		}
	}

	// profiles.addDevice(device)
	ds.devices[device.Name] = device

	return nil
}

// FIXME: !threadsafe - none of the compare methods are threadsafe
// as other code can access the struct instances and potentially
// modify them while they're being compared.
func compareCommands(a []models.Command, b []models.Command) bool {
	if len(a) != len(b) {
		return false
	}

	for i, _ := range a {
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

	for i, _ := range a {
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

	for i, _ := range a {
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

	for i, _ := range a {
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

	for i, _ := range a {
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
