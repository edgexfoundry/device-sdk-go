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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func (c *V2HttpController) Discovery(writer http.ResponseWriter, request *http.Request) {
	ds := container.DeviceServiceFrom(c.dic.Get)
	if ds.AdminState == contract.Locked {
		err := edgexErr.NewCommonEdgeX(edgexErr.KindServiceLocked, "service locked", nil)
		c.sendEdgexError(writer, request, err, sdkCommon.APIV2DiscoveryRoute)
		return
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	if !configuration.Device.Discovery.Enabled {
		err := edgexErr.NewCommonEdgeX(edgexErr.KindServiceUnavailable, "device discovery disabled", nil)
		c.sendEdgexError(writer, request, err, sdkCommon.APIV2DiscoveryRoute)
		return
	}

	discovery := container.ProtocolDiscoveryFrom(c.dic.Get)
	if discovery == nil {
		err := edgexErr.NewCommonEdgeX(edgexErr.KindNotImplemented, "protocolDiscovery not implemented", nil)
		c.sendEdgexError(writer, request, err, sdkCommon.APIV2DiscoveryRoute)
		return
	}

	go autodiscovery.DiscoveryWrapper(discovery, c.lc)
	c.sendResponse(writer, request, sdkCommon.APIV2DiscoveryRoute, nil, http.StatusAccepted)
}
