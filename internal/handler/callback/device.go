// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func handleDevice(method string, id string) common.AppError {
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	switch method {
	case http.MethodPost:
		handleAddDevice(ctx, id)
	case http.MethodPut:
		handleUpdateDevice(ctx, id)
	case http.MethodDelete:
		handleDeleteDevice(id)
	default:
		common.LoggingClient.Error(fmt.Sprintf("Invalid device method type: %s", method))
		appErr := common.NewBadRequestError("Invalid device method", nil)
		return appErr
	}

	return nil
}

func handleAddDevice(ctx context.Context, id string) common.AppError {
	device, err := common.DeviceClient.Device(ctx, id)
	if err != nil {
		appErr := common.NewBadRequestError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Cannot find the device %s from Core Metadata: %v", id, err))
		return appErr
	}

	err = updateSpecifiedProfile(device.Profile)
	if err != nil {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Couldn't add device profile %s: %v", device.Profile.Name, err.Error()))
		return appErr
	}

	err = cache.Devices().Add(device)
	if err == nil {
		common.LoggingClient.Info(fmt.Sprintf("Added device: %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Couldn't add device %s: %v", device.Name, err.Error()))
		return appErr
	}

	err = common.Driver.AddDevice(device.Name, device.Protocols, device.AdminState)
	if err == nil {
		common.LoggingClient.Debug(fmt.Sprintf("Invoked driver.AddDevice callback for %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Invoked driver.AddDevice callback failed for %s: %v", device.Name, err.Error()))
		return appErr
	}

	common.LoggingClient.Debug(fmt.Sprintf("Handler - starting AutoEvents for device %s", device.Name))
	autoevent.GetManager().RestartForDevice(device.Name)

	return nil
}

func handleUpdateDevice(ctx context.Context, id string) common.AppError {
	device, err := common.DeviceClient.Device(ctx, id)
	if err != nil {
		appErr := common.NewBadRequestError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Cannot find the device %s from Core Metadata: %v", id, err))
		return appErr
	}

	err = updateSpecifiedProfile(device.Profile)
	if err != nil {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Couldn't add device profile %s: %v", device.Profile.Name, err.Error()))
		return appErr
	}

	err = cache.Devices().Update(device)
	if err == nil {
		common.LoggingClient.Info(fmt.Sprintf("Updated device: %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Couldn't update device %s: %v", device.Name, err.Error()))
		return appErr
	}

	err = common.Driver.UpdateDevice(device.Name, device.Protocols, device.AdminState)
	if err == nil {
		common.LoggingClient.Debug(fmt.Sprintf("Invoked driver.UpdateDevice callback for %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Invoked driver.UpdateDevice callback failed for %s: %v", device.Name, err.Error()))
		return appErr
	}

	common.LoggingClient.Debug(fmt.Sprintf("Handler - restarting AutoEvents for updated device %s", device.Name))
	autoevent.GetManager().RestartForDevice(device.Name)

	return nil
}

func handleDeleteDevice(id string) common.AppError {
	device, ok := cache.Devices().ForId(id)
	if ok {
		common.LoggingClient.Debug(fmt.Sprintf("Handler - stopping AutoEvents for updated device %s", device.Name))
		autoevent.GetManager().StopForDevice(device.Name)
	}

	err := cache.Devices().Remove(id)
	if err == nil {
		common.LoggingClient.Info(fmt.Sprintf("Removed device: %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Couldn't remove device %s: %v", device.Name, err.Error()))
		return appErr
	}

	err = common.Driver.RemoveDevice(device.Name, device.Protocols)
	if err == nil {
		common.LoggingClient.Debug(fmt.Sprintf("Invoked driver.RemoveDevice callback for %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Invoked driver.RemoveDevice callback failed for %s: %v", device.Name, err.Error()))
		return appErr
	}

	return nil
}

func updateSpecifiedProfile(profile contract.DeviceProfile) error {
	_, exist := cache.Profiles().ForName(profile.Name)
	if exist == false {
		err := cache.Profiles().Add(profile)
		if err == nil {
			provision.CreateDescriptorsFromProfile(&profile)
			common.LoggingClient.Info(fmt.Sprintf("Added device profile: %s", profile.Name))
		} else {
			return err
		}
	} else {
		err := cache.Profiles().Update(profile)
		if err != nil {
			common.LoggingClient.Warn(fmt.Sprintf("Unable to update profile %s in cache, using the original one", profile.Name))
		}
	}
	return nil
}
