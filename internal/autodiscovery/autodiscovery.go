// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autodiscovery

import (
	"context"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

func BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	_ startup.Timer,
	dic *di.Container) bool {
	discovery := container.ProtocolDiscoveryFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	var runDiscovery bool = true

	if configuration.Device.Discovery.Enabled == false {
		lc.Info("AutoDiscovery stopped: disabled by configuration")
		runDiscovery = false
	}
	duration, err := time.ParseDuration(configuration.Device.Discovery.Interval)
	if err != nil || duration <= 0 {
		lc.Info("AutoDiscovery stopped: interval error in configuration")
		runDiscovery = false
	}
	if discovery == nil {
		lc.Info("AutoDiscovery stopped: ProtocolDiscovery not implemented")
		runDiscovery = false
	}

	if runDiscovery {
		go func() {
			wg.Add(1)
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(duration):
					lc.Debug("Auto-discovery triggered")
					handler.DiscoveryHandler(nil, discovery, lc)
				}
			}
		}()
	}

	return true
}
