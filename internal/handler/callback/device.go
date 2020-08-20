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
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func handleDevice(method string, id string, dic *di.Container) common.AppError {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	switch method {
	case http.MethodPost:
		handleAddDevice(ctx, id, dic)
	case http.MethodPut:
		handleUpdateDevice(ctx, id, dic)
	case http.MethodDelete:
		handleDeleteDevice(id, dic)
	default:
		lc.Error(fmt.Sprintf("Invalid device method type: %s", method))
		appErr := common.NewBadRequestError("Invalid device method", nil)
		return appErr
	}

	return nil
}

func handleAddDevice(ctx context.Context, id string, dic *di.Container) common.AppError {
	dc := container.MetadataDeviceClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	device, err := dc.Device(ctx, id)
	if err != nil {
		appErr := common.NewBadRequestError(err.Error(), err)
		lc.Error(fmt.Sprintf("Cannot find the device %s from Core Metadata: %v", id, err))
		return appErr
	}

	err = updateSpecifiedProfile(
		device.Profile,
		lc,
		container.GeneralClientFrom(dic.Get),
		container.CoredataValueDescriptorClientFrom(dic.Get))
	if err != nil {
		appErr := common.NewServerError(err.Error(), err)
		lc.Error(fmt.Sprintf("Couldn't add device profile %s: %v", device.Profile.Name, err.Error()))
		return appErr
	}

	err = cache.Devices().Add(device)
	if err == nil {
		lc.Info(fmt.Sprintf("Added device: %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		lc.Error(fmt.Sprintf("Couldn't add device %s: %v", device.Name, err.Error()))
		return appErr
	}

	driver := container.ProtocolDriverFrom(dic.Get)
	err = driver.AddDevice(device.Name, device.Protocols, device.AdminState)
	if err == nil {
		lc.Debug(fmt.Sprintf("Invoked driver.AddDevice callback for %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		lc.Error(fmt.Sprintf("Invoked driver.AddDevice callback failed for %s: %v", device.Name, err.Error()))
		return appErr
	}

	lc.Debug(fmt.Sprintf("Handler - starting AutoEvents for device %s", device.Name))
	autoevent.GetManager().RestartForDevice(device.Name, dic)

	return nil
}

func handleUpdateDevice(ctx context.Context, id string, dic *di.Container) common.AppError {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dc := container.MetadataDeviceClientFrom(dic.Get)

	device, err := dc.Device(ctx, id)
	if err != nil {
		appErr := common.NewBadRequestError(err.Error(), err)
		lc.Error(fmt.Sprintf("Cannot find the device %s from Core Metadata: %v", id, err))
		return appErr
	}

	err = updateSpecifiedProfile(
		device.Profile,
		lc,
		container.GeneralClientFrom(dic.Get),
		container.CoredataValueDescriptorClientFrom(dic.Get))
	if err != nil {
		appErr := common.NewServerError(err.Error(), err)
		lc.Error(fmt.Sprintf("Couldn't add device profile %s: %v", device.Profile.Name, err.Error()))
		return appErr
	}

	err = cache.Devices().Update(device)
	if err == nil {
		lc.Info(fmt.Sprintf("Updated device: %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		lc.Error(fmt.Sprintf("Couldn't update device %s: %v", device.Name, err.Error()))
		return appErr
	}

	driver := container.ProtocolDriverFrom(dic.Get)
	err = driver.UpdateDevice(device.Name, device.Protocols, device.AdminState)
	if err == nil {
		lc.Debug(fmt.Sprintf("Invoked driver.UpdateDevice callback for %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		lc.Error(fmt.Sprintf("Invoked driver.UpdateDevice callback failed for %s: %v", device.Name, err.Error()))
		return appErr
	}

	lc.Debug(fmt.Sprintf("Handler - restarting AutoEvents for updated device %s", device.Name))
	autoevent.GetManager().RestartForDevice(device.Name, dic)

	return nil
}

func handleDeleteDevice(id string, dic *di.Container) common.AppError {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	device, ok := cache.Devices().ForId(id)
	if ok {
		lc.Debug(fmt.Sprintf("Handler - stopping AutoEvents for updated device %s", device.Name))
		autoevent.GetManager().StopForDevice(device.Name)
	}

	err := cache.Devices().Remove(id)
	if err == nil {
		lc.Info(fmt.Sprintf("Removed device: %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		lc.Error(fmt.Sprintf("Couldn't remove device %s: %v", device.Name, err.Error()))
		return appErr
	}

	driver := container.ProtocolDriverFrom(dic.Get)
	err = driver.RemoveDevice(device.Name, device.Protocols)
	if err == nil {
		lc.Debug(fmt.Sprintf("Invoked driver.RemoveDevice callback for %s", device.Name))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		lc.Error(fmt.Sprintf("Invoked driver.RemoveDevice callback failed for %s: %v", device.Name, err.Error()))
		return appErr
	}

	return nil
}

func updateSpecifiedProfile(
	profile contract.DeviceProfile,
	lc logger.LoggingClient,
	gc general.GeneralClient,
	vdc coredata.ValueDescriptorClient) error {
	_, exist := cache.Profiles().ForName(profile.Name)
	if exist == false {
		err := cache.Profiles().Add(profile)
		if err == nil {
			provision.CreateDescriptorsFromProfile(&profile, lc, gc, vdc)
			lc.Info(fmt.Sprintf("Added device profile: %s", profile.Name))
		} else {
			return err
		}
	} else {
		err := cache.Profiles().Update(profile)
		if err != nil {
			lc.Warn(fmt.Sprintf("Unable to update profile %s in cache, using the original one", profile.Name))
		}
	}
	return nil
}
