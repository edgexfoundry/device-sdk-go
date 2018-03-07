// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
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
package cache

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
	pcOnce       sync.Once
	cache        *Profiles

	// TODO: grab settings from daemon-config.json OR Consul
	dataPort            string = ":48080"
	dataHost            string = "localhost"
	dataValueDescUrl    string = "http://" + dataHost + dataPort + "/api/v1/valuedescriptor"
)

type Profiles struct {
	profiles    map[string]models.Device
	vdc         coredataclients.ValueDescriptorClient
	descriptors []models.ValueDescriptor
        commands    map[string]map[string]map[string][]models.ResourceOperation
	objects     map[string]map[string]models.DeviceObject
}

// Create a singleton ProfileStore instance
func NewProfiles() *Profiles {

	pcOnce.Do(func() {
		cache = &Profiles{}
		cache.vdc = coredataclients.NewValueDescriptorClient(dataValueDescUrl)
		cache.objects = make(map[string]map[string]models.DeviceObject)
		cache.commands = make(map[string]map[string]map[string][]models.ResourceOperation)
	})

	return cache
}

// TODO: this function is based on the original Java device-sdk-tools,
// and is too large & complicated; re-factor for simplicity, testability!
func (p *Profiles) addDevice(device models.Device) error {
	fmt.Fprintf(os.Stdout, "pstore: device: %s\n", device.Name)

	// map[resource name]map[get|set][]models.ResourceOperation
	var deviceOps = make(map[string]map[string][]models.ResourceOperation)

	// TODO: need to vet size & capacity for both...
	var ops = make([]models.ResourceOperation, 0, 1024)
	var usedDescriptors = make([]string, 0, 512)

	// TODO: this should be done once, and changes watched...
	// get current value descriptors from core-data
	// ignore err, zero-value slice returned by default
	descriptors, _ := p.vdc.ValueDescriptors()
	fmt.Fprintf(os.Stderr, "pstore: %d valuedescriptors returned\n", len(descriptors))
	fmt.Fprintf(os.Stderr, "pstore: valuedescriptors: %v\n", descriptors)

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

	fmt.Fprintf(os.Stderr, "pstore: usedDescriptors: %v\n", usedDescriptors)

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

			// TODO: note, Resource.Index isn't being set to 1 here...
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=1],
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=null]
			fmt.Fprintf(os.Stderr, "pstore: adding Get ro: %v to ops\n", ro)
			ops = append(ops, ro)
		}

		for _, ro := range resource.Set {

			// TODO: note, Resource.Index isn't being set to 1 here...
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=1],
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=null]
			fmt.Fprintf(os.Stderr, "pstore: adding Set ro: %v to ops\n", ro)
			ops = append(ops, ro)
		}
	}

	fmt.Fprintf(os.Stderr, "\n\npstore: ops: %v\n\n", ops)
	fmt.Fprintf(os.Stderr, "\n\npstore: deviceOps: %v\n\n", deviceOps)

	// put the device's profile objects in the objects map
	// put the device's profile objects in the commands map if no resource exists
	// TODO: for now, just store DeviceObject; protocol code can extract the
	// attributes from DeviceObject.Attributes map directly
	//Map<String, ${Protocol name}Object> deviceObjects = new HashMap<>();
	deviceObjects := make(map[string]models.DeviceObject)

	fmt.Fprintf(os.Stderr, "\npstore: start-->DeviceResources\n\n")

	for _, object := range device.Profile.DeviceResources {
		value := object.Properties.Value
		fmt.Fprintf(os.Stderr, "pstore: deviceobject: %v\n", object)

		deviceObjects[object.Name] = object

		// if there is no resource defined for an object, create one based on the
		// RW parameters

		name := strings.ToLower(object.Name)

		if _, ok := deviceOps[name]; !ok {
			readWrite := strings.ToLower(value.ReadWrite)

			fmt.Fprintf(os.Stderr, "pstore: couldn't find %s in deviceOps; readWrite: %s\n", name, readWrite)
			operations := make(map[string][]models.ResourceOperation)

			if strings.Contains(readWrite, "r") {
				resource := &models.ResourceOperation{
					Index: "1",
					Object: object.Name,
					Operation: "get",
					Parameter: object.Name,
					Property: "value",
					Secondary: []string{},
				}
				getOp := []models.ResourceOperation{*resource}
				key := strings.ToLower(resource.Operation)

				fmt.Fprintf(os.Stderr, "pstore: created new get operation %s: %v\n", key, getOp)

				operations[key] = getOp
				ops = append(ops, *resource)
			}

			if strings.Contains(readWrite, "w") {
				resource := &models.ResourceOperation{
					Index: "1",
					Object: object.Name,
					Operation: "set",
					Parameter: object.Name,
					Property: "value",
					Secondary: []string{},
				}

				setOp := []models.ResourceOperation{*resource}
				key := strings.ToLower(resource.Operation)

				fmt.Fprintf(os.Stderr, "pstore: created new get operation %s: %v\n", key, setOp)

				operations[key] = setOp
				ops = append(ops, *resource)
			}

			// TODO: can a deviceresource have no operations?
			deviceOps[name] = operations
		}
	}

	fmt.Fprintf(os.Stdout, "\npstore: done w/deviceresources\n\n")
	fmt.Fprintf(os.Stdout, "pstore: ops: %v\n\n", ops)
	fmt.Fprintf(os.Stderr, "\n\npstore: deviceOps: %v\n\n", deviceOps)

	p.objects[device.Name] = deviceObjects
	p.commands[device.Name] = deviceOps

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
			var found bool

			// TODO: ask Tyler or Jim why this check is needed...
			for _, used := range usedDescriptors {
				if op.Parameter == used {
					found = true
				}
			}

			if !found {
				continue
			}

			for _, v := range device.Profile.DeviceResources {
				fmt.Fprintf(os.Stdout, "ps: addDevice: op.Object: %s v.Name: %s\n", op.Object, v.Name)
				if op.Object == v.Name {
					object = &v
					break
				}
			}

			desc = p.createDescriptor(op.Parameter, *object)
			if desc == nil {
				// TODO: should the whole thing unwind due to this failure?

			}
		}

		p.descriptors = append(p.descriptors, *desc)
		descriptors = append(descriptors, *desc)
	}

	return nil
}

func (p *Profiles) updateDevice(device models.Device) {
	p.removeDevice(device)
	p.addDevice(device)
}

func (p *Profiles) removeDevice(device models.Device) {
    delete(p.objects, device.Name)
    delete(p.commands, device.Name)
}

func (p *Profiles) createDescriptor(name string, object models.DeviceObject) *models.ValueDescriptor {
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

	id, err := p.vdc.Add(descriptor)
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
