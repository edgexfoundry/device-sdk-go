//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/application"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
)

const SystemEventTopic = "SystemEventTopic"

func DeviceCallback(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageBus
	systemEventTopic := messageBusInfo.Topics[SystemEventTopic]

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{
			Topic:    systemEventTopic,
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
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", systemEventTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("System event received on message queue. Topic: %s, Correlation-id: %s ", systemEventTopic, msgEnvelope.CorrelationID)

				var systemEvent dtos.SystemEvent
				err := json.Unmarshal(msgEnvelope.Payload, &systemEvent)
				if err != nil {
					lc.Errorf("failed to JSON decoding system event: %s", err.Error())
					continue
				}

				serviceName := container.DeviceServiceFrom(dic.Get).Name
				if systemEvent.Owner != serviceName {
					lc.Errorf("unmatched system event owner %s with service name %s", systemEvent.Owner, serviceName)
					continue
				}

				if systemEvent.Type != common.DeviceSystemEventType {
					lc.Errorf("unknown system event type %s", systemEvent.Type)
					continue
				}

				var device dtos.Device
				err = systemEvent.DecodeDetails(&device)
				if err != nil {
					lc.Errorf("failed to decode system event details: %s", err.Error())
					continue
				}

				switch systemEvent.Action {
				case common.DeviceSystemEventActionAdd:
					err = application.AddDevice(requests.NewAddDeviceRequest(device), dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				case common.DeviceSystemEventActionUpdate:
					deviceModel := dtos.ToDeviceModel(device)
					updateDeviceDTO := dtos.FromDeviceModelToUpdateDTO(deviceModel)
					err = application.UpdateDevice(requests.NewUpdateDeviceRequest(updateDeviceDTO), dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				case common.DeviceSystemEventActionDelete:
					err = application.DeleteDevice(device.Name, dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				default:
					lc.Errorf("unknown device system event action %s", systemEvent.Action)
				}
			}
		}
	}()

	return nil
}
