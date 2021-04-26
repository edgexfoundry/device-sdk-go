// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/gorilla/mux"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/controller/http/correlation"
)

type RestController struct {
	lc             logger.LoggingClient
	router         *mux.Router
	reservedRoutes map[string]bool
	dic            *di.Container
}

func NewRestController(r *mux.Router, dic *di.Container) *RestController {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	return &RestController{
		lc:             lc,
		router:         r,
		reservedRoutes: make(map[string]bool),
		dic:            dic,
	}
}

func (c *RestController) InitRestRoutes() {
	c.lc.Info("Registering v2 routes...")
	// common
	c.addReservedRoute(v2.ApiPingRoute, c.Ping).Methods(http.MethodGet)
	c.addReservedRoute(v2.ApiVersionRoute, c.Version).Methods(http.MethodGet)
	c.addReservedRoute(v2.ApiConfigRoute, c.Config).Methods(http.MethodGet)
	c.addReservedRoute(v2.ApiMetricsRoute, c.Metrics).Methods(http.MethodGet)
	// secret
	c.addReservedRoute(sdkCommon.APIV2SecretRoute, c.Secret).Methods(http.MethodPost)
	// discovery
	c.addReservedRoute(v2.ApiDiscoveryRoute, c.Discovery).Methods(http.MethodPost)
	// device command
	c.addReservedRoute(v2.ApiDeviceNameCommandNameRoute, c.Command).Methods(http.MethodPut, http.MethodGet)
	// callback
	c.addReservedRoute(v2.ApiDeviceCallbackRoute, c.AddDevice).Methods(http.MethodPost)
	c.addReservedRoute(v2.ApiDeviceCallbackRoute, c.UpdateDevice).Methods(http.MethodPut)
	c.addReservedRoute(v2.ApiDeviceCallbackNameRoute, c.DeleteDevice).Methods(http.MethodDelete)
	c.addReservedRoute(v2.ApiProfileCallbackRoute, c.UpdateProfile).Methods(http.MethodPut)
	c.addReservedRoute(v2.ApiWatcherCallbackRoute, c.AddProvisionWatcher).Methods(http.MethodPost)
	c.addReservedRoute(v2.ApiWatcherCallbackRoute, c.UpdateProvisionWatcher).Methods(http.MethodPut)
	c.addReservedRoute(v2.ApiWatcherCallbackNameRoute, c.DeleteProvisionWatcher).Methods(http.MethodDelete)
	c.addReservedRoute(v2.ApiServiceCallbackRoute, c.UpdateDeviceService).Methods(http.MethodPut)

	c.router.Use(correlation.RequestLimitMiddleware(container.ConfigurationFrom(c.dic.Get).Service.MaxRequestSize))
	c.router.Use(correlation.ManageHeader)
	c.router.Use(correlation.LoggingMiddleware(c.lc))
}

func (c *RestController) addReservedRoute(route string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	c.reservedRoutes[route] = true
	return c.router.HandleFunc(route, handler)
}

func (c *RestController) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) errors.EdgeX {
	if c.reservedRoutes[route] {
		return errors.NewCommonEdgeX(errors.KindServerError, "route is reserved", nil)
	}

	c.router.HandleFunc(route, handler).Methods(methods...)
	c.lc.Debug("Route added", "route", route, "methods", fmt.Sprintf("%v", methods))

	return nil
}

func (c *RestController) Router() *mux.Router {
	return c.router
}

// sendResponse puts together the response packet for the V2 API
func (c *RestController) sendResponse(
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

// sendEventResponse puts together the EventResponse packet for the V2 API
func (c *RestController) sendEventResponse(
	writer http.ResponseWriter,
	request *http.Request,
	response responses.EventResponse,
	statusCode int) {

	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)
	data, encoding, err := response.Encode()
	if err != nil {
		c.lc.Errorf("Unable to marshal EventResponse: %s; %s: %s", err.Error(), clients.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set(sdkCommon.CorrelationHeader, correlationID)
	writer.Header().Set(clients.ContentType, encoding)
	writer.WriteHeader(statusCode)

	_, err = writer.Write(data)
	if err != nil {
		c.lc.Errorf("Unable to write DeviceCommand response: %s; %s: %s", err.Error(), clients.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *RestController) sendEdgexError(
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
