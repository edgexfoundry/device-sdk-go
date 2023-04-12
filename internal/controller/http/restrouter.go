// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/controller/http/correlation"
)

type RestController struct {
	serviceName    string
	router         *mux.Router
	reservedRoutes map[string]bool
	customConfig   interfaces.UpdatableConfig
	lc             logger.LoggingClient
	dic            *di.Container
}

func NewRestController(r *mux.Router, dic *di.Container, serviceName string) *RestController {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	return &RestController{
		lc:             lc,
		router:         r,
		reservedRoutes: make(map[string]bool),
		dic:            dic,
		serviceName:    serviceName,
	}
}

// SetCustomConfigInfo sets the custom configuration, which is used to include the service's custom config in the /config endpoint response.
func (c *RestController) SetCustomConfigInfo(customConfig interfaces.UpdatableConfig) {
	c.customConfig = customConfig
}

func (c *RestController) InitRestRoutes() {
	c.lc.Info("Registering v2 routes...")
	// router.UseEncodedPath() tells the router to match the encoded original path to the routes
	c.router.UseEncodedPath()

	lc := container.LoggingClientFrom(c.dic.Get)
	secretProvider := container.SecretProviderExtFrom(c.dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// common
	c.addReservedRoute(common.ApiPingRoute, c.Ping).Methods(http.MethodGet)
	c.addReservedRoute(common.ApiVersionRoute, authenticationHook(c.Version)).Methods(http.MethodGet)
	c.addReservedRoute(common.ApiConfigRoute, authenticationHook(c.Config)).Methods(http.MethodGet)
	// secret
	c.addReservedRoute(common.ApiSecretRoute, authenticationHook(c.Secret)).Methods(http.MethodPost)
	// discovery
	c.addReservedRoute(common.ApiDiscoveryRoute, authenticationHook(c.Discovery)).Methods(http.MethodPost)
	// device command
	c.addReservedRoute(common.ApiDeviceNameCommandNameRoute, authenticationHook(c.GetCommand)).Methods(http.MethodGet)
	c.addReservedRoute(common.ApiDeviceNameCommandNameRoute, authenticationHook(c.SetCommand)).Methods(http.MethodPut)

	c.router.Use(correlation.ManageHeader)
	c.router.Use(correlation.LoggingMiddleware(c.lc))
	c.router.Use(correlation.UrlDecodeMiddleware(c.lc))
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

	correlationID := request.Header.Get(common.CorrelationHeader)

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	if response != nil {
		data, err := json.Marshal(response)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), common.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(data)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), common.CorrelationHeader, correlationID)
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

	correlationID := request.Header.Get(common.CorrelationHeader)
	data, encoding, err := response.Encode()
	if err != nil {
		c.lc.Errorf("Unable to marshal EventResponse: %s; %s: %s", err.Error(), common.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, encoding)
	writer.WriteHeader(statusCode)

	_, err = writer.Write(data)
	if err != nil {
		c.lc.Errorf("Unable to write DeviceCommand response: %s; %s: %s", err.Error(), common.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *RestController) sendEdgexError(
	writer http.ResponseWriter,
	request *http.Request,
	err errors.EdgeX,
	api string) {
	correlationID := request.Header.Get(common.CorrelationHeader)
	c.lc.Error(err.Error(), common.CorrelationHeader, correlationID)
	c.lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationID)
	response := commonDTO.NewBaseResponse("", err.Error(), err.Code())
	c.sendResponse(writer, request, api, response, err.Code())
}
