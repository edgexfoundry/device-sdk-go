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
	"github.com/gorilla/mux"
)

type RestController struct {
	router         *mux.Router
	reservedRoutes map[string]bool
}

func NewRestController(r *mux.Router) RestController {
	return RestController{
		router:         r,
		reservedRoutes: make(map[string]bool),
	}
}

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	c := NewRestController(r)
	c.InitRestRoutes()
	dic.Update(di.ServiceConstructorMap{
		"RestController": func(dic di.Get) interface{} {
			return c
		},
	})
}

func (c RestController) InitRestRoutes() {
	// Status
	c.addReservedRoute(common.APIPingRoute, statusFunc).Methods(http.MethodGet)
	// Version
	c.addReservedRoute(common.APIVersionRoute, versionFunc).Methods(http.MethodGet)
	// Command
	c.addReservedRoute(common.APIAllCommandRoute, commandAllFunc).Methods(http.MethodGet, http.MethodPut)
	c.addReservedRoute(common.APIIdCommandRoute, commandFunc).Methods(http.MethodGet, http.MethodPut)
	c.addReservedRoute(common.APINameCommandRoute, commandFunc).Methods(http.MethodGet, http.MethodPut)
	// Callback
	c.addReservedRoute(common.APICallbackRoute, callbackFunc)
	// Discovery and Transform
	c.addReservedRoute(common.APIDiscoveryRoute, discoveryFunc).Methods(http.MethodPost)
	c.addReservedRoute(common.APITransformRoute, transformFunc).Methods(http.MethodGet)
	// Metric and Config
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
