//
// Copyright (c) 2021 Intel Corporation
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

package app

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/config"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/handlers"
)

// ConfigUpdateProcessor contains the data need to process configuration updates
type ConfigUpdateProcessor struct {
	svc *Service
}

// NewConfigUpdateProcessor creates a new ConfigUpdateProcessor which processes configuration updates triggered from
// the Configuration Provider
func NewConfigUpdateProcessor(svc *Service) *ConfigUpdateProcessor {
	return &ConfigUpdateProcessor{svc: svc}
}

// WaitForConfigUpdates waits for signal that configuration has been updated (triggered from by Configuration Provider)
// and then determines what was updated and does any special processing, if needed, for the updates.
func (processor *ConfigUpdateProcessor) WaitForConfigUpdates(configUpdated config.UpdatedStream) {
	svc := processor.svc
	svc.ctx.appWg.Add(1)

	go func() {
		defer svc.ctx.appWg.Done()
		lc := svc.LoggingClient()
		lc.Info("Waiting for App Service configuration updates...")

		previousWriteable := svc.config.Writable

		for {
			select {
			case <-svc.ctx.appCtx.Done():
				lc.Info("Exiting waiting for App Service configuration updates")
				return

			case <-configUpdated:
				currentWritable := svc.config.Writable
				lc.Info("Processing App Service configuration updates")

				// Note: Updates occur one setting at a time so only have to look for single changes
				switch {
				case previousWriteable.StoreAndForward.MaxRetryCount != currentWritable.StoreAndForward.MaxRetryCount:
					if currentWritable.StoreAndForward.MaxRetryCount < 0 {
						lc.Warn("StoreAndForward MaxRetryCount can not be less than 0, defaulting to 1")
						currentWritable.StoreAndForward.MaxRetryCount = 1
					}
					lc.Infof("StoreAndForward MaxRetryCount changed to %d", currentWritable.StoreAndForward.MaxRetryCount)

				case previousWriteable.StoreAndForward.RetryInterval != currentWritable.StoreAndForward.RetryInterval:
					if _, err := time.ParseDuration(currentWritable.StoreAndForward.RetryInterval); err != nil {
						lc.Errorf("StoreAndForward RetryInterval not change: %s", err.Error())
						currentWritable.StoreAndForward.RetryInterval = previousWriteable.StoreAndForward.RetryInterval
						continue
					}

					processor.processConfigChangedStoreForwardRetryInterval()
					lc.Infof("StoreAndForward RetryInterval changed to %s", currentWritable.StoreAndForward.RetryInterval)

				case previousWriteable.StoreAndForward.Enabled != currentWritable.StoreAndForward.Enabled:
					processor.processConfigChangedStoreForwardEnabled()
					lc.Infof("StoreAndForward Enabled changed to %v", currentWritable.StoreAndForward.Enabled)

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

func (processor *ConfigUpdateProcessor) processConfigChangedStoreForwardRetryInterval() {
	sdk := processor.svc

	if sdk.config.Writable.StoreAndForward.Enabled {
		sdk.stopStoreForward()
		sdk.startStoreForward()
	}
}

func (processor *ConfigUpdateProcessor) processConfigChangedStoreForwardEnabled() {
	sdk := processor.svc

	if sdk.config.Writable.StoreAndForward.Enabled {
		storeClient := container.StoreClientFrom(sdk.dic.Get)
		// StoreClient must be set up for StoreAndForward
		if storeClient == nil {
			var err error
			startupTimer := startup.NewStartUpTimer(sdk.serviceKey)
			secretProvider := bootstrapContainer.SecretProviderFrom(sdk.dic.Get)
			storeClient, err = handlers.InitializeStoreClient(secretProvider, sdk.config, startupTimer, sdk.LoggingClient())
			if err != nil {
				// Error already logged
				sdk.config.Writable.StoreAndForward.Enabled = false
				return
			}

			sdk.dic.Update(di.ServiceConstructorMap{
				container.StoreClientName: func(get di.Get) interface{} {
					return storeClient
				},
			})
		}

		sdk.startStoreForward()
	} else {
		sdk.stopStoreForward()
	}
}

func (processor *ConfigUpdateProcessor) processConfigChangedPipeline() {
	sdk := processor.svc

	if sdk.usingConfigurablePipeline {
		pipelines, err := sdk.LoadConfigurableFunctionPipelines()
		if err != nil {
			sdk.LoggingClient().Errorf("unable to reload Configurable Pipeline(s) from new configuration: %s: all pipelines have been disabled", err.Error())
			// Clear the pipeline transforms so error occurs when attempting to execute the pipeline(s).
			sdk.runtime.ClearAllFunctionsPipelineTransforms()
			return
		}

		sdk.runtime.TargetType = sdk.targetType

		// Update the pipelines with their new transforms
		for _, pipeline := range pipelines {
			sdk.runtime.SetFunctionsPipelineTransforms(pipeline.Id, pipeline.Transforms)
		}

		sdk.LoggingClient().Info("Configurable Pipeline successfully reloaded from new configuration")
	}
}

func (svc *Service) startStoreForward() {
	var storeForwardEnabledCtx context.Context
	svc.ctx.storeForwardWg = &sync.WaitGroup{}
	storeForwardEnabledCtx, svc.ctx.storeForwardCancelCtx = context.WithCancel(context.Background())
	svc.runtime.StartStoreAndForward(svc.ctx.appWg, svc.ctx.appCtx, svc.ctx.storeForwardWg, storeForwardEnabledCtx, svc.serviceKey)
}

func (svc *Service) stopStoreForward() {
	svc.LoggingClient().Info("Canceling Store and Forward retry loop")
	svc.ctx.storeForwardCancelCtx()
	svc.ctx.storeForwardWg.Wait()
}

func (svc *Service) findMatchingFunction(configurable reflect.Value, functionName string) (reflect.Value, reflect.Type, error) {
	var functionValue reflect.Value
	count := configurable.Type().NumMethod()

	for index := 0; index < count; index++ {
		method := configurable.Type().Method(index)
		// If the target configuration function name starts with actual method name then it is a match
		if strings.Index(functionName, method.Name) == 0 {
			functionValue = configurable.MethodByName(method.Name)
			break
		}
	}

	if functionValue.Kind() == reflect.Invalid {
		return functionValue, nil, fmt.Errorf("function %s is not a built in SDK function", functionName)
	} else if functionValue.IsNil() {
		return functionValue, nil, fmt.Errorf("invalid/missing configuration for %s", functionName)
	}

	functionType := functionValue.Type()
	return functionValue, functionType, nil
}
