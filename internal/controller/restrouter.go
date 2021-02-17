// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/controller/correlation"
	v2 "github.com/edgexfoundry/device-sdk-go/v2/internal/v2/controller/http"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2/v2"

	"github.com/gorilla/mux"
)

type RestController struct {
	LoggingClient    logger.LoggingClient
	router           *mux.Router
	reservedRoutes   map[string]bool
	v2HttpController *v2.V2HttpController
	dic              *di.Container
}

func NewRestController(r *mux.Router, dic *di.Container) *RestController {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	return &RestController{
		LoggingClient:    lc,
		router:           r,
		reservedRoutes:   make(map[string]bool),
		v2HttpController: v2.NewV2HttpController(dic),
		dic:              dic,
	}
}

func (c *RestController) InitRestRoutes() {
	c.LoggingClient.Info("Registering v2 routes...")
	// common
	c.addReservedRoute(contractsV2.ApiPingRoute, c.v2HttpController.Ping).Methods(http.MethodGet)
	c.addReservedRoute(contractsV2.ApiVersionRoute, c.v2HttpController.Version).Methods(http.MethodGet)
	c.addReservedRoute(contractsV2.ApiConfigRoute, c.v2HttpController.Config).Methods(http.MethodGet)
	c.addReservedRoute(contractsV2.ApiMetricsRoute, c.v2HttpController.Metrics).Methods(http.MethodGet)
	// secret
	c.addReservedRoute(sdkCommon.APIV2SecretRoute, c.v2HttpController.Secret).Methods(http.MethodPost)
	// discovery
	c.addReservedRoute(contractsV2.ApiDiscoveryRoute, c.v2HttpController.Discovery).Methods(http.MethodPost)
	// device command
	c.addReservedRoute(contractsV2.ApiDeviceNameCommandNameRoute, c.v2HttpController.Command).Methods(http.MethodPut, http.MethodGet)
	// callback
	c.addReservedRoute(contractsV2.ApiDeviceCallbackRoute, c.v2HttpController.AddDevice).Methods(http.MethodPost)
	c.addReservedRoute(contractsV2.ApiDeviceCallbackRoute, c.v2HttpController.UpdateDevice).Methods(http.MethodPut)
	c.addReservedRoute(contractsV2.ApiDeviceCallbackNameRoute, c.v2HttpController.DeleteDevice).Methods(http.MethodDelete)
	c.addReservedRoute(contractsV2.ApiProfileCallbackRoute, c.v2HttpController.UpdateProfile).Methods(http.MethodPut)
	c.addReservedRoute(contractsV2.ApiProvisionWatcherRoute, c.v2HttpController.AddProvisionWatcher).Methods(http.MethodPost)
	c.addReservedRoute(contractsV2.ApiProvisionWatcherRoute, c.v2HttpController.UpdateProvisionWatcher).Methods(http.MethodPut)
	c.addReservedRoute(contractsV2.ApiProvisionWatcherByNameRoute, c.v2HttpController.DeleteProvisionWatcher).Methods(http.MethodDelete)
	c.addReservedRoute(contractsV2.ApiServiceCallbackRoute, c.v2HttpController.UpdateDeviceService).Methods(http.MethodPut)

	c.router.Use(correlation.ManageHeader)
	c.router.Use(correlation.OnResponseComplete)
	c.router.Use(correlation.OnRequestBegin)
}

func (c *RestController) addReservedRoute(route string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	c.reservedRoutes[route] = true
	return c.router.HandleFunc(
		route,
		func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), bootstrapContainer.LoggingClientInterfaceName, c.LoggingClient)
			handler(
				w,
				r.WithContext(ctx))
		})
}

func (c *RestController) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) error {
	if c.reservedRoutes[route] {
		return errors.New("route is reserved")
	}

	c.router.HandleFunc(
		route,
		func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), bootstrapContainer.LoggingClientInterfaceName, c.LoggingClient)
			handler(
				w,
				r.WithContext(ctx))
		}).Methods(methods...)
	c.LoggingClient.Debug("Route added", "route", route, "methods", fmt.Sprintf("%v", methods))

	return nil
}

func (c *RestController) Router() *mux.Router {
	return c.router
}
