// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	gometrics "github.com/rcrowley/go-metrics"
)

const (
	eventsSentName           = "EventsSent"
	readingsSentName         = "ReadingsSent"
	DeviceServiceEventPrefix = "device"
)

// TODO: Refactor code in 3.0 to encapsulate this in a struct, factory func and
var eventsSent gometrics.Counter
var readingsSent gometrics.Counter

func UpdateLastConnected(name string, lc logger.LoggingClient, dc interfaces.DeviceClient) {
	t := time.Now().UnixNano() / int64(time.Millisecond)
	device := dtos.UpdateDevice{
		Name:          &name,
		LastConnected: &t,
	}

	req := requests.NewUpdateDeviceRequest(device)
	_, err := dc.Update(context.Background(), []requests.UpdateDeviceRequest{req})
	if err != nil {
		lc.Errorf("failed to update LastConnected for Device %s in Core Metadata: %v", name, err)
	}
}

func UpdateOperatingState(name string, state string, lc logger.LoggingClient, dc interfaces.DeviceClient) {
	device := dtos.UpdateDevice{
		Name:           &name,
		OperatingState: &state,
	}

	req := requests.NewUpdateDeviceRequest(device)
	_, err := dc.Update(context.Background(), []requests.UpdateDeviceRequest{req})
	if err != nil {
		lc.Errorf("failed to update OperatingState for Device %s in Core Metadata: %v", name, err)
	}
}

func SendEvent(event *dtos.Event, correlationID string, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, correlationID) // nolint: staticcheck
	req := requests.NewAddEventRequest(*event)

	bytes, encoding, err := req.Encode()
	if err != nil {
		lc.Error(err.Error())
	}

	// Check event size in kilobytes
	if configuration.MaxEventSize > 0 && int64(len(bytes)) > configuration.MaxEventSize*1024 {
		lc.Errorf("event size exceed MaxEventSize(%d KB)", configuration.MaxEventSize)
		return
	}

	sent := false
	mc := bootstrapContainer.MessagingClientFrom(dic.Get)
	ctx = context.WithValue(ctx, common.ContentType, encoding) // nolint: staticcheck
	envelope := types.NewMessageEnvelope(bytes, ctx)
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	publishTopic := common.BuildTopic(configuration.MessageBus.GetBaseTopicPrefix(), common.EventsPublishTopic, DeviceServiceEventPrefix, serviceName, event.ProfileName, event.DeviceName, event.SourceName)
	err = mc.Publish(envelope, publishTopic)
	if err != nil {
		lc.Errorf("Failed to publish event to MessageBus: %s", err)
		return
	}
	lc.Debugf("Event(profileName: %s, deviceName: %s, sourceName: %s, id: %s) published to MessageBus on topic: %s",
		event.ProfileName, event.DeviceName, event.SourceName, event.Id, publishTopic)
	sent = true

	if sent {
		eventsSent.Inc(1)
		readingsSent.Inc(int64(len(event.Readings)))
	}
}

func InitializeSentMetrics(lc logger.LoggingClient, dic *di.Container) {
	eventsSent = gometrics.NewCounter()
	readingsSent = gometrics.NewCounter()

	metricsManager := bootstrapContainer.MetricsManagerFrom(dic.Get)
	if metricsManager != nil {
		registerMetric(metricsManager, lc, eventsSentName, eventsSent)
		registerMetric(metricsManager, lc, readingsSentName, readingsSent)
	} else {
		lc.Warn("MetricsManager not available to register Event/Reading Sent metrics")
	}
}

func registerMetric(metricsManager bootstrapInterfaces.MetricsManager, lc logger.LoggingClient, name string, metric interface{}) {
	err := metricsManager.Register(name, metric, nil)
	if err != nil {
		lc.Errorf("unable to register %s metric. Metric will not be reported: %v", name, err)
	} else {
		lc.Debugf("%s metric has been registered and will be reported (if enabled)", name)
	}
}

func AddEventTags(event *dtos.Event) {
	if event.Tags == nil {
		event.Tags = make(map[string]interface{})
	}
	cmd, cmdExist := cache.Profiles().DeviceCommand(event.ProfileName, event.SourceName)
	if cmdExist && len(cmd.Tags) > 0 {
		for k, v := range cmd.Tags {
			event.Tags[k] = v
		}
	}
	device, deviceExist := cache.Devices().ForName(event.DeviceName)
	if deviceExist && len(device.Tags) > 0 {
		for k, v := range device.Tags {
			event.Tags[k] = v
		}
	}
}

func AddReadingTags(reading *dtos.BaseReading) {
	dr, drExist := cache.Profiles().DeviceResource(reading.ProfileName, reading.ResourceName)
	if drExist && len(dr.Tags) > 0 {
		if reading.Tags == nil {
			reading.Tags = make(map[string]interface{})
		}
		for k, v := range dr.Tags {
			reading.Tags[k] = v
		}
	}
}
