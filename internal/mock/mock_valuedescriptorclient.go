// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

var (
	ValueDescriptorEnableRandomization = contract.ValueDescriptor{}
	ValueDescriptorBool                = contract.ValueDescriptor{}
	ValueDescriptorInt8                = contract.ValueDescriptor{}
	ValueDescriptorInt16               = contract.ValueDescriptor{}
	ValueDescriptorInt32               = contract.ValueDescriptor{}
	ValueDescriptorInt64               = contract.ValueDescriptor{}
	ValueDescriptorUint8               = contract.ValueDescriptor{}
	ValueDescriptorUint16              = contract.ValueDescriptor{}
	ValueDescriptorUint32              = contract.ValueDescriptor{}
	ValueDescriptorUint64              = contract.ValueDescriptor{}
	ValueDescriptorFloat32             = contract.ValueDescriptor{}
	ValueDescriptorFloat64             = contract.ValueDescriptor{}
	//ValueDescriptorString              = contract.ValueDescriptor{}
	NewValueDescriptor            = contract.ValueDescriptor{}
	DuplicateValueDescriptorInt16 = contract.ValueDescriptor{}
	descMap                       = make(map[string]contract.ValueDescriptor, 0)
)

type ValueDescriptorMock struct{}

func (ValueDescriptorMock) ValueDescriptorsUsage(names []string, ctx context.Context) (map[string]bool, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptors(ctx context.Context) ([]contract.ValueDescriptor, error) {
	err := populateValueDescriptorMock()
	if err != nil {
		return nil, err
	}
	return []contract.ValueDescriptor{
		ValueDescriptorEnableRandomization,
		ValueDescriptorBool,
		ValueDescriptorInt8,
		ValueDescriptorInt16,
		ValueDescriptorInt32,
		ValueDescriptorInt64,
		ValueDescriptorUint8,
		ValueDescriptorUint16,
		ValueDescriptorUint32,
		ValueDescriptorUint64,
		ValueDescriptorFloat32,
		ValueDescriptorFloat64,
		//ValueDescriptorString,
	}, nil
}

func (ValueDescriptorMock) ValueDescriptor(id string, ctx context.Context) (contract.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorForName(name string, ctx context.Context) (contract.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsByLabel(label string, ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsForDevice(deviceId string, ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsForDeviceByName(deviceName string, ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsByUomLabel(uomLabel string, ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) Add(vdr *contract.ValueDescriptor, ctx context.Context) (string, error) {
	panic("implement me")
}

func (ValueDescriptorMock) Update(vdr *contract.ValueDescriptor, ctx context.Context) error {
	panic("implement me")
}

func (ValueDescriptorMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}

func (ValueDescriptorMock) DeleteByName(name string, ctx context.Context) error {
	panic("implement me")
}

func populateValueDescriptorMock() error {
	common.DeviceProfileClient = &DeviceProfileClientMock{}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	dps, _ := common.DeviceProfileClient.DeviceProfiles(ctx)

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	profiles, err := loadData(basepath + "/data/deviceprofile")
	if err != nil {
		return err
	}
	json.Unmarshal(profiles["New-Device"], &NewDeviceProfile)
	dps = append(dps, NewDeviceProfile)

	for _, dp := range dps {
		if err := CreateDescriptorsFromProfile(dp); err != nil {
			return err
		}
	}

	ValueDescriptorBool = descMap[ResourceObjectBool]
	ValueDescriptorInt8 = descMap[ResourceObjectInt8]
	ValueDescriptorInt16 = descMap[ResourceObjectInt16]
	ValueDescriptorInt32 = descMap[ResourceObjectInt32]
	ValueDescriptorInt64 = descMap[ResourceObjectInt64]
	ValueDescriptorUint8 = descMap[ResourceObjectUint8]
	ValueDescriptorUint16 = descMap[ResourceObjectUint16]
	ValueDescriptorUint32 = descMap[ResourceObjectUint32]
	ValueDescriptorUint64 = descMap[ResourceObjectUint64]
	ValueDescriptorFloat32 = descMap[ResourceObjectFloat32]
	ValueDescriptorFloat64 = descMap[ResourceObjectFloat64]
	DuplicateValueDescriptorInt16 = descMap[ResourceObjectInt16]
	NewValueDescriptor = descMap[ResourceObjectRandFloat32]

	return nil
}

func CreateDescriptorsFromProfile(profile contract.DeviceProfile) error {
	dcs := profile.DeviceCommands
	for _, dc := range dcs {
		for _, op := range dc.Get {
			if err := createDescriptorFromResourceOperation(profile.Name, profile.DeviceResources, op); err != nil {
				return err
			}
		}
		for _, op := range dc.Set {
			if err := createDescriptorFromResourceOperation(profile.Name, profile.DeviceResources, op); err != nil {
				return err
			}
		}
	}
	return nil
}

func createDescriptorFromResourceOperation(profileName string, drs []contract.DeviceResource, op contract.ResourceOperation) error {
	if _, ok := descMap[op.DeviceResource]; ok {
		return nil
	}
	drMap := make(map[string]map[string]contract.DeviceResource, 0)
	drMap[profileName] = deviceResourceSliceToMap(drs)
	dr := drMap[profileName][op.DeviceResource]

	desc, err := createDescriptor(op.DeviceResource, dr)
	if err != nil {
		return err
	}
	descMap[op.DeviceResource] = desc
	return nil
}

func createDescriptor(name string, dr contract.DeviceResource) (contract.ValueDescriptor, error) {
	value := dr.Properties.Value
	units := dr.Properties.Units

	desc := contract.ValueDescriptor{
		Name:          name,
		Min:           value.Minimum,
		Max:           value.Maximum,
		Type:          value.Type,
		UomLabel:      units.DefaultValue,
		DefaultValue:  value.DefaultValue,
		Formatting:    "%s",
		Description:   dr.Description,
		FloatEncoding: value.FloatEncoding,
		MediaType:     value.MediaType,
	}

	id := uuid.New()
	desc.Id = id.String()

	return desc, nil
}

func deviceResourceSliceToMap(deviceResources []contract.DeviceResource) map[string]contract.DeviceResource {
	result := make(map[string]contract.DeviceResource, len(deviceResources))
	for _, dr := range deviceResources {
		result[dr.Name] = dr
	}
	return result
}
