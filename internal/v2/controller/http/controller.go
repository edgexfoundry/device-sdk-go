// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
)

// V2HttpController controller for V2 REST APIs
type V2HttpController struct {
	dic *di.Container
	lc  logger.LoggingClient
}

// NewV2HttpController creates and initializes an V2HttpController
func NewV2HttpController(dic *di.Container) *V2HttpController {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	return &V2HttpController{
		dic: dic,
		lc:  lc,
	}
}

// sendResponse puts together the response packet for the V2 API
func (c *V2HttpController) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	api string,
	response interface{},
	statusCode int) {

	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)

	writer.Header().Set(sdkCommon.CorrelationHeader, correlationID)
	writer.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	if response != nil {
		data, err := json.Marshal(response)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(data)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (c *V2HttpController) sendEdgexError(
	writer http.ResponseWriter,
	request *http.Request,
	err errors.EdgeX,
	api string) {
	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)
	c.lc.Error(err.Error(), sdkCommon.CorrelationHeader, correlationID)
	c.lc.Debug(err.DebugMessages(), sdkCommon.CorrelationHeader, correlationID)
	response := common.NewBaseResponse("", err.Error(), err.Code())
	c.sendResponse(writer, request, api, response, err.Code())
}
