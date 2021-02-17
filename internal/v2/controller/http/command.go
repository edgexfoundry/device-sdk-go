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

	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/gorilla/mux"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/application"
)

func (c *V2HttpController) Command(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var body string
	var sendEvent bool
	var err edgexErr.EdgeX
	var reserved url.Values
	vars := mux.Vars(request)
	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)

	// read request body for PUT command, or parse query parameters for GET command.
	if request.Method == http.MethodPut {
		body, err = readBodyAsString(request)
	} else if request.Method == http.MethodGet {
		body, reserved, err = filterQueryParams(request.URL.RawQuery)
	}
	if err != nil {
		c.sendEdgexError(writer, request, err, v2.ApiDeviceNameCommandNameRoute)
		return
	}

	// push event to CoreData if specified (default no)
	if ok, exist := reserved[v2.PushEvent]; exist && ok[0] == v2.ValueYes {
		sendEvent = true
	}
	isRead := request.Method == http.MethodGet
	event, edgexErr := application.CommandHandler(isRead, sendEvent, correlationID, vars, body, c.dic)
	if edgexErr != nil {
		c.sendEdgexError(writer, request, edgexErr, v2.ApiDeviceNameCommandNameRoute)
		return
	}

	var res interface{}
	if event.Id != "" {
		res = responses.NewEventResponse(correlationID, "", http.StatusOK, event)
	} else {
		res = common.NewBaseResponse("", "", http.StatusOK)
	}

	// return event in http response if specified (default yes)
	if ok, exist := reserved[v2.ReturnEvent]; !exist || ok[0] == v2.ValueYes {
		// TODO: the usage of CBOR encoding for binary reading is under discussion
		c.sendResponse(writer, request, v2.ApiDeviceNameCommandNameRoute, res, http.StatusOK)
	}
}

func readBodyAsString(req *http.Request) (string, edgexErr.EdgeX) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to read request body", err)
	}

	if len(body) == 0 && req.Method == http.MethodPut {
		return "", edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "no request body provided for PUT command", nil)
	}

	return string(body), nil
}

func filterQueryParams(queryParams string) (string, url.Values, edgexErr.EdgeX) {
	m, err := url.ParseQuery(queryParams)
	if err != nil {
		edgexErr := edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to parse query parameter", err)
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
