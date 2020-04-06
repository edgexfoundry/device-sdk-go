// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// DeviceServiceName contains the name of device service implementation in the DIC.
var DeviceServiceName = di.TypeInstanceToName(contract.DeviceService{})

// DeviceServiceFrom helper function queries the DIC and returns device service implementation.
func DeviceServiceFrom(get di.Get) *contract.DeviceService {
	return get(DeviceServiceName).(*contract.DeviceService)
}
