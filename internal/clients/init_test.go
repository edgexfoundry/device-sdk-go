// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"net"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestCheckServiceAvailableByPingWithTimeoutError(test *testing.T) {
	var clientConfig = map[string]bootstrapConfig.ClientInfo{
		common.ClientData: bootstrapConfig.ClientInfo{
			Host:     "www.google.com",
			Port:     81,
			Protocol: "http",
		},
	}
	config := &common.ConfigurationStruct{Clients: clientConfig}
	lc := logger.NewClientStdOut("device-sdk-test", false, "DEBUG")
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return config
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	err := checkServiceAvailableByPing(common.ClientData, dic)

	if err, ok := err.(net.Error); ok && !err.Timeout() {
		test.Fatal("Should be timeout error")
	}
}
