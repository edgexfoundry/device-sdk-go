// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type ScheduleCache interface {
	ForName(name string) (models.Schedule, bool)
	All() []models.Schedule
	Add(sch models.Schedule) error
	Update(sch models.Schedule) error
	Remove(id string) error
	RemoveByName(name string) error
}

var (
	scCache *scheduleCache
)

// ScheduleCache is a local cache of Schedules,
// usually loaded from Core Metadata, and existing scheduleCache
// Schedules can be used to seed this cache.
type scheduleCache struct {
	scMap   map[string]models.Schedule // key is Schedule name
	nameMap map[string]string          // key is id, and value is Schedule name
}

func (s *scheduleCache) ForName(name string) (models.Schedule, bool) {
	sc, ok := s.scMap[name]
	return sc, ok
}

func (s *scheduleCache) All() []models.Schedule {
	sches := make([]models.Schedule, len(s.scMap))
	i := 0
	for _, sch := range s.scMap {
		sches[i] = sch
		i++
	}
	return sches
}

func (s *scheduleCache) Add(sch models.Schedule) error {
	if _, ok := s.scMap[sch.Name]; ok {
		return fmt.Errorf("schedule %s has already existed in cache", sch.Name)
	}
	s.scMap[sch.Name] = sch
	s.nameMap[sch.Id.Hex()] = sch.Name
	return nil
}

func (s *scheduleCache) Update(sch models.Schedule) error {
	if err := s.Remove(sch.Id.Hex()); err != nil {
		return err
	}
	return s.Add(sch)
}

func (s *scheduleCache) Remove(id string) error {
	name, ok := s.nameMap[id]
	if !ok {
		return fmt.Errorf("schedule %s does not exist in cache", id)
	}

	return s.RemoveByName(name)
}

// RemoveByName removes the specified schedule by name from the cache.
func (s *scheduleCache) RemoveByName(name string) error {
	sch, ok := s.scMap[name]
	if !ok {
		return fmt.Errorf("schedule %s does not exist in cache", name)
	}

	delete(s.nameMap, sch.Id.Hex())
	delete(s.scMap, name)
	return nil
}

// Creates a singleton Schedule Cache instance.
func newScheduleCache(schMap map[string]models.Schedule) ScheduleCache {
	nameMap := make(map[string]string, len(schMap)*2)
	for _, sc := range schMap {
		nameMap[sc.Id.Hex()] = sc.Name
	}
	scCache = &scheduleCache{scMap: schMap, nameMap: nameMap}
	return scCache
}

func Schedules() ScheduleCache {
	if scCache == nil {
		InitCache()
	}
	return scCache
}
