// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"errors"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type ScheduleEventClientMock struct{}

func (ScheduleEventClientMock) Add(dev *models.ScheduleEvent, ctx context.Context) (string, error) {
	return "", nil
}

func (ScheduleEventClientMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}

func (ScheduleEventClientMock) DeleteByName(name string, ctx context.Context) error {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEvent(id string, ctx context.Context) (models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEventForName(name string, ctx context.Context) (models.ScheduleEvent, error) {
	var scheduleEvent = models.ScheduleEvent{Name: name}
	var err error = nil
	if name == "" {
		err = errors.New("scheduleEvent not exist")
	}
	return scheduleEvent, err
}

func (ScheduleEventClientMock) ScheduleEvents(ctx context.Context) ([]models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEventsForAddressable(name string, ctx context.Context) ([]models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEventsForAddressableByName(name string, ctx context.Context) ([]models.ScheduleEvent, error) {
	panic("implement me")
}

func (ScheduleEventClientMock) ScheduleEventsForServiceByName(name string, ctx context.Context) ([]models.ScheduleEvent, error) {
	return []models.ScheduleEvent{}, nil
}

func (ScheduleEventClientMock) Update(dev models.ScheduleEvent, ctx context.Context) error {
	panic("implement me")
}
