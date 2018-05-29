// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package cache

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/tonyespy/gxds"
)

var (
	ocOnce  sync.Once
	objects *Objects
)

// TODO: re-name to 'values'?
type Objects struct {
	c             *gxds.Config
	objects       map[string]map[string][]string
	responses     map[string]map[string][]*models.Reading
	cacheSize     int
	transformData bool
	mu            sync.RWMutex
	lc            logger.LoggingClient
}

func NewObjects(c *gxds.Config, lc logger.LoggingClient) *Objects {

	ocOnce.Do(func() {
		objects = &Objects{c: c, lc: lc}

		objects.objects = make(map[string]map[string][]string)
		objects.responses = make(map[string]map[string][]*models.Reading)
	})

	return objects
}

// GetDeviceObject...
func (o *Objects) GetDeviceObject(d *models.Device, op *models.ResourceOperation) *models.DeviceObject {
	var devObj models.DeviceObject
	devObjs := profiles.GetDeviceObjects(d.Name)

	if op != nil && devObjs != nil {
		devObj, ok := devObjs[op.Object]

		if !ok {
			return nil
		}

		if profiles.descriptorExists(op.Parameter) {
			devObj.Name = op.Parameter
		}
	}

	return &devObj
}

func (o *Objects) createObjectList(d *models.Device, op *models.ResourceOperation) []models.DeviceObject {
	devObjs := profiles.GetDeviceObjects(d.Name)
	objs := make([]models.DeviceObject, 0, 64)

	if op != nil && devObjs != nil {
		do := devObjs[op.Object]

		if profiles.descriptorExists(op.Parameter) {
			do.Name = op.Parameter
			objs = append(objs, do)
		} else if profiles.descriptorExists(do.Name) {
			objs = append(objs, do)
		}
	}

	return objs
}

func (o *Objects) transformResult(d *models.Device, op *models.ResourceOperation, do *models.DeviceObject, val string) string {
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
func (o *Objects) ReadingsExist(d *models.Device, op *models.ResourceOperation) bool {
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
func (o *Objects) AddReading(d *models.Device, op *models.ResourceOperation, val string) []*models.Reading {
	if val == "" || val == "{}" {
		return nil
	}

	objs := o.createObjectList(d, op)
	id := d.Id.Hex()

	readings := make([]*models.Reading, 0, o.c.Device.MaxCmdOps)

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
func (o *Objects) Readings(d *models.Device, op *models.ResourceOperation) []models.Reading {
	return nil
}

// TransformData returns a bool which indicates if data read from a device
// or sensor should be tranformed before using it to create a reading.
// TODO: default values comes from Config: DataTransform
func (o *Objects) TransformData() bool {
	return false
}

// SetTransformData returns a bool which indicates if data read from a device
// or sensor should be tranformed before using it to create a reading.
func (o *Objects) SetTransformData(transform bool) {
}
