// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package device

import (
	"errors"
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type ScheduleCacheInterface interface {
	GetScheduleByName(name string) (*models.Schedule, error)
	GetAllSchedules() *[]models.Schedule
	AddSchedule(schedule *models.Schedule)
	UpdateSchedule(schedule *models.Schedule) error
	RemoveSchedule(schedule *models.Schedule) error

	GetScheduleEventByName(name string) (*models.ScheduleEvent, error)
	GetAllScheduleEvents() *[]models.ScheduleEvent
	AddScheduleEvent(scheduleEvent *models.ScheduleEvent)
	UpdateScheduleEvent(scheduleEvent *models.ScheduleEvent) error
	RemoveScheduleEvent(scheduleEvent *models.ScheduleEvent) error
}

// ScheduleCache is a local cache of scheduleCache and scheduleevents,
// usually loaded into Core Metadata, however existing scheduleCache
// scheduleevents can be used to seed this cache.
type ScheduleCache struct {
	schedules      []models.Schedule
	scheduleEvents []models.ScheduleEvent
}

var (
	scOnce        sync.Once
	scheduleCache *ScheduleCache
)

// Creates a singleton Schedule Cache instance.
func getScheduleCache(config *Config) *ScheduleCache {
	scOnce.Do(func() {
		scheduleCache = &ScheduleCache{
			schedules:      config.Schedules,
			scheduleEvents: config.ScheduleEvents,
		}

		addSchedules(config.Schedules)
		addScheduleEvents(config.ScheduleEvents)
	})

	return scheduleCache
}

func addSchedules(schedules []models.Schedule) {
	for i := 0; i < len(schedules); i++ {
		schedule := schedules[i]

		if isScheduleExist(schedule.Name) {
			svc.lc.Info(fmt.Sprintf("Schedule (%v) exist.", schedule.Name))
			continue
		}

		id, err := svc.scc.Add(&schedule)
		if err != nil {
			svc.lc.Error(fmt.Sprintf("Add schedule (%v) fail: %v", schedule.Name, err.Error()))
			continue
		}
		schedule.Id = bson.ObjectIdHex(id)

		svc.lc.Info(fmt.Sprintf(fmt.Sprintf("Add schedule (%v) successful", schedule.Name)))
	}
}

func isScheduleExist(scheduleName string) bool {
	isExist := true
	schedule, _ := svc.scc.ScheduleForName(scheduleName)
	if schedule.Name == "" {
		isExist = false
	}
	return isExist
}

func addScheduleEvents(scheduleEvents []models.ScheduleEvent) {
	for i := 0; i < len(scheduleEvents); i++ {
		scheduleEvent := scheduleEvents[i]
		if scheduleEvent.Service == "" {
			scheduleEvent.Service = svc.Name
		}

		if isScheduleEventExist(scheduleEvent.Name) {
			svc.lc.Info(fmt.Sprintf("Schedule evnt (%v) exist", scheduleEvent.Name))
			continue
		}

		err := addScheduleEventAddressable(&scheduleEvent)
		if err != nil {
			svc.lc.Error(fmt.Sprintf("Add schedule event addressable (%v) fail: %v", scheduleEvent.Addressable.Name, err.Error()))
			continue
		}

		id, err := svc.scec.Add(&scheduleEvent)
		if err != nil {
			svc.lc.Error(fmt.Sprintf("Add schedule event (%v) fail: %v", scheduleEvent.Name, err.Error()))
			continue
		}
		scheduleEvent.Id = bson.ObjectIdHex(id)

		svc.lc.Info(fmt.Sprintf(fmt.Sprintf("Add schedule event (%v) successful", scheduleEvent.Name)))

	}
}

func addScheduleEventAddressable(scheduleEvent *models.ScheduleEvent) error {
	scheduleEvent.Addressable.Name = fmt.Sprintf("addressable-%v", scheduleEvent.Name)

	if isScheduleEventAddressableExist(scheduleEvent.Addressable.Name) {
		svc.lc.Info(fmt.Sprintf("Schedule evnt addressable (%v) exist", scheduleEvent.Addressable.Name))
		return nil
	}

	scheduleEvent.Addressable.Protocol = svc.ds.Addressable.Protocol
	scheduleEvent.Addressable.Address = svc.ds.Addressable.Address
	scheduleEvent.Addressable.Port = svc.ds.Addressable.Port

	addressableId, err := svc.ac.Add(&scheduleEvent.Addressable)
	if err != nil {
		return err
	}

	scheduleEvent.Addressable.Id = bson.ObjectIdHex(addressableId)

	return nil
}

func isScheduleEventAddressableExist(addressableName string) bool {
	isExist := true
	addressable, _ := svc.ac.AddressableForName(addressableName)
	if addressable.Name == "" {
		isExist = false
	}
	return isExist
}

func isScheduleEventExist(scheduleEventName string) bool {
	isExist := true
	scheduleEvent, _ := svc.scec.ScheduleEventForName(scheduleEventName)
	if scheduleEvent.Name == "" {
		isExist = false
	}
	return isExist
}

func (s *ScheduleCache) GetScheduleByName(name string) (*models.Schedule, error) {
	var schedule *models.Schedule
	var err error = nil
	for _, sc := range s.schedules {
		if sc.Name == name {
			schedule = &sc
			break
		}
	}
	if schedule == nil {
		err = errors.New(fmt.Sprintf("schedule not found : %v", name))
	}
	return schedule, err
}

func (s *ScheduleCache) GetAllSchedules() *[]models.Schedule {
	return &s.schedules
}

func (s *ScheduleCache) AddSchedule(schedule *models.Schedule) {
	s.schedules = append(s.schedules, *schedule)
}

func (s *ScheduleCache) UpdateSchedule(schedule *models.Schedule) error {
	err := s.RemoveSchedule(schedule)
	if err != nil {
		return errors.New("update schedule fail: " + err.Error())
	}
	s.AddSchedule(schedule)
	return nil
}

func (s *ScheduleCache) RemoveSchedule(schedule *models.Schedule) error {
	var removable = false
	for i, sc := range s.schedules {
		if sc.Name == schedule.Name {
			s.schedules = append(s.schedules[:i], s.schedules[i+1:]...)
			removable = true
			break
		}
	}
	if removable {
		return nil
	} else {
		return errors.New(fmt.Sprintf("schedule not found : %v", schedule.Name))
	}
}

func (s *ScheduleCache) GetScheduleEventByName(name string) (*models.ScheduleEvent, error) {
	var scheduleEvent *models.ScheduleEvent
	var err error = nil
	for _, se := range s.scheduleEvents {
		if se.Name == name {
			scheduleEvent = &se
			break
		}
	}
	if scheduleEvent == nil {
		err = errors.New(fmt.Sprintf("schedule not found : %v", name))
	}
	return scheduleEvent, err
}

func (s *ScheduleCache) GetAllScheduleEvents() *[]models.ScheduleEvent {
	return &s.scheduleEvents
}

func (s *ScheduleCache) AddScheduleEvent(scheduleEvent *models.ScheduleEvent) {
	s.scheduleEvents = append(s.scheduleEvents, *scheduleEvent)
}

func (s *ScheduleCache) UpdateScheduleEvent(scheduleEvent *models.ScheduleEvent) error {
	err := s.RemoveScheduleEvent(scheduleEvent)
	if err != nil {
		return errors.New("update schedule fail: " + err.Error())
	}
	s.AddScheduleEvent(scheduleEvent)
	return nil
}

func (s *ScheduleCache) RemoveScheduleEvent(scheduleEvent *models.ScheduleEvent) error {
	var removable = false
	for i, se := range s.scheduleEvents {
		if se.Name == scheduleEvent.Name {
			s.scheduleEvents = append(s.scheduleEvents[:i], s.scheduleEvents[i+1:]...)
			removable = true
			break
		}
	}
	if removable {
		return nil
	} else {
		return errors.New(fmt.Sprintf("scheduleEvent not found : %v", scheduleEvent.Name))
	}

}
