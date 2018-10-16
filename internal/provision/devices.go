// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

func LoadDevices(deviceList []common.DeviceConfig) error {
	common.LoggingClient.Debug(fmt.Sprintf("Loading pre-define Devices from configuration: %v", deviceList))
	for _, d := range deviceList {
		if _, ok := cache.Devices().ForName(d.Name); ok {
			common.LoggingClient.Debug(fmt.Sprintf("Device %s exists, using the existing one", d.Name))
			continue
		} else {
			common.LoggingClient.Debug(fmt.Sprintf("Device %s doesn't exist, creating a new one", d.Name))
			err := createDevice(d)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("creating Device from config failed: %v", d))
				return err
			}
		}
	}
	return nil
}

func createDevice(dc common.DeviceConfig) error {
	prf, ok := cache.Profiles().ForName(dc.Profile)
	if !ok {
		errMsg := fmt.Sprintf("Device Profile %s doesn't exist for Device %v", dc.Profile, dc)
		common.LoggingClient.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	addr, err := common.MakeAddressable(dc.Name, &dc.Addressable)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("makeAddressable failed: %v", err))
		return err
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	device := &models.Device{
		Name:           dc.Name,
		Profile:        prf,
		Addressable:    *addr,
		Labels:         dc.Labels,
		Service:        common.CurrentDeviceService,
		AdminState:     models.Unlocked,
		OperatingState: models.Enabled,
	}
	device.Origin = millis
	device.Description = dc.Description
	common.LoggingClient.Debug(fmt.Sprintf("Adding Device: %v", device))
	id, err := common.DeviceClient.Add(device)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Add Device failed %v, error: %v", device, err))
		return err
	}
	if err = common.VerifyIdFormat(id, "Device"); err != nil {
		return err
	}
	device.Id = bson.ObjectIdHex(id)
	cache.Devices().Add(*device)

	return nil
}
