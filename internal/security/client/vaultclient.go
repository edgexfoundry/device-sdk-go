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
	"fmt"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/pkg"
	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"
	"github.com/edgexfoundry/go-mod-secrets/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/pkg/token/fileioperformer"
)

// Vault is the structure to get the secret client from go-mod-secrets vault package
type Vault struct {
	ctx    context.Context
	config vault.SecretConfig
	lc     logger.LoggingClient
}

// NewVault is the constructor for Vault in order to get the Vault secret client
func NewVault(ctx context.Context, config vault.SecretConfig, lc logger.LoggingClient) Vault {
	return Vault{
		ctx:    ctx,
		config: config,
		lc:     lc,
	}
}

// Get is the getter for Vault secret client from go-mod-secrets
func (c Vault) Get(secretStoreInfo bootstrapConfig.SecretStoreInfo) (pkg.SecretClient, error) {
	return vault.NewSecretClientFactory().NewSecretClient(
		c.ctx,
		c.config,
		c.lc,
		c.getDefaultTokenExpiredCallback(secretStoreInfo))
}

// getDefaultTokenExpiredCallback is the default implementation of tokenExpiredCallback function
// It utilizes the tokenFile to re-read the token and enable retry if any update from the expired token
func (c Vault) getDefaultTokenExpiredCallback(
	secretStoreInfo bootstrapConfig.SecretStoreInfo) func(expiredToken string) (replacementToken string, retry bool) {
	// if there is no tokenFile, then no replacement token can be used and hence no callback
	if secretStoreInfo.TokenFile == "" {
		return nil
	}

	tokenFile := secretStoreInfo.TokenFile

	return func(expiredToken string) (replacementToken string, retry bool) {
		// during the callback, we want to re-read the token from the disk
		// specified by tokenFile and set the retry to true if a new token
		// is different from the expiredToken
		fileIoPerformer := fileioperformer.NewDefaultFileIoPerformer()
		authTokenLoader := authtokenloader.NewAuthTokenLoader(fileIoPerformer)
		reReadToken, err := authTokenLoader.Load(tokenFile)

		if err != nil {
			c.lc.Error(fmt.Sprintf("fail to load auth token from tokenFile %s: %v", tokenFile, err))
			return "", false
		}

		if reReadToken == expiredToken {
			c.lc.Error("No new replacement token found for the expired token")
			return reReadToken, false
		}

		return reReadToken, true
	}
}
