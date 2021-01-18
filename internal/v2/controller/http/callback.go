// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/application"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/gorilla/mux"
)

func (c *V2HttpController) DeleteDevice(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars[common.IdVar]

	err := application.DeleteDevice(id, c.dic)
	if err == nil {
		res := commonDTO.NewBaseResponse("", "", http.StatusOK)
		c.sendResponse(writer, request, v2.ApiDeviceCallbackIdRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, err, v2.ApiDeviceCallbackIdRoute)
	}
}

func (c *V2HttpController) AddDevice(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var addDeviceRequest requests.AddDeviceRequest

	err := json.NewDecoder(request.Body).Decode(&addDeviceRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, v2.ApiDeviceCallbackRoute)
		return
	}

	edgexErr := application.AddDevice(addDeviceRequest, c.dic)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(addDeviceRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, v2.ApiDeviceCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, v2.ApiDeviceCallbackRoute)
	}
}

func (c *V2HttpController) UpdateDevice(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var updateDeviceRequest requests.UpdateDeviceRequest

	err := json.NewDecoder(request.Body).Decode(&updateDeviceRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, v2.ApiDeviceCallbackRoute)
		return
	}

	edgexErr := application.UpdateDevice(updateDeviceRequest, c.dic)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(updateDeviceRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, v2.ApiDeviceCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, v2.ApiDeviceCallbackRoute)
	}
}

func (c *V2HttpController) DeleteProfile(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars[common.IdVar]

	err := application.DeleteProfile(id, c.lc)
	if err == nil {
		res := commonDTO.NewBaseResponse("", "", http.StatusOK)
		c.sendResponse(writer, request, v2.ApiProfileCallbackIdRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, err, v2.ApiProfileCallbackIdRoute)
	}
}

func (c *V2HttpController) AddUpdateProfile(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var edgexErr errors.EdgeX
	var profileRequest requests.DeviceProfileRequest

	err := json.NewDecoder(request.Body).Decode(&profileRequest)
	if err != nil {
		edgexErr = errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, v2.ApiProfileCallbackRoute)
		return
	}

	switch request.Method {
	case http.MethodPost:
		edgexErr = application.AddProfile(profileRequest, c.lc)
	case http.MethodPut:
		edgexErr = application.UpdateProfile(profileRequest, c.lc)
	}

	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(profileRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, v2.ApiProfileCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, v2.ApiProfileCallbackRoute)
	}
}

func (c *V2HttpController) DeleteProvisionWatcher(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	id := vars[common.IdVar]

	err := application.DeleteProvisionWatcher(id, c.lc)
	if err == nil {
		res := commonDTO.NewBaseResponse("", "", http.StatusOK)
		c.sendResponse(writer, request, v2.ApiWatcherCallbackIdRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, err, v2.ApiWatcherCallbackIdRoute)
	}
}

func (c *V2HttpController) AddProvisionWatcher(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var addProvisionWatcherRequest requests.AddProvisionWatcherRequest

	err := json.NewDecoder(request.Body).Decode(&addProvisionWatcherRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, v2.ApiWatcherCallbackRoute)
		return
	}

	edgexErr := application.AddProvisionWatcher(addProvisionWatcherRequest, c.lc)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse("", "", http.StatusOK)
		c.sendResponse(writer, request, v2.ApiWatcherCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, v2.ApiWatcherCallbackRoute)
	}
}

func (c *V2HttpController) UpdateProvisionWatcher(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var updateProvisionWatcherRequest requests.UpdateProvisionWatcherRequest

	err := json.NewDecoder(request.Body).Decode(&updateProvisionWatcherRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, v2.ApiWatcherCallbackRoute)
		return
	}

	edgexErr := application.UpdateProvisionWatcher(updateProvisionWatcherRequest, c.lc)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse("", "", http.StatusOK)
		c.sendResponse(writer, request, v2.ApiWatcherCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, v2.ApiWatcherCallbackRoute)
	}
}
