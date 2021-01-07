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
//

package runtime

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/contracts"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

const (
	defaultMinRetryInterval = time.Duration(1 * time.Second)
)

type storeForwardInfo struct {
	runtime      *GolangRuntime
	storeClient  interfaces.StoreClient
	pipelineHash string
}

func (sf *storeForwardInfo) startStoreAndForwardRetryLoop(
	appWg *sync.WaitGroup,
	appCtx context.Context,
	enabledWg *sync.WaitGroup,
	enabledCtx context.Context,
	serviceKey string,
	config *common.ConfigurationStruct,
	edgeXClients common.EdgeXClients) {

	appWg.Add(1)
	enabledWg.Add(1)

	go func() {
		defer appWg.Done()
		defer enabledWg.Done()

		retryInterval, err := time.ParseDuration(config.Writable.StoreAndForward.RetryInterval)
		if err != nil {
			edgeXClients.LoggingClient.Warn(
				fmt.Sprintf("StoreAndForward RetryInterval failed to parse, defaulting to %s",
					defaultMinRetryInterval.String()))
			retryInterval = defaultMinRetryInterval
		} else if retryInterval < defaultMinRetryInterval {
			edgeXClients.LoggingClient.Warn(
				fmt.Sprintf("StoreAndForward RetryInterval value %s is less than the allowed minimum value, defaulting to %s",
					retryInterval.String(), defaultMinRetryInterval.String()))
			retryInterval = defaultMinRetryInterval
		}

		if config.Writable.StoreAndForward.MaxRetryCount < 0 {
			edgeXClients.LoggingClient.Warn(
				fmt.Sprintf("StoreAndForward MaxRetryCount can not be less than 0, defaulting to 1"))
			config.Writable.StoreAndForward.MaxRetryCount = 1
		}

		edgeXClients.LoggingClient.Info(
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
				sf.retryStoredData(serviceKey, config, edgeXClients)
			}
		}

		edgeXClients.LoggingClient.Info("Exiting StoreAndForward Retry Loop")
	}()
}

func (sf *storeForwardInfo) storeForLaterRetry(payload []byte,
	edgexcontext *appcontext.Context,
	pipelinePosition int) {

	item := contracts.NewStoredObject(sf.runtime.ServiceKey, payload, pipelinePosition, sf.pipelineHash)
	item.CorrelationID = edgexcontext.CorrelationID

	edgexcontext.LoggingClient.Trace("Storing data for later retry",
		clients.CorrelationHeader, edgexcontext.CorrelationID)

	if !edgexcontext.Configuration.Writable.StoreAndForward.Enabled {
		edgexcontext.LoggingClient.Error(
			"Failed to store item for later retry", "error", "StoreAndForward not enabled",
			clients.CorrelationHeader, item.CorrelationID)
		return
	}

	if _, err := sf.storeClient.Store(item); err != nil {
		edgexcontext.LoggingClient.Error("Failed to store item for later retry",
			"error", err,
			clients.CorrelationHeader, item.CorrelationID)
	}
}

func (sf *storeForwardInfo) retryStoredData(serviceKey string,
	config *common.ConfigurationStruct,
	edgeXClients common.EdgeXClients) {

	items, err := sf.storeClient.RetrieveFromStore(serviceKey)
	if err != nil {
		edgeXClients.LoggingClient.Error("Unable to load store and forward items from DB", "error", err)
		return
	}

	edgeXClients.LoggingClient.Debug(fmt.Sprintf(" %d stored data items found for retrying", len(items)))

	if len(items) > 0 {
		itemsToRemove, itemsToUpdate := sf.processRetryItems(items, config, edgeXClients)

		edgeXClients.LoggingClient.Debug(
			fmt.Sprintf(" %d stored data items will be removed post retry", len(itemsToRemove)))
		edgeXClients.LoggingClient.Debug(
			fmt.Sprintf(" %d stored data items will be update post retry", len(itemsToUpdate)))

		for _, item := range itemsToRemove {
			if err := sf.storeClient.RemoveFromStore(item); err != nil {
				edgeXClients.LoggingClient.Error(
					"Unable to remove stored data item from DB",
					"error", err,
					"objectID", item.ID,
					clients.CorrelationHeader, item.CorrelationID)
			}
		}

		for _, item := range itemsToUpdate {
			if err := sf.storeClient.Update(item); err != nil {
				edgeXClients.LoggingClient.Error("Unable to update stored data item in DB",
					"error", err,
					"objectID", item.ID,
					clients.CorrelationHeader, item.CorrelationID)
			}
		}
	}
}

func (sf *storeForwardInfo) processRetryItems(items []contracts.StoredObject,
	config *common.ConfigurationStruct,
	edgeXClients common.EdgeXClients) ([]contracts.StoredObject, []contracts.StoredObject) {

	var itemsToRemove []contracts.StoredObject
	var itemsToUpdate []contracts.StoredObject

	for _, item := range items {
		if item.Version == sf.calculatePipelineHash() {
			if !sf.retryExportFunction(item, config, edgeXClients) {
				item.RetryCount++
				if config.Writable.StoreAndForward.MaxRetryCount == 0 ||
					item.RetryCount < config.Writable.StoreAndForward.MaxRetryCount {
					edgeXClients.LoggingClient.Trace("Export retry failed. Incrementing retry count",
						"retries",
						item.RetryCount,
						clients.CorrelationHeader,
						item.CorrelationID)
					itemsToUpdate = append(itemsToUpdate, item)
					continue
				}

				edgeXClients.LoggingClient.Trace(
					"Max retries exceeded. Removing item from DB", "retries",
					item.RetryCount,
					clients.CorrelationHeader,
					item.CorrelationID)
				// Note that item will be removed for DB below.
			} else {
				edgeXClients.LoggingClient.Trace(
					"Export retry successful. Removing item from DB",
					clients.CorrelationHeader,
					item.CorrelationID)
			}
		} else {
			edgeXClients.LoggingClient.Error(
				"Stored data item's Function Pipeline Version doesn't match current Function Pipeline Version. Removing item from DB",
				clients.CorrelationHeader,
				item.CorrelationID)
		}

		// Item will be remove from store if:
		//    - successfully retried
		//    - max retries exceeded
		//    - version no longer matches current Pipeline
		// Item will not be removed if retry failed and more retries available (hit 'continue' above)
		itemsToRemove = append(itemsToRemove, item)
	}

	return itemsToRemove, itemsToUpdate
}

func (sf *storeForwardInfo) retryExportFunction(item contracts.StoredObject, config *common.ConfigurationStruct,
	edgeXClients common.EdgeXClients) bool {
	edgexContext := &appcontext.Context{
		CorrelationID:         item.CorrelationID,
		Configuration:         config,
		LoggingClient:         edgeXClients.LoggingClient,
		EventClient:           edgeXClients.EventClient,
		ValueDescriptorClient: edgeXClients.ValueDescriptorClient,
		CommandClient:         edgeXClients.CommandClient,
		NotificationsClient:   edgeXClients.NotificationsClient,
	}

	edgexContext.LoggingClient.Trace("Retrying stored data", clients.CorrelationHeader, edgexContext.CorrelationID)

	return sf.runtime.ExecutePipeline(
		item.Payload,
		"",
		edgexContext,
		sf.runtime.transforms,
		item.PipelinePosition,
		true) == nil
}

func (sf *storeForwardInfo) calculatePipelineHash() string {
	hash := "Pipeline-functions: "
	for _, item := range sf.runtime.transforms {
		name := runtime.FuncForPC(reflect.ValueOf(item).Pointer()).Name()
		hash = hash + " " + name
	}

	return hash
}
