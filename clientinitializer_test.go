//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"github.com/edgexfoundry/device-sdk-go/common"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"net"
	"testing"
)

func TestInitializeLoggingClientByFile(test *testing.T) {
	var loggingConfig = common.LoggingInfo{File: "./device-simple.log", RemoteURL: ""}
	var config = common.Config{Logging: loggingConfig}
	svc = &Service{config: &config}

	initializeLoggingClient()

	if svc.lc == nil {
		test.Fatal("New file logging fail")
	}

}

func TestCheckServiceAvailableByPingWithTimeoutError(test *testing.T) {
	var clientConfig = map[string]common.RegisteredService{common.ClientData: common.RegisteredService{Host: "www.google.com", Port: 81, Timeout: 3000}}
	var config = common.Config{Clients: clientConfig}
	svc = &Service{config: &config}
	svc.lc = logger.NewClient("test_service", false, svc.config.Logging.File)

	err := checkServiceAvailableByPing(common.ClientData)

	if err, ok := err.(net.Error); ok && !err.Timeout() {
		test.Fatal("Should be timeout error")
	}

}
