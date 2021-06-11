// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"testing"
	"time"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
)

func TestCheckServiceAvailableByPingWithTimeoutError(test *testing.T) {
	var clientConfig = map[string]bootstrapConfig.ClientInfo{
		common.CoreDataServiceKey: {
			Host:     "www.google.com",
			Port:     81,
			Protocol: "http",
		},
	}
	config := &config.ConfigurationStruct{Clients: clientConfig}
	lc := logger.NewMockClient()

	res := checkServiceAvailableByPing(common.CoreDataServiceKey, time.Duration(1000), config, lc)
	assert.Equal(test, res, false, "request should be timeout and return false")
}
