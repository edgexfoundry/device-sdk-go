//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// ProfileScanRequest is the struct for requesting a profile for a specified device.
type ProfileScanRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	DeviceName            string                               `json:"deviceName" validate:"required"`
	ProfileName           string                               `json:"profileName"`
	Options               any                                  `json:"options"`
	Protocols             map[string]models.ProtocolProperties `json:"protocols"`
}
