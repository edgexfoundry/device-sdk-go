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
package data

import (
	"github.com/edgexfoundry/core-clients-go/coredataclients"
	"github.com/edgexfoundry/core-domain-go/models"
)

var (
	// TODO: grab settings from daemon-config.json OR Consul
	dataPort            string = ":48080"
	dataHost            string = "localhost"
	dataValueDescUrl    string = "http://" + dataHost + dataPort + "/api/v1/valuedescriptor"
)


type profileStore struct {
	profiles    map[string]models.Device
	vdc         coredataclients.ValueDescriptorClient
}

func (ps *profileStore) Init() {
	ps.vdc = coredataclients.NewValueDescriptorClient(dataValueDescUrl)
}

// TODO: re-factor to make this a singleton
func newProfileStore() (*profileStore, error) {
	return &profileStore{}, nil
}

// TODO: this function is based on the original Java device-sdk-tools,
// and is too large & complicated; re-factor for simplicity, testability!
func (ps *profileStore) addDevice(device models.Device) error {
        var vdNames = new(map[string]string)

	// use ValueDescriptionClient to grab a list of descriptors
	_, err := ps.vdc.ValueDescriptors()
	if err != nil {
		return err
	}

	if len(device.Profile.DeviceResources) == 0 {
		// try to find existing profile by name
		// set the profile into the device
		// recursive call:
		// call addDevice(device)
		// all done
	}

	// for each command in the profile, append the associated value
	// descriptors of each command to a single list of used descriptors
	for _, command := range device.Profile.Commands {
		command.AllAssociatedValueDescriptors(vdNames)
	}

	return nil
}
