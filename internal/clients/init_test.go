// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"testing"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
)

func TestCheckServiceAvailableByPingWithTimeoutError(test *testing.T) {
	var clientConfig = map[string]bootstrapConfig.ClientInfo{
		clients.CoreDataServiceKey: {
			Host:     "www.google.com",
			Port:     81,
			Protocol: "http",
		},
	}
	config := &config.ConfigurationStruct{Service: config.ServiceInfo{Timeout: 1000}, Clients: clientConfig}
	lc := logger.NewMockClient()

	res := checkServiceAvailableByPing(clients.CoreDataServiceKey, config, lc)
	assert.Equal(test, res, false, "request should be timeout and return false")
}
