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

package http

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
)

type SecretProviderMock struct {
	config          *common.ConfigurationStruct
	mockSecretStore map[string]map[string]string // secret's path, key, value

	//used to track when secrets have last been retrieved
	secretsLastUpdated time.Time
}

// NewSecretProviderMock returns a new mock secret provider
func NewSecretProviderMock(config *common.ConfigurationStruct) *SecretProviderMock {
	sp := &SecretProviderMock{}
	sp.config = config
	sp.mockSecretStore = make(map[string]map[string]string)
	return sp
}

// Initialize does nothing.
func (s *SecretProviderMock) Initialize(_ context.Context) bool {
	return true
}

// StoreSecrets saves secrets to the mock secret store.
func (s *SecretProviderMock) StoreSecrets(path string, secrets map[string]string) error {
	testFullPath := s.config.SecretStoreExclusive.Path + path
	// Base path should not have any leading slashes, only trailing or none, for this test to work
	if strings.Contains(testFullPath, "//") || !strings.Contains(testFullPath, "/") {
		return fmt.Errorf("Path is malformed: path=%s", path)
	}

	if !s.isSecurityEnabled() {
		return fmt.Errorf("Storing secrets is not supported when running in insecure mode")
	}
	s.mockSecretStore[path] = secrets
	return nil
}

// GetSecrets retrieves secrets from a mock secret store.
func (s *SecretProviderMock) GetSecrets(path string, _ ...string) (map[string]string, error) {
	secrets, ok := s.mockSecretStore[path]
	if !ok {
		return nil, fmt.Errorf("no secrets for path '%s' found", path)
	}
	return secrets, nil
}

// GetDatabaseCredentials retrieves the login credentials for the database from mock secret store
func (s *SecretProviderMock) GetDatabaseCredentials(database db.DatabaseInfo) (common.Credentials, error) {
	credentials, ok := s.mockSecretStore[database.Type]
	if !ok {
		return common.Credentials{}, fmt.Errorf("no credentials for type '%s' found", database.Type)
	}

	return common.Credentials{
		Username: credentials["username"],
		Password: credentials["password"],
	}, nil
}

// InsecureSecretsUpdated resets LastUpdate is not running in secure mode.If running in secure mode, changes to
// InsecureSecrets have no impact and are not used.
func (s *SecretProviderMock) InsecureSecretsUpdated() {
	s.secretsLastUpdated = time.Now()
}

// SecretsLastUpdated returns the time stamp when the provider secrets cache was latest updated
func (s *SecretProviderMock) SecretsLastUpdated() time.Time {
	return s.secretsLastUpdated
}

// isSecurityEnabled determines if security has been enabled.
func (s *SecretProviderMock) isSecurityEnabled() bool {
	env := os.Getenv(security.EnvSecretStore)
	return env != "false"
}
