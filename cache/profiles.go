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

	"github.com/edgexfoundry/edgex-go/core/clients/coredata"
	"github.com/edgexfoundry/edgex-go/core/clients/types"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/tonyespy/gxds"
	"gopkg.in/mgo.v2/bson"
)

const (
	colon      = ":"
	httpScheme = "http://"
	v1Valuedescriptor = "/api/v1/valuedescriptor"
)

// Profiles is a local cache of devices seeded from Core Metadata.
type Profiles struct {
	config *gxds.Config
	vdc    coredata.ValueDescriptorClient
	// TODO: descriptors should be a map of vds.name to vds!!!
	descriptors []models.ValueDescriptor
	commands    map[string]map[string]map[string][]models.ResourceOperation
	objects     map[string]map[string]models.DeviceObject // TODO: make *models.DeviceObject?
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
func NewProfiles(c *gxds.Config, lc logger.LoggingClient, useRegistry bool) *Profiles {

	pcOnce.Do(func() {
		profiles = &Profiles{config: c, lc: lc}

		// TODO: move all client init code into service
		dataHost := c.Clients[gxds.ClientData].Host
		dataPort := strconv.Itoa(c.Clients[gxds.ClientData].Port)
		dataAddr := httpScheme + dataHost + colon + dataPort
		dataPath := v1Valuedescriptor
		dataURL := dataAddr + dataPath

		params := types.EndpointParams{
			// TODO: Can't use edgex-go internal constants!
			//ServiceKey:internal.CoreDataServiceKey,
			ServiceKey:  "edgex-core-data",
			Path:        dataPath,
			UseRegistry: useRegistry,
			Url:         dataURL}

		profiles.vdc = coredata.NewValueDescriptorClient(params, types.Endpoint{})

		profiles.objects = make(map[string]map[string]models.DeviceObject)
		profiles.commands = make(map[string]map[string]map[string][]models.ResourceOperation)
	})

	return profiles
}

func (p *Profiles) descriptorExists(name string) bool {
	var exists bool

	// TODO: make this a map?
	for _, desc := range p.descriptors {
		if desc.Name == name {
			exists = true
			break
		}
	}

	return exists
}

// GetDeviceObjects returns a map of object names to DeviceObject instances.
func (p *Profiles) GetDeviceObjects(devName string) map[string]models.DeviceObject {
	devObjs := p.objects[devName]
	return devObjs
}

// CommandExists returns a bool indicating whether the specified command exists for the
// specified (by name) device. If the specified device doesn't exist, an error is returned.
// Note - this command currently checks that a deviceprofile *resource* with the given name
// exists, it's not actually checking that a deviceprofile *command* with this name exists.
// See addDevice() for more details.
func (p *Profiles) CommandExists(devName string, cmd string) (bool, error) {
	devOps, ok := p.commands[devName]
	if !ok {
		err := fmt.Errorf("profiles: CommandExists: specified dev: %s not found", devName)
		return false, err
	}

	if _, ok := devOps[strings.ToLower(cmd)]; !ok {
		return false, nil
	}

	return true, nil
}

// GetResourceOperation...
func (p *Profiles) GetResourceOperations(devName string, cmd string, method string) ([]models.ResourceOperation, error) {
	var err error

	devOps, ok := p.commands[devName]
	if !ok {
		err = fmt.Errorf("profiles: GetResourceOperations: specified dev: %s not found", devName)
		return nil, err
	}

	cmdOps, ok := devOps[strings.ToLower(cmd)]
	if !ok {
		err = fmt.Errorf("profiles: GetResourceOperations: specified cmd: %s not found", cmd)
		return nil, err
	}

	resOps, ok := cmdOps[strings.ToLower(method)]
	if !ok {
		err = fmt.Errorf("profiles: GetResourceOperations: specified cmd method: %s not found", method)
		return nil, err
	}

	return resOps, nil
}

// TODO: this function is based on the original Java device-sdk-tools,
// and is too large & complicated; re-factor for simplicity, testability!
func (p *Profiles) addDevice(d *models.Device) error {
	p.lc.Debug(fmt.Sprintf("profiles: dev: %s\n", d.Name))

	var devOps = make(map[string]map[string][]models.ResourceOperation)

	// TODO: need to vet size & capacity for both...
	var ops = make([]models.ResourceOperation, 0, 1024)
	var usedDescs = make([]string, 0, 512)

	// TODO: this should be done once, and changes watched...
	// get current value descriptors from core-data
	// ignore err, zero-value slice returned by default
	descs, _ := p.vdc.ValueDescriptors()
	p.lc.Debug(fmt.Sprintf("profiles: valuedescriptors: %v\n", descs))

	// TODO: deviceprofiles with no device resources aren't supported, unlike
	// the Java SDK-based DSs.
	if len(d.Profile.DeviceResources) == 0 {
		// try to find existing profile by name
		// set the profile into the device
		// recursive call:
		// call addDevice(dev)

		err := fmt.Errorf("profiles: dev: %s has no device resources", d.Name)
		return err
	}

	// ** Commands **

	// for each command in the profile, append the associated value
	// descriptors of each command to a single list of used descriptors
	vdNames := make(map[string]string)
	for _, cmd := range d.Profile.Commands {
		p.lc.Debug(fmt.Sprintf("profiles: cmd: %s\n", cmd.Name))
		cmd.AllAssociatedValueDescriptors(&vdNames)
	}

	for name, _ := range vdNames {
		usedDescs = append(usedDescs, name)

	}

	p.lc.Debug(fmt.Sprintf("profiles: usedDescriptors: %v\n", usedDescs))

	// ** Resources **

	for _, r := range d.Profile.Resources {
		profOps := make(map[string][]models.ResourceOperation)
		p.lc.Debug(fmt.Sprintf("\nprofiles: resource: %s\n", r.Name))

		profOps["get"] = r.Get
		profOps["set"] = r.Set

		name := strings.ToLower(r.Name)

		devOps[name] = profOps
		p.lc.Debug(fmt.Sprintf("profiles: profOps: %v\n\n", profOps))

		// NOTE - Java uses ArrayList.addAll, which gets rid of duplicates!

		for _, ro := range r.Get {

			// TODO: note, Resource.Index isn't being set to 1 here...
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=1],
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=null]
			p.lc.Debug(fmt.Sprintf("profiles: adding Get ro: %v to ops\n", ro))
			ops = append(ops, ro)
		}

		for _, ro := range r.Set {

			// TODO: note, Resource.Index isn't being set to 1 here...
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=1],
			//  [operation=get, object=HoldingRegister_8455, property=value, parameter=HoldingRegister_8455, mappings={}, index=null]
			p.lc.Debug(fmt.Sprintf("profiles: adding Set ro: %v to ops\n", ro))
			ops = append(ops, ro)
		}
	}

	p.lc.Debug(fmt.Sprintf("\n\nprofiles: ops: %v\n\n", ops))
	p.lc.Debug(fmt.Sprintf("\n\nprofiles: devOps: %v\n\n", devOps))

	// put the device's profile objects in the objects map
	// put the device's profile objects in the commands map if no resource exists
	// TODO: for now, just store DeviceObject; protocol code can extract the
	// attributes from DeviceObject.Attributes map directly
	devObjs := make(map[string]models.DeviceObject)

	p.lc.Debug(fmt.Sprintf("\nprofiles: start-->DeviceResources\n\n"))

	for _, dr := range d.Profile.DeviceResources {
		value := dr.Properties.Value
		p.lc.Debug(fmt.Sprintf("profiles: devobject: %v\n", dr))

		devObjs[dr.Name] = dr

		// if there is no resource defined for an object, create one based on the
		// RW parameters

		name := strings.ToLower(dr.Name)

		if _, ok := devOps[name]; !ok {
			rw := strings.ToLower(value.ReadWrite)

			p.lc.Debug(fmt.Sprintf("profiles: couldn't find %s in devOps; rw: %s\n", name, rw))
			resOps := make(map[string][]models.ResourceOperation)

			if strings.Contains(rw, "r") {
				res := &models.ResourceOperation{
					Index:     "1",
					Object:    dr.Name,
					Operation: "get",
					Parameter: dr.Name,
					Property:  "value",
					Secondary: []string{},
				}
				getOp := []models.ResourceOperation{*res}
				key := strings.ToLower(res.Operation)

				p.lc.Debug(fmt.Sprintf("profiles: created new get operation %s: %v\n", key, getOp))

				resOps[key] = getOp
				ops = append(ops, *res)
			}

			if strings.Contains(rw, "w") {
				res := &models.ResourceOperation{
					Index:     "1",
					Object:    dr.Name,
					Operation: "set",
					Parameter: dr.Name,
					Property:  "value",
					Secondary: []string{},
				}

				setOp := []models.ResourceOperation{*res}
				key := strings.ToLower(res.Operation)

				p.lc.Debug(fmt.Sprintf("profiles: created new get operation %s: %v\n", key, setOp))

				resOps[key] = setOp
				ops = append(ops, *res)
			}

			// TODO: can a deviceresource have no operations?
			devOps[name] = resOps
		}
	}

	p.lc.Debug(fmt.Sprintf("\nprofiles: done w/devresources\n\n"))
	p.lc.Debug(fmt.Sprintf("profiles: ops: %v\n\n", ops))
	p.lc.Debug(fmt.Sprintf("\n\nprofiles: deviceOps: %v\n\n", devOps))

	p.objects[d.Name] = devObjs
	p.commands[d.Name] = devOps

	// Create a value descriptor for each parameter using its underlying object
	for _, op := range ops {
		var desc *models.ValueDescriptor
		var devObj *models.DeviceObject

		p.lc.Debug(fmt.Sprintf("profiles: op: %v\n", op))

		// descs is []models.ValueDescriptor
		for _, v := range descs {
			p.lc.Debug(fmt.Sprintf("profiles: addDevice: op.Parameter: %s v.Name: %s\n", op.Parameter, v.Name))
			if op.Parameter == v.Name {
				desc = &v
				break
			}
		}

		if desc == nil {
			var found bool

			// TODO: ask Tyler or Jim why this check is needed...
			for _, used := range usedDescs {
				if op.Parameter == used {
					found = true
				}
			}

			if !found {
				continue
			}

			for _, dr := range d.Profile.DeviceResources {
				p.lc.Debug(fmt.Sprintf("ps: addDevice: op.Object: %s dr.Name: %s\n", op.Object, dr.Name))
				if op.Object == dr.Name {
					devObj = &dr
					break
				}
			}

			desc = p.createDescriptor(op.Parameter, *devObj)
			if desc == nil {
				// TODO: should the whole thing unwind due to this failure?

			}
		}

		p.descriptors = append(p.descriptors, *desc)
		descs = append(descs, *desc)
	}

	return nil
}

func (p *Profiles) updateDevice(d *models.Device) {
	p.removeDevice(d)
	p.addDevice(d)
}

func (p *Profiles) removeDevice(d *models.Device) {
	delete(p.objects, d.Name)
	delete(p.commands, d.Name)
}

func (p *Profiles) createDescriptor(name string, devObj models.DeviceObject) *models.ValueDescriptor {
	value := devObj.Properties.Value
	units := devObj.Properties.Units

	p.lc.Debug(fmt.Sprintf("ps: createDescriptor: %v value: %v units: %s\n", name, value, units))

	desc := &models.ValueDescriptor{
		Name: name,
		Min:  value.Minimum,
		Max:  value.Maximum,
		// TODO: fix this --> IoTType.valueOf(value.getType().substring(0,1))
		Type:         "0",
		UomLabel:     units.DefaultValue,
		DefaultValue: value.DefaultValue,
		Formatting:   "%s",
		Description:  devObj.Description,
	}

	id, err := p.vdc.Add(desc)
	if err != nil {
		p.lc.Error(fmt.Sprintf("profiles: Add ValueDescriptor failed: %v\n", err))
		return nil
	}

	if !bson.IsObjectIdHex(id) {
		// TODO: should probably be an assertion?
		p.lc.Error(fmt.Sprintf("profiles: Add ValueDescriptor returned invalid Id: %s\n", id))
		return nil
	} else {
		desc.Id = bson.ObjectIdHex(id)
		p.lc.Debug(fmt.Sprintf("profiles: createDescriptor id: %s\n", id))
	}

	return desc
}

// UpdateProfile updates the specified device profile in
// the local cache, as well as updating all devices that
// in the local cache, and in Core Metadata with the
// updated profile.
func (p *Profiles) UpdateProfile(id string) bool {
	return true
}
