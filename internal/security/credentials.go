/********************************************************************************
 *  Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package security

import (
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
)

func (s *SecretProvider) GetDatabaseCredentials(database db.DatabaseInfo) (common.Credentials, error) {
	// If security is disabled or the database is Redis then we are to use the credentials supplied by the
	// configuration. The reason we do this for Redis is because Redis does not have an authentication nor an
	// authorization mechanism.
	if !s.isSecurityEnabled() || database.Type == db.RedisDB {
		return common.Credentials{
			Username: database.Username,
			Password: database.Password,
		}, nil
	}

	secrets, err := s.secretClient.GetSecrets(database.Type, "username", "password")
	if err != nil {
		return common.Credentials{}, err
	}

	return common.Credentials{
		Username: secrets["username"],
		Password: secrets["password"],
	}, nil
}

// GetSecrets retrieves secrets from a secret store.
// path specifies the type or location of the secrets to retrieve.
// keys specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the
// specified path will be returned.
func (s *SecretProvider) GetSecrets(path string, keys ...string) (map[string]string, error) {

	// check cache for keys
	allKeysExistsInCache := false
	cachedSecrets, cacheExists := s.secrets[path]

	if cacheExists {
		for _, key := range keys {
			if _, allKeysExistsInCache = cachedSecrets[key]; !allKeysExistsInCache {
				break
			}
		}

		// return cached secrets if they exist in cache
		if allKeysExistsInCache {
			return cachedSecrets, nil
		}
	}

	newSecrets, err := s.secretClient.GetSecrets(path, keys...)
	if err != nil {
		return nil, err
	}

	s.secrets[path] = newSecrets
	return s.secrets[path], nil
}

// StoreSecrets stores the secrets to a secret store.
// it sets the values requested at provided keys
// path specifies the type or location of the secrets to store
// secrets map specifies the "key": "value" pairs of secrets to store
func (s *SecretProvider) StoreSecrets(path string, secrets map[string]string) error {
	return nil // TODO: returning nil until interface is implemented
}
