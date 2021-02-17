//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/stretchr/testify/assert"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

var d = dsModels.DiscoveredDevice{
	Name: "device-sdk-test",
}

func Test_whitelistPass(t *testing.T) {
	lc := logger.NewMockClient()
	pw := models.ProvisionWatcher{
		Name: "test-watcher",
		Identifiers: map[string]string{
			"host": "localhost",
			"port": "3[0-9]{2}",
		},
	}

	onlyOneMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "localhost",
			"port": "301",
		},
	}
	oneOfProtocolsMatch := map[string]models.ProtocolProperties{
		"tcp": {
			"host": "localhost",
			"port": "80",
		},
		"http": {
			"host": "localhost",
			"port": "301",
		},
	}
	noIdentifiersMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "192.168.0.1",
			"port": "400",
		},
	}
	someIdentifiersMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "127.0.0.1",
			"port": "301",
		},
		"tcp": {
			"host": "localhost",
			"port": "80",
		},
	}
	noMatchInSingleIdentifier := map[string]models.ProtocolProperties{
		"http": {
			"port": "301",
		},
		"tcp": {
			"host": "localhost",
		},
	}

	tests := []struct {
		name      string
		protocols map[string]models.ProtocolProperties
		expected  bool
	}{
		{"pass - match found", onlyOneMatch, true},
		{"pass - one match found in multiple protocol", oneOfProtocolsMatch, true},
		{"fail - none of identifier match in one protocol", noIdentifiersMatch, false},
		{"fail - only partial of identifiers match in one protocol", someIdentifiersMatch, false},
		{"fail - all of the identifiers match but across different protocol", noMatchInSingleIdentifier, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			d.Protocols = testCase.protocols
			result := whitelistPass(d, pw, lc)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func Test_blacklistPass(t *testing.T) {
	lc := logger.NewMockClient()
	pw := models.ProvisionWatcher{
		Name: "test-watcher",
		BlockingIdentifiers: map[string][]string{
			"port": []string{"399", "398", "397"},
		},
	}

	noBlockingIdentifierFound := map[string]models.ProtocolProperties{
		"http": {
			"host": "localhost",
		},
		"tcp": {
			"host": "127.0.0.1",
		},
	}
	noBlockingIdentifierMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "localhost",
			"port": "400",
		},
		"tcp": {
			"host": "localhost",
			"port": "80",
		},
	}
	blockingIdentifierMatch := map[string]models.ProtocolProperties{
		"http": {
			"host": "localhost",
			"port": "399",
		},
		"tcp": {
			"host": "localhost",
			"port": "80",
		},
	}

	tests := []struct {
		name      string
		protocols map[string]models.ProtocolProperties
		expected  bool
	}{
		{"pass - no blocking identifier found", noBlockingIdentifierFound, true},
		{"pass - blocking identifier found but not match", noBlockingIdentifierMatch, true},
		{"fail - blocking identifier match", blockingIdentifierMatch, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			d.Protocols = testCase.protocols
			result := blacklistPass(d, pw, lc)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
