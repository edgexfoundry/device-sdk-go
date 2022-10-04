// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"os"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/controller/messaging"
)

const EnvInstanceName = "EDGEX_INSTANCE_NAME"

var instanceName string

func Main(serviceName string, serviceVersion string, proto interface{}, ctx context.Context, cancel context.CancelFunc, router *mux.Router) {
	startupTimer := startup.NewStartUpTimer(serviceName)

	additionalUsage :=
		"    -i, --instance                  Provides a service name suffix which allows unique instance to be created\n" +
			"                                    If the option is provided, service name will be replaced with \"<name>_<instance>\"\n"
	sdkFlags := flags.NewWithUsage(additionalUsage)
	sdkFlags.FlagSet.StringVar(&instanceName, "instance", "", "")
	sdkFlags.FlagSet.StringVar(&instanceName, "i", "", "")
	sdkFlags.Parse(os.Args[1:])

	serviceName = setServiceName(serviceName)
	ds = &DeviceService{}
	ds.Initialize(serviceName, serviceVersion, proto)

	ds.flags = sdkFlags

	ds.dic = di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return ds.config
		},
		container.DeviceServiceName: func(get di.Get) interface{} {
			return ds.deviceService
		},
		container.ProtocolDriverName: func(get di.Get) interface{} {
			return ds.driver
		},
		container.ProtocolDiscoveryName: func(get di.Get) interface{} {
			return ds.discovery
		},
		container.DeviceValidatorName: func(get di.Get) interface{} {
			return ds.validator
		},
	})

	httpServer := handlers.NewHttpServer(router, true)

	bootstrap.Run(
		ctx,
		cancel,
		sdkFlags,
		ds.ServiceName,
		common.ConfigStemDevice,
		ds.config,
		startupTimer,
		ds.dic,
		true,
		[]interfaces.BootstrapHandler{
			httpServer.BootstrapHandler,
			messageBusBootstrapHandler,
			handlers.NewServiceMetrics(ds.ServiceName).BootstrapHandler, // Must be after Messaging
			clientBootstrapHandler,
			autoevent.BootstrapHandler,
			NewBootstrap(router).BootstrapHandler,
			autodiscovery.BootstrapHandler,
			handlers.NewStartMessage(serviceName, serviceVersion).BootstrapHandler,
		})

	ds.Stop(false)
}

func clientBootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	configuration := container.ConfigurationFrom(dic.Get)
	if configuration.Device.UseMessageBus {
		delete(configuration.Clients, common.CoreDataServiceKey)
	}

	if !handlers.NewClientsBootstrap().BootstrapHandler(ctx, wg, startupTimer, dic) {
		return false
	}

	return true
}

func messageBusBootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	configuration := container.ConfigurationFrom(dic.Get)
	if configuration.Device.UseMessageBus {
		if !handlers.MessagingBootstrapHandler(ctx, wg, startupTimer, dic) {
			return false
		}
		err := messaging.SubscribeCommands(ctx, dic)
		if err != nil {
			lc := bootstrapContainer.LoggingClientFrom(dic.Get)
			lc.Errorf("Failed to subscribe internal command request: %v", err)
			return false
		}
	}

	return true
}

func setServiceName(name string) string {
	envValue := os.Getenv(EnvInstanceName)
	if len(envValue) > 0 {
		instanceName = envValue
	}

	if len(instanceName) > 0 {
		name = name + "_" + instanceName
	}

	return name
}
