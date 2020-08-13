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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
)

// GetDatabaseCredentials retrieves the login credentials for the database
// If security is disabled then we use the insecure credentials supplied by the configuration.
func (s *SecretProviderImpl) GetDatabaseCredentials(database db.DatabaseInfo) (common.Credentials, error) {
	var credentials map[string]string
	var err error

	// If security is disabled then we are to use the insecure credentials supplied by the configuration.
	if !s.isSecurityEnabled() {
		credentials, err = s.getInsecureSecrets(database.Type, "username", "password")
		// TODO: Remove for release version v2.0 when DB Credentials are only in new Insecure Secrets
		if err != nil {
			// Not found in new InsecureSecrets,so just use old V1 Database settings.
			return common.Credentials{
				Username: database.Username,
				Password: database.Password,
			}, nil
		}
	} else {
		if s.SharedSecretClient == nil {
			return common.Credentials{}, errors.New("SharedSecretClient is required but not configured")
		}

		credentials, err = s.SharedSecretClient.GetSecrets(database.Type, "username", "password")
	}

	if err != nil {
		return common.Credentials{}, err
	}

	return common.Credentials{
		Username: credentials["username"],
		Password: credentials["password"],
	}, nil
}

// GetSecrets retrieves secrets from a secret store.
// path specifies the type or location of the secrets to retrieve.
// keys specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the
// specified path will be returned.
func (s *SecretProviderImpl) GetSecrets(path string, keys ...string) (map[string]string, error) {
	if !s.isSecurityEnabled() {
		return s.getInsecureSecrets(path, keys...)
	}

	if cachedSecrets := s.getSecretsCache(path, keys...); cachedSecrets != nil {
		return cachedSecrets, nil
	}

	if s.ExclusiveSecretClient == nil {
		return nil, errors.New("can't get secret(s), exclusive secret client is not properly initialized")
	}

	secrets, err := s.ExclusiveSecretClient.GetSecrets(path, keys...)
	if err != nil {
		return nil, err
	}

	s.updateSecretsCache(path, secrets)
	return secrets, nil
}

// GetInsecureSecrets retrieves secrets from the Writable.InsecureSecrets section of the configuration
// path specifies the type or location of the secrets to retrieve.
// keys specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the
// specified path will be returned.
func (s *SecretProviderImpl) getInsecureSecrets(path string, keys ...string) (map[string]string, error) {
	secrets := make(map[string]string)
	pathExists := false
	var missingKeys []string

	for _, insecureSecrets := range s.configuration.Writable.InsecureSecrets {
		if insecureSecrets.Path == path {
			if len(keys) == 0 {
				// If no keys are provided then all the keys associated with the specified path will be returned
				for k, v := range insecureSecrets.Secrets {
					secrets[k] = v
				}
				return secrets, nil
			}

			pathExists = true
			for _, key := range keys {
				value, keyExists := insecureSecrets.Secrets[key]
				if !keyExists {
					missingKeys = append(missingKeys, key)
					continue
				}
				secrets[key] = value
			}
		}
	}

	if len(missingKeys) > 0 {
		err := fmt.Errorf("No value for the keys: [%s] exists", strings.Join(missingKeys, ","))
		return nil, err
	}

	if !pathExists {
		// if path is not in secret store
		err := fmt.Errorf("Error, path (%v) doesn't exist in secret store", path)
		return nil, err
	}

	return secrets, nil
}

func (s *SecretProviderImpl) getSecretsCache(path string, keys ...string) map[string]string {
	secrets := make(map[string]string)

	// Synchronize cache access
	s.cacheMuxtex.Lock()
	defer s.cacheMuxtex.Unlock()

	// check cache for keys
	allKeysExistInCache := false
	cachedSecrets, cacheExists := s.secretsCache[path]

	if cacheExists {
		for _, key := range keys {
			value, allKeysExistInCache := cachedSecrets[key]
			if !allKeysExistInCache {
				return nil
			}
			secrets[key] = value
		}

		// return secrets if the requested keys exist in cache
		if allKeysExistInCache {
			return secrets
		}
	}

	return nil
}

func (s *SecretProviderImpl) updateSecretsCache(path string, secrets map[string]string) {
	// Synchronize cache access
	s.cacheMuxtex.Lock()
	defer s.cacheMuxtex.Unlock()

	if _, cacheExists := s.secretsCache[path]; !cacheExists {
		s.secretsCache[path] = secrets
	}

	for key, value := range secrets {
		s.secretsCache[path][key] = value
	}
}

// StoreSecrets stores the secrets to a secret store.
// it sets the values requested at provided keys
// path specifies the type or location of the secrets to store
// secrets map specifies the "key": "value" pairs of secrets to store
func (s *SecretProviderImpl) StoreSecrets(path string, secrets map[string]string) error {
	if !s.isSecurityEnabled() {
		return errors.New("Storing secrets is not supported when running in insecure mode")
	}

	if s.ExclusiveSecretClient == nil {
		return errors.New("can't store secret(s) 'SecretProvider' is not properly initialized")
	}

	err := s.ExclusiveSecretClient.StoreSecrets(path, secrets)
	if err != nil {
		return err
	}

	// Synchronize cache access before clearing
	s.cacheMuxtex.Lock()
	// Clearing cache because adding a new secret(s) possibly invalidates the previous cache
	s.secretsCache = make(map[string]map[string]string)
	s.cacheMuxtex.Unlock()
	//indicate to the SDK that the cache has been invalidated
	s.LastUpdated = time.Now()
	return nil
}
