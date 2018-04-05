// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package cache

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/core/clients/coredataclients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/tonyespy/gxds"
	"gopkg.in/mgo.v2/bson"
)

// Profiles is a local cache of devices seeded from Core Metadata.
type Profiles struct {
	config      *gxds.Config
	profiles    map[string]models.Device
	vdc         coredataclients.ValueDescriptorClient
	descriptors []models.ValueDescriptor
	commands    map[string]map[string]map[string][]models.ResourceOperation
	objects     map[string]map[string]models.DeviceObject
	lc          logger.LoggingClient
}

var (
	pcOnce   sync.Once
	profiles *Profiles
)

// Create a singleton Profile cache instance. The cache
// actually stores copies of the objects contained within
// a device profile vs. the profiles themselves, although
// it can be used to update and existing profile.
func NewProfiles(config *gxds.Config, lc logger.LoggingClient) *Profiles {

	pcOnce.Do(func() {
		profiles = &Profiles{config: config, lc: lc}

		dataPort := strconv.Itoa(config.DataPort)
		profiles.vdc = coredataclients.NewValueDescriptorClient("http://" +
			config.DataHost + ":" + dataPort + "/api/v1/valuedescriptor")

		profiles.objects = make(map[string]map[string]models.DeviceObject)
		profiles.commands = make(map[string]map[string]map[string][]models.ResourceOperation)
	})

	return profiles
}

// CommandExists returns a bool indicating whether the specified command exists for the
// specified (by name) device. If the specified device doesn't exist, an error is returned.
func (p *Profiles) CommandExists(deviceName string, command string) (exists bool, err error) {
	devOps, ok := p.commands[deviceName]

	if !ok {
		err = fmt.Errorf("profiles: CommandExists: specified deviceName: %s not found", deviceName)
		return
	}

	if _, ok := devOps[command]; !ok {
		return
	}

	exists = true
	return
}

// TODO: this function is based on the original Java device-sdk-tools,
// and is too large & complicated; re-factor for simplicity, testability!
func (p *Profiles) addDevice(device *models.Device) error {
	p.lc.Debug(fmt.Sprintf("profiles: device: %s\n", device.Name))

	// map[resource name]map[get|set][]models.ResourceOperation
	var deviceOps = make(map[string]map[string][]models.ResourceOperation)

	// TODO: need to vet size & capacity for both...
	var ops = make([]models.ResourceOperation, 0, 1024)
	var usedDescriptors = make([]string, 0, 512)

	// TODO: this should be done once, and changes watched...
	// get current value descriptors from core-data
	// ignore err, zero-value slice returned by default
	descriptors, _ := p.vdc.ValueDescriptors()
	p.lc.Debug(fmt.Sprintf("profiles: %d valuedescriptors returned\n", len(descriptors)))
	p.lc.Debug(fmt.Sprintf("profiles: valuedescriptors: %v\n", descriptors))

	// TODO: if profile is not complete, update it
	if len(device.Profile.DeviceResources) == 0 {
		// try to find existing profile by name
		// set the profile into the device
		// recursive call:
		// call addDevice(device)
		// all done

		p.lc.Error(fmt.Sprintf("profiles: NO DeviceResources; failed state!!!\n"))
	}

	// ** Commands **

	// for each command in the profile, append the associated value
	// descriptors of each command to a single list of used descriptors
	vdNames := make(map[string]string)
	for _, command := range device.Profile.Commands {
		p.lc.Debug(fmt.Sprintf("profiles: command: %s\n", command.Name))
		command.AllAssociatedValueDescriptors(&vdNames)
	}

	for name, _ := range vdNames {
		usedDescriptors = append(usedDescriptors, name)

	}

	p.lc.Debug(fmt.Sprintf("profiles: usedDescriptors: %v\n", usedDescriptors))

	// ** Resources **

	for _, resource := range device.Profile.Resources {
		profileOps := make(map[string][]models.ResourceOperation)
		p.lc.Debug(fmt.Sprintf("\nprofiles: resource: %s\n", resource.Name))

		profileOps["get"] = resource.Get
		profileOps["set"] = resource.Set

		name := strings.ToLower(resource.Name)

		deviceOps[name] = profileOps
		p.lc.Debug(fmt.Sprintf("profiles: profileOps: %v\n\n", profileOps))

		// NOTE - Java uses ArrayList.addAll, which gets rid of duplicates!

		for _, ro := range resource.Get {

			// TODO: note, Resource.Index isn't being set to 1 here...
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=1],
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=null]
			p.lc.Debug(fmt.Sprintf("profiles: adding Get ro: %v to ops\n", ro))
			ops = append(ops, ro)
		}

		for _, ro := range resource.Set {

			// TODO: note, Resource.Index isn't being set to 1 here...
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=1],
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=null]
			p.lc.Debug(fmt.Sprintf("profiles: adding Set ro: %v to ops\n", ro))
			ops = append(ops, ro)
		}
	}

	p.lc.Debug(fmt.Sprintf("\n\nprofiles: ops: %v\n\n", ops))
	p.lc.Debug(fmt.Sprintf("\n\nprofiles: deviceOps: %v\n\n", deviceOps))

	// put the device's profile objects in the objects map
	// put the device's profile objects in the commands map if no resource exists
	// TODO: for now, just store DeviceObject; protocol code can extract the
	// attributes from DeviceObject.Attributes map directly
	//Map<String, ${Protocol name}Object> deviceObjects = new HashMap<>();
	deviceObjects := make(map[string]models.DeviceObject)

	p.lc.Debug(fmt.Sprintf("\nprofiles: start-->DeviceResources\n\n"))

	for _, object := range device.Profile.DeviceResources {
		value := object.Properties.Value
		p.lc.Debug(fmt.Sprintf("profiles: deviceobject: %v\n", object))

		deviceObjects[object.Name] = object

		// if there is no resource defined for an object, create one based on the
		// RW parameters

		name := strings.ToLower(object.Name)

		if _, ok := deviceOps[name]; !ok {
			readWrite := strings.ToLower(value.ReadWrite)

			p.lc.Debug(fmt.Sprintf("profiles: couldn't find %s in deviceOps; readWrite: %s\n", name, readWrite))
			operations := make(map[string][]models.ResourceOperation)

			if strings.Contains(readWrite, "r") {
				resource := &models.ResourceOperation{
					Index:     "1",
					Object:    object.Name,
					Operation: "get",
					Parameter: object.Name,
					Property:  "value",
					Secondary: []string{},
				}
				getOp := []models.ResourceOperation{*resource}
				key := strings.ToLower(resource.Operation)

				p.lc.Debug(fmt.Sprintf("profiles: created new get operation %s: %v\n", key, getOp))

				operations[key] = getOp
				ops = append(ops, *resource)
			}

			if strings.Contains(readWrite, "w") {
				resource := &models.ResourceOperation{
					Index:     "1",
					Object:    object.Name,
					Operation: "set",
					Parameter: object.Name,
					Property:  "value",
					Secondary: []string{},
				}

				setOp := []models.ResourceOperation{*resource}
				key := strings.ToLower(resource.Operation)

				p.lc.Debug(fmt.Sprintf("profiles: created new get operation %s: %v\n", key, setOp))

				operations[key] = setOp
				ops = append(ops, *resource)
			}

			// TODO: can a deviceresource have no operations?
			deviceOps[name] = operations
		}
	}

	p.lc.Debug(fmt.Sprintf("\nprofiles: done w/deviceresources\n\n"))
	p.lc.Debug(fmt.Sprintf("profiles: ops: %v\n\n", ops))
	p.lc.Debug(fmt.Sprintf("\n\nprofiles: deviceOps: %v\n\n", deviceOps))

	p.objects[device.Name] = deviceObjects
	p.commands[device.Name] = deviceOps

	// Create a value descriptor for each parameter using its underlying object
	for _, op := range ops {
		var desc *models.ValueDescriptor
		var object *models.DeviceObject

		p.lc.Debug(fmt.Sprintf("profiles: op: %v\n", op))

		// descriptors is []models.ValueDescriptor
		for _, v := range descriptors {
			p.lc.Debug(fmt.Sprintf("profiles: addDevice: op.Parameter: %s v.Name: %s\n", op.Parameter, v.Name))
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
				p.lc.Debug(fmt.Sprintf("ps: addDevice: op.Object: %s v.Name: %s\n", op.Object, v.Name))
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

func (p *Profiles) updateDevice(device *models.Device) {
	p.removeDevice(device)
	p.addDevice(device)
}

func (p *Profiles) removeDevice(device *models.Device) {
	delete(p.objects, device.Name)
	delete(p.commands, device.Name)
}

func (p *Profiles) createDescriptor(name string, object models.DeviceObject) *models.ValueDescriptor {
	value := object.Properties.Value
	units := object.Properties.Units

	p.lc.Debug(fmt.Sprintf("ps: createDescriptor: %v value: %v units: %s\n", name, value, units))

	descriptor := &models.ValueDescriptor{
		Name: name,
		Min:  value.Minimum,
		Max:  value.Maximum,
		// TODO: fix this --> IoTType.valueOf(value.getType().substring(0,1))
		Type:         "0",
		UomLabel:     units.DefaultValue,
		DefaultValue: value.DefaultValue,
		Formatting:   "%s",
		Description:  object.Description,
	}

	id, err := p.vdc.Add(descriptor)
	if err != nil {
		p.lc.Error(fmt.Sprintf("profiles: Add ValueDescriptor failed: %v\n", err))
		return nil
	}

	if !bson.IsObjectIdHex(id) {
		// TODO: should probably be an assertion?
		p.lc.Error(fmt.Sprintf("profiles: Add ValueDescriptor returned invalid Id: %s\n", id))
		return nil
	} else {
		descriptor.Id = bson.ObjectIdHex(id)
		p.lc.Debug(fmt.Sprintf("profiles: createDescriptor id: %s\n", id))
	}

	return descriptor
}

// UpdateProfile updates the specified device profile in
// the local cache, as well as updating all devices that
// in the local cache, and in Core Metadata with the
// updated profile.
func (p *Profiles) UpdateProfile(profileId string) bool {
	return true
}
