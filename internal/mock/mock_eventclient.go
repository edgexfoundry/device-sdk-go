// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type EventClientMock struct{}

func (EventClientMock) Events(ctx context.Context) ([]models.Event, error) {
	panic("implement me")
}

func (EventClientMock) Event(id string, ctx context.Context) (models.Event, error) {
	panic("implement me")
}

func (EventClientMock) EventCount(ctx context.Context) (int, error) {
	panic("implement me")
}

func (EventClientMock) EventCountForDevice(deviceId string, ctx context.Context) (int, error) {
	panic("implement me")
}

func (EventClientMock) EventsForDevice(id string, limit int, ctx context.Context) ([]models.Event, error) {
	panic("implement me")
}

func (EventClientMock) EventsForInterval(start int, end int, limit int, ctx context.Context) ([]models.Event, error) {
	panic("implement me")
}

func (EventClientMock) EventsForDeviceAndValueDescriptor(deviceId string, vd string, limit int, ctx context.Context) ([]models.Event, error) {
	panic("implement me")
}

func (EventClientMock) Add(event *models.Event, ctx context.Context) (string, error) {
	return "", nil
}

func (EventClientMock) DeleteForDevice(id string, ctx context.Context) error {
	panic("implement me")
}

func (EventClientMock) DeleteOld(age int, ctx context.Context) error {
	panic("implement me")
}

func (EventClientMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}

func (EventClientMock) MarkPushed(id string, ctx context.Context) error {
	panic("implement me")
}
