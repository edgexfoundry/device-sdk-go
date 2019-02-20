// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"net"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
)

func TestInitializeLoggingClientByFile(test *testing.T) {
	var loggingConfig = common.LoggingInfo{File: "./device-simple.log", EnableRemote: false}
	var config = common.Config{Logging: loggingConfig}
	common.CurrentConfig = &config

	initializeLoggingClient()

	if common.LoggingClient == nil {
		test.Fatal("New file logging fail")
	}

}

func TestCheckServiceAvailableByPingWithTimeoutError(test *testing.T) {
	var clientConfig = map[string]common.ClientInfo{common.ClientData: common.ClientInfo{Protocol: "http", Host: "www.google.com", Port: 81, Timeout: 3000}}
	var config = common.Config{Clients: clientConfig}
	common.CurrentConfig = &config
	common.LoggingClient = logger.NewClient("test_service", false, "./device-simple.log", "DEBUG")

	err := checkServiceAvailableByPing(common.ClientData)

	if err, ok := err.(net.Error); ok && !err.Timeout() {
		test.Fatal("Should be timeout error")
	}

}
