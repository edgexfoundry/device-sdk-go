// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/device-sdk-go/internal/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

// RestControllerName contains the name of RestController instance in the DIC.
var RestControllerName = "RestController"

// RestControllerFrom helper function queries the DIC and returns RestController instance.
func RestControllerFrom(get di.Get) controller.RestController {
	return get(RestControllerName).(controller.RestController)
}
