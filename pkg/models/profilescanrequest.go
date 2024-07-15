//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "github.com/edgexfoundry/go-mod-core-contracts/v3/models"

// ProfileScanRequest is the struct for requesting a profile for a specified device.
type ProfileScanRequest struct {
	DeviceName  string
	ProfileName string
	Options     any
	Protocols   map[string]models.ProtocolProperties
}
