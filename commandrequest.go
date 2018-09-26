// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import "github.com/edgexfoundry/edgex-go/pkg/models"

type CommandRequest struct {
	// RO is a ResourceOperation
	RO models.ResourceOperation
	// DeviceObject (aka device resource) represents the device resource
	// to be read or set. It can be used to access the attributes map,
	// PropertyValue, and PropertyUnit structs.
	DeviceObject models.DeviceObject
}
