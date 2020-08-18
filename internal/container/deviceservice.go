// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// DeviceServiceName contains the name of device service implementation in the DIC.
var DeviceServiceName = di.TypeInstanceToName(contract.DeviceService{})
var ProtocolDiscoveryName = di.TypeInstanceToName((*models.ProtocolDiscovery)(nil))
var ProtocolDriverName = di.TypeInstanceToName((*models.ProtocolDriver)(nil))

// DeviceServiceFrom helper function queries the DIC and returns device service implementation.
func DeviceServiceFrom(get di.Get) contract.DeviceService {
	return get(DeviceServiceName).(contract.DeviceService)
}

func ProtocolDiscoveryFrom(get di.Get) models.ProtocolDiscovery {
	return get(ProtocolDiscoveryName).(models.ProtocolDiscovery)
}

func ProtocolDriverFrom(get di.Get) models.ProtocolDriver {
	return get(ProtocolDriverName).(models.ProtocolDriver)
}
