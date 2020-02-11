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
)

func Run() {
	enabled := common.CurrentConfig.Device.Discovery.Enabled
	if !enabled {
		common.LoggingClient.Info("AutoDiscovery stopped: disabled by configuration")
		return
	}
	duration, err := time.ParseDuration(common.CurrentConfig.Device.Discovery.Interval)
	if err != nil || duration <= 0 {
		common.LoggingClient.Info("AutoDiscovery stopped: interval error in configuration")
		return
	}
	if common.Discovery == nil {
		common.LoggingClient.Info("AutoDiscovery stopped: ProtocolDiscovery not implemented")
		return
	}

	for {
		time.Sleep(duration)

		common.LoggingClient.Debug("Auto-discovery triggered")
		handler.DiscoveryHandler(nil)
	}
}
