// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"

	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

var (
	DeviceProfileRandomBoolGenerator           = contract.DeviceProfile{}
	DeviceProfileRandomIntegerGenerator        = contract.DeviceProfile{}
	DeviceProfileRandomUnsignedGenerator       = contract.DeviceProfile{}
	DeviceProfileRandomFloatGenerator          = contract.DeviceProfile{}
	DuplicateDeviceProfileRandomFloatGenerator = contract.DeviceProfile{}
	NewDeviceProfile                           = contract.DeviceProfile{}
)

type DeviceProfileClientMock struct{}

func (DeviceProfileClientMock) Add(_ context.Context, _ *contract.DeviceProfile) (string, error) {
	panic("implement me")
}

func (DeviceProfileClientMock) Delete(_ context.Context, _ string) error {
	panic("implement me")
}

func (DeviceProfileClientMock) DeleteByName(_ context.Context, _ string) error {
	panic("implement me")
}

func (DeviceProfileClientMock) DeviceProfile(_ context.Context, _ string) (contract.DeviceProfile, error) {
	panic("implement me")
}

func (DeviceProfileClientMock) DeviceProfiles(_ context.Context) ([]contract.DeviceProfile, error) {
	err := populateDeviceProfileMock()
	if err != nil {
		return nil, err
	}
	return []contract.DeviceProfile{
		DeviceProfileRandomBoolGenerator,
		DeviceProfileRandomIntegerGenerator,
		DeviceProfileRandomUnsignedGenerator,
		DeviceProfileRandomFloatGenerator,
	}, nil
}

func (DeviceProfileClientMock) DeviceProfileForName(_ context.Context, _ string) (contract.DeviceProfile, error) {
	panic("implement me")
}

func (DeviceProfileClientMock) Update(_ context.Context, _ contract.DeviceProfile) error {
	panic("implement me")
}

func (DeviceProfileClientMock) Upload(_ context.Context, _ string) (string, error) {
	panic("implement me")
}

func (DeviceProfileClientMock) UploadFile(_ context.Context, _ string) (string, error) {
	panic("implement me")
}

func populateDeviceProfileMock() error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	profiles, err := loadData(basepath + "/data/deviceprofile")
	if err != nil {
		return err
	}
	_ = json.Unmarshal(profiles[ProfileBool], &DeviceProfileRandomBoolGenerator)
	_ = json.Unmarshal(profiles[ProfileInt], &DeviceProfileRandomIntegerGenerator)
	_ = json.Unmarshal(profiles[ProfileUint], &DeviceProfileRandomUnsignedGenerator)
	_ = json.Unmarshal(profiles[ProfileFloat], &DeviceProfileRandomFloatGenerator)
	_ = json.Unmarshal(profiles[ProfileFloat], &DuplicateDeviceProfileRandomFloatGenerator)
	_ = json.Unmarshal(profiles[ProfileNew], &NewDeviceProfile)

	return nil
}
