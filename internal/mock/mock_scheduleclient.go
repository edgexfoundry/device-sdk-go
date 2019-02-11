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

type ScheduleClientMock struct {
}

func (s *ScheduleClientMock) Add(dev *models.Schedule, ctx context.Context) (string, error) {
	return "", nil
}

func (s *ScheduleClientMock) Delete(id string, ctx context.Context) error {
	return nil
}

func (s *ScheduleClientMock) DeleteByName(name string, ctx context.Context) error {
	return nil
}

func (s *ScheduleClientMock) Schedule(id string, ctx context.Context) (models.Schedule, error) {
	return models.Schedule{}, nil
}

func (s *ScheduleClientMock) Schedules(ctx context.Context) ([]models.Schedule, error) {
	return []models.Schedule{}, nil
}

func (s *ScheduleClientMock) Update(dev models.Schedule, ctx context.Context) error {
	return nil
}

func (s *ScheduleClientMock) ScheduleForName(name string, ctx context.Context) (models.Schedule, error) {
	var schedule = models.Schedule{Name: name}
	var err error = nil
	if name == "" {
		err = errors.New("schedule not exist")
	}
	return schedule, err
}
