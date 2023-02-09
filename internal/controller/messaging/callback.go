//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"fmt"

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

const MetadataSystemEventTopic = "MetadataSystemEventTopic"

func MetadataSystemEventCallback(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageBus
	metadataSystemEventTopic := messageBusInfo.Topics[MetadataSystemEventTopic]

	messages := make(chan types.MessageEnvelope)
	messageErrors := make(chan error)
	topics := []types.TopicChannel{
		{
			Topic:    metadataSystemEventTopic,
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
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", metadataSystemEventTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("System event received on message queue. Topic: %s, Correlation-id: %s ", metadataSystemEventTopic, msgEnvelope.CorrelationID)

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

				switch systemEvent.Type {
				case common.DeviceSystemEventType:
					err = deviceSystemEventAction(systemEvent, dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				case common.DeviceProfileSystemEventType:
					err = deviceProfileSystemEventAction(systemEvent, dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				default:
					lc.Errorf("unknown system event type %s", systemEvent.Type)
					continue
				}
			}
		}
	}()

	return nil
}

func deviceSystemEventAction(systemEvent dtos.SystemEvent, dic *di.Container) error {
	var device dtos.Device
	err := systemEvent.DecodeDetails(&device)
	if err != nil {
		return fmt.Errorf("failed to decode %s system event details: %s", systemEvent.Type, err.Error())
	}

	switch systemEvent.Action {
	case common.SystemEventActionAdd:
		err = application.AddDevice(requests.NewAddDeviceRequest(device), dic)
	case common.SystemEventActionUpdate:
		deviceModel := dtos.ToDeviceModel(device)
		updateDeviceDTO := dtos.FromDeviceModelToUpdateDTO(deviceModel)
		err = application.UpdateDevice(requests.NewUpdateDeviceRequest(updateDeviceDTO), dic)
	case common.SystemEventActionDelete:
		err = application.DeleteDevice(device.Name, dic)
	default:
		return fmt.Errorf("unknown %s system event action %s", systemEvent.Type, systemEvent.Action)
	}

	return err
}

func deviceProfileSystemEventAction(systemEvent dtos.SystemEvent, dic *di.Container) error {
	var deviceProfile dtos.DeviceProfile
	err := systemEvent.DecodeDetails(&deviceProfile)
	if err != nil {
		return fmt.Errorf("failed to decode %s system event details: %s", systemEvent.Type, err.Error())
	}

	switch systemEvent.Action {
	case common.SystemEventActionUpdate:
		err = application.UpdateProfile(requests.NewDeviceProfileRequest(deviceProfile), dic)
	// there is no action needed for Device Profile Add and Delete in Device Service
	case common.SystemEventActionAdd, common.SystemEventActionDelete:
		break
	default:
		return fmt.Errorf("unknown %s system event action %s", systemEvent.Type, systemEvent.Action)
	}

	return err
}
