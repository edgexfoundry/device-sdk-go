// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func LoadDevices(deviceList []common.DeviceConfig, dic *di.Container) error {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	lc.Debug("Loading pre-define Devices from configuration")
	for _, d := range deviceList {
		if _, ok := cache.Devices().ForName(d.Name); ok {
			lc.Debug(fmt.Sprintf("Device %s exists, using the existing one", d.Name))
			continue
		} else {
			lc.Debug(fmt.Sprintf("Device %s doesn't exist, creating a new one", d.Name))
			err := createDevice(
				d,
				lc,
				container.DeviceServiceFrom(dic.Get),
				container.MetadataDeviceClientFrom(dic.Get))
			if err != nil {
				lc.Error(fmt.Sprintf("creating Device %s from config failed", d.Name))
				return err
			}
		}
	}
	return nil
}

func createDevice(
	dc common.DeviceConfig,
	lc logger.LoggingClient,
	ds contract.DeviceService,
	mdc metadata.DeviceClient) error {
	prf, ok := cache.Profiles().ForName(dc.Profile)
	if !ok {
		errMsg := fmt.Sprintf("Device Profile %s doesn't exist for Device %s", dc.Profile, dc.Name)
		lc.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	device := &contract.Device{
		Name:           dc.Name,
		Profile:        prf,
		Protocols:      dc.Protocols,
		Labels:         dc.Labels,
		Service:        ds,
		AdminState:     contract.Unlocked,
		OperatingState: contract.Enabled,
		AutoEvents:     dc.AutoEvents,
	}
	device.Origin = millis
	device.Description = dc.Description
	lc.Debug(fmt.Sprintf("Adding Device: %v", device))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err := mdc.Add(ctx, device)
	if err != nil {
		return err
	}
	if err = common.VerifyIdFormat(id, "Device"); err != nil {
		return err
	}
	device.Id = id

	return nil
}
