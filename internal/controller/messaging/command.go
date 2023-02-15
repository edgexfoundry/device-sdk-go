//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/application"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v3/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
)

func SubscribeCommands(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageBus
	deviceService := container.DeviceServiceFrom(dic.Get)

	requestSubscribeTopic := common.BuildTopic(messageBusInfo.GetBaseTopicPrefix(), common.CommandRequestSubscribeTopic, deviceService.Name, "#")
	lc.Infof("Subscribing to command requests on topic: %s", requestSubscribeTopic)

	responsePublishTopicPrefix := common.BuildTopic(messageBusInfo.GetBaseTopicPrefix(), common.ResponseTopic, deviceService.Name)
	lc.Infof("Responses to command requests will be published on topic: %s/<requestId>", responsePublishTopicPrefix)

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{
			Topic:    requestSubscribeTopic,
			Messages: messages,
		},
	}

	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	err := messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", requestSubscribeTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("Command request received on message queue. Topic: %s, Correlation-id: %s", msgEnvelope.ReceivedTopic, msgEnvelope.CorrelationID)

				// expected command request topic scheme: #/<service-name>/<device-name>/<command-name>/<method>
				topicLevels := strings.Split(msgEnvelope.ReceivedTopic, "/")
				length := len(topicLevels)
				if length < 4 {
					lc.Error("Failed to parse and construct command response topic scheme, expected request topic scheme: '#/<service-name>/<device-name>/<command-name>/<method>'")
					continue
				}

				// expected command response topic scheme: #/<service-name>/<device-name>/<command-name>/<method>
				deviceName := topicLevels[length-3]
				commandName := topicLevels[length-2]
				method := topicLevels[length-1]

				responsePublishTopic := common.BuildTopic(responsePublishTopicPrefix, msgEnvelope.RequestID)

				switch strings.ToUpper(method) {
				case "GET":
					getCommand(ctx, msgEnvelope, responsePublishTopic, deviceName, commandName, dic)
				case "SET":
					setCommand(ctx, msgEnvelope, responsePublishTopic, deviceName, commandName, dic)
				default:
					lc.Errorf("unknown command method '%s', only 'get' or 'set' is allowed", method)
					continue
				}

				lc.Debugf("Command response published on message queue. Topic: %s, Correlation-id: %s", responsePublishTopic, msgEnvelope.CorrelationID)
			}
		}
	}()

	return nil
}

func getCommand(ctx context.Context, msgEnvelope types.MessageEnvelope, responseTopic string, deviceName string, commandName string, dic *di.Container) {
	var responseEnvelope types.MessageEnvelope

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	rawQuery, pushEvent, returnEvent := filterQueryParams(msgEnvelope.QueryParams)

	// TODO: fix properly in EdgeX 3.0
	ctx = context.WithValue(ctx, common.CorrelationHeader, msgEnvelope.CorrelationID) // nolint: staticcheck
	event, edgexErr := application.GetCommand(ctx, deviceName, commandName, rawQuery, dic)
	if edgexErr != nil {
		lc.Errorf("Failed to process get device command %s for device %s: %s", commandName, deviceName, edgexErr.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, edgexErr.Error())
		err := messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command error response: %s", err.Error())
		}
		return
	}

	var err error
	var encoding string
	var eventResponseBytes []byte
	if returnEvent {
		eventResponse := responses.NewEventResponse(msgEnvelope.RequestID, "", http.StatusOK, *event)
		eventResponseBytes, encoding, err = eventResponse.Encode()
		if err != nil {
			lc.Errorf("Failed to encode event response: %s", err.Error())
			responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
			err = messageBus.Publish(responseEnvelope, responseTopic)
			if err != nil {
				lc.Errorf("Failed to publish command error response: %s", err.Error())
			}
			return
		}
	} else {
		eventResponseBytes = nil
		encoding = common.ContentTypeJSON
	}

	responseEnvelope, err = types.NewMessageEnvelopeForResponse(eventResponseBytes, msgEnvelope.RequestID, msgEnvelope.CorrelationID, encoding)
	if err != nil {
		lc.Errorf("Failed to create response message envelope: %s", err.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command error response: %s", err.Error())
		}
		return
	}

	err = messageBus.Publish(responseEnvelope, responseTopic)
	if err != nil {
		lc.Errorf("Failed to publish command response: %s", err.Error())
		return
	}

	if pushEvent {
		go sdkCommon.SendEvent(event, msgEnvelope.CorrelationID, dic)
	}

}

func setCommand(ctx context.Context, msgEnvelope types.MessageEnvelope, responseTopic string, deviceName string, commandName string, dic *di.Container) {
	var responseEnvelope types.MessageEnvelope

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	rawQuery, _, _ := filterQueryParams(msgEnvelope.QueryParams)

	requestPayload := make(map[string]any)
	err := json.Unmarshal(msgEnvelope.Payload, &requestPayload)
	if err != nil {
		lc.Errorf("Failed to decode set command request payload: %s", err.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command response: %s", err.Error())
		}
		return
	}

	// TODO: fix properly in EdgeX 3.0
	ctx = context.WithValue(ctx, common.CorrelationHeader, msgEnvelope.CorrelationID) // nolint: staticcheck
	edgexErr := application.SetCommand(ctx, deviceName, commandName, rawQuery, requestPayload, dic)
	if edgexErr != nil {
		lc.Errorf("Failed to process set device command %s for device %s: %s", commandName, deviceName, edgexErr.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, edgexErr.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command response: %s", err.Error())
		}
		return
	}

	responseEnvelope, err = types.NewMessageEnvelopeForResponse(nil, msgEnvelope.RequestID, msgEnvelope.CorrelationID, common.ContentTypeJSON)
	if err != nil {
		lc.Errorf("Failed to create response message envelope: %s", err.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command response: %s", err.Error())
		}
		return
	}

	err = messageBus.Publish(responseEnvelope, responseTopic)
	if err != nil {
		lc.Errorf("Failed to publish command response: %s", err.Error())
		return
	}
}

func filterQueryParams(queries map[string]string) (string, bool, bool) {
	var rawQuery []string
	pushEvent, returnEvent := false, true
	for k, v := range queries {
		if k == common.PushEvent && v == common.ValueTrue {
			pushEvent = true
			continue
		}
		if k == common.ReturnEvent && v == common.ValueFalse {
			returnEvent = false
			continue
		}
		if strings.HasPrefix(k, sdkCommon.SDKReservedPrefix) {
			continue
		}
		rawQuery = append(rawQuery, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(rawQuery, "&"), pushEvent, returnEvent
}
