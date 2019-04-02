// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	InvalidDeviceId   = "1ef435eb-5060-49b0-8d55-8d4e43239800"
)

var (
	ValidDeviceRandomBoolGenerator            = models.Device{}
	ValidDeviceRandomIntegerGenerator         = models.Device{}
	ValidDeviceRandomUnsignedIntegerGenerator = models.Device{}
	ValidDeviceRandomFloatGenerator           = models.Device{}
	DuplicateDeviceRandomFloatGenerator       = models.Device{}
	NewValidDevice                            = models.Device{}
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
	if id == InvalidDeviceId {
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
	populateDeviceMock()
	return []models.Device{
		ValidDeviceRandomBoolGenerator,
		ValidDeviceRandomIntegerGenerator,
		ValidDeviceRandomUnsignedIntegerGenerator,
		ValidDeviceRandomFloatGenerator,
	}, nil
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

func populateDeviceMock() {
	deviceDataRandomBoolGenerator := `{"created":1551711642756,"modified":1551711642756,"origin":1551711642745,"description":"Example of Device Virtual","id":"f9bcbc53-5a36-41f5-a749-c8cfdb6bd4a9","name":"Random-Boolean-Generator01","adminState":"UNLOCKED","operatingState":"ENABLED","addressable":{"created":1551711642745,"modified":0,"origin":1551711642745,"id":"6e4c48c7-3fb4-4e47-9f7e-fb20bfdf75a3","name":"Random-Boolean-Generator01","protocol":"OTHER","method":null,"address":"device-virtual-bool-01","port":0,"path":null,"publisher":null,"user":null,"password":null,"topic":null,"baseURL":"OTHER://device-virtual-bool-01:0","url":"OTHER://device-virtual-bool-01:0"},"lastConnected":0,"lastReported":0,"labels":["device-virtual-example"],"location":null,"service":{"created":1551711642565,"modified":1551711642565,"origin":1551711642563,"description":"","id":"9c1bb898-cfd9-46b8-93c6-1d197e50df48","name":"device-virtual","lastConnected":0,"lastReported":0,"operatingState":"ENABLED","labels":[],"addressable":{"created":1551711642549,"modified":0,"origin":1551711642541,"id":"9946109e-2592-478b-9f5a-99c2aaa3683f","name":"device-virtual","protocol":"HTTP","method":"POST","address":"device-virtual","port":49988,"path":"/api/v1/callback","publisher":null,"user":null,"password":null,"topic":null,"baseURL":"HTTP://device-virtual:49988","url":"HTTP://device-virtual:49988/api/v1/callback"},"adminState":"UNLOCKED"},"profileName":"Random-Boolean-Generator"}`
	deviceDataRandomIntegerGenerator := `{"created":1551711642770,"modified":1551711642770,"origin":1551711642762,"description":"Example of Device Virtual","id":"5f8f2f75-41bf-4f8a-8397-7421d8009b93","name":"Random-Integer-Generator01","adminState":"UNLOCKED","operatingState":"ENABLED","addressable":{"created":1551711642761,"modified":0,"origin":1551711642761,"id":"90b4ff6a-2049-4810-b43c-ca16cbe7107c","name":"Random-Integer-Generator01","protocol":"OTHER","method":null,"address":"device-virtual-int-01","port":0,"path":null,"publisher":null,"user":null,"password":null,"topic":null,"baseURL":"OTHER://device-virtual-int-01:0","url":"OTHER://device-virtual-int-01:0"},"lastConnected":0,"lastReported":0,"labels":["device-virtual-example"],"location":null,"service":{"created":1551711642565,"modified":1551711642565,"origin":1551711642563,"description":"","id":"7775e6bf-4e44-4166-9224-75749a9a12aa","name":"device-virtual","lastConnected":0,"lastReported":0,"operatingState":"ENABLED","labels":[],"addressable":{"created":1551711642549,"modified":0,"origin":1551711642541,"id":"2262b4f7-52d1-443f-a744-f38c9fe10008","name":"device-virtual","protocol":"HTTP","method":"POST","address":"device-virtual","port":49988,"path":"/api/v1/callback","publisher":null,"user":null,"password":null,"topic":null,"baseURL":"HTTP://device-virtual:49988","url":"HTTP://device-virtual:49988/api/v1/callback"},"adminState":"UNLOCKED"},"profileName":"Random-Integer-Generator"}`
	deviceDataRandomUnsignedIntegerGenerator := `{"created":1551711642783,"modified":1551711642783,"origin":1551711642774,"description":"Example of Device Virtual","id":"ae18c544-f4b7-4aa8-851c-792c2031de31","name":"Random-UnsignedInteger-Generator01","adminState":"UNLOCKED","operatingState":"ENABLED","addressable":{"created":1551711642774,"modified":0,"origin":1551711642774,"id":"366acaa0-b0b1-4ff2-9692-51e81c744e0d","name":"Random-UnsignedInteger-Generator01","protocol":"OTHER","method":null,"address":"device-virtual-uint-01","port":0,"path":null,"publisher":null,"user":null,"password":null,"topic":null,"baseURL":"OTHER://device-virtual-uint-01:0","url":"OTHER://device-virtual-uint-01:0"},"lastConnected":0,"lastReported":0,"labels":["device-virtual-example"],"location":null,"service":{"created":1551711642565,"modified":1551711642565,"origin":1551711642563,"description":"","id":"d8b39d48-f13a-4091-9318-b3048a45d919","name":"device-virtual","lastConnected":0,"lastReported":0,"operatingState":"ENABLED","labels":[],"addressable":{"created":1551711642549,"modified":0,"origin":1551711642541,"id":"ae3a7cb0-23b2-4727-9be4-6ed34bba6269","name":"device-virtual","protocol":"HTTP","method":"POST","address":"device-virtual","port":49988,"path":"/api/v1/callback","publisher":null,"user":null,"password":null,"topic":null,"baseURL":"HTTP://device-virtual:49988","url":"HTTP://device-virtual:49988/api/v1/callback"},"adminState":"UNLOCKED"},"profileName":"Random-UnsignedInteger-Generator"}`
	deviceDataRandomFloatGenerator := `{"created":1551711642791,"modified":1551711642791,"origin":1551711642786,"description":"Example of Device Virtual","id":"af8b1799-dc70-427c-ad95-1b8fad4fca0a","name":"Random-Float-Generator01","adminState":"UNLOCKED","operatingState":"ENABLED","addressable":{"created":1551711642785,"modified":0,"origin":1551711642785,"id":"16dc581e-f79b-4a2e-8859-afd1dd081ac6","name":"Random-Float-Generator01","protocol":"OTHER","method":null,"address":"device-virtual-float-01","port":0,"path":null,"publisher":null,"user":null,"password":null,"topic":null,"baseURL":"OTHER://device-virtual-float-01:0","url":"OTHER://device-virtual-float-01:0"},"lastConnected":0,"lastReported":0,"labels":["device-virtual-example"],"location":null,"service":{"created":1551711642565,"modified":1551711642565,"origin":1551711642563,"description":"","id":"994520b8-1307-4e61-b67e-5091545dad6e","name":"device-virtual","lastConnected":0,"lastReported":0,"operatingState":"ENABLED","labels":[],"addressable":{"created":1551711642549,"modified":0,"origin":1551711642541,"id":"0b6caba6-a16f-4b80-a622-9888caf188e5","name":"device-virtual","protocol":"HTTP","method":"POST","address":"device-virtual","port":49988,"path":"/api/v1/callback","publisher":null,"user":null,"password":null,"topic":null,"baseURL":"HTTP://device-virtual:49988","url":"HTTP://device-virtual:49988/api/v1/callback"},"adminState":"UNLOCKED"},"profileName":"Random-Float-Generator"}`
	newValidDeviceData := `{"created":1552129341782,"modified":1552129341782,"origin":1552129341676,"description":"Auto-generate this virtual device. GS1 AC Drive","id":"37ee8e20-d914-4588-95e6-0088fceb4a99","name":"GS1-AC-Drive01","adminState":"UNLOCKED","operatingState":"ENABLED","addressable":{"created":1552129341674,"modified":0,"origin":1552129341642,"id":"a09cf6c1-427c-4517-bc28-b4cb37df97e3","name":"GS1-AC-Drive01-virtual-addressable","protocol":"OTHER","method":"POST","address":"edgex-device-virtual","port":49990,"path":null,"publisher":null,"user":null,"password":null,"topic":null,"baseURL":"OTHER://edgex-device-virtual:49990","url":"OTHER://edgex-device-virtual:49990"},"lastConnected":0,"lastReported":0,"labels":["modbus","industrial"],"location":null,"service":{"created":1552129332730,"modified":1552129332730,"origin":1552129332645,"description":"","id":"2ce0393b-9814-4bd1-9ea6-3e95140c4b72","name":"edgex-device-virtual","lastConnected":0,"lastReported":0,"operatingState":"ENABLED","labels":["virtual"],"addressable":{"created":1552129332610,"modified":0,"origin":1552129332114,"id":"35fa45f1-f531-4789-b5f6-ec640dc1670c","name":"edgex-device-virtual","protocol":"HTTP","method":"POST","address":"edgex-device-virtual","port":49990,"path":"/api/v1/callback","publisher":null,"user":null,"password":null,"topic":null,"baseURL":"HTTP://edgex-device-virtual:49990","url":"HTTP://edgex-device-virtual:49990/api/v1/callback"},"adminState":"UNLOCKED"},"profileName":"GS1-AC-Drive"}`
	json.Unmarshal([]byte(deviceDataRandomBoolGenerator), &ValidDeviceRandomBoolGenerator)
	json.Unmarshal([]byte(deviceDataRandomIntegerGenerator), &ValidDeviceRandomIntegerGenerator)
	json.Unmarshal([]byte(deviceDataRandomUnsignedIntegerGenerator), &ValidDeviceRandomUnsignedIntegerGenerator)
	json.Unmarshal([]byte(deviceDataRandomFloatGenerator), &ValidDeviceRandomFloatGenerator)
	json.Unmarshal([]byte(deviceDataRandomFloatGenerator), &DuplicateDeviceRandomFloatGenerator)
	json.Unmarshal([]byte(newValidDeviceData), &NewValidDevice)
}
