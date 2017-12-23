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
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/edgexfoundry/core-clients-go/coredataclients"
	"github.com/edgexfoundry/core-domain-go/models"
	"gopkg.in/mgo.v2/bson"
)

var (
	psOnce      sync.Once
	profileStore *ProfileStore

	// TODO: grab settings from daemon-config.json OR Consul
	dataPort            string = ":48080"
	dataHost            string = "localhost"
	dataValueDescUrl    string = "http://" + dataHost + dataPort + "/api/v1/valuedescriptor"
)

type ProfileStore struct {
	profiles    map[string]models.Device
	vdc         coredataclients.ValueDescriptorClient
	descriptors []models.ValueDescriptor
        commands    map[string]map[string]map[string][]models.ResourceOperation
	objects     map[string]map[string]models.DeviceObject
}

// Create a singleton ProfileStore instance
func NewProfileStore() *ProfileStore {

	psOnce.Do(func() {
		profileStore = &ProfileStore{}
		profileStore.vdc = coredataclients.NewValueDescriptorClient(dataValueDescUrl)
		profileStore.objects = make(map[string]map[string]models.DeviceObject)
		profileStore.commands = make(map[string]map[string]map[string][]models.ResourceOperation)
	})

	return profileStore
}

// TODO: this function is based on the original Java device-sdk-tools,
// and is too large & complicated; re-factor for simplicity, testability!
func (ps *ProfileStore) addDevice(device models.Device) error {
	fmt.Fprintf(os.Stdout, "pstore: device: %s\n", device.Name)

	var deviceOps = make(map[string]map[string][]models.ResourceOperation)

	// TODO: need to vet size & capacity for both...
	var ops = make([]models.ResourceOperation, 0, 1024)
	var usedDescriptors = make([]string, 0, 512)

	// TODO: this should be done once, and changes watched...
	// ignore err, zero-value slice returned by default
	descriptors, _ := ps.vdc.ValueDescriptors()
	fmt.Fprintf(os.Stderr, "pstore: %d valuedescriptors returned\n", len(descriptors))

	// TODO: if profile is not complete, update it
	if len(device.Profile.DeviceResources) == 0 {
		// try to find existing profile by name
		// set the profile into the device
		// recursive call:
		// call addDevice(device)
		// all done

		fmt.Fprintf(os.Stderr, "pstore: NO DeviceResources; failed state!!!\n")
	}

	// ** Commands **

	// for each command in the profile, append the associated value
	// descriptors of each command to a single list of used descriptors
	vdNames := make(map[string]string)
	for _, command := range device.Profile.Commands {
		fmt.Fprintf(os.Stderr, "pstore: command: %s\n", command.Name)
		command.AllAssociatedValueDescriptors(&vdNames)
	}

	for name, _ := range vdNames {
		usedDescriptors = append(usedDescriptors, name)
	}

	//fmt.Fprintf(os.Stderr, "pstore: usedDescriptors: %v\n", usedDescriptors)

	// ** Resources **

	for _, resource := range device.Profile.Resources {
		profileOps := make(map[string][]models.ResourceOperation)
		fmt.Fprintf(os.Stderr, "\npstore: resource: %s\n", resource.Name)

		profileOps["get"] = resource.Get
		profileOps["set"] = resource.Set

		name := strings.ToLower(resource.Name)

		deviceOps[name] = profileOps
		fmt.Fprintf(os.Stderr, "pstore: profileOps: %v\n\n", profileOps)

		// NOTE - Java uses ArrayList.addAll, which gets rid of duplicates!

		for _, ro := range resource.Get {
			fmt.Fprintf(os.Stderr, "pstore: adding Get ro: %v to ops\n", ro)
			ops = append(ops, ro)
		}

		for _, ro := range resource.Set {
			fmt.Fprintf(os.Stderr, "pstore: adding Set ro: %v to ops\n", ro)
			ops = append(ops, ro)
		}
	}

	//fmt.Fprintf(os.Stderr, "\n\npstore: ops: %v\n\n", ops)

	// put the device's profile objects in the objects map
	// put the device's profile objects in the commands map if no resource exists
	// TODO: for now, just store DeviceObject; protocol code can extract the
	// attributes from DeviceObject.Attributes map directly
	//Map<String, ${Protocol name}Object> deviceObjects = new HashMap<>();
	deviceObjects := make(map[string]models.DeviceObject)

	fmt.Fprintf(os.Stderr, "pstore: start-->DeviceResources\n")

	// DeviceResources are not being handled correctly, entries in ops
	// should be like this:
        //  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=1],
	//  [operation=get, object=HoldingRegister_8454, property=value, parameter=HoldingRegister_8454, mappings={}, index=1],
	//  [operation=get, object=HoldingRegister_2331, property=value, parameter=HoldingRegister_2331, mappings={}, index=1],
	//  [operation=set, object=enableRandomization, property=value, parameter=enableRandomization, mappings={}, index=1],
        //  [operation=set, object=collectionFrequency, property=value, parameter=collectionFrequency, mappings={}, index=1]]
	//
	// but instead looks like this:
	//
        //  ["index":null,"operation":"get","object":"HoldingRegister_8455","property":null,"parameter":null,"resource":null,"secondary":null,"mappings":null]
        //  ["index":null,"operation":"get","object":"HoldingRegister_8454","property":null,"parameter":null,"resource":null,"secondary":null,"mappings":null]
        //  ["index":null,"operation":"get","object":"HoldingRegister_2331","property":null,"parameter":null,"resource":null,"secondary":null,"mappings":null]
	//  ["index":null,"operation":"set","object":"enableRandomization","property":null,"parameter":null,"resource":null,"secondary":null,"mappings":null]
	//  ["index":null,"operation":"set","object":"collectionFrequency","property":null,"parameter":null,"resource":null,"secondary":null,"mappings":null]
	//
	// NOTE - parameter for each is "null"
	//
	// also note, that for DeviceResources at op(#1) have an empty slice for secondary, whereas these above are null:
	//
	// ["index":null,"operation":"set","object":"collectionFrequency","property":"value","parameter":"collectionFrequency","resource":null,"secondary":[],
	//  "mappings":{}]

	for _, object := range device.Profile.DeviceResources {
		value := object.Properties
		fmt.Fprintf(os.Stderr, "pstore: deviceobject: %v\n", object)

		deviceObjects[object.Name] = object

		// if there is no resource defined for an object, create one based on the
		// RW parameters

		name := strings.ToLower(object.Name)

		if _, ok := deviceOps[name]; !ok {
			readWrite := strings.ToLower(value.Value.ReadWrite)

			fmt.Fprintf(os.Stderr, "pstore: couldn't find %s in deviceOps; readWrite: %s\n", name, readWrite)
			operations := make(map[string][]models.ResourceOperation)

			if strings.Contains(readWrite, "r") {
				resource := &models.ResourceOperation{Operation: "get", Object: object.Name}
				getOp := []models.ResourceOperation{*resource}
				key := strings.ToLower(resource.Operation)

				operations[key] = getOp
				ops = append(ops, *resource)
			}

			if strings.Contains(readWrite, "w") {
				resource := &models.ResourceOperation{Operation: "set", Object: object.Name}
				setOp := []models.ResourceOperation{*resource}
				key := strings.ToLower(resource.Operation)

				operations[key] = setOp
				ops = append(ops, *resource)
			}

			deviceOps[name] = operations
		}
	}

	fmt.Fprintf(os.Stdout, "pstore: done w/deviceresources\n")
	fmt.Fprintf(os.Stdout, "pstore: ops: %v\n", ops)

	ps.objects[device.Name] = deviceObjects
	ps.commands[device.Name] = deviceOps

	// Create a value descriptor for each parameter using its underlying object
//	for (ResourceOperation op: ops) {
//		ValueDescriptor descriptor = descriptors.stream().filter(d -> d.getName()
//			.equals(op.getParameter())).findAny().orElse(null);

//		if (descriptor == null) {
//			if (!usedDescriptors.contains(op.getParameter())) {
//				continue;
//			}

	// Create a value descriptor for each parameter using its underlying object
	for _, op := range ops {
		var desc *models.ValueDescriptor
		var object *models.DeviceObject

		fmt.Fprintf(os.Stdout, "pstore: op: %v\n", op)

		// descriptors is []models.ValueDescriptor
		for _, v := range descriptors {
			fmt.Fprintf(os.Stdout, "pstore: addDevice: op.Parameter: %s v.Name: %s\n", op.Parameter, v.Name)
			if op.Parameter == v.Name {
				desc = &v
				break
			}
		}

		if desc == nil {
			// TODO: fix this
			//if (!usedDescriptors.contains(op.getParameter())) {
			//continue;
			for _, v := range device.Profile.DeviceResources {
				fmt.Fprintf(os.Stdout, "ps: addDevice: op.Object: %s v.Name: %s\n", op.Object, v.Name)
				if op.Object == v.Name {
					object = &v
					break
				}
			}

			desc = ps.createDescriptor(op.Parameter, *object)
			if desc == nil {
				// TODO: should the whole thing unwind due to this failure?

			}
		}

		ps.descriptors = append(ps.descriptors, *desc)
		descriptors = append(descriptors, *desc)
	}

	return nil
}

func (ps *ProfileStore) updateDevice(device models.Device) {
	ps.removeDevice(device)
	ps.addDevice(device)
}

func (ps *ProfileStore) removeDevice(device models.Device) {
    delete(ps.objects, device.Name)
    delete(ps.commands, device.Name)
}

func (ps *ProfileStore) createDescriptor(name string, object models.DeviceObject) *models.ValueDescriptor {
	value := object.Properties.Value
	units := object.Properties.Units

	fmt.Fprintf(os.Stdout, "ps: createDescriptor: %v value: %v units: %s\n", name, value, units)

	descriptor := &models.ValueDescriptor{
		Name: name,
		Min: value.Minimum,
		Max: value.Maximum,
		// TODO: fix this --> IoTType.valueOf(value.getType().substring(0,1))
		Type: "0",
		UomLabel: units.DefaultValue,
		DefaultValue: value.DefaultValue,
		Formatting: "%s",
		Description: object.Description,
	}

	id, err := ps.vdc.Add(descriptor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Add ValueDescriptor failed: %v\n", err)
		return nil
	}

	if !bson.IsObjectIdHex(id) {
		// TODO: should probably be an assertion?
		fmt.Fprintf(os.Stderr, "Add ValueDescriptor returned invalid Id: %s\n", id)
		return nil
	} else {
		descriptor.Id = bson.ObjectIdHex(id)
		fmt.Fprintf(os.Stdout, "createDescriptor id: %s\n", id)
	}

	return descriptor
}
