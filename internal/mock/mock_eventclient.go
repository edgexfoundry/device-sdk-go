// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"

	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type EventClientMock struct{}

func (e EventClientMock) Events(_ context.Context) ([]contract.Event, error) {
	panic("implement me")
}

func (e EventClientMock) Event(_ context.Context, _ string) (contract.Event, error) {
	panic("implement me")
}

func (e EventClientMock) EventCount(_ context.Context) (int, error) {
	panic("implement me")
}

func (e EventClientMock) EventCountForDevice(_ context.Context, _ string) (int, error) {
	panic("implement me")
}

func (e EventClientMock) EventsForDevice(_ context.Context, _ string, _ int) ([]contract.Event, error) {
	panic("implement me")
}

func (e EventClientMock) EventsForInterval(_ context.Context, _ int, _ int, _ int) ([]contract.Event, error) {
	panic("implement me")
}

func (e EventClientMock) EventsForDeviceAndValueDescriptor(_ context.Context, _ string, _ string, _ int) ([]contract.Event, error) {
	panic("implement me")
}

func (e EventClientMock) Add(_ context.Context, _ *contract.Event) (string, error) {
	panic("implement me")
}

func (e EventClientMock) AddBytes(_ context.Context, _ []byte) (string, error) {
	panic("implement me")
}

func (e EventClientMock) DeleteForDevice(_ context.Context, _ string) error {
	panic("implement me")
}

func (e EventClientMock) DeleteOld(_ context.Context, _ int) error {
	panic("implement me")
}

func (e EventClientMock) Delete(_ context.Context, _ string) error {
	panic("implement me")
}

func (e EventClientMock) MarkPushed(_ context.Context, _ string) error {
	panic("implement me")
}

func (e EventClientMock) MarkPushedByChecksum(_ context.Context, _ string) error {
	panic("implement me")
}

func (e EventClientMock) MarshalEvent(_ contract.Event) ([]byte, error) {
	panic("implement me")
}
