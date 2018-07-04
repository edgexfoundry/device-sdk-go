// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package service

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

var (
	ocOnce  sync.Once
	oc      *objectCache
)

type objectCache struct {
	objects       map[string]map[string][]string
	responses     map[string]map[string][]*models.Reading
	cacheSize     int
	transformData bool
	mu            sync.RWMutex
}

func newObjectCache() {

	ocOnce.Do(func() {
		oc = &objectCache{}

		oc.objects = make(map[string]map[string][]string)
		oc.responses = make(map[string]map[string][]*models.Reading)
	})
}

// GetDeviceObject...
func (o *objectCache) GetDeviceObject(d *models.Device, op *models.ResourceOperation) *models.DeviceObject {
	var devObj models.DeviceObject
	devObjs := pc.GetDeviceObjects(d.Name)

	if op != nil && devObjs != nil {
		devObj, ok := devObjs[op.Object]

		if !ok {
			return nil
		}

		if pc.descriptorExists(op.Parameter) {
			devObj.Name = op.Parameter
		}
	}

	return &devObj
}

func (o *objectCache) createObjectList(d *models.Device, op *models.ResourceOperation) []models.DeviceObject {
	devObjs := pc.GetDeviceObjects(d.Name)
	objs := make([]models.DeviceObject, 0, 64)

	if op != nil && devObjs != nil {
		do := devObjs[op.Object]

		if pc.descriptorExists(op.Parameter) {
			do.Name = op.Parameter
			objs = append(objs, do)
		} else if pc.descriptorExists(do.Name) {
			objs = append(objs, do)
		}
	}

	return objs
}

func (o *objectCache) transformResult(d *models.Device, op *models.ResourceOperation, do *models.DeviceObject, val string) string {
	// TODO: implement
	return val
}

// ReadingsExist returns a bool indicating whether or not the cache
// has any readings for the specified device and ResourceOperation.
//
// Note - in the Java SDK ObjectStore, this function was called 'get'
// and returned a JSONObject, however the only client, ProtocolHandler
// just checked for a nil retval.
//
// TODO - this probably can get removed, as a caller could just uses
// the Readings() method instead, and check for an empty list.
func (o *objectCache) ReadingsExist(d *models.Device, op *models.ResourceOperation) bool {
	return false
}

func buildOpId(objs []models.DeviceObject) string {
	buffer := bytes.NewBufferString("")
	fmt.Fprint(buffer, "[")

	sz := len(objs)
	for _, o := range objs {
		fmt.Fprintf(buffer, o.Name)

		if sz > 1 {
			fmt.Fprintf(buffer, ",")
		}
	}

	fmt.Fprint(buffer, "]")
	return buffer.String()
}

// AddReading adds a result from the specified ResourceOperation result to the cache.
func (o *objectCache) AddReading(d *models.Device, op *models.ResourceOperation, val string) []*models.Reading {
	if val == "" || val == "{}" {
		return nil
	}

	objs := o.createObjectList(d, op)
	id := d.Id.Hex()

	readings := make([]*models.Reading, 0, svc.c.Device.MaxCmdOps)

	for _, do := range objs {
		result := o.transformResult(d, op, &do, val)

		reading := &models.Reading{Name: do.Name, Value: result, Device: d.Name}
		readings = append(readings, reading)

		// TODO: add sync on Mutex
		o.mu.RLock()

		if o.objects[id] == nil {
			o.objects[id] = make(map[string][]string)
		}

		if o.objects[id][do.Name] == nil {
			o.objects[id][do.Name] = make([]string, 0, 64)
		}

		o.objects[id][do.Name] = append(o.objects[id][do.Name], result)

		// TODO: handle cacheSize check
		//
		// if the matching list == cacheSize (a config setting: ObjectsCacheSize)
		//if len(o.objects[id][do.Name]) == cacheSize {
		//   // remove one element
		//}

		o.mu.RUnlock()
	}

	opId := buildOpId(objs)

	// TODO: is it better to use two Mutexes?
	o.mu.RLock()

	if o.responses[id] == nil {
		o.responses[id] = make(map[string][]*models.Reading)
	}

	o.responses[id][opId] = readings

	o.mu.RUnlock()

	return readings
}

// Responses returns a list of readings from the cache for the specified
// device and ResourceOperation.
func (o *objectCache) Readings(d *models.Device, op *models.ResourceOperation) []models.Reading {
	return nil
}

// TransformData returns a bool which indicates if data read from a device
// or sensor should be tranformed before using it to create a reading.
// TODO: default values comes from Config: DataTransform
func (o *objectCache) TransformData() bool {
	return false
}

// SetTransformData returns a bool which indicates if data read from a device
// or sensor should be tranformed before using it to create a reading.
func (o *objectCache) SetTransformData(transform bool) {
}
