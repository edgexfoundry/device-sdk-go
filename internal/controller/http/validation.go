//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

func (c *RestController) ValidateDevice(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var deviceRequest requests.AddDeviceRequest
	validator := container.DeviceValidatorFrom(c.dic.Get)
	if validator != nil {
		err := json.NewDecoder(request.Body).Decode(&deviceRequest)
		if err != nil {
			edgexErr := errors.NewCommonEdgeX(errors.KindContractInvalid, "json decoding failed", err)
			c.sendEdgexError(writer, request, edgexErr, common.ApiDeviceValidationRoute)
			return
		}

		err = validator.ValidateDevice(dtos.ToDeviceModel(deviceRequest.Device))
		if err != nil {
			edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "Device validation failed", err)
			c.sendEdgexError(writer, request, edgexErr, common.ApiDeviceValidationRoute)
			return
		}
	}

	res := commonDTO.NewBaseResponse(deviceRequest.RequestId, "", http.StatusOK)
	c.sendResponse(writer, request, common.ApiDeviceValidationRoute, res, http.StatusOK)
}
