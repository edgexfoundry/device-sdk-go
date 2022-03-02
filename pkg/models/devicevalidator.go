//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "github.com/edgexfoundry/go-mod-core-contracts/v2/models"

// DeviceValidator is a low-level device-specific interface implemented
// by device services that validate device's protocol properties.
type DeviceValidator interface {
	// ValidateDevice triggers device's protocol properties validation, returns error
	// if validation failed and the incoming device will not be added into EdgeX.
	ValidateDevice(device models.Device) error
}
