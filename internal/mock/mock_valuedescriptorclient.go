// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import "github.com/edgexfoundry/edgex-go/pkg/models"

type ValueDescriptorMock struct{}

func (ValueDescriptorMock) ValueDescriptors() ([]models.ValueDescriptor, error) {
	return []models.ValueDescriptor{}, nil
}

func (ValueDescriptorMock) ValueDescriptor(id string) (models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorForName(name string) (models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsForDevice(deviceId string) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsForDeviceByName(deviceName string) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) Add(vdr *models.ValueDescriptor) (string, error) {
	panic("implement me")
}

func (ValueDescriptorMock) Update(vdr *models.ValueDescriptor) error {
	panic("implement me")
}

func (ValueDescriptorMock) Delete(id string) error {
	panic("implement me")
}

func (ValueDescriptorMock) DeleteByName(name string) error {
	panic("implement me")
}
