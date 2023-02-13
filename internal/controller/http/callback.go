// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/application"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
)

func (c *RestController) UpdateDeviceService(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var updateDeviceServiceRequest requests.UpdateDeviceServiceRequest

	err := json.NewDecoder(request.Body).Decode(&updateDeviceServiceRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.ApiServiceCallbackRoute)
		return
	}

	edgexErr := application.UpdateDeviceService(updateDeviceServiceRequest, c.dic)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(updateDeviceServiceRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiServiceCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.ApiServiceCallbackRoute)
	}
}
