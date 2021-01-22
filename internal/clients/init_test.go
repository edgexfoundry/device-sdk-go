// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"net"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func TestCheckServiceAvailableByPingWithTimeoutError(test *testing.T) {
	var clientConfig = map[string]bootstrapConfig.ClientInfo{
		common.ClientData: {
			Host:     "www.google.com",
			Port:     81,
			Protocol: "http",
		},
	}
	config := &common.ConfigurationStruct{Clients: clientConfig}
	lc := logger.NewMockClient()

	err := checkServiceAvailableByPing(common.ClientData, config, lc)
	if err, ok := err.(net.Error); ok && !err.Timeout() {
		test.Fatal("Should be timeout error")
	}
}
