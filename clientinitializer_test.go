//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"net"
	"testing"

	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

func TestInitializeLoggingClientByFile(test *testing.T) {
	var loggingConfig = LoggingInfo{File: "./device-simple.log", RemoteURL: ""}
	var config = Config{Logging: loggingConfig}
	svc = &Service{c: &config}

	initializeLoggingClient()

	if svc.lc == nil {
		test.Fatal("New file logging fail")
	}

}

func TestCheckServiceAvailableByPingWithTimeoutError(test *testing.T) {
	var clientsConfig = map[string]service{ClientData: service{Host: "www.google.com", Port: 81, Timeout: 3000}}
	var config = Config{Clients: clientsConfig}
	svc = &Service{c: &config}
	svc.lc = logger.NewClient("test_service", false, svc.c.Logging.File)

	err := checkServiceAvailableByPing(ClientData)

	if err, ok := err.(net.Error); ok && !err.Timeout() {
		test.Fatal("Should be timeout error")
	}

}
