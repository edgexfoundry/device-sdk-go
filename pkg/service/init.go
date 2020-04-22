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
	"os"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/clients"
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

func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) (success bool) {
	common.CurrentConfig = container.ConfigurationFrom(dic.Get)
	common.LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)
	common.RegistryClient = bootstrapContainer.RegistryFrom(dic.Get)

	// init svc and autoevent manager in the beginning so that if there's
	// error in following bootstrap process the device service can correctly
	// call svc.Stop and gracefully shut down.
	svc = newService(dic)
	autoevent.NewManager(ctx, wg)

	if svc.svcInfo.EnableAsyncReadings {
		svc.asyncCh = make(chan *dsModels.AsyncValues, svc.svcInfo.AsyncBufferSize)
		go processAsyncResults(ctx, wg)
	}

	svc.deviceCh = make(chan []dsModels.DiscoveredDevice)
	go processAsyncFilterAndAdd(ctx, wg)

	err := clients.InitDependencyClients(ctx, wg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return false
	}

	err = selfRegister()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Couldn't register to metadata service: %v\n", err)
		return false
	}

	// initialize devices, deviceResources, provisionwatcheres & profiles
	cache.InitCache()

	err = common.Driver.Initialize(common.LoggingClient, svc.asyncCh, svc.deviceCh)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Driver.Initialize failed: %v\n", err)
		return false
	}
	svc.initiazlied = true

	err = provision.LoadProfiles(common.CurrentConfig.Device.ProfilesDir)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to create the pre-defined Device Profiles: %v\n", err)
		return false
	}

	err = provision.LoadDevices(common.CurrentConfig.DeviceList)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to create the pre-defined Devices: %v\n", err)
		return false
	}

	go autodiscovery.Run()
	autoevent.GetManager().StartAutoEvents()
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(common.CurrentConfig.Service.Timeout), "Request timed out")

	return true
}
