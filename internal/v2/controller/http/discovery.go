// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/autodiscovery"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/errors"
)

func (c *V2HttpController) Discovery(writer http.ResponseWriter, request *http.Request) {
	ds := container.DeviceServiceFrom(c.dic.Get)
	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)

	err := checkServiceLocked(request, ds.AdminState)
	if err != nil {
		c.sendError(writer, request, edgexErr.KindServiceLocked, "Service Locked", err, sdkCommon.APIV2DiscoveryRoute, correlationID)
		return
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	if !configuration.Device.Discovery.Enabled {
		c.sendError(writer, request, edgexErr.KindServiceUnavailable, "Device discovery disabled", nil, sdkCommon.APIV2DiscoveryRoute, correlationID)
		return
	}

	discovery := container.ProtocolDiscoveryFrom(c.dic.Get)
	if discovery == nil {
		c.sendError(writer, request, edgexErr.KindNotImplemented, "ProtocolDiscovery not implemented", nil, sdkCommon.APIV2DiscoveryRoute, correlationID)
		return
	}

	go autodiscovery.DiscoveryWrapper(discovery, c.lc)
	c.sendResponse(writer, request, sdkCommon.APIV2DiscoveryRoute, nil, http.StatusAccepted)
}
