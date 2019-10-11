// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/controller/correlation"
	"github.com/gorilla/mux"
)

type RestController struct {
	router         *mux.Router
	reservedRoutes map[string]bool
}

func NewRestController() RestController {
	return RestController{
		router:         mux.NewRouter(),
		reservedRoutes: make(map[string]bool),
	}
}

func (c RestController) InitRestRoutes() {
	common.LoggingClient.Debug("init status rest controller")
	c.addReservedRoute(common.APIPingRoute, statusFunc).Methods(http.MethodGet)

	common.LoggingClient.Debug("init version rest controller")
	c.addReservedRoute(common.APIVersionRoute, versionFunc).Methods(http.MethodGet)

	common.LoggingClient.Debug("init command rest controller")
	c.addReservedRoute(common.APIAllCommandRoute, commandAllFunc).Methods(http.MethodGet, http.MethodPut)
	c.addReservedRoute(common.APIIdCommandRoute, commandFunc).Methods(http.MethodGet, http.MethodPut)
	c.addReservedRoute(common.APINameCommandRoute, commandFunc).Methods(http.MethodGet, http.MethodPut)

	common.LoggingClient.Debug("init callback rest controller")
	c.addReservedRoute(common.APICallbackRoute, callbackFunc)

	common.LoggingClient.Debug("init other rest controller")
	c.addReservedRoute(common.APIDiscoveryRoute, discoveryFunc).Methods(http.MethodPost)
	c.addReservedRoute(common.APITransformRoute, transformFunc).Methods(http.MethodGet)

	common.LoggingClient.Debug("init the metrics and config rest controller each")
	c.addReservedRoute(common.APIMetricsRoute, metricsHandler).Methods(http.MethodGet)
	c.addReservedRoute(common.APIConfigRoute, configHandler).Methods(http.MethodGet)

	c.router.Use(correlation.ManageHeader)
	c.router.Use(correlation.OnResponseComplete)
	c.router.Use(correlation.OnRequestBegin)
}

func (c RestController) addReservedRoute(route string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	c.reservedRoutes[route] = true
	return c.router.HandleFunc(route, handler)
}

func (c RestController) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) error {
	if c.reservedRoutes[route] {
		return errors.New("route is reserved")
	}

	c.router.HandleFunc(route, handler).Methods(methods...)
	common.LoggingClient.Debug("Route added", "route", route, "methods", fmt.Sprintf("%v", methods))

	return nil
}

func (c RestController) Router() *mux.Router {
	return c.router
}
