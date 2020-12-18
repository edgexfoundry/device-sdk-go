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
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	v2cache "github.com/edgexfoundry/device-sdk-go/internal/v2/cache"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
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
	ds.UpdateFromContainer(b.router, dic)
	autoevent.NewManager(ctx, wg, ds.config.Service.AsyncBufferSize, dic)

	err := ds.selfRegister()
	if err != nil {
		ds.LoggingClient.Error(fmt.Sprintf("Couldn't register to metadata service: %v\n", err))
		return false
	}

	// initialize devices, deviceResources, provisionWatchers & profiles cache
	cache.InitCache(
		ds.ServiceName,
		ds.LoggingClient,
		container.CoredataValueDescriptorClientFrom(dic.Get),
		container.MetadataDeviceClientFrom(dic.Get),
		container.MetadataProvisionWatcherClientFrom(dic.Get))
	v2cache.InitV2Cache()

	if ds.AsyncReadings() {
		ds.asyncCh = make(chan *dsModels.AsyncValues, ds.config.Service.AsyncBufferSize)
		go ds.processAsyncResults(ctx, wg)
	}
	if ds.DeviceDiscovery() {
		ds.deviceCh = make(chan []dsModels.DiscoveredDevice, 1)
		go ds.processAsyncFilterAndAdd(ctx, wg)
	}

	err = ds.driver.Initialize(ds.LoggingClient, ds.asyncCh, ds.deviceCh)
	if err != nil {
		ds.LoggingClient.Error(fmt.Sprintf("Driver.Initialize failed: %v\n", err))
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

	ds.controller.InitRestRoutes()

	err = provision.LoadProfiles(ds.config.Device.ProfilesDir, dic)
	if err != nil {
		ds.LoggingClient.Error(fmt.Sprintf("Failed to create the pre-defined device profiles: %v\n", err))
		return false
	}

	err = provision.LoadDevices(ds.config.DeviceList, dic)
	if err != nil {
		ds.LoggingClient.Error(fmt.Sprintf("Failed to create the pre-defined devices: %v\n", err))
		return false
	}

	autoevent.GetManager().StartAutoEvents(dic)
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(ds.config.Service.Timeout), "Request timed out")

	return true
}
