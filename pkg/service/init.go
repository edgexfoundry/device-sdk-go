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

	"github.com/edgexfoundry/device-sdk-go/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
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

func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) (success bool) {
	// TODO: remove these after refactor are done.
	common.CurrentConfig = container.ConfigurationFrom(dic.Get)
	common.LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)
	common.RegistryClient = bootstrapContainer.RegistryFrom(dic.Get)

	ds.UpdateFromContainer(b.router, dic)
	ds.controller.InitRestRoutes(dic)

	lc := ds.LoggingClient
	configuration := ds.config
	autoevent.NewManager(ctx, wg)

	if ds.AsyncReadings() {
		ds.asyncCh = make(chan *dsModels.AsyncValues, configuration.Service.AsyncBufferSize)
		go ds.processAsyncResults(ctx, wg)
	}
	if ds.DeviceDiscovery() {
		ds.deviceCh = make(chan []dsModels.DiscoveredDevice)
		go ds.processAsyncFilterAndAdd(ctx, wg)
	}

	err := ds.selfRegister()
	if err != nil {
		lc.Error(fmt.Sprintf("Couldn't register to metadata service: %v\n", err))
		return false
	}

	// initialize devices, deviceResources, provisionwatcheres & profiles cache
	cache.InitCache()

	err = ds.driver.Initialize(lc, ds.asyncCh, ds.deviceCh)
	if err != nil {
		lc.Error(fmt.Sprintf("Driver.Initialize failed: %v\n", err))
		return false
	}
	ds.initialized = true

	dic.Update(di.ServiceConstructorMap{
		container.DeviceServiceName: func(get di.Get) interface{} {
			return ds.deviceService
		},
		container.ProtocolDiscoveryName: func(get di.Get) interface{} {
			return ds.discovery
		},
		container.ProtocolDriverName: func(get di.Get) interface{} {
			return ds.driver
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

	autoevent.GetManager().StartAutoEvents(dic)
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration.Service.Timeout), "Request timed out")

	return true
}
