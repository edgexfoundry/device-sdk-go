// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package startup

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/service"
)

func Bootstrap(serviceKey string, serviceVersion string, driver interfaces.ProtocolDriver) {
	service, err := service.NewDeviceService(serviceKey, serviceVersion, driver)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
		os.Exit(-1)
	}

	err = service.Run()
	if err != nil {
		service.LoggingClient().Errorf("Device Service %s", err.Error())
		os.Exit(-1)
	}

	os.Exit(0)
}
