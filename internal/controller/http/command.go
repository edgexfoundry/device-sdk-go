// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/application"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
)

func (c *RestController) GetCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceName := vars[common.Name]
	commandName := vars[common.Command]

	ctx := r.Context()
	correlationId := utils.FromContext(ctx, common.CorrelationHeader)

	// parse query parameter
	queryParams, reserved, err := filterQueryParams(r.URL.RawQuery)
	if err != nil {
		c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
		return
	}

	event, err := application.GetCommand(ctx, deviceName, commandName, queryParams, c.dic)
	if err != nil {
		c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
		return
	}

	// push event to CoreData if specified (default no)
	if ok, exist := reserved[common.PushEvent]; exist && ok[0] == common.ValueYes {
		go sdkCommon.SendEvent(event, correlationId, c.dic)
	}

	// return event in http response if specified (default yes)
	if ok, exist := reserved[common.ReturnEvent]; !exist || ok[0] == common.ValueYes {
		res := responses.NewEventResponse("", "", http.StatusOK, *event)
		c.sendEventResponse(w, r, res, http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *RestController) SetCommand(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	deviceName := vars[common.Name]
	commandName := vars[common.Command]

	// parse query parameter
	queryParams, _, err := filterQueryParams(r.URL.RawQuery)
	if err != nil {
		c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
		return
	}

	requestParamsMap, err := parseRequestBody(r)
	if err != nil {
		c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
		return
	}

	err = application.SetCommand(ctx, deviceName, commandName, queryParams, requestParamsMap, c.dic)
	if err != nil {
		c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
		return
	}

	res := commonDTO.NewBaseResponse("", "", http.StatusOK)
	c.sendResponse(w, r, common.ApiDeviceNameCommandNameRoute, res, http.StatusOK)
}

func parseRequestBody(req *http.Request) (map[string]interface{}, errors.EdgeX) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to read request body", err)
	}

	paramMap := make(map[string]interface{})
	if len(body) == 0 {
		return paramMap, nil
	}

	err = json.Unmarshal(body, &paramMap)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to parse request body", err)
	}

	return paramMap, nil
}

func filterQueryParams(rawQuery string) (string, url.Values, errors.EdgeX) {
	queryParams, err := url.ParseQuery(rawQuery)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to parse query parameter", err)
		return "", nil, edgexErr
	}

	var reserved = make(url.Values)
	// Separate parameters with SDK reserved prefix
	for k := range queryParams {
		if strings.HasPrefix(k, sdkCommon.SDKReservedPrefix) {
			reserved.Set(k, queryParams.Get(k))
			delete(queryParams, k)
		}
	}

	return queryParams.Encode(), reserved, nil
}
