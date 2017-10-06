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

	"bitbucket.org/clientcto/go-core-clients/metadataclients"
	"bitbucket.org/clientcto/go-core-domain/models"
)

type DeviceStore struct {
	initialized bool
	devices     map[string]models.Device
	ac          metadataclients.AddressableClient
	dc          metadataclients.DeviceClient
}

// TODO: should return error instead of boolean?
func (ds *DeviceStore) add(device models.Device) error {

	// if device already exists in devices, delete & re-add
	if _, ok := ds.devices[device.Name]; ok {
		delete(ds.devices, device.Name)

		// TODO: remove from profiles too...(why?)
	}

	fmt.Fprintf(os.Stdout, "Adding managed device: : %v\n", device)

	metaDevice := ds.addDeviceToMetadata(device)

	fmt.Fprintf(os.Stdout, "metaDevice: : %v\n", metaDevice)

	// TODO:
	// if (metaDevice == null) {
	//	remove(device);
	//	return false;
	// }

	// If opState == enabled, call Protocol.initializeDevice
	// TODO: can Protocol.initializeDevice fail???
	//if (metaDevice.getOperatingState().equals(OperatingState.enabled)) {
	//	${Protocol name}.initializeDevice(metaDevice);
	//}

	return nil
}

// TODO: should this return *models.Device (mimcs Java version)?
// is it really needed?

func (ds *DeviceStore) addDeviceToMetadata(device models.Device) *models.Device {
	// Create a new addressable Object with the devicename + last 6 digits of MAC address.
	// Assume this to be unique

	// check for addressable
	//
	// TODO: fix metadataclients to indicate !found, vs. returned zeroed struct!
	fmt.Fprintf(os.Stderr, "Trying to find addressable for: %s\n", d.config.Name)
	addr, err := ds.ac.AddressableForName(device.Addressable.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AddressableForName: %s; failed: %v\n", d.config.Name, err)
		return nil
	}

	// This is the best test for not-found for now...
	if addr.Name != device.Addressable.Name {
		// TODO: create new addressable

	}

	// TODO: lookup deviceForName
	// if found, set local device Id
	// if operating states of returned device don't match input device, update in metadata

	// if not found, create a new device and add to metadata

	// profiles.addDevice(device)
	// add device to devices Hashmap

	return device
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

	for _, device := range metaDevices {
		err = ds.dc.UpdateOpState(device.Id.Hex(), "disabled")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Update metadata DeviceOpState failed: %s; error: %v",
				device.Name, err)
		}

		ds.add(device)
	}

	// TODO: is this needed?
	ds.initialized = true

	return err
}

// New DeviceStore
// TODO: re-factor to make this a singleton
func NewDeviceStore() (*DeviceStore, error) {
	return &DeviceStore{}, nil
}
