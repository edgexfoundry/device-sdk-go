// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
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
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/gorilla/mux"
)

type RestController struct {
	LoggingClient  logger.LoggingClient
	router         *mux.Router
	reservedRoutes map[string]bool
}

func NewRestController(r *mux.Router, lc logger.LoggingClient) *RestController {
	return &RestController{
		LoggingClient:  lc,
		router:         r,
		reservedRoutes: make(map[string]bool),
	}
}

func (c RestController) InitRestRoutes(dic *di.Container) {
	// Callback
	c.addReservedRoute(common.APICallbackRoute, c.callbackFunc, dic)
	// Command
	c.addReservedRoute(common.APIAllCommandRoute, c.commandAllFunc, dic).Methods(http.MethodGet, http.MethodPut)
	c.addReservedRoute(common.APIIdCommandRoute, c.commandFunc, dic).Methods(http.MethodGet, http.MethodPut)
	c.addReservedRoute(common.APINameCommandRoute, c.commandFunc, dic).Methods(http.MethodGet, http.MethodPut)
	// Discovery and Transform
	c.addReservedRoute(common.APIDiscoveryRoute, c.discoveryFunc, dic).Methods(http.MethodPost)
	c.addReservedRoute(common.APITransformRoute, c.transformFunc, dic).Methods(http.MethodGet)
	// Status
	c.addReservedRoute(common.APIPingRoute, c.statusFunc, dic).Methods(http.MethodGet)
	// Version
	c.addReservedRoute(common.APIVersionRoute, c.versionFunc, dic).Methods(http.MethodGet)
	// Metric and Config
	c.addReservedRoute(common.APIMetricsRoute, c.metricsFunc, dic).Methods(http.MethodGet)
	c.addReservedRoute(common.APIConfigRoute, c.configFunc, dic).Methods(http.MethodGet)

	c.router.Use(correlation.ManageHeader)
	c.router.Use(correlation.OnResponseComplete)
	c.router.Use(correlation.OnRequestBegin)
}

func (c RestController) addReservedRoute(route string, handler func(http.ResponseWriter, *http.Request, *di.Container), dic *di.Container) *mux.Route {
	c.reservedRoutes[route] = true
	return c.router.HandleFunc(
		route,
		func(w http.ResponseWriter, r *http.Request) {
			handler(
				w,
				r,
				dic)
		})
}

func (c RestController) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) error {
	if c.reservedRoutes[route] {
		return errors.New("route is reserved")
	}

	c.router.HandleFunc(route, handler).Methods(methods...)
	c.LoggingClient.Debug("Route added", "route", route, "methods", fmt.Sprintf("%v", methods))

	return nil
}

func (c RestController) Router() *mux.Router {
	return c.router
}
