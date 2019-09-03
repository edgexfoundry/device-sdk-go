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

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
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

func (DeviceProfileClientMock) Add(dp *contract.DeviceProfile, ctx context.Context) (string, error) {
	panic("implement me")
}

func (DeviceProfileClientMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}

func (DeviceProfileClientMock) DeleteByName(name string, ctx context.Context) error {
	panic("implement me")
}

func (DeviceProfileClientMock) DeviceProfile(id string, ctx context.Context) (contract.DeviceProfile, error) {
	panic("implement me")
}

func (DeviceProfileClientMock) DeviceProfiles(ctx context.Context) ([]contract.DeviceProfile, error) {
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

func (DeviceProfileClientMock) DeviceProfileForName(name string, ctx context.Context) (contract.DeviceProfile, error) {
	panic("implement me")
}

func (DeviceProfileClientMock) Update(dp contract.DeviceProfile, ctx context.Context) error {
	panic("implement me")
}

func (DeviceProfileClientMock) Upload(yamlString string, ctx context.Context) (string, error) {
	panic("implement me")
}

func (DeviceProfileClientMock) UploadFile(yamlFilePath string, ctx context.Context) (string, error) {
	panic("implement me")
}

func populateDeviceProfileMock() error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	profiles, err := loadData(basepath + "/data/deviceprofile")
	if err != nil {
		return err
	}
	json.Unmarshal(profiles[ProfileBool], &DeviceProfileRandomBoolGenerator)
	json.Unmarshal(profiles[ProfileInt], &DeviceProfileRandomIntegerGenerator)
	json.Unmarshal(profiles[ProfileUint], &DeviceProfileRandomUnsignedGenerator)
	json.Unmarshal(profiles[ProfileFloat], &DeviceProfileRandomFloatGenerator)
	json.Unmarshal(profiles[ProfileFloat], &DuplicateDeviceProfileRandomFloatGenerator)
	json.Unmarshal(profiles[ProfileNew], &NewDeviceProfile)

	return nil
}
