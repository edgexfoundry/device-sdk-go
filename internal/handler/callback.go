// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func CallbackHandler(cbAlert models.CallbackAlert, method string) common.AppError {
	if (cbAlert.Id == "") || (cbAlert.ActionType == "") {
		appErr := common.NewBadRequestError("Missing parameters", nil)
		common.LoggingClient.Error(fmt.Sprintf("Missing callback parameters"))
		return appErr
	}

	if cbAlert.ActionType == models.DEVICE {
		return handleDevice(method, cbAlert.Id)
	} else if cbAlert.ActionType == models.PROFILE {
		return handleProfile(method, cbAlert.Id)
	}

	common.LoggingClient.Error(fmt.Sprintf("Invalid callback action type: %s", cbAlert.ActionType))
	appErr := common.NewBadRequestError("Invalid callback action type", nil)
	return appErr
}

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
	} else if method == http.MethodPut {
		dev, err := common.DeviceClient.Device(id, ctx)
		if err != nil {
			appErr := common.NewBadRequestError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Cannot find the device %s from Core Metadata: %v", id, err))
			return appErr
		}

		err = cache.Devices().Update(dev)
		if err == nil {
			common.LoggingClient.Info(fmt.Sprintf("Updated device %s", id))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Couldn't update device %s: %v", id, err.Error()))
			return appErr
		}
	} else if method == http.MethodDelete {
		err := cache.Devices().Remove(id)
		if err == nil {
			common.LoggingClient.Info(fmt.Sprintf("Removed device %s", id))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Couldn't remove device %s: %v", id, err.Error()))
			return appErr
		}
	} else {
		common.LoggingClient.Error(fmt.Sprintf("Invalid device method type: %s", method))
		appErr := common.NewBadRequestError("Invalid device method", nil)
		return appErr
	}

	return nil
}

func handleProfile(method string, id string) common.AppError {
	if method == http.MethodPut {
		ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
		profile, err := common.DeviceProfileClient.DeviceProfile(id, ctx)
		if err != nil {
			appErr := common.NewBadRequestError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Cannot find the device profile %s from Core Metadata: %v", id, err))
			return appErr
		}

		err = cache.Profiles().Update(profile)
		if err == nil {
			provision.CreateDescriptorsFromProfile(&profile)
			common.LoggingClient.Info(fmt.Sprintf("Updated device profile %s", id))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Couldn't update device profile %s: %v", id, err.Error()))
			return appErr
		}
	} else {
		common.LoggingClient.Error(fmt.Sprintf("Invalid device profile method: %s", method))
		appErr := common.NewBadRequestError("Invalid device profile method", nil)
		return appErr
	}

	return nil
}
