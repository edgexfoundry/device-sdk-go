// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"os"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/httpserver"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/message"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/gorilla/mux"
)

func Main(serviceName string, serviceVersion string, driver dsModels.ProtocolDriver, ctx context.Context, cancel context.CancelFunc, router *mux.Router, readyStream chan<- bool) {
	startupTimer := startup.NewStartUpTimer(common.BootRetrySecondsDefault, common.BootTimeoutSecondsDefault)

	if len(serviceName) == 0 || len(serviceVersion) == 0 || driver == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify correct device service informations")
		os.Exit(1)
	}
	common.Driver = driver
	common.ServiceName = serviceName
	common.ServiceVersion = serviceVersion

	f := flags.New()
	f.Parse(os.Args[1:])

	configuration := &common.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	httpServer := httpserver.NewBootstrap(router, true)

	bootstrap.Run(
		ctx,
		cancel,
		f,
		serviceName,
		common.ConfigStemCore+common.ConfigMajorVersion,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			// secret.NewSecret().BootstrapHandler,
			NewBootstrap(router).BootstrapHandler,
			// telemetry.BootstrapHandler,
			httpServer.BootstrapHandler,
			message.NewBootstrap(serviceName, serviceVersion).BootstrapHandler,
		})

	svc.Stop(false)
}
