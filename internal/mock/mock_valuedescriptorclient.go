// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type ValueDescriptorMock struct{}

func (ValueDescriptorMock) ValueDescriptors(ctx context.Context) ([]models.ValueDescriptor, error) {
	return []models.ValueDescriptor{}, nil
}

func (ValueDescriptorMock) ValueDescriptor(id string, ctx context.Context) (models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorForName(name string, ctx context.Context) (models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsByLabel(label string, ctx context.Context) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsForDevice(deviceId string, ctx context.Context) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsForDeviceByName(deviceName string, ctx context.Context) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsByUomLabel(uomLabel string, ctx context.Context) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) Add(vdr *models.ValueDescriptor, ctx context.Context) (string, error) {
	panic("implement me")
}

func (ValueDescriptorMock) Update(vdr *models.ValueDescriptor, ctx context.Context) error {
	panic("implement me")
}

func (ValueDescriptorMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}

func (ValueDescriptorMock) DeleteByName(name string, ctx context.Context) error {
	panic("implement me")
}
