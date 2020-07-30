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
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"

	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/bootstrap/handlers"
)

// ConfigUpdateProcessor contains the data need to process configuration updates
type ConfigUpdateProcessor struct {
	sdk *AppFunctionsSDK
}

// NewConfigUpdateProcessor creates a new ConfigUpdateProcessor which process configuration updates triggered from
// the Configuration Provider
func NewConfigUpdateProcessor(sdk *AppFunctionsSDK) *ConfigUpdateProcessor {
	return &ConfigUpdateProcessor{sdk: sdk}
}

// WaitForConfigUpdates waits for signal that configuration has been updated (triggered from by Configuration Provider)
// and then determines what was updated and does any special processing, if needed, for the updates.
func (processor *ConfigUpdateProcessor) WaitForConfigUpdates(configUpdated config.UpdatedStream) {
	sdk := processor.sdk
	sdk.appWg.Add(1)

	go func() {
		defer sdk.appWg.Done()

		sdk.LoggingClient.Info("Waiting for App Service configuration updates...")

		previousWriteable := sdk.config.Writable

		for {
			select {
			case <-sdk.appCtx.Done():
				sdk.LoggingClient.Info("Exiting waiting for App Service configuration updates")
				return

			case <-configUpdated:
				currentWritable := sdk.config.Writable
				sdk.LoggingClient.Info("Processing App Service configuration updates")

				// Note: Updates occur one setting at a time so only have to look for single changes
				switch {
				case previousWriteable.LogLevel != currentWritable.LogLevel:
					_ = sdk.LoggingClient.SetLogLevel(currentWritable.LogLevel)
					sdk.LoggingClient.Info(fmt.Sprintf("Logging level changed to %s", currentWritable.LogLevel))

				// InsecureSecrets (map) will be nil if not in the original TOML used to seed the Config Provider,
				// so ignore it if this is the case.
				case currentWritable.InsecureSecrets != nil &&
					!reflect.DeepEqual(previousWriteable.InsecureSecrets, currentWritable.InsecureSecrets):
					sdk.secretProvider.InsecureSecretsUpdated()
					sdk.LoggingClient.Info("Insecure Secrets have been updated")

				case previousWriteable.StoreAndForward.MaxRetryCount != currentWritable.StoreAndForward.MaxRetryCount:
					if currentWritable.StoreAndForward.MaxRetryCount < 0 {
						sdk.LoggingClient.Warn(
							fmt.Sprintf("StoreAndForward MaxRetryCount can not be less than 0, defaulting to 1"))
						currentWritable.StoreAndForward.MaxRetryCount = 1
					}
					sdk.LoggingClient.Info(
						fmt.Sprintf(
							"StoreAndForward MaxRetryCount changed to %d",
							currentWritable.StoreAndForward.MaxRetryCount))

				case previousWriteable.StoreAndForward.RetryInterval != currentWritable.StoreAndForward.RetryInterval:
					if _, err := time.ParseDuration(currentWritable.StoreAndForward.RetryInterval); err != nil {
						sdk.LoggingClient.Error(fmt.Sprintf("StoreAndForward RetryInterval not change: %s", err.Error()))
						currentWritable.StoreAndForward.RetryInterval = previousWriteable.StoreAndForward.RetryInterval
						continue
					}

					processor.processConfigChangedStoreForwardRetryInterval()
					sdk.LoggingClient.Info(
						fmt.Sprintf(
							"StoreAndForward RetryInterval changed to %s",
							currentWritable.StoreAndForward.RetryInterval))

				case previousWriteable.StoreAndForward.Enabled != currentWritable.StoreAndForward.Enabled:
					processor.processConfigChangedStoreForwardEnabled()
					sdk.LoggingClient.Info(
						fmt.Sprintf(
							"StoreAndForward Enabled changed to %v",
							currentWritable.StoreAndForward.Enabled))

				default:
					// Assume change is in the pipeline since all others have been checked appropriately
					processor.processConfigChangedPipeline()
				}

				// grab new copy of the writeable configuration for comparing against when next update occurs
				previousWriteable = currentWritable
			}
		}
	}()
}

// processConfigChangedStoreForwardRetryInterval handles when the Store and Forward RetryInterval setting has been updated
func (processor *ConfigUpdateProcessor) processConfigChangedStoreForwardRetryInterval() {
	sdk := processor.sdk

	if sdk.config.Writable.StoreAndForward.Enabled {
		sdk.stopStoreForward()
		sdk.startStoreForward()
	}
}

// processConfigChangedStoreForwardEnabled handles when the Store and Forward Enabled setting has been updated
func (processor *ConfigUpdateProcessor) processConfigChangedStoreForwardEnabled() {
	sdk := processor.sdk

	if sdk.config.Writable.StoreAndForward.Enabled {
		// StoreClient must be set up for StoreAndForward
		if sdk.storeClient == nil {
			var err error
			startupTimer := startup.NewStartUpTimer(internal.BootRetrySecondsDefault, internal.BootTimeoutSecondsDefault)
			sdk.storeClient, err = handlers.InitializeStoreClient(sdk.secretProvider, sdk.config, startupTimer, sdk.LoggingClient)
			if err != nil {
				// Error already logged
				sdk.config.Writable.StoreAndForward.Enabled = false
				return
			}

			sdk.runtime.Initialize(sdk.storeClient, sdk.secretProvider)
		}

		sdk.startStoreForward()
	} else {
		sdk.stopStoreForward()
	}
}

// processConfigChangedPipeline handles when any of the Pipeline settings have been updated
func (processor *ConfigUpdateProcessor) processConfigChangedPipeline() {
	sdk := processor.sdk

	if sdk.usingConfigurablePipeline {
		transforms, err := sdk.LoadConfigurablePipeline()
		if err != nil {
			sdk.LoggingClient.Error("unable to reload Configurable Pipeline from new configuration: " + err.Error())
			return
		}
		err = sdk.SetFunctionsPipeline(transforms...)
		if err != nil {
			sdk.LoggingClient.Error("unable to set Configurable Pipeline functions from new configuration: " + err.Error())
			return
		}

		sdk.LoggingClient.Info("Configurable Pipeline successfully reloaded from new configuration")
	}
}

// startStoreForward starts the Store and Forward processing
func (sdk *AppFunctionsSDK) startStoreForward() {
	var storeForwardEnabledCtx context.Context
	sdk.storeForwardWg = &sync.WaitGroup{}
	storeForwardEnabledCtx, sdk.storeForwardCancelCtx = context.WithCancel(context.Background())
	sdk.runtime.StartStoreAndForward(
		sdk.appWg,
		sdk.appCtx,
		sdk.storeForwardWg,
		storeForwardEnabledCtx,
		sdk.ServiceKey,
		sdk.config,
		sdk.edgexClients)
}

// stopStoreForward stops the Store and Forward processing
func (sdk *AppFunctionsSDK) stopStoreForward() {
	sdk.LoggingClient.Info("Canceling Store and Forward retry loop")
	sdk.storeForwardCancelCtx()
	sdk.storeForwardWg.Wait()
}
