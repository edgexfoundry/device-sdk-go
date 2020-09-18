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

// DeviceServiceName contains the name of device service struct in the DIC.
var DeviceServiceName = di.TypeInstanceToName(contract.DeviceService{})

// ProtocolDiscoveryName contains the name of protocol discovery implementation in the DIC.
var ProtocolDiscoveryName = di.TypeInstanceToName((*models.ProtocolDiscovery)(nil))

// ProtocolDriverName contains the name of protocol driver implementation in the DIC.
var ProtocolDriverName = di.TypeInstanceToName((*models.ProtocolDriver)(nil))

// DeviceServiceFrom helper function queries the DIC and returns device service struct.
func DeviceServiceFrom(get di.Get) contract.DeviceService {
	return get(DeviceServiceName).(contract.DeviceService)
}

// ProtocolDiscoveryFrom helper function queries the DIC and returns protocol discovery implementation.
func ProtocolDiscoveryFrom(get di.Get) models.ProtocolDiscovery {
	casted, ok := get(ProtocolDiscoveryName).(models.ProtocolDiscovery)
	if ok {
		return casted
	}
	return nil
}

// ProtocolDriverFrom helper function queries the DIC and returns protocol driver implementation.
func ProtocolDriverFrom(get di.Get) models.ProtocolDriver {
	return get(ProtocolDriverName).(models.ProtocolDriver)
}
