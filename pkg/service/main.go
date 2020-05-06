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
	"github.com/edgexfoundry/device-sdk-go/internal/controller"
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

var serviceNameOverride string

func Main(serviceName string, serviceVersion string, proto interface{}, ctx context.Context, cancel context.CancelFunc, router *mux.Router, readyStream chan<- bool) {
	startupTimer := startup.NewStartUpTimer(common.BootRetrySecondsDefault, common.BootTimeoutSecondsDefault)

	additionalUsage :=
		"    -n, --serviceName               Overrides the service name to be stored in metadata, used with Registry and/or Configuration Providers\n" +
			"                                    If the profile is also provided, name will be replaced with \"<name>;<profile>\"\n"
	sdkFlags := flags.NewWithUsage(additionalUsage)
	sdkFlags.FlagSet.StringVar(&serviceNameOverride, "serviceName", "", "")
	sdkFlags.FlagSet.StringVar(&serviceNameOverride, "n", "", "")
	sdkFlags.Parse(os.Args[1:])

	serviceName = setServiceName(serviceName, sdkFlags.Profile())
	if serviceName == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify device service name")
		os.Exit(1)
	}
	common.ServiceName = serviceName
	if serviceVersion == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify device service version")
		os.Exit(1)
	}
	common.ServiceVersion = serviceVersion
	if driver, ok := proto.(dsModels.ProtocolDriver); ok {
		common.Driver = driver
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Please implement and specify the protocoldriver")
		os.Exit(1)
	}
	if discovery, ok := proto.(dsModels.ProtocolDiscovery); ok {
		common.Discovery = discovery
	} else {
		common.Discovery = nil
	}

	configuration := &common.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	httpServer := httpserver.NewBootstrap(router, true)
	controller.LoadRestRoutes(router, dic)

	bootstrap.Run(
		ctx,
		cancel,
		sdkFlags,
		serviceName,
		common.ConfigStemDevice+common.ConfigMajorVersion,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			httpServer.BootstrapHandler,
			NewBootstrap(router).BootstrapHandler,
			message.NewBootstrap(serviceName, serviceVersion).BootstrapHandler,
		})

	svc.Stop(false)
}

func setServiceName(name string, profile string) string {
	envValue := os.Getenv(common.EnvServiceName)
	if len(envValue) > 0 {
		serviceNameOverride = envValue
	}

	if len(serviceNameOverride) > 0 {
		name = serviceNameOverride
	}

	if len(profile) > 0 {
		name = name + ";" + profile
	}

	return name
}
