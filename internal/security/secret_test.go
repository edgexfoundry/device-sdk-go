//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0
//

package security

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security/mock"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecretProvider(t *testing.T) {
	secretProvider := NewSecretProvider()
	assert.NotNil(t, secretProvider, "secretProvider from NewSecretProvider should not be nil")
}

func TestCreateClientFromSecretProvider(t *testing.T) {
	// setup
	tokenPeriod := 6
	tokenDataMap := initTokenData(tokenPeriod)

	server := mock.GetMockTokenServer(tokenDataMap)
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoErrorf(t, err, "error on parsing server url %s: %s", server.URL, err)

	host, port, _ := net.SplitHostPort(serverURL.Host)
	portNum, _ := strconv.Atoi(port)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	lc := logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")

	testSecretStoreInfo := common.SecretStoreInfo{
		SecretConfig: vault.SecretConfig{
			Host:       host,
			Port:       portNum,
			Protocol:   "http",
			ServerName: "mockVaultServer",
		},
	}

	secretProvider := NewSecretProvider()
	tests := []struct {
		name        string
		tokenFile   string
		expectError bool
	}{
		{
			name:        "Create client with test-token",
			tokenFile:   "client/testdata/testToken.json",
			expectError: false,
		},
		{
			name:        "Create client with expired token, no TTL remaining",
			tokenFile:   "client/testdata/expiredToken.json",
			expectError: true,
		},
		{
			name:        "Create client with non-existing TokenFile path",
			tokenFile:   "client/testdata/non-existing.json",
			expectError: true,
		},
		{
			name:        "New secret client with no TokenFile",
			tokenFile:   "",
			expectError: true,
		},
	}

	for _, test := range tests {
		// inject testing data
		testSecretStoreInfo.TokenFile = test.tokenFile

		config := common.ConfigurationStruct{
			SecretStore: testSecretStoreInfo,
		}

		// pinned local test variables to avoid scopelint warnings
		currentTest := test

		t.Run(test.name, func(t *testing.T) {
			ok := secretProvider.CreateClient(ctx, lc, config)

			if currentTest.expectError {
				assert.False(t, ok, "Expect error but none was received")
			} else {
				assert.True(t, ok, "Expect no error but got not ok")
			}
		})
	}
	// wait for some time to allow renewToken to be run if any
	time.Sleep(7 * time.Second)
}

func initTokenData(tokenPeriod int) *sync.Map {
	var tokenDataMap sync.Map

	tokenDataMap.Store("test-token", vault.TokenLookupMetadata{
		Renewable: true,
		Ttl:       tokenPeriod * 7 / 10,
		Period:    tokenPeriod,
	})
	// expired token
	tokenDataMap.Store("expired-token", vault.TokenLookupMetadata{
		Renewable: true,
		Ttl:       0,
		Period:    tokenPeriod,
	})
	// not renewable token
	tokenDataMap.Store("unrenewable-token", vault.TokenLookupMetadata{
		Renewable: false,
		Ttl:       0,
		Period:    tokenPeriod,
	})

	return &tokenDataMap
}
