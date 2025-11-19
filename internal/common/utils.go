// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/localqueue"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

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
	encoding := req.GetEncodingContentType()
	ctx = context.WithValue(ctx, common.ContentType, encoding) // nolint: staticcheck
	envelope := types.NewMessageEnvelope(req, ctx)

	sent := false
	mc := bootstrapContainer.MessagingClientFrom(dic.Get)
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	publishTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(configuration.MessageBus.GetBaseTopicPrefix()).SetPath(common.EventsPublishTopic).SetPath(DeviceServiceEventPrefix).
		SetNameFieldPath(serviceName).SetNameFieldPath(event.ProfileName).SetNameFieldPath(event.DeviceName).SetNameFieldPath(event.SourceName).BuildPath()
	err := mc.PublishWithSizeLimit(envelope, publishTopic, configuration.MaxEventSize)
	if err != nil {
		lc.Errorf("Failed to publish event to MessageBus: %s", err)
		// If local queue is enabled, persist the event for retry
		if configuration.LocalQueue.Enabled {
			// marshal the AddEventRequest so we can restore it later
			data, mErr := json.Marshal(req)
			if mErr != nil {
				lc.Errorf("Failed to marshal event for local queue: %v", mErr)
				return
			}
			// initialize queue singleton and worker
			ensureLocalQueueAndWorker(configuration, dic)
			if localQ != nil {
				if _, qErr := localQ.Enqueue(data); qErr != nil {
					lc.Errorf("Failed to enqueue event to local queue: %v", qErr)
				} else {
					lc.Infof("Event enqueued to local queue for later delivery. Event id: %s", event.Id)
				}
			}
		}
		return
	}
	lc.Debugf("Event(profileName: %s, deviceName: %s, sourceName: %s, id: %s) published to MessageBus on topic: %s",
		event.ProfileName, event.DeviceName, event.SourceName, event.Id, publishTopic)
	sent = true

	if sent && eventsSent != nil && readingsSent != nil {
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

// local queue singleton and worker
var (
	localQ     *localqueue.LocalQueue
	localQOnce sync.Once
)

func ensureLocalQueueAndWorker(configuration *config.ConfigurationStruct, dic *di.Container) {
	localQOnce.Do(func() {
		dbPath := configuration.LocalQueue.DBPath
		if dbPath == "" {
			dbPath = "localqueue.db"
		}
		maxItems := configuration.LocalQueue.MaxItems
		if maxItems <= 0 {
			maxItems = 10000
		}
		itemTTL := time.Duration(configuration.LocalQueue.ItemTTLSeconds) * time.Second
		if itemTTL <= 0 {
			itemTTL = 7 * 24 * time.Hour // Default 7 days
		}
		q, err := localqueue.NewLocalQueue(dbPath, maxItems, itemTTL)
		if err != nil {
			bootstrapContainer.LoggingClientFrom(dic.Get).Errorf("failed to open local queue DB: %v", err)
			return
		}
		localQ = q

		// start background worker
		retrySec := configuration.LocalQueue.RetryIntervalSeconds
		if retrySec <= 0 {
			retrySec = 5
		}
		go func() {
			lc := bootstrapContainer.LoggingClientFrom(dic.Get)
			mc := bootstrapContainer.MessagingClientFrom(dic.Get)
			for {
				items, err := localQ.DequeuePending(10)
				if err != nil {
					lc.Errorf("local queue dequeue error: %v", err)
					time.Sleep(time.Duration(retrySec) * time.Second)
					continue
				}
				if len(items) == 0 {
					time.Sleep(time.Duration(retrySec) * time.Second)
					continue
				}
				for _, it := range items {
					var req requests.AddEventRequest
					if err := json.Unmarshal(it.Payload, &req); err != nil {
						lc.Errorf("failed to unmarshal queued event: %v", err)
						// mark sent to avoid tight loop on bad payload
						_ = localQ.MarkSent(it.ID)
						continue
					}
					// rebuild envelope and try publish
					// use the stored ctx ContentType if available
					enc := req.GetEncodingContentType()
					ctx := context.WithValue(context.Background(), common.ContentType, enc)
					envelope := types.NewMessageEnvelope(req, ctx)
					publishTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
						SetPath(configuration.MessageBus.GetBaseTopicPrefix()).SetPath(common.EventsPublishTopic).SetPath(DeviceServiceEventPrefix).
						SetNameFieldPath(container.DeviceServiceFrom(dic.Get).Name).SetNameFieldPath(req.Event.ProfileName).SetNameFieldPath(req.Event.DeviceName).SetNameFieldPath(req.Event.SourceName).BuildPath()
					if err := mc.PublishWithSizeLimit(envelope, publishTopic, configuration.MaxEventSize); err != nil {
						lc.Errorf("retry publish failed for queued event %s: %v", it.ID, err)
						continue
					}
					lc.Infof("queued event %s published to MessageBus", it.ID)
					_ = localQ.MarkSent(it.ID)
				}
			}
		}()
	})
}
