// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/controller/correlation"
	"github.com/gorilla/mux"
)

func InitRestRoutes() *mux.Router {
	r := mux.NewRouter()

	common.LoggingClient.Debug("init status rest controller")
	r.HandleFunc(common.APIPingRoute, statusFunc).Methods(http.MethodGet)

	common.LoggingClient.Debug("init version rest controller")
	r.HandleFunc(common.APIVersionRoute, versionFunc).Methods(http.MethodGet)

	common.LoggingClient.Debug("init command rest controller")
	r.HandleFunc(common.APIAllCommandRoute, commandAllFunc).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc(common.APIIdCommandRoute, commandFunc).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc(common.APINameCommandRoute, commandFunc).Methods(http.MethodGet, http.MethodPut)

	common.LoggingClient.Debug("init callback rest controller")
	r.HandleFunc(common.APICallbackRoute, callbackFunc)

	common.LoggingClient.Debug("init other rest controller")
	r.HandleFunc(common.APIDiscoveryRoute, discoveryFunc).Methods(http.MethodPost)
	r.HandleFunc(common.APITransformRoute, transformFunc).Methods(http.MethodGet)

	common.LoggingClient.Debug("init the metrics and config rest controller each")
	r.HandleFunc(common.APIMetricsRoute, metricsHandler).Methods(http.MethodGet)
	r.HandleFunc(common.APIConfigRoute, configHandler).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}
