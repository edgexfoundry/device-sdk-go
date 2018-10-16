// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func CallbackHandler(cbAlert models.CallbackAlert, method string) common.AppError {
	if (cbAlert.Id == "") || (cbAlert.ActionType == "") {
		appErr := common.NewBadRequestError("Missing parameters", nil)
		common.LoggingClient.Error(fmt.Sprintf("Missing callback parameters"))
		return appErr
	}

	if (cbAlert.ActionType == models.DEVICE) && (method == http.MethodPut) {
		dev, err := common.DeviceClient.Device(cbAlert.Id)
		if err != nil {
			appErr := common.NewBadRequestError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("cannot find the Device %s from Core Metadata: %v", cbAlert.Id, err))
			return appErr
		}

		err = cache.Devices().UpdateAdminState(cbAlert.Id, dev.AdminState)
		if err == nil {
			common.LoggingClient.Info(fmt.Sprintf("Updated device %s admin state", cbAlert.Id))
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Couldn't update device %s admin state: %v", cbAlert.Id, err.Error()))
			return appErr
		}
	} else {
		common.LoggingClient.Error(fmt.Sprintf("Invalid device method and/or action type: %s - %s", method, cbAlert.ActionType))
		appErr := common.NewBadRequestError("Invalid device method and/or action type", nil)
		return appErr
	}

	return nil
}
