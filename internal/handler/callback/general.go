// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package callback

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func CallbackHandler(cbAlert contract.CallbackAlert, method string) common.AppError {
	if (cbAlert.Id == "") || (cbAlert.ActionType == "") {
		appErr := common.NewBadRequestError("Missing parameters", nil)
		common.LoggingClient.Error(fmt.Sprintf("Missing callback parameters"))
		return appErr
	}

	if cbAlert.ActionType == contract.DEVICE {
		return handleDevice(method, cbAlert.Id)
	} else if cbAlert.ActionType == contract.PROFILE {
		return handleProfile(method, cbAlert.Id)
	}

	common.LoggingClient.Error(fmt.Sprintf("Invalid callback action type: %s", cbAlert.ActionType))
	appErr := common.NewBadRequestError("Invalid callback action type", nil)
	return appErr
}
