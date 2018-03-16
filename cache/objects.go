// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package cache

import (
	"sync"
	
	"github.com/edgexfoundry/core-domain-go/models"
)

var (
	ocOnce      sync.Once
	objects     *Objects
)

type Objects struct {
	objects       map[string]map[string][]string
	responses     map[string]map[string][]models.Reading
	cacheSize     int
	transformData bool
}

func NewObjects() *Objects {

	ocOnce.Do(func() {
		objects = &Objects{}
	})

	return objects
}

// Object returns a string containing the result of a ResourceOperation
// performed on a specific device.  It's used by the ProtocolHandler to
// ... (TODO: complete this)
func (o *Objects) Object(device models.Device, op models.ResourceOperation) string {
	return ""
}

// add a ResourceOperation result to the cache.
func (o *Objects) add(device models.Device, op models.ResourceOperation, value string) {
}

// Responses returns a list of readings from the cache for the specified
// device and ResourceOperation.
func (o *Objects) Responses(device models.Device, op models.ResourceOperation) []models.Reading {
	return nil
}

// TransformData returns a bool which indicates if data read from a device
// or sensor should be tranformed before using it to create a reading.
func (o *Objects) TransformData() bool {
	return false
}

// SetTransformData returns a bool which indicates if data read from a device
// or sensor should be tranformed before using it to create a reading.
func (o *Objects) SetTransformData(transform bool) {
}
