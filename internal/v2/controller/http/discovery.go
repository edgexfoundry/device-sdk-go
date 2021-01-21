// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
)

func (c *V2HttpController) Discovery(writer http.ResponseWriter, request *http.Request) {
	ds := container.DeviceServiceFrom(c.dic.Get)
	if ds.AdminState == contract.Locked {
		err := edgexErr.NewCommonEdgeX(edgexErr.KindServiceLocked, "service locked", nil)
		c.sendEdgexError(writer, request, err, v2.ApiDiscoveryRoute)
		return
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	if !configuration.Device.Discovery.Enabled {
		err := edgexErr.NewCommonEdgeX(edgexErr.KindServiceUnavailable, "device discovery disabled", nil)
		c.sendEdgexError(writer, request, err, v2.ApiDiscoveryRoute)
		return
	}

	discovery := container.ProtocolDiscoveryFrom(c.dic.Get)
	if discovery == nil {
		err := edgexErr.NewCommonEdgeX(edgexErr.KindNotImplemented, "protocolDiscovery not implemented", nil)
		c.sendEdgexError(writer, request, err, v2.ApiDiscoveryRoute)
		return
	}

	go autodiscovery.DiscoveryWrapper(discovery, c.lc)
	c.sendResponse(writer, request, v2.ApiDiscoveryRoute, nil, http.StatusAccepted)
}
