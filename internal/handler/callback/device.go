// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package callback

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	"github.com/google/uuid"
)

func handleDevice(method string, id string) common.AppError {
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	if method == http.MethodPost {
		device, err := common.DeviceClient.Device(id, ctx)
		if err != nil {
			appErr := common.NewBadRequestError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Cannot find the device %s from Core Metadata: %v", id, err))
			return appErr
		}

		_, exist := cache.Profiles().ForName(device.Profile.Name)
		if exist == false {
			err = cache.Profiles().Add(device.Profile)
			if err == nil {
				provision.CreateDescriptorsFromProfile(&device.Profile)
				common.LoggingClient.Info(fmt.Sprintf("Added device profile %s", device.Profile.Id))
			} else {
				appErr := common.NewServerError(err.Error(), err)
				common.LoggingClient.Error(fmt.Sprintf("Couldn't add device profile %s: %v", device.Profile.Name, err.Error()))
				return appErr
			}
		}

		err = cache.Devices().Add(device)
		if err == nil {
			common.LoggingClient.Info(fmt.Sprintf("Added device %s", id))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Couldn't add device %s: %v", id, err.Error()))
			return appErr
		}

		err = common.Driver.AddDevice(device.Name, device.Protocols, device.AdminState)
		if err == nil {
			common.LoggingClient.Debug(fmt.Sprintf("Invoked driver.AddDevice callback for %s", device.Name))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Invoked driver.AddDevice callback failed for %s: %v", id, err.Error()))
			return appErr
		}

		common.LoggingClient.Debug(fmt.Sprintf("Handler - starting AutoEvents for device %s", device.Name))
		autoevent.GetManager().RestartForDevice(device.Name)
	} else if method == http.MethodPut {
		device, err := common.DeviceClient.Device(id, ctx)
		if err != nil {
			appErr := common.NewBadRequestError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Cannot find the device %s from Core Metadata: %v", id, err))
			return appErr
		}

		err = cache.Devices().Update(device)
		if err == nil {
			common.LoggingClient.Info(fmt.Sprintf("Updated device %s", id))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Couldn't update device %s: %v", id, err.Error()))
			return appErr
		}

		err = common.Driver.UpdateDevice(device.Name, device.Protocols, device.AdminState)
		if err == nil {
			common.LoggingClient.Debug(fmt.Sprintf("Invoked driver.UpdateDevice callback for %s", device.Name))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Invoked driver.UpdateDevice callback failed for %s: %v", id, err.Error()))
			return appErr
		}

		common.LoggingClient.Debug(fmt.Sprintf("Handler - restarting AutoEvents for updated device %s", device.Name))
		autoevent.GetManager().RestartForDevice(device.Name)
	} else if method == http.MethodDelete {
		device, ok := cache.Devices().ForId(id)
		if ok {
			common.LoggingClient.Debug(fmt.Sprintf("Handler - stopping AutoEvents for updated device %s", device.Name))
			autoevent.GetManager().StopForDevice(device.Name)
		}

		err := cache.Devices().Remove(id)
		if err == nil {
			common.LoggingClient.Info(fmt.Sprintf("Removed device %s", id))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Couldn't remove device %s: %v", id, err.Error()))
			return appErr
		}

		err = common.Driver.RemoveDevice(device.Name, device.Protocols)
		if err == nil {
			common.LoggingClient.Debug(fmt.Sprintf("Invoked driver.RemoveDevice callback for %s", device.Name))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Invoked driver.RemoveDevice callback failed for %s: %v", id, err.Error()))
			return appErr
		}
	} else {
		common.LoggingClient.Error(fmt.Sprintf("Invalid device method type: %s", method))
		appErr := common.NewBadRequestError("Invalid device method", nil)
		return appErr
	}

	return nil
}
