// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/v2/application"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/gorilla/mux"
)

const SDKPostEventReserved = "ds-postevent"
const SDKReturnEventReserved = "ds-returnevent"

func (c *V2HttpController) Command(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	vars := mux.Vars(request)
	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)

	var err error
	var body string
	var sendEvent bool
	var reserved url.Values
	if request.Method == http.MethodPut {
		// read request body for PUT command
		body, err = readBodyAsString(request)
		if err != nil {
			c.sendError(writer, request, edgexErr.KindServerError, "failed to read request body", err, sdkCommon.APIV2NameCommandRoute, correlationID)
			return
		}
	} else if request.Method == http.MethodGet {
		// filter out the SDK reserved parameters and save the result for GET command
		body, reserved = filterQueryParams(request.URL.RawQuery, c.lc)
	}

	// push event to coredata if specified in GET command query parameters (default no)
	if ok, exist := reserved[SDKPostEventReserved]; exist && ok[0] == "yes" {
		sendEvent = true
	}
	isRead := request.Method == http.MethodGet
	event, edgexErr := application.CommandHandler(isRead, sendEvent, correlationID, vars, body, c.dic)
	if edgexErr != nil {
		c.sendEdgexError(writer, request, edgexErr, sdkCommon.APIV2NameCommandRoute, correlationID)
		return
	}

	// return http response based on SDK reserved query parameters (default yes)
	if ok, exist := reserved[SDKReturnEventReserved]; !exist || ok[0] == "yes" {
		// TODO: the usage of CBOR encoding for binary reading is under discussion
		// make the desired change when we have conclusion
		c.sendResponse(writer, request, sdkCommon.APIV2NameCommandRoute, event, http.StatusOK)
	}
}

func readBodyAsString(req *http.Request) (string, error) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	if len(body) == 0 && req.Method == http.MethodPut {
		return "", errors.New("no request body provided for PUT command")
	}

	return string(body), nil
}

func filterQueryParams(queryParams string, lc logger.LoggingClient) (string, url.Values) {
	m, err := url.ParseQuery(queryParams)
	if err != nil {
		lc.Error("Error parsing query parameters: %s\n", err)
		return "", nil
	}
	var reserved = make(url.Values)
	// Separate parameters with SDK reserved prefix
	for k := range m {
		if strings.HasPrefix(k, sdkCommon.SDKReservedPrefix) {
			reserved.Set(k, m.Get(k))
			delete(m, k)
		}
	}

	return m.Encode(), reserved
}
