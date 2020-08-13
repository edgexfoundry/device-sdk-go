//
// Copyright (c) 2020 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package appsdk

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security"
)

func TestWaitForConfigUpdates_InsecureSecrets(t *testing.T) {
	_ = os.Setenv(security.EnvSecretStore, "false")
	defer os.Clearenv()

	sdk := &AppFunctionsSDK{}
	sdk.secretProvider = &security.SecretProviderImpl{}
	sdk.secretProvider.InsecureSecretsUpdated()
	expected := sdk.secretProvider.SecretsLastUpdated()

	// Create all dependencies that are going to be required for this test
	sdk.LoggingClient = logger.NewMockClient()
	sdk.appCtx, sdk.appCancelCtx = context.WithCancel(context.Background())
	sdk.appWg = &sync.WaitGroup{}
	sdk.config = &common.ConfigurationStruct{}

	target := NewConfigUpdateProcessor(sdk)
	configUpdated := make(config.UpdatedStream)

	// This starts a go func that listens on configUpdated for signal that something in Writable has changed
	target.WaitForConfigUpdates(configUpdated)
	// Give the go func time to spin up.
	time.Sleep(1 * time.Second)
	defer sdk.appCtx.Done()

	// Add a new InsecureSecret entry so it is detected as changed. Must make a new map so
	// it doesn't add to the old map.
	sdk.config.Writable.InsecureSecrets = make(map[string]common.InsecureSecretsInfo)
	sdk.config.Writable.InsecureSecrets["New"] = common.InsecureSecretsInfo{
		Path:    "New",
		Secrets: make(map[string]string),
	}

	// Enable security mode so change will be ignored by SecretProvider
	_ = os.Setenv(security.EnvSecretStore, "true")

	// Signal update occurred and give it time to process
	configUpdated <- struct{}{}
	time.Sleep(1 * time.Second)
	assert.Equal(t, expected, sdk.secretProvider.SecretsLastUpdated(), "LastUpdated should not have changed")

	// Add another new InsecureSecret entry so it is detected as changed. Must make a new map so
	// it doesn't add to the old map.
	sdk.config.Writable.InsecureSecrets = make(map[string]common.InsecureSecretsInfo)
	sdk.config.Writable.InsecureSecrets["New2"] = common.InsecureSecretsInfo{
		Path:    "New2",
		Secrets: make(map[string]string),
	}

	// Disable security mode so change is not ignored by SecretProvider
	_ = os.Setenv(security.EnvSecretStore, "false")

	// Signal update occurred and give it time to process
	configUpdated <- struct{}{}
	time.Sleep(1 * time.Second)
	assert.NotEqual(t, expected, sdk.secretProvider.SecretsLastUpdated(), "LastUpdated should have changed")
}
