// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/v2/application"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/gorilla/mux"
)

func (c *V2HttpController) DeleteDevice(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars[common.IdVar]

	err := application.DeleteDevice(id, c.dic)
	if err == nil {
		c.sendResponse(writer, request, common.APIV2DeviceCallbackIdRoute, nil, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, err, common.APIV2DeviceCallbackIdRoute)
	}
}

func (c *V2HttpController) AddDevice(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var addDeviceRequest requests.AddDeviceRequest

	err := json.NewDecoder(request.Body).Decode(&addDeviceRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.APIV2DeviceCallbackRoute)
		return
	}

	edgexErr := application.AddDevice(addDeviceRequest, c.dic)
	if edgexErr == nil {
		c.sendResponse(writer, request, common.APIV2DeviceCallbackRoute, nil, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.APIV2DeviceCallbackRoute)
	}
}

func (c *V2HttpController) UpdateDevice(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var updateDeviceRequest requests.UpdateDeviceRequest

	err := json.NewDecoder(request.Body).Decode(&updateDeviceRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.APIV2DeviceCallbackRoute)
		return
	}

	edgexErr := application.UpdateDevice(updateDeviceRequest, c.dic)
	if edgexErr == nil {
		c.sendResponse(writer, request, common.APIV2DeviceCallbackRoute, nil, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.APIV2DeviceCallbackRoute)
	}
}

func (c *V2HttpController) DeleteProfile(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars[common.IdVar]

	err := application.DeleteProfile(id, c.dic)
	if err == nil {
		c.sendResponse(writer, request, common.APIV2ProfileCallbackIdRoute, nil, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, err, common.APIV2ProfileCallbackIdRoute)
	}
}

func (c *V2HttpController) AddUpdateProfile(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var edgexErr errors.EdgeX
	var profileRequest requests.DeviceProfileRequest

	err := json.NewDecoder(request.Body).Decode(&profileRequest)
	if err != nil {
		edgexErr = errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.APIV2ProfileCallbackRoute)
		return
	}

	switch request.Method {
	case http.MethodPost:
		edgexErr = application.AddProfile(profileRequest, c.lc)
	case http.MethodPut:
		edgexErr = application.UpdateProfile(profileRequest, c.lc)
	}

	if edgexErr == nil {
		c.sendResponse(writer, request, common.APIV2ProfileCallbackRoute, nil, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.APIV2ProfileCallbackRoute)
	}
}
