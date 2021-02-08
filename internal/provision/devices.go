// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/cache"
)

func LoadDevices(deviceList []common.DeviceConfig, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	var addDevicesReq []requests.AddDeviceRequest

	lc.Debug("loading pre-defined devices from configuration")
	for _, d := range deviceList {
		if _, ok := cache.Devices().ForName(d.Name); ok {
			lc.Debugf("device %s exists, using the existing one", d.Name)
			continue
		} else {
			lc.Debugf("device %s doesn't exist, creating a new one", d.Name)
			deviceDTO, err := createDeviceDTO(serviceName, d, dic)
			if err != nil {
				return err
			}
			req := requests.NewAddDeviceRequest(deviceDTO)
			addDevicesReq = append(addDevicesReq, req)
		}
	}

	if len(addDevicesReq) == 0 {
		return nil
	}
	dc := container.MetadataDeviceClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, err := dc.Add(ctx, addDevicesReq)
	return err
}

func createDeviceDTO(name string, dc common.DeviceConfig, dic *di.Container) (deviceDTO dtos.Device, err errors.EdgeX) {
	dpc := container.MetadataDeviceProfileClientFrom(dic.Get)
	_, err = dpc.DeviceProfileByName(context.Background(), dc.Profile)
	if err != nil {
		return
	}

	device := models.Device{
		Name:           dc.Name,
		Description:    dc.Description,
		Protocols:      dc.Protocols,
		Labels:         dc.Labels,
		ProfileName:    dc.Profile,
		ServiceName:    name,
		AdminState:     models.Unlocked,
		OperatingState: models.Up,
		AutoEvents:     dc.AutoEvents,
	}

	return dtos.FromDeviceModelToDTO(device), nil
}
