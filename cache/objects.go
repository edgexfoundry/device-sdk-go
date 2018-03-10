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
