// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"fmt"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const Int8Value = int8(123)

type DriverMock struct{}

func (DriverMock) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	panic("implement me")
}

func (DriverMock) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues) error {
	panic("implement me")
}

func (DriverMock) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	res = make([]*dsModels.CommandValue, len(reqs))
	now := time.Now().UnixNano() / int64(time.Millisecond)
	var v *dsModels.CommandValue
	for i, req := range reqs {
		switch deviceName {
		case "Random-Boolean-Generator01":
			if req.RO.Object == "RandomValue_Bool" {
				v, _ = dsModels.NewBoolValue(&req.RO, now, true)
			} else {
				err = fmt.Errorf("error occurred in HandleReadCommands")
			}
		case "Random-Integer-Generator01":
			switch req.RO.Object {
			default:
				v, _ = dsModels.NewInt8Value(&req.RO, now, Int8Value)
			case "NoDeviceResourceForResult":
				ro := models.ResourceOperation{Object: ""}
				v, _ = dsModels.NewInt8Value(&ro, now, Int8Value)
			case "Error":
				err = fmt.Errorf("error occurred in HandleReadCommands")
			}
		case "Random-UnsignedInteger-Generator01":
			if req.RO.Object == "RandomValue_Uint8" {
				v, _ = dsModels.NewUint8Value(&req.RO, now, uint8(123))
			} else {
				err = fmt.Errorf("error occurred in HandleReadCommands")
			}
		case "Random-Float-Generator01":
			if req.RO.Object == "RandomValue_Float32" {
				v, _ = dsModels.NewFloat32Value(&req.RO, now, float32(123))
			} else {
				err = fmt.Errorf("error occurred in HandleReadCommands")
			}
		}
		res[i] = v
	}
	return res, err
}

func (DriverMock) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {
	for _, req := range reqs {
		if req.RO.Object == "Error" {
			return fmt.Errorf("error occurred in HandleReadCommands")
		}
	}
	return nil
}

func (DriverMock) Stop(force bool) error {
	panic("implement me")
}
