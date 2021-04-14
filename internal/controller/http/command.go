// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/application"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
)

func (c *RestController) Command(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var requestBody string
	var queryParams string
	var err errors.EdgeX
	var reserved url.Values
	vars := mux.Vars(request)
	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)
	if correlationID == "" {
		correlationID, _ = request.Context().Value(sdkCommon.CorrelationHeader).(string)
	}

	// read request body for SET command
	if request.Method == http.MethodPut {
		requestBody, err = readBodyAsString(request)
	}
	// parse query parameter
	queryParams, reserved, err = filterQueryParams(request.URL.RawQuery)
	if err != nil {
		c.sendEdgexError(writer, request, err, v2.ApiDeviceNameCommandNameRoute)
		return
	}

	var sendEvent bool
	// push event to CoreData if specified (default no)
	if ok, exist := reserved[v2.PushEvent]; exist && ok[0] == v2.ValueYes {
		sendEvent = true
	}
	isRead := request.Method == http.MethodGet
	eventDTO, edgexErr := application.CommandHandler(isRead, sendEvent, correlationID, vars, requestBody, queryParams, c.dic)
	if edgexErr != nil {
		c.sendEdgexError(writer, request, edgexErr, v2.ApiDeviceNameCommandNameRoute)
		return
	}

	var res interface{}
	if eventDTO != nil {
		res = responses.NewEventResponse("", "", http.StatusOK, *eventDTO)
	} else {
		res = common.NewBaseResponse("", "", http.StatusOK)
	}

	// return event in http response if specified (default yes)
	if ok, exist := reserved[v2.ReturnEvent]; !exist || ok[0] == v2.ValueYes {
		// TODO: the usage of CBOR encoding for binary reading is under discussion
		c.sendResponse(writer, request, v2.ApiDeviceNameCommandNameRoute, res, http.StatusOK)
	}
}

func readBodyAsString(req *http.Request) (string, errors.EdgeX) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "failed to read request body", err)
	}

	if len(body) == 0 {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "no request body provided for SET command", nil)
	}

	return string(body), nil
}

func filterQueryParams(queryParams string) (string, url.Values, errors.EdgeX) {
	m, err := url.ParseQuery(queryParams)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to parse query parameter", err)
		return "", nil, edgexErr
	}

	var reserved = make(url.Values)
	// Separate parameters with SDK reserved prefix
	for k := range m {
		if strings.HasPrefix(k, sdkCommon.SDKReservedPrefix) {
			reserved.Set(k, m.Get(k))
			delete(m, k)
		}
	}

	return m.Encode(), reserved, nil
}
