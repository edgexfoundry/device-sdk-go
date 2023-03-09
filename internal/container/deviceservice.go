// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
)

// DeviceServiceName contains the name of device service struct in the DIC.
var DeviceServiceName = di.TypeInstanceToName(models.DeviceService{})

// ProtocolDiscoveryName contains the name of protocol discovery implementation in the DIC.
var ProtocolDiscoveryName = di.TypeInstanceToName((*interfaces.ProtocolDiscovery)(nil))

// ProtocolDriverName contains the name of protocol driver implementation in the DIC.
var ProtocolDriverName = di.TypeInstanceToName((*interfaces.ProtocolDriver)(nil))

// AutoEventManagerName contains the name of autoevent manager implementation in the DIC
var AutoEventManagerName = di.TypeInstanceToName((*interfaces.AutoEventManager)(nil))

// DeviceValidatorName contains the name of device validator implementation in the DIC.
var DeviceValidatorName = di.TypeInstanceToName((*interfaces.DeviceValidator)(nil))

// DeviceServiceFrom helper function queries the DIC and returns device service struct.
func DeviceServiceFrom(get di.Get) *models.DeviceService {
	return get(DeviceServiceName).(*models.DeviceService)
}

// ProtocolDiscoveryFrom helper function queries the DIC and returns protocol discovery implementation.
func ProtocolDiscoveryFrom(get di.Get) interfaces.ProtocolDiscovery {
	casted, ok := get(ProtocolDiscoveryName).(interfaces.ProtocolDiscovery)
	if ok {
		return casted
	}
	return nil
}

// DeviceValidatorFrom helper function queries the DIC and returns device validator implementation.
func DeviceValidatorFrom(get di.Get) interfaces.DeviceValidator {
	casted, ok := get(DeviceValidatorName).(interfaces.DeviceValidator)
	if ok {
		return casted
	}
	return nil
}

// ProtocolDriverFrom helper function queries the DIC and returns protocol driver implementation.
func ProtocolDriverFrom(get di.Get) interfaces.ProtocolDriver {
	return get(ProtocolDriverName).(interfaces.ProtocolDriver)
}

// AutoEventManagerFrom helper function queries the DIC and returns autoevent manager implementation
func AutoEventManagerFrom(get di.Get) interfaces.AutoEventManager {
	return get(AutoEventManagerName).(interfaces.AutoEventManager)
}
