// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package startup

import (
	"context"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/gorilla/mux"
)

func Bootstrap(serviceName string, serviceVersion string, driver dsModels.ProtocolDriver) {
	ctx, cancel := context.WithCancel(context.Background())
	service.Main(serviceName, serviceVersion, driver, ctx, cancel, mux.NewRouter(), nil)
}
