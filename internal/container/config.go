// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

// ConfigurationName contains the name of device service's ConfigurationStruct implementation in the DIC.
var ConfigurationName = di.TypeInstanceToName(common.ConfigurationStruct{})

// ConfigurationFrom helper function queries the DIC and returns device service's ConfigurationStruct implementation.
func ConfigurationFrom(get di.Get) *common.ConfigurationStruct {
	return get(ConfigurationName).(*common.ConfigurationStruct)
}
