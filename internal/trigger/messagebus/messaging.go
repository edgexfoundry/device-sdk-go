//
// Copyright (c) 2019 Intel Corporation
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

package messagebus

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-messaging/messaging"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

// Trigger implements Trigger to support MessageBusData
type Trigger struct {
	Configuration common.ConfigurationStruct
	Runtime       *runtime.GolangRuntime
	client        messaging.MessageClient
	topics        []types.TopicChannel
	EdgeXClients  common.EdgeXClients
}

// Initialize ...
func (trigger *Trigger) Initialize(appWg *sync.WaitGroup, appCtx context.Context) error {
	var err error
	logger := trigger.EdgeXClients.LoggingClient

	logger.Info(fmt.Sprintf("Initializing Message Bus Trigger. Subscribing to topic: %s on port %d , Publish Topic: %s on port %d", trigger.Configuration.Binding.SubscribeTopic, trigger.Configuration.MessageBus.SubscribeHost.Port, trigger.Configuration.Binding.PublishTopic, trigger.Configuration.MessageBus.PublishHost.Port))

	trigger.client, err = messaging.NewMessageClient(trigger.Configuration.MessageBus)
	if err != nil {
		return err
	}
	trigger.topics = []types.TopicChannel{{Topic: trigger.Configuration.Binding.SubscribeTopic, Messages: make(chan types.MessageEnvelope)}}
	messageErrors := make(chan error)

	trigger.client.Subscribe(trigger.topics, messageErrors)
	receiveMessage := true

	appWg.Add(1)

	go func() {
		defer appWg.Done()

		for receiveMessage {
			select {
			case <-appCtx.Done():
				return

			case msgErr := <-messageErrors:
				logger.Error(fmt.Sprintf("Failed to receive ZMQ Message, %v", msgErr))

			case msgs := <-trigger.topics[0].Messages:
				go func() {
					logger.Trace("Received message from bus", "topic", trigger.Configuration.Binding.SubscribeTopic, clients.CorrelationHeader, msgs.CorrelationID)

					edgexContext := &appcontext.Context{
						CorrelationID:         msgs.CorrelationID,
						Configuration:         trigger.Configuration,
						LoggingClient:         trigger.EdgeXClients.LoggingClient,
						EventClient:           trigger.EdgeXClients.EventClient,
						ValueDescriptorClient: trigger.EdgeXClients.ValueDescriptorClient,
						CommandClient:         trigger.EdgeXClients.CommandClient,
						NotificationsClient:   trigger.EdgeXClients.NotificationsClient,
					}

					messageError := trigger.Runtime.ProcessMessage(edgexContext, msgs)
					if messageError != nil {
						// ProcessMessage logs the error, so no need to log it here.
						return
					}

					if edgexContext.OutputData != nil {
						outputEnvelope := types.MessageEnvelope{
							CorrelationID: edgexContext.CorrelationID,
							Payload:       edgexContext.OutputData,
							ContentType:   clients.ContentTypeJSON,
						}
						err := trigger.client.Publish(outputEnvelope, trigger.Configuration.Binding.PublishTopic)
						if err != nil {
							logger.Error(fmt.Sprintf("Failed to publish Message to bus, %v", err))
						}

						logger.Trace("Published message to bus", "topic", trigger.Configuration.Binding.PublishTopic, clients.CorrelationHeader, msgs.CorrelationID)
					}
				}()
			}
		}
	}()

	return nil
}
