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
//

package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/contracts"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

const (
	defaultMinRetryInterval = 1 * time.Second
)

type storeForwardInfo struct {
	runtime *GolangRuntime
	dic     *di.Container
}

func (sf *storeForwardInfo) startStoreAndForwardRetryLoop(
	appWg *sync.WaitGroup,
	appCtx context.Context,
	enabledWg *sync.WaitGroup,
	enabledCtx context.Context,
	serviceKey string) {

	appWg.Add(1)
	enabledWg.Add(1)

	config := container.ConfigurationFrom(sf.dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(sf.dic.Get)

	go func() {
		defer appWg.Done()
		defer enabledWg.Done()

		retryInterval, err := time.ParseDuration(config.Writable.StoreAndForward.RetryInterval)
		if err != nil {
			lc.Warn(
				fmt.Sprintf("StoreAndForward RetryInterval failed to parse, defaulting to %s",
					defaultMinRetryInterval.String()))
			retryInterval = defaultMinRetryInterval
		} else if retryInterval < defaultMinRetryInterval {
			lc.Warn(
				fmt.Sprintf("StoreAndForward RetryInterval value %s is less than the allowed minimum value, defaulting to %s",
					retryInterval.String(), defaultMinRetryInterval.String()))
			retryInterval = defaultMinRetryInterval
		}

		if config.Writable.StoreAndForward.MaxRetryCount < 0 {
			lc.Warn("StoreAndForward MaxRetryCount can not be less than 0, defaulting to 1")
			config.Writable.StoreAndForward.MaxRetryCount = 1
		}

		lc.Info(
			fmt.Sprintf("Starting StoreAndForward Retry Loop with %s RetryInterval and %d max retries",
				retryInterval.String(), config.Writable.StoreAndForward.MaxRetryCount))

	exit:
		for {
			select {

			case <-appCtx.Done():
				// Exit the loop and function when application service is terminating.
				break exit

			case <-enabledCtx.Done():
				// Exit the loop and function when Store and Forward has been disabled.
				break exit

			case <-time.After(retryInterval):
				sf.retryStoredData(serviceKey)
			}
		}

		lc.Info("Exiting StoreAndForward Retry Loop")
	}()
}

func (sf *storeForwardInfo) storeForLaterRetry(
	payload []byte,
	appContext interfaces.AppFunctionContext,
	pipeline *interfaces.FunctionPipeline,
	pipelinePosition int) {

	item := contracts.NewStoredObject(sf.runtime.ServiceKey, payload, pipeline.Id, pipelinePosition, pipeline.Hash, appContext.GetAllValues())
	item.CorrelationID = appContext.CorrelationID()

	appContext.LoggingClient().Tracef("Storing data for later retry for pipeline '%s' (%s=%s)",
		pipeline.Id,
		common.CorrelationHeader,
		appContext.CorrelationID())

	config := container.ConfigurationFrom(sf.dic.Get)
	if !config.Writable.StoreAndForward.Enabled {
		appContext.LoggingClient().Errorf("Failed to store item for later retry for pipeline '%s': StoreAndForward not enabled", pipeline.Id)
		return
	}

	storeClient := container.StoreClientFrom(sf.dic.Get)

	if _, err := storeClient.Store(item); err != nil {
		appContext.LoggingClient().Errorf("Failed to store item for later retry for pipeline '%s': %s", pipeline.Id, err.Error())
	}
}

func (sf *storeForwardInfo) retryStoredData(serviceKey string) {

	storeClient := container.StoreClientFrom(sf.dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(sf.dic.Get)

	items, err := storeClient.RetrieveFromStore(serviceKey)
	if err != nil {
		lc.Errorf("Unable to load store and forward items from DB: %s", err.Error())
		return
	}

	lc.Debugf("%d stored data items found for retrying", len(items))

	if len(items) > 0 {
		itemsToRemove, itemsToUpdate := sf.processRetryItems(items)

		lc.Debugf(" %d stored data items will be removed post retry", len(itemsToRemove))
		lc.Debugf(" %d stored data items will be update post retry", len(itemsToUpdate))

		for _, item := range itemsToRemove {
			if err := storeClient.RemoveFromStore(item); err != nil {
				lc.Errorf("Unable to remove stored data item for pipeline '%s' from DB, objectID=%s: %s",
					item.PipelineId,
					err.Error(),
					item.ID)
			}
		}

		for _, item := range itemsToUpdate {
			if err := storeClient.Update(item); err != nil {
				lc.Errorf("Unable to update stored data item for pipeline '%s' from DB, objectID=%s: %s",
					item.PipelineId,
					err.Error(),
					item.ID)
			}
		}
	}
}

func (sf *storeForwardInfo) processRetryItems(items []contracts.StoredObject) ([]contracts.StoredObject, []contracts.StoredObject) {
	lc := bootstrapContainer.LoggingClientFrom(sf.dic.Get)
	config := container.ConfigurationFrom(sf.dic.Get)

	var itemsToRemove []contracts.StoredObject
	var itemsToUpdate []contracts.StoredObject

	// Item will be removed from store if:
	//    - successfully retried
	//    - max retries exceeded
	//    - version no longer matches current Pipeline
	// Item will not be removed if retry failed and more retries available (hit 'continue' above)
	for _, item := range items {
		pipeline := sf.runtime.GetPipelineById(item.PipelineId)

		if pipeline == nil {
			lc.Errorf("Stored data item's pipeline '%s' no longer exists. Removing item from DB", item.PipelineId)
			itemsToRemove = append(itemsToRemove, item)
			continue
		}

		if item.Version != pipeline.Hash {
			lc.Error("Stored data item's pipeline Version doesn't match '%s' pipeline's Version. Removing item from DB", item.PipelineId)
			itemsToRemove = append(itemsToRemove, item)
			continue
		}

		if !sf.retryExportFunction(item, pipeline) {
			item.RetryCount++
			if config.Writable.StoreAndForward.MaxRetryCount == 0 ||
				item.RetryCount < config.Writable.StoreAndForward.MaxRetryCount {
				lc.Tracef("Export retry failed for pipeline '%s'. retries=%d, Incrementing retry count (%s=%s)",
					item.PipelineId,
					item.RetryCount,
					common.CorrelationHeader,
					item.CorrelationID)
				itemsToUpdate = append(itemsToUpdate, item)
				continue
			}

			lc.Tracef("Max retries exceeded for pipeline '%s'. retries=%d, Removing item from DB (%s=%s)",
				item.PipelineId,
				item.RetryCount,
				common.CorrelationHeader,
				item.CorrelationID)
			itemsToRemove = append(itemsToRemove, item)

			// Note that item will be removed for DB below.
		} else {
			lc.Tracef("Retry successful for pipeline '%s'. Removing item from DB (%s=%s)",
				item.PipelineId,
				common.CorrelationHeader,
				item.CorrelationID)
			itemsToRemove = append(itemsToRemove, item)
		}
	}

	return itemsToRemove, itemsToUpdate
}

func (sf *storeForwardInfo) retryExportFunction(item contracts.StoredObject, pipeline *interfaces.FunctionPipeline) bool {
	appContext := appfunction.NewContext(item.CorrelationID, sf.dic, "")

	for k, v := range item.ContextData {
		appContext.AddValue(strings.ToLower(k), v)
	}

	appContext.LoggingClient().Tracef("Retrying stored data for pipeline '%s' (%s=%s)",
		item.PipelineId,
		common.CorrelationHeader,
		appContext.CorrelationID())

	return sf.runtime.ExecutePipeline(
		item.Payload,
		"",
		appContext,
		pipeline,
		item.PipelinePosition,
		true) == nil
}
