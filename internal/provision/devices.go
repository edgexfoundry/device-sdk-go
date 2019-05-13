// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func LoadDevices(deviceList []common.DeviceConfig) error {
	common.LoggingClient.Debug("Loading pre-define Devices from configuration")
	for _, d := range deviceList {
		if _, ok := cache.Devices().ForName(d.Name); ok {
			common.LoggingClient.Debug(fmt.Sprintf("Device %s exists, using the existing one", d.Name))
			continue
		} else {
			common.LoggingClient.Debug(fmt.Sprintf("Device %s doesn't exist, creating a new one", d.Name))
			err := createDevice(d)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("creating Device from config failed: %s", d.Name))
				return err
			}
		}
	}
	return nil
}

func createDevice(dc common.DeviceConfig) error {
	prf, ok := cache.Profiles().ForName(dc.Profile)
	if !ok {
		errMsg := fmt.Sprintf("Device Profile %s doesn't exist for Device %s", dc.Profile, dc.Name)
		common.LoggingClient.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	device := &contract.Device{
		Name:           dc.Name,
		Profile:        prf,
		Protocols:      dc.Protocols,
		Labels:         dc.Labels,
		Service:        common.CurrentDeviceService,
		AdminState:     contract.Unlocked,
		OperatingState: contract.Enabled,
		AutoEvents:     dc.AutoEvents,
	}
	device.Origin = millis
	device.Description = dc.Description
	common.LoggingClient.Debug(fmt.Sprintf("Adding Device: %v", device))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err := common.DeviceClient.Add(device, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Add Device failed %s, error: %v", device.Name, err))
		return err
	}
	if err = common.VerifyIdFormat(id, "Device"); err != nil {
		return err
	}
	device.Id = id
	cache.Devices().Add(*device)

	return nil
}
