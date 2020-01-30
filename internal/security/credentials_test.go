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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
}

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

var getSecretsTestData = []getSecretsTestObj{

	getSecretsTestObj{
		testName:          "Empty keys",
		path:              "db_secrets",
		keys:              []string{"", ""},
		expectedSecrets:   nil,
		expectedErr:       errors.New("Couldn't get secrets: No value for the keys: [] exists"),
		resetSecretsCache: true,
	},

	getSecretsTestObj{
		testName:          "One valid key, one empty key",
		path:              "db_secrets",
		keys:              []string{"key1", ""},
		expectedSecrets:   nil,
		expectedErr:       errors.New("Couldn't get secrets: No value for the keys: [] exists"),
		resetSecretsCache: true,
	},
	getSecretsTestObj{
		testName:          "One valid key one not found key",
		path:              "db_secrets",
		keys:              []string{"key1", "notFoundKey"},
		expectedSecrets:   nil,
		expectedErr:       errors.New("Couldn't get secrets: No value for the keys: [notFoundKey] exists"),
		resetSecretsCache: true,
	},
	getSecretsTestObj{
		testName:          "Not found key",
		path:              "db_secrets",
		keys:              []string{"notFoundKey"},
		expectedSecrets:   nil,
		expectedErr:       errors.New("Couldn't get secrets: No value for the keys: [notFoundKey] exists"),
		resetSecretsCache: true,
	},
	getSecretsTestObj{
		testName:          "Valid key",
		path:              "db_secrets",
		keys:              []string{"key1"},
		expectedSecrets:   map[string]string{"key1": "value1"},
		resetSecretsCache: true,
		expectedErr:       nil,
	},
	getSecretsTestObj{
		testName:          "Two valid keys",
		path:              "db_secrets",
		keys:              []string{"key1", "key2"},
		expectedSecrets:   map[string]string{"key1": "value1", "key2": "value2"},
		resetSecretsCache: true,
		expectedErr:       nil,
	},
	getSecretsTestObj{
		testName:          "Valid key (key1 not cached)",
		path:              "db_secrets",
		keys:              []string{"key1"},
		expectedSecrets:   map[string]string{"key1": "value1"},
		expectedErr:       nil,
		resetSecretsCache: false,
	},
	getSecretsTestObj{
		testName:          "One valid key (key1 cached)",
		path:              "db_secrets",
		keys:              []string{"key1"},
		expectedSecrets:   map[string]string{"key1": "value1"},
		expectedErr:       nil,
		resetSecretsCache: false,
	},
	getSecretsTestObj{
		testName:          "Two valid keys (key1 cached, key2 not cached)",
		path:              "db_secrets",
		keys:              []string{"key1", "key2"},
		expectedSecrets:   map[string]string{"key1": "value1", "key2": "value2"},
		expectedErr:       nil,
		resetSecretsCache: false,
	},
}

func TestGetSecrets(t *testing.T) {

	secretProvider := newMockSecretProvider()

	for i, test := range getSecretsTestData {

		testNameInfo := fmt.Sprintf("Test: %v", test.testName)
		secretProvider.secretClient.(*mockSecretClient).testIndex = i

		path := test.path
		keys := test.keys

		secrets, err := secretProvider.GetSecrets(path, keys...)

		assert.Equal(t, test.expectedErr, err, testNameInfo)
		assert.Equal(t, test.expectedSecrets, secrets, testNameInfo)

		// not re-newing the secretProvider will test the cache for the next item in the getSecretsTestData slice
		if test.resetSecretsCache {
			secretProvider = newMockSecretProvider()
		}
	}
}

func newMockSecretProvider() *SecretProvider {
	return &SecretProvider{secretClient: &mockSecretClient{}, secrets: make(map[string]map[string]string)}
}

func (s *mockSecretClient) GetSecrets(path string, keys ...string) (map[string]string, error) {

	return getSecretsTestData[s.testIndex].expectedSecrets, getSecretsTestData[s.testIndex].expectedErr
}

func (s *mockSecretClient) StoreSecrets(path string, secrets map[string]string) error {

	return nil
}
