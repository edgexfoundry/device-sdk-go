// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package startup

import (
	"context"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Bootstrap(serviceName string, serviceVersion string, driver interface{}) {
	ctx, cancel := context.WithCancel(context.Background())
	//add promethues metrics，modified by jacktian
	r := mux.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	service.Main(serviceName, serviceVersion, driver, ctx, cancel, mux.NewRouter())
}
