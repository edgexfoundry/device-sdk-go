// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "github.com/edgexfoundry/go-mod-core-contracts/models"

type CommandRequest struct {
	// RO is a ResourceOperation
	RO models.ResourceOperation
	// DeviceResource represents the device resource
	// to be read or set. It can be used to access the attributes map,
	// PropertyValue, and PropertyUnit structs.
	DeviceResource models.DeviceResource
}
