// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	vdc *valueDescriptorCache
)

type ValueDescriptorCache interface {
	ForName(name string) (models.ValueDescriptor, bool)
	All() []models.ValueDescriptor
	Add(descriptor models.ValueDescriptor) error
	Update(descriptor models.ValueDescriptor) error
	Remove(id string) error
	RemoveByName(name string) error
}

type valueDescriptorCache struct {
	vdMap   map[string]models.ValueDescriptor // key is ValueDescriptor name
	nameMap map[string]string                 // key is id, and value is ValueDescriptor name
}

func (v *valueDescriptorCache) ForName(name string) (models.ValueDescriptor, bool) {
	vd, ok := v.vdMap[name]
	return vd, ok
}

func (v *valueDescriptorCache) All() []models.ValueDescriptor {
	vds := make([]models.ValueDescriptor, len(v.vdMap))
	i := 0
	for _, vd := range v.vdMap {
		vds[i] = vd
		i++
	}
	return vds
}

func (v *valueDescriptorCache) Add(descriptor models.ValueDescriptor) error {
	_, ok := v.vdMap[descriptor.Name]
	if ok {
		return fmt.Errorf("value descriptor %s has already existed in cache", descriptor.Name)
	}
	v.vdMap[descriptor.Name] = descriptor
	v.nameMap[descriptor.Id] = descriptor.Name
	return nil
}

func (v *valueDescriptorCache) Update(descriptor models.ValueDescriptor) error {
	if err := v.Remove(descriptor.Id); err != nil {
		return err
	}
	return v.Add(descriptor)
}

func (v *valueDescriptorCache) Remove(id string) error {
	name, ok := v.nameMap[id]
	if !ok {
		return fmt.Errorf("value descriptor %s does not exist in cache", id)
	}

	return v.RemoveByName(name)
}

func (v *valueDescriptorCache) RemoveByName(name string) error {
	vd, ok := v.vdMap[name]
	if !ok {
		return fmt.Errorf("value descriptor %s does not exist in cache", name)
	}
	delete(v.nameMap, vd.Id)
	delete(v.vdMap, name)
	return nil
}

func newValueDescriptorCache(descriptors []models.ValueDescriptor) ValueDescriptorCache {
	defaultSize := len(descriptors) * 2
	vdMap := make(map[string]models.ValueDescriptor, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	for _, vd := range descriptors {
		vdMap[vd.Name] = vd
		nameMap[vd.Id] = vd.Name
	}
	vdc = &valueDescriptorCache{vdMap: vdMap, nameMap: nameMap}
	return vdc
}

func ValueDescriptors() ValueDescriptorCache {
	if vdc == nil {
		InitCache()
	}
	return vdc
}
