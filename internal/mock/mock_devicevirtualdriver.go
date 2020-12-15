// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"fmt"
	"math"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

const Int8Value = int8(123)

type DriverMock struct{}

func (DriverMock) DisconnectDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {
	panic("implement me")
}

func (DriverMock) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceCh chan<- []dsModels.DiscoveredDevice) error {
	panic("implement me")
}

func (DriverMock) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	res = make([]*dsModels.CommandValue, len(reqs))
	now := time.Now().UnixNano()
	var v *dsModels.CommandValue
	for i, req := range reqs {
		switch deviceName {
		case "Random-Boolean-Generator01":
			if req.DeviceResourceName == "RandomValue_Bool" {
				v, _ = dsModels.NewBoolValue(req.DeviceResourceName, now, true)
			} else {
				err = fmt.Errorf("error occurred in HandleReadCommands")
			}
		case "Random-Integer-Generator01":
			switch req.DeviceResourceName {
			default:
				v, _ = dsModels.NewInt8Value(req.DeviceResourceName, now, Int8Value)
			case "NoDeviceResourceForResult":
				ro := contract.ResourceOperation{DeviceResource: ""}
				v, _ = dsModels.NewInt8Value(ro.DeviceResource, now, Int8Value)
			case "Error":
				err = fmt.Errorf("error occurred in HandleReadCommands")
			}
		case "Random-UnsignedInteger-Generator01":
			if req.DeviceResourceName == "RandomValue_Uint8" {
				v, _ = dsModels.NewUint8Value(req.DeviceResourceName, now, uint8(123))
			} else {
				err = fmt.Errorf("error occurred in HandleReadCommands")
			}
		case "Random-Float-Generator01":
			switch req.DeviceResourceName {
			case ResourceObjectFloat32:
				v, _ = dsModels.NewFloat32Value(req.DeviceResourceName, now, float32(123))
			case ResourceObjectNaNFloat32:
				v, _ = dsModels.NewFloat32Value(req.DeviceResourceName, now, float32(math.NaN()))
			case ResourceObjectNaNFloat64:
				v, _ = dsModels.NewFloat64Value(req.DeviceResourceName, now, math.NaN())
			default:
				err = fmt.Errorf("error occurred in HandleReadCommands")
			}
		}
		res[i] = v
	}
	return res, err
}

func (DriverMock) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {
	for _, req := range reqs {
		if req.DeviceResourceName == "Error" {
			return fmt.Errorf("error occurred in HandleReadCommands")
		}
	}
	return nil
}

func (DriverMock) Stop(force bool) error {
	panic("implement me")
}

func (DriverMock) AddDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	return nil
}

func (DriverMock) UpdateDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	return nil
}

func (DriverMock) RemoveDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {
	return nil
}
