// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autodiscovery

import (
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func Run(discovery models.ProtocolDiscovery, lc logger.LoggingClient, configuration *common.ConfigurationStruct) {
	enabled := configuration.Device.Discovery.Enabled
	if !enabled {
		lc.Info("AutoDiscovery stopped: disabled by configuration")
		return
	}
	duration, err := time.ParseDuration(configuration.Device.Discovery.Interval)
	if err != nil || duration <= 0 {
		lc.Info("AutoDiscovery stopped: interval error in configuration")
		return
	}
	if discovery == nil {
		lc.Info("AutoDiscovery stopped: ProtocolDiscovery not implemented")
		return
	}

	for {
		time.Sleep(duration)

		lc.Debug("Auto-discovery triggered")
		handler.DiscoveryHandler(nil, discovery, lc)
	}
}
