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

package client

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/security/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"
)

func TestGetVaultClient(t *testing.T) {
	// setup
	tokenPeriod := 6
	tokenDataMap := initTokenData(tokenPeriod)

	server := mock.GetMockTokenServer(tokenDataMap)
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoErrorf(t, err, "error on parsing server url %s: %s", server.URL, err)

	host, port, _ := net.SplitHostPort(serverURL.Host)
	portNum, _ := strconv.Atoi(port)

	bkgCtx := context.Background()
	lc := logger.NewMockClient()

	testSecretStoreInfo := config.SecretStoreInfo{
		Host:       host,
		Port:       portNum,
		Path:       "/test",
		Protocol:   "http",
		ServerName: "mockVaultServer",
	}

	tests := []struct {
		name                string
		authToken           string
		tokenFile           string
		expectedNewToken    string
		expectedNilCallback bool
		expectedRetry       bool
		expectError         bool
	}{
		{
			name:                "New secret client with testToken1, more than half of TTL remaining",
			authToken:           "testToken1",
			tokenFile:           "testdata/replacement.json",
			expectedNilCallback: false,
			expectedNewToken:    "replacement-token",
			expectedRetry:       true,
			expectError:         false,
		},
		{
			name:                "New secret client with the same first token again",
			authToken:           "testToken1",
			tokenFile:           "testdata/replacement.json",
			expectedNilCallback: false,
			expectedNewToken:    "replacement-token",
			expectedRetry:       true,
			expectError:         false,
		},
		{
			name:                "New secret client with testToken2, half of TTL remaining",
			authToken:           "testToken2",
			expectedNilCallback: false,
			tokenFile:           "testdata/replacement.json",
			expectedNewToken:    "replacement-token",
			expectedRetry:       true,
			expectError:         false,
		},
		{
			name:                "New secret client with testToken3, less than half of TTL remaining",
			authToken:           "testToken3",
			tokenFile:           "testdata/replacement.json",
			expectedNilCallback: false,
			expectedNewToken:    "replacement-token",
			expectedRetry:       true,
			expectError:         false,
		},
		{
			name:                "New secret client with expired token, no TTL remaining",
			authToken:           "expiredToken",
			tokenFile:           "testdata/replacement.json",
			expectedNilCallback: false,
			expectedNewToken:    "replacement-token",
			expectedRetry:       true,
			expectError:         true,
		},
		{
			name:                "New secret client with expired token, non-existing TokenFile path",
			authToken:           "expiredToken",
			tokenFile:           "testdata/non-existing.json",
			expectedNilCallback: false,
			expectedNewToken:    "",
			expectedRetry:       false,
			expectError:         true,
		},
		{
			name:                "New secret client with expired test token, but same expired replacement token",
			authToken:           "test-token",
			tokenFile:           "testdata/testToken.json",
			expectedNilCallback: false,
			expectedNewToken:    "test-token",
			expectedRetry:       false,
			expectError:         true,
		},
		{
			name:                "New secret client with unauthenticated token",
			authToken:           "test-token",
			expectedNilCallback: true,
			expectedNewToken:    "",
			expectedRetry:       false,
			expectError:         true,
		},
		{
			name:                "New secret client with unrenewable token",
			authToken:           "unrenewableToken",
			expectedNilCallback: true,
			expectedNewToken:    "",
			expectedRetry:       true,
			expectError:         false,
		},
		{
			name:                "New secret client with no TokenFile",
			authToken:           "testToken1",
			tokenFile:           "",
			expectedNilCallback: true,
			expectedNewToken:    "",
			expectedRetry:       false,
			expectError:         false,
		},
	}

	for _, test := range tests {
		testSecretStoreInfo.TokenFile = test.tokenFile
		cfgHTTP := vault.SecretConfig{
			Host:           host,
			Port:           portNum,
			Protocol:       "http",
			Authentication: vault.AuthenticationInfo{AuthToken: test.authToken},
		}

		// pinned local test variable to avoid scopelint warnings
		currentTest := test

		t.Run(test.name, func(t *testing.T) {
			sclient := NewVault(bkgCtx, cfgHTTP, lc)
			_, err := sclient.Get(testSecretStoreInfo)

			if currentTest.expectError {
				require.Error(t, err, "Expect error but none was received")
			} else {
				require.NoError(t, err, "Expect no error but found some error")
			}

			tokenCallback := sclient.getDefaultTokenExpiredCallback(testSecretStoreInfo)
			if currentTest.expectedNilCallback {
				assert.Nil(t, tokenCallback, "found token expired callback func not nil")
			} else {
				require.NotNil(t, tokenCallback, "token expired callback func is nil")

				repToken, retry := tokenCallback(currentTest.authToken)

				assert.Equal(t, currentTest.expectedNewToken, repToken,
					"replacement token not as expected from callback func")

				assert.Equal(t, currentTest.expectedRetry, retry, "retry not as expected from callback func")
			}
		})
	}

	// wait for some time to allow renewToken to be run if any
	time.Sleep(7 * time.Second)
}

func initTokenData(tokenPeriod int) *sync.Map {
	var tokenDataMap sync.Map

	// ttl > half of period
	tokenDataMap.Store("testToken1", vault.TokenLookupMetadata{
		Renewable: true,
		Ttl:       tokenPeriod * 7 / 10,
		Period:    tokenPeriod,
	})
	// ttl = half of period
	tokenDataMap.Store("testToken2", vault.TokenLookupMetadata{
		Renewable: true,
		Ttl:       tokenPeriod / 2,
		Period:    tokenPeriod,
	})
	// ttl < half of period
	tokenDataMap.Store("testToken3", vault.TokenLookupMetadata{
		Renewable: true,
		Ttl:       tokenPeriod * 3 / 10,
		Period:    tokenPeriod,
	})
	// to be expired token
	tokenDataMap.Store("toToExpiredToken", vault.TokenLookupMetadata{
		Renewable: true,
		Ttl:       1,
		Period:    tokenPeriod,
	})
	// expired token
	tokenDataMap.Store("expiredToken", vault.TokenLookupMetadata{
		Renewable: true,
		Ttl:       0,
		Period:    tokenPeriod,
	})
	// not renewable token
	tokenDataMap.Store("unrenewableToken", vault.TokenLookupMetadata{
		Renewable: false,
		Ttl:       0,
		Period:    tokenPeriod,
	})

	return &tokenDataMap
}
