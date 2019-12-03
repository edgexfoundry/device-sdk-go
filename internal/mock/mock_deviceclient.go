// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	InvalidDeviceId = "1ef435eb-5060-49b0-8d55-8d4e43239800"
)

var (
	ValidDeviceRandomBoolGenerator            = contract.Device{}
	ValidDeviceRandomIntegerGenerator         = contract.Device{}
	ValidDeviceRandomUnsignedIntegerGenerator = contract.Device{}
	ValidDeviceRandomFloatGenerator           = contract.Device{}
	DuplicateDeviceRandomFloatGenerator       = contract.Device{}
	NewValidDevice                            = contract.Device{}
	OperatingStateDisabled                    = contract.Device{}
)

type DeviceClientMock struct{}

func (dc *DeviceClientMock) Add(dev *contract.Device, ctx context.Context) (string, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}

func (dc *DeviceClientMock) DeleteByName(name string, ctx context.Context) error {
	panic("implement me")
}

func (dc *DeviceClientMock) CheckForDevice(token string, ctx context.Context) (contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Device(id string, ctx context.Context) (contract.Device, error) {
	if id == InvalidDeviceId {
		return contract.Device{}, fmt.Errorf("invalid id")
	}
	return contract.Device{}, nil
}

func (dc *DeviceClientMock) DeviceForName(name string, ctx context.Context) (contract.Device, error) {
	var device = contract.Device{Id: "5b977c62f37ba10e36673802", Name: name}
	var err error = nil
	if name == "" {
		err = errors.New("Item not found")
	}

	return device, err
}

func (dc *DeviceClientMock) Devices(ctx context.Context) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesByLabel(label string, ctx context.Context) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForProfile(profileid string, ctx context.Context) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForProfileByName(profileName string, ctx context.Context) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForService(serviceid string, ctx context.Context) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForServiceByName(serviceName string, ctx context.Context) ([]contract.Device, error) {
	err := populateDeviceMock()
	if err != nil {
		return nil, err
	}
	return []contract.Device{
		ValidDeviceRandomBoolGenerator,
		ValidDeviceRandomIntegerGenerator,
		ValidDeviceRandomUnsignedIntegerGenerator,
		ValidDeviceRandomFloatGenerator,
		OperatingStateDisabled,
	}, nil
}

func (dc *DeviceClientMock) Update(dev contract.Device, ctx context.Context) error {
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

func populateDeviceMock() error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	devices, err := loadData(basepath + "/data/device")
	if err != nil {
		return err
	}
	profiles, err := loadData(basepath + "/data/deviceprofile")
	if err != nil {
		return err
	}
	json.Unmarshal(devices[DeviceBool], &ValidDeviceRandomBoolGenerator)
	json.Unmarshal(profiles[DeviceBool], &ValidDeviceRandomBoolGenerator.Profile)
	json.Unmarshal(devices[DeviceInt], &ValidDeviceRandomIntegerGenerator)
	json.Unmarshal(profiles[DeviceInt], &ValidDeviceRandomIntegerGenerator.Profile)
	json.Unmarshal(devices[DeviceUint], &ValidDeviceRandomUnsignedIntegerGenerator)
	json.Unmarshal(profiles[DeviceUint], &ValidDeviceRandomUnsignedIntegerGenerator.Profile)
	json.Unmarshal(devices[DeviceFloat], &ValidDeviceRandomFloatGenerator)
	json.Unmarshal(profiles[DeviceFloat], &ValidDeviceRandomFloatGenerator.Profile)
	json.Unmarshal(devices[DeviceFloat], &DuplicateDeviceRandomFloatGenerator)
	json.Unmarshal(profiles[DeviceFloat], &DuplicateDeviceRandomFloatGenerator.Profile)
	json.Unmarshal(devices[DeviceNew], &NewValidDevice)
	json.Unmarshal(profiles[DeviceNew], &NewValidDevice.Profile)
	json.Unmarshal(devices[DeviceNew02], &OperatingStateDisabled)
	json.Unmarshal(profiles[DeviceNew], &OperatingStateDisabled.Profile)

	return nil
}
