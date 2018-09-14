//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"errors"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

type DeviceClientMock struct{}

func (dc *DeviceClientMock) Add(dev *models.Device) (string, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Delete(id string) error {
	panic("implement me")
}

func (dc *DeviceClientMock) DeleteByName(name string) error {
	panic("implement me")
}

func (dc *DeviceClientMock) CheckForDevice(token string) (models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Device(id string) (models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DeviceForName(name string) (models.Device, error) {
	var device = models.Device{Id: bson.ObjectIdHex("5b977c62f37ba10e36673802"), Name: name}
	var err error = nil
	if name == "" {
		err = errors.New("Item not found")
	}

	return device, err
}

func (dc *DeviceClientMock) Devices() ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesByLabel(label string) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForAddressable(addressableid string) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForAddressableByName(addressableName string) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForProfile(profileid string) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForProfileByName(profileName string) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForService(serviceid string) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForServiceByName(serviceName string) ([]models.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Update(dev models.Device) error {
	return nil
}

func (dc *DeviceClientMock) UpdateAdminState(id string, adminState string) error {
	return nil
}

func (dc *DeviceClientMock) UpdateAdminStateByName(name string, adminState string) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastConnected(id string, time int64) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastConnectedByName(name string, time int64) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastReported(id string, time int64) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastReportedByName(name string, time int64) error {
	return nil
}

func (dc *DeviceClientMock) UpdateOpState(id string, opState string) error {
	return nil
}

func (dc *DeviceClientMock) UpdateOpStateByName(name string, opState string) error {
	return nil
}
