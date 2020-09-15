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
// SPDX-License-Identifier: Apache-2.0'
//

package security

import (
	"errors"
	"os"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/require"
)

type mockSecretClient struct {
	testIndex int
}

type getSecretsTestObj struct {
	testName          string
	path              string
	keys              []string
	expectedSecrets   map[string]string
	expectedErr       error
	resetSecretsCache bool
}

type secretTestData []getSecretsTestObj

func getSecretsTestData() secretTestData {
	return []getSecretsTestObj{
		{
			testName:          "Empty path",
			path:              "",
			keys:              []string{"key1", "key2"},
			expectedSecrets:   map[string]string{"key1": "value1", "key2": "value2"},
			expectedErr:       nil,
			resetSecretsCache: true,
		},
		{
			testName:          "Empty path, keys list empty",
			path:              "",
			keys:              []string{},
			expectedSecrets:   map[string]string{"key1": "value1", "key2": "value2"},
			expectedErr:       nil,
			resetSecretsCache: true,
		},
		{
			testName:          "Fake path",
			path:              "fakepath",
			keys:              []string{"key1", "key2"},
			expectedSecrets:   nil,
			expectedErr:       errors.New("Error, path (fakepath) doesn't exist in secret store"),
			resetSecretsCache: true,
		},
		{
			testName:          "Empty keys",
			path:              "db_secrets",
			keys:              []string{"", ""},
			expectedSecrets:   nil,
			expectedErr:       errors.New("No value for the keys: [,] exists"),
			resetSecretsCache: true,
		},
		{
			testName:          "One valid key, one empty key",
			path:              "db_secrets",
			keys:              []string{"key1", ""},
			expectedSecrets:   nil,
			expectedErr:       errors.New("No value for the keys: [] exists"),
			resetSecretsCache: true,
		},
		{
			testName:          "One valid key one not found key",
			path:              "db_secrets",
			keys:              []string{"key1", "notFoundKey"},
			expectedSecrets:   nil,
			expectedErr:       errors.New("No value for the keys: [notFoundKey] exists"),
			resetSecretsCache: true,
		},
		{
			testName:          "Not found key",
			path:              "db_secrets",
			keys:              []string{"notFoundKey"},
			expectedSecrets:   nil,
			expectedErr:       errors.New("No value for the keys: [notFoundKey] exists"),
			resetSecretsCache: true,
		},
		{
			testName:          "Two missing keys",
			path:              "db_secrets",
			keys:              []string{"notFoundKey1", "notFoundKey2"},
			expectedSecrets:   nil,
			expectedErr:       errors.New("No value for the keys: [notFoundKey1,notFoundKey2] exists"),
			resetSecretsCache: true,
		},
		{
			testName:          "Valid key",
			path:              "db_secrets",
			keys:              []string{"key1"},
			expectedSecrets:   map[string]string{"key1": "value1"},
			expectedErr:       nil,
			resetSecretsCache: true,
		},
		{
			testName:          "Two valid keys",
			path:              "db_secrets",
			keys:              []string{"key1", "key2"},
			expectedSecrets:   map[string]string{"key1": "value1", "key2": "value2"},
			expectedErr:       nil,
			resetSecretsCache: true,
		},
		{
			testName:          "Valid key (key1 not cached)",
			path:              "db_secrets",
			keys:              []string{"key1"},
			expectedSecrets:   map[string]string{"key1": "value1"},
			expectedErr:       nil,
			resetSecretsCache: false,
		},
		{
			testName:          "One valid key (key1 already cached)",
			path:              "db_secrets",
			keys:              []string{"key1"},
			expectedSecrets:   map[string]string{"key1": "value1"},
			expectedErr:       nil,
			resetSecretsCache: false,
		},
		{
			testName:          "Two valid keys (key1 cached, key2 not cached)",
			path:              "db_secrets",
			keys:              []string{"key1", "key2"},
			expectedSecrets:   map[string]string{"key1": "value1", "key2": "value2"},
			expectedErr:       nil,
			resetSecretsCache: false,
		},
	}
}

func TestGetSecrets(t *testing.T) {
	os.Clearenv()
	secretProvider := newMockSecretProvider(nil)

	for i, test := range getSecretsTestData() {
		i := i
		test := test
		t.Run(test.testName, func(t *testing.T) {
			secretProvider.ExclusiveSecretClient.(*mockSecretClient).testIndex = i
			secrets, err := secretProvider.GetSecrets(test.path, test.keys...)

			require.Equal(t, test.expectedErr, err)
			require.Equal(t, test.expectedSecrets, secrets)

			// not re-newing the secretProvider will test the cache for the next item in the getSecretsTestData slice
			if test.resetSecretsCache {
				secretProvider = newMockSecretProvider(nil)
			}
		})
	}
}

func TestGetInsecureSecrets(t *testing.T) {
	os.Clearenv()
	secretProvider, origEnv := setupGetInsecureSecrets(t)

	for _, test := range getSecretsTestData() {
		test := test
		t.Run(test.testName, func(t *testing.T) {
			secrets, err := secretProvider.getInsecureSecrets(test.path, test.keys...)

			require.Equal(t, test.expectedErr, err)
			require.Equal(t, test.expectedSecrets, secrets)
		})
	}

	tearDownGetInsecureSecrets(t, origEnv)
}

func setupGetInsecureSecrets(t *testing.T) (sp *SecretProviderImpl, origEnv string) {
	insecureSecrets := common.InsecureSecrets{
		"no_path": common.InsecureSecretsInfo{
			Path: "",
			Secrets: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		"db_secrets": common.InsecureSecretsInfo{
			Path: "db_secrets",
			Secrets: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	configuration := common.ConfigurationStruct{
		Writable: common.WritableInfo{
			InsecureSecrets: insecureSecrets,
		},
	}
	secretProvider := newMockSecretProvider(&configuration)

	origEnv = os.Getenv(EnvSecretStore)

	//	disable secure store
	if err := os.Setenv(EnvSecretStore, "false"); err != nil {
		t.Fatalf("Failed to set env variable: " + EnvSecretStore)
	}

	return secretProvider, origEnv
}

func tearDownGetInsecureSecrets(t *testing.T, origEnv string) {
	if err := os.Setenv(EnvSecretStore, origEnv); err != nil {
		t.Fatalf("Failed to set env variable: %s back to original value", EnvSecretStore)
	}
}

func newMockSecretProvider(configuration *common.ConfigurationStruct) *SecretProviderImpl {
	logClient := logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
	mockSP := NewSecretProvider(logClient, configuration)
	mockSP.SharedSecretClient = &mockSecretClient{}
	mockSP.ExclusiveSecretClient = &mockSecretClient{}
	return mockSP
}

func (s *mockSecretClient) GetSecrets(path string, keys ...string) (map[string]string, error) {
	secretTestData := getSecretsTestData()
	return secretTestData[s.testIndex].expectedSecrets, secretTestData[s.testIndex].expectedErr
}

func (s *mockSecretClient) StoreSecrets(path string, secrets map[string]string) error {
	return nil
}
