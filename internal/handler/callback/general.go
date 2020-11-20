// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package callback

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func CallbackHandler(cbAlert contract.CallbackAlert, method string, dic *di.Container) common.AppError {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	if (cbAlert.Id == "") || (cbAlert.ActionType == "") {
		appErr := common.NewBadRequestError("Missing parameters", nil)
		lc.Error(fmt.Sprintf("Missing callback parameters"))
		return appErr
	}

	if cbAlert.ActionType == contract.DEVICE {
		return handleDevice(method, cbAlert.Id, dic)
	} else if cbAlert.ActionType == contract.SERVICE {
		return handleService(method, cbAlert.Id, dic)
	} else if cbAlert.ActionType == contract.PROFILE {
		return handleProfile(method, cbAlert.Id, dic)
	} else if cbAlert.ActionType == contract.PROVISIONWATCHER {
		return handleProvisionWatcher(method, cbAlert.Id, dic)
	}

	lc.Error(fmt.Sprintf("Invalid callback action type: %s", cbAlert.ActionType))
	appErr := common.NewBadRequestError("Invalid callback action type", nil)
	return appErr
}
