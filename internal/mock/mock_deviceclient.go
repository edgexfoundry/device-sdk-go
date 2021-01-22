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

	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/requests/states/admin"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/requests/states/operating"
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

func (dc *DeviceClientMock) Add(_ context.Context, _ *contract.Device) (string, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Delete(_ context.Context, _ string) error {
	panic("implement me")
}

func (dc *DeviceClientMock) DeleteByName(_ context.Context, _ string) error {
	panic("implement me")
}

func (dc *DeviceClientMock) CheckForDevice(_ context.Context, _ string) (contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) Device(_ context.Context, id string) (contract.Device, error) {
	if id == InvalidDeviceId {
		return contract.Device{}, fmt.Errorf("invalid id")
	}
	return contract.Device{}, nil
}

func (dc *DeviceClientMock) DeviceForName(_ context.Context, name string) (contract.Device, error) {
	var device = contract.Device{Id: "5b977c62f37ba10e36673802", Name: name}
	var err error = nil
	if name == "" {
		err = errors.New("item not found")
	}

	return device, err
}

func (dc *DeviceClientMock) Devices(_ context.Context) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesByLabel(_ context.Context, _ string) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForProfile(_ context.Context, _ string) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForProfileByName(_ context.Context, _ string) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForService(_ context.Context, _ string) ([]contract.Device, error) {
	panic("implement me")
}

func (dc *DeviceClientMock) DevicesForServiceByName(_ context.Context, _ string) ([]contract.Device, error) {
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

func (dc *DeviceClientMock) Update(_ context.Context, _ contract.Device) error {
	return nil
}

func (dc *DeviceClientMock) UpdateAdminState(_ context.Context, _ string, _ admin.UpdateRequest) error {
	return nil
}

func (dc *DeviceClientMock) UpdateAdminStateByName(_ context.Context, _ string, _ admin.UpdateRequest) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastConnected(_ context.Context, _ string, _ int64) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastConnectedByName(_ context.Context, _ string, _ int64) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastReported(_ context.Context, _ string, _ int64) error {
	return nil
}

func (dc *DeviceClientMock) UpdateLastReportedByName(_ context.Context, _ string, _ int64) error {
	return nil
}

func (dc *DeviceClientMock) UpdateOpState(_ context.Context, _ string, _ operating.UpdateRequest) error {
	return nil
}

func (dc *DeviceClientMock) UpdateOpStateByName(_ context.Context, _ string, _ operating.UpdateRequest) error {
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
	_ = json.Unmarshal(devices[DeviceBool], &ValidDeviceRandomBoolGenerator)
	_ = json.Unmarshal(profiles[DeviceBool], &ValidDeviceRandomBoolGenerator.Profile)
	_ = json.Unmarshal(devices[DeviceInt], &ValidDeviceRandomIntegerGenerator)
	_ = json.Unmarshal(profiles[DeviceInt], &ValidDeviceRandomIntegerGenerator.Profile)
	_ = json.Unmarshal(devices[DeviceUint], &ValidDeviceRandomUnsignedIntegerGenerator)
	_ = json.Unmarshal(profiles[DeviceUint], &ValidDeviceRandomUnsignedIntegerGenerator.Profile)
	_ = json.Unmarshal(devices[DeviceFloat], &ValidDeviceRandomFloatGenerator)
	_ = json.Unmarshal(profiles[DeviceFloat], &ValidDeviceRandomFloatGenerator.Profile)
	_ = json.Unmarshal(devices[DeviceFloat], &DuplicateDeviceRandomFloatGenerator)
	_ = json.Unmarshal(profiles[DeviceFloat], &DuplicateDeviceRandomFloatGenerator.Profile)
	_ = json.Unmarshal(devices[DeviceNew], &NewValidDevice)
	_ = json.Unmarshal(profiles[DeviceNew], &NewValidDevice.Profile)
	_ = json.Unmarshal(devices[DeviceNew02], &OperatingStateDisabled)
	_ = json.Unmarshal(profiles[DeviceNew], &OperatingStateDisabled.Profile)

	return nil
}
