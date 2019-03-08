// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"errors"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	invalidDeviceId = "5b9a4f9a64562a2f966fdb0b"
)

type DeviceClientMock struct{}

func (dc *DeviceClientMock) Add(dev *models.Device, ctx context.Context) (string, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}

func (dc *DeviceClientMock) DeleteByName(name string, ctx context.Context) error {
	panic("implement me")
}

func (dc *DeviceClientMock) CheckForDevice(token string, ctx context.Context) (models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Device(id string, ctx context.Context) (models.Device, error) {
	if id == invalidDeviceId {
		return models.Device{}, fmt.Errorf("invalid id")
	}
	return models.Device{}, nil
}

func (dc *DeviceClientMock) DeviceForName(name string, ctx context.Context) (models.Device, error) {
	var device = models.Device{Id: "5b977c62f37ba10e36673802", Name: name}
	var err error = nil
	if name == "" {
		err = errors.New("Item not found")
	}

	return device, err
}

func (dc *DeviceClientMock) Devices(ctx context.Context) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesByLabel(label string, ctx context.Context) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForProfile(profileid string, ctx context.Context) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForProfileByName(profileName string, ctx context.Context) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForService(serviceid string, ctx context.Context) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForServiceByName(serviceName string, ctx context.Context) ([]models.Device, error) {
	return []models.Device{}, nil
}

func (dc *DeviceClientMock) Update(dev models.Device, ctx context.Context) error {
	return nil
}

func (dc *DeviceClientMock) UpdateAdminState(id string, adminState string, ctx context.Context) error {
	return nil
}

func (dc *DeviceClientMock) UpdateAdminStateByName(name string, adminState string, ctx context.Context) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastConnected(id string, time int64, ctx context.Context) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastConnectedByName(name string, time int64, ctx context.Context) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastReported(id string, time int64, ctx context.Context) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastReportedByName(name string, time int64, ctx context.Context) error {
	return nil
}

func (dc *DeviceClientMock) UpdateOpState(id string, opState string, ctx context.Context) error {
	return nil
}

func (dc *DeviceClientMock) UpdateOpStateByName(name string, opState string, ctx context.Context) error {
	return nil
}
