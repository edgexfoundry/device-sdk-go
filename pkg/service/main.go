// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"os"

	"github.com/edgexfoundry/device-sdk-go/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/internal/clients"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/httpserver"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/message"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/gorilla/mux"
)

var instanceName string

func Main(serviceName string, serviceVersion string, proto interface{}, ctx context.Context, cancel context.CancelFunc, router *mux.Router, readyStream chan<- bool) {
	startupTimer := startup.NewStartUpTimer(serviceName)

	additionalUsage :=
		"    -i, --instance                  Provides a service name suffix which allows unique instance to be created\n" +
			"                                    If the option is provided, service name will be replaced with \"<name>_<instance>\"\n"
	sdkFlags := flags.NewWithUsage(additionalUsage)
	sdkFlags.FlagSet.StringVar(&instanceName, "instance", "", "")
	sdkFlags.FlagSet.StringVar(&instanceName, "i", "", "")
	sdkFlags.Parse(os.Args[1:])

	serviceName = setServiceName(serviceName, sdkFlags.Profile())
	ds = &DeviceService{}
	ds.Initialize(serviceName, serviceVersion, proto)

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return ds.config
		},
	})

	httpServer := httpserver.NewBootstrap(router, true)

	bootstrap.Run(
		ctx,
		cancel,
		sdkFlags,
		ds.ServiceName,
		common.ConfigStemDevice+common.ConfigMajorVersion,
		ds.config,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			httpServer.BootstrapHandler,
			clients.NewClients().BootstrapHandler,
			NewBootstrap(router).BootstrapHandler,
			autodiscovery.BootstrapHandler,
			message.NewBootstrap(serviceName, serviceVersion).BootstrapHandler,
		})

	ds.Stop(false)
}

func setServiceName(name string, profile string) string {
	envValue := os.Getenv(common.EnvInstanceName)
	if len(envValue) > 0 {
		instanceName = envValue
	}

	if len(instanceName) > 0 {
		name = name + "_" + instanceName
	}

	return name
}
