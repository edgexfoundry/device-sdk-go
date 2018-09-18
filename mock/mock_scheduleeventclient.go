//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"errors"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type ScheduleEventClientMock struct{}

func (ScheduleEventClientMock) Add(dev *models.ScheduleEvent) (string, error) {
	return "", nil
}

func (ScheduleEventClientMock) Delete(id string) error {
	panic("implement me")
}

func (ScheduleEventClientMock) DeleteByName(name string) error {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEvent(id string) (models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEventForName(name string) (models.ScheduleEvent, error) {
	var scheduleEvent = models.ScheduleEvent{Name: name}
	var err error = nil
	if name == "" {
		err = errors.New("scheduleEvent not exist")
	}
	return scheduleEvent, err
}

func (ScheduleEventClientMock) ScheduleEvents() ([]models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEventsForAddressable(name string) ([]models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEventsForAddressableByName(name string) ([]models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEventsForServiceByName(name string) ([]models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) Update(dev models.ScheduleEvent) error {
	panic("implement me")
}
