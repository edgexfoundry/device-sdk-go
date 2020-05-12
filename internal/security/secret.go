/********************************************************************************
 *  Copyright 2019 Dell Inc.
 *  Copyright 2020 Dell Intel Corporation.
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
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security/authtokenloader"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security/client"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security/fileioperformer"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/pkg"
	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"
)

// SecretProvider cache storage for the secrets
type SecretProvider struct {
	SharedSecretClient    pkg.SecretClient
	ExclusiveSecretClient pkg.SecretClient
	secretsCache          map[string]map[string]string // secret's path, key, value
	configuration         *common.ConfigurationStruct
	cacheMuxtex           *sync.Mutex
	loggingClient         logger.LoggingClient
	//used to track when secrets have last been retrieved
	LastUpdated time.Time
}

// NewSecretProvider returns a new secret provider
func NewSecretProvider(loggingClient logger.LoggingClient, configuration *common.ConfigurationStruct) *SecretProvider {
	sp := &SecretProvider{
		secretsCache:  make(map[string]map[string]string),
		cacheMuxtex:   &sync.Mutex{},
		configuration: configuration,
		loggingClient: loggingClient,
		LastUpdated:   time.Now(),
	}

	return sp
}

// Initialize creates SecretClients to be used for obtaining secrets from a secrets store manager.
func (s *SecretProvider) Initialize(ctx context.Context) bool {
	var err error

	// initialize shared secret client if configured
	if s.SharedSecretClient, err = s.initializeSecretClient(ctx, s.configuration.SecretStore); err != nil {
		s.loggingClient.Error(fmt.Sprintf("unable to create shared secret client : %s", err.Error()))
		return false
	}

	// initialize exclusive secret client if configured
	if s.ExclusiveSecretClient, err = s.initializeSecretClient(ctx, s.configuration.SecretStoreExclusive); err != nil {
		s.loggingClient.Error(fmt.Sprintf("unable to create exclusive secret client : %s", err.Error()))
		return false
	}

	return true
}

func (s *SecretProvider) initializeSecretClient(
	ctx context.Context,
	secretStoreInfo bootstrapConfig.SecretStoreInfo) (pkg.SecretClient, error) {
	var secretClient pkg.SecretClient

	// secretStoreInfo is optional so that secret config can be empty
	secretConfig, secretStoreEmpty, err := s.getSecretConfig(secretStoreInfo)
	if err != nil {
		return nil, err
	}

	// no secret client to be created
	if secretStoreEmpty {
		return nil, nil
	}

	if s.isSecurityEnabled() {
		secretClient, err = client.NewVault(ctx, secretConfig, s.loggingClient).Get(secretStoreInfo)

		if err == nil || secretConfig.AdditionalRetryAttempts <= 0 {
			return secretClient, err
		}

		// retries some more times if secretConfig.AdditionalRetryAttempts is > 0
		waitTIme, parseErr := time.ParseDuration(secretConfig.RetryWaitPeriod)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid retry wait period for secret config: %s", parseErr.Error())
		}

		for retry := 0; retry < secretConfig.AdditionalRetryAttempts; retry++ {
			time.Sleep(waitTIme)

			secretClient, err = client.NewVault(ctx, secretConfig, s.loggingClient).Get(secretStoreInfo)

			if err == nil {
				break
			}
		}

		// check whehter the last retry is failed?
		if err != nil {
			return nil, err
		}
	}

	return secretClient, nil
}

// getSecretConfig creates a SecretConfig based on the SecretStoreInfo configuration properties.
// If a tokenfile is present it will override the Authentication.AuthToken value.
// the return boolean is used to indicate whether the secret store configuration is empty or not
func (s *SecretProvider) getSecretConfig(secretStoreInfo bootstrapConfig.SecretStoreInfo) (vault.SecretConfig, bool, error) {
	emptySecretStore := bootstrapConfig.SecretStoreInfo{}
	if secretStoreInfo == emptySecretStore {
		return vault.SecretConfig{}, true, nil
	}

	secretConfig := vault.SecretConfig{
		Host:                    secretStoreInfo.Host,
		Port:                    secretStoreInfo.Port,
		Path:                    secretStoreInfo.Path,
		Protocol:                secretStoreInfo.Protocol,
		Namespace:               secretStoreInfo.Namespace,
		RootCaCertPath:          secretStoreInfo.RootCaCertPath,
		ServerName:              secretStoreInfo.ServerName,
		Authentication:          secretStoreInfo.Authentication,
		AdditionalRetryAttempts: secretStoreInfo.AdditionalRetryAttempts,
		RetryWaitPeriod:         secretStoreInfo.RetryWaitPeriod,
	}

	if !s.isSecurityEnabled() || secretStoreInfo.TokenFile == "" {
		return secretConfig, false, nil
	}

	// only bother getting a token if security is enabled and the configuration-provided tokenfile is not empty.
	fileIoPerformer := fileioperformer.NewDefaultFileIoPerformer()
	authTokenLoader := authtokenloader.NewAuthTokenLoader(fileIoPerformer)

	token, err := authTokenLoader.Load(secretStoreInfo.TokenFile)
	if err != nil {
		return secretConfig, false, err
	}

	secretConfig.Authentication.AuthToken = token

	return secretConfig, false, nil
}

// isSecurityEnabled determines if security has been enabled.
func (s *SecretProvider) isSecurityEnabled() bool {
	env := os.Getenv("EDGEX_SECURITY_SECRET_STORE")
	return env != "false"
}
