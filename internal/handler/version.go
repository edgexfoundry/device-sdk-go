// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import "github.com/edgexfoundry/device-sdk-go/internal/common"

func VersionHandler() string {
	return common.ServiceVersion
}
