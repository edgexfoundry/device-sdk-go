// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/gorilla/mux"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router *mux.Router
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *mux.Router) *Bootstrap {
	return &Bootstrap{
		router: router,
	}
}

func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) (success bool) {
	sdk.UpdateFromContainer(b.router, dic)
	sdk.controller.InitRestRoutes(dic)

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	// init autoevent manager in the beginning so that if there's error
	// in following bootstrap process the device service can correctly
	// call svc.Stop and gracefully shut down.
	autoevent.NewManager(ctx, wg, dic)

	// invoke async goroutine
	if sdk.AsyncReadings() {
		sdk.asyncCh = make(chan *dsModels.AsyncValues, sdk.svcInfo.AsyncBufferSize)
		go sdk.processAsyncResults(ctx, wg)
	}

	sdk.deviceCh = make(chan []dsModels.DiscoveredDevice)
	go sdk.processAsyncFilterAndAdd(ctx, wg)

	err := sdk.selfRegister()
	if err != nil {
		lc.Error(fmt.Sprintf("Couldn't register to metadata service: %v\n", err))
		return false
	}

	// initialize devices, deviceResources, provisionwatcheres & profiles
	cache.InitCache(
		sdk.ServiceName,
		bootstrapContainer.LoggingClientFrom(dic.Get),
		container.CoredataValueDescriptorClientFrom(dic.Get),
		container.MetadataDeviceClientFrom(dic.Get),
		container.MetadataProvisionWatcherClientFrom(dic.Get))

	err = sdk.driver.Initialize(sdk.LoggingClient, sdk.asyncCh, sdk.deviceCh)
	if err != nil {
		lc.Error(fmt.Sprintf("Driver.Initialize failed: %v\n", err))
		return false
	}
	sdk.initialized = true

	dic.Update(di.ServiceConstructorMap{
		container.DeviceServiceName: func(get di.Get) interface{} {
			return sdk.deviceService
		},
		container.ProtocolDiscoveryName: func(get di.Get) interface{} {
			return sdk.discovery
		},
		container.ProtocolDriverName: func(get di.Get) interface{} {
			return sdk.driver
		},
	})

	err = provision.LoadProfiles(configuration.Device.ProfilesDir, dic)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to create the pre-defined device profiles: %v\n", err))
		return false
	}

	err = provision.LoadDevices(configuration.DeviceList, dic)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to create the pre-defined devices: %v\n", err))
		return false
	}

	go autodiscovery.Run(sdk.discovery, sdk.LoggingClient, sdk.config)
	autoevent.GetManager().StartAutoEvents()
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.Service.Timeout), "Request timed out")

	return true
}
