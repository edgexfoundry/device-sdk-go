//
// Copyright (c) 2020 Intel Corporation
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
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security/mock"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecretProvider(t *testing.T) {
	config := &common.ConfigurationStruct{}
	lc := logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")

	secretProvider := NewSecretProvider(lc, config)
	assert.NotNil(t, secretProvider, "secretProvider from NewSecretProvider should not be nil")
}

func TestInitializeClientFromSecretProvider(t *testing.T) {
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

	testSecretStoreInfo := config.SecretStoreInfo{
		Host:                    host,
		Port:                    portNum,
		Protocol:                "http",
		ServerName:              "mockVaultServer",
		AdditionalRetryAttempts: 2,
		RetryWaitPeriod:         "100ms",
	}

	emptySecretStoreInfo := config.SecretStoreInfo{}

	tests := []struct {
		name                             string
		tokenFileForShared               string
		tokenFileForExclusive            string
		sharedSecretStore                config.SecretStoreInfo
		exclusiveSecretStore             config.SecretStoreInfo
		expectError                      bool
		expectSharedSecretClientEmpty    bool
		expectExclusiveSecretClientEmpty bool
	}{
		{
			name:                             "Create client with test-token",
			tokenFileForShared:               "client/testdata/testToken.json",
			tokenFileForExclusive:            "client/testdata/testToken.json",
			sharedSecretStore:                testSecretStoreInfo,
			exclusiveSecretStore:             testSecretStoreInfo,
			expectError:                      false,
			expectSharedSecretClientEmpty:    false,
			expectExclusiveSecretClientEmpty: false,
		},
		{
			name:                             "Create client with expired token, no TTL remaining",
			tokenFileForShared:               "client/testdata/expiredToken.json",
			tokenFileForExclusive:            "client/testdata/expiredToken.json",
			sharedSecretStore:                testSecretStoreInfo,
			exclusiveSecretStore:             testSecretStoreInfo,
			expectError:                      true,
			expectSharedSecretClientEmpty:    true,
			expectExclusiveSecretClientEmpty: true,
		},
		{
			name:                             "Create client with non-existing TokenFile path",
			tokenFileForShared:               "client/testdata/non-existing.json",
			tokenFileForExclusive:            "client/testdata/non-existing.json",
			sharedSecretStore:                testSecretStoreInfo,
			exclusiveSecretStore:             testSecretStoreInfo,
			expectError:                      true,
			expectSharedSecretClientEmpty:    true,
			expectExclusiveSecretClientEmpty: true,
		},
		{
			name:                             "New secret client with no TokenFile",
			sharedSecretStore:                testSecretStoreInfo,
			exclusiveSecretStore:             testSecretStoreInfo,
			expectError:                      true,
			expectSharedSecretClientEmpty:    true,
			expectExclusiveSecretClientEmpty: true,
		},
		{
			name:                             "empty shared secret store",
			tokenFileForExclusive:            "client/testdata/testToken.json",
			sharedSecretStore:                emptySecretStoreInfo,
			exclusiveSecretStore:             testSecretStoreInfo,
			expectError:                      false,
			expectSharedSecretClientEmpty:    true,
			expectExclusiveSecretClientEmpty: false,
		},
		{
			name:                             "empty exclusive secret store",
			tokenFileForShared:               "client/testdata/testToken.json",
			sharedSecretStore:                testSecretStoreInfo,
			exclusiveSecretStore:             emptySecretStoreInfo,
			expectError:                      false,
			expectSharedSecretClientEmpty:    false,
			expectExclusiveSecretClientEmpty: true,
		},
		{
			name:                             "both empty secret stores",
			sharedSecretStore:                emptySecretStoreInfo,
			exclusiveSecretStore:             emptySecretStoreInfo,
			expectError:                      false,
			expectSharedSecretClientEmpty:    true,
			expectExclusiveSecretClientEmpty: true,
		},
	}

	for _, test := range tests {
		// pinned local test variables to avoid scopelint warnings
		currentTest := test

		t.Run(test.name, func(t *testing.T) {
			currentTest.sharedSecretStore.TokenFile = currentTest.tokenFileForShared
			currentTest.exclusiveSecretStore.TokenFile = currentTest.tokenFileForExclusive

			config := &common.ConfigurationStruct{
				SecretStore:          currentTest.sharedSecretStore,
				SecretStoreExclusive: currentTest.exclusiveSecretStore,
			}

			secretProvider := NewSecretProvider(lc, config)
			ok := secretProvider.Initialize(ctx)

			if currentTest.expectError {
				assert.False(t, ok, "Expect error but none was received")
			} else {
				assert.True(t, ok, "Expect no error but got not ok")
			}

			if currentTest.expectSharedSecretClientEmpty {
				assert.Nil(t, secretProvider.SharedSecretClient, "shared secret client should be empty")
			} else {
				assert.NotNil(t, secretProvider.SharedSecretClient, "shared secret client should NOT be empty")
			}

			if currentTest.expectExclusiveSecretClientEmpty {
				assert.Nil(t, secretProvider.ExclusiveSecretClient, "exclusive secret client should be empty")
			} else {
				assert.NotNil(t, secretProvider.ExclusiveSecretClient, "exclusive secret client should NOT be empty")
			}
		})
	}
	// wait for some time to allow renewToken to be run if any
	time.Sleep(7 * time.Second)
}

func TestInsecureSecretsUpdated(t *testing.T) {

	expected := time.Now()
	target := SecretProvider{
		LastUpdated: expected,
	}

	os.Setenv(EnvSecretStore, "true")
	target.InsecureSecretsUpdated()
	assert.Equal(t, expected, target.LastUpdated, "LastUpdated should not have changed")

	// Give a little time between tests so LastUpdated will be significantly different
	time.Sleep(1 * time.Second)
	os.Setenv(EnvSecretStore, "false")
	target.InsecureSecretsUpdated()
	assert.NotEqual(t, expected, target.LastUpdated, "LastUpdated should have changed")
}

func TestConfigAdditonalRetryAttempts(t *testing.T) {
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

	origEnv := os.Getenv(EnvSecretStore)

	defer func() {
		_ = os.Setenv(EnvSecretStore, origEnv)
	}()

	os.Setenv(EnvSecretStore, "true")

	testSecretStoreInfo := config.SecretStoreInfo{
		// configuration with AdditionalRetryAttempts omitted
		Host:       host,
		Port:       portNum,
		Protocol:   "http",
		ServerName: "mockVaultServer",
		TokenFile:  "client/testdata/testToken.json",
	}

	config := &common.ConfigurationStruct{
		SecretStore: testSecretStoreInfo,
	}

	tests := []struct {
		name                   string
		additonalRetryAttempts int
	}{
		{
			name:                   "Create client with 0 retry",
			additonalRetryAttempts: 0,
		},
		{
			name:                   "Create client with 1 retry",
			additonalRetryAttempts: 1,
		},
		{
			name:                   "Create client with 2 (or more) retries",
			additonalRetryAttempts: 2,
		},
	}

	for _, test := range tests {
		// pinned local test variables to avoid scopelint warnings
		currentTest := test

		t.Run(test.name, func(t *testing.T) {
			// inject the test data
			config.SecretStore.AdditionalRetryAttempts = currentTest.additonalRetryAttempts

			secrClient, err := NewSecretProvider(lc, config).initializeSecretClient(ctx, config.SecretStore)

			require.NoError(t, err)

			require.NotEmpty(t, secrClient, "should have client created even AdditionalRetryAttempts = 0")
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
