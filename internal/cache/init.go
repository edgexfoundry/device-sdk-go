// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

var (
	initOnce sync.Once
)

// Init basic state for cache
func InitCache() {
	initOnce.Do(func() {
		ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())

		vds, err := common.ValueDescriptorClient.ValueDescriptors(ctx)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("Value Descriptor cache initialization failed: %v", err))
			vds = make([]models.ValueDescriptor, 0)
		}
		newValueDescriptorCache(vds)

		ds, err := common.DeviceClient.DevicesForServiceByName(common.ServiceName, ctx)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("Device cache initialization failed: %v", err))
			ds = make([]models.Device, 0)
		}
		newDeviceCache(ds)

		dps := make([]models.DeviceProfile, len(ds))
		for i, d := range ds {
			dps[i] = d.Profile
		}
		newProfileCache(dps)

		ses, err := common.ScheduleEventClient.ScheduleEventsForServiceByName(common.ServiceName, ctx)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("Schedule Event cache initialization failed: %v", err))
			ses = make([]models.ScheduleEvent, 0)
		}
		newScheduleEventCache(ses)

		schMap := make(map[string]models.Schedule, len(ses))
		for _, se := range ses {
			if _, ok := schMap[se.Schedule]; !ok {
				sc, err := common.ScheduleClient.ScheduleForName(se.Schedule, ctx)
				if err != nil {
					common.LoggingClient.Error(fmt.Sprintf("Schedule %s cannot be found in Core Metadata", se.Schedule))
					continue
				}
				schMap[sc.Name] = sc
			}
		}
		newScheduleCache(schMap)
	})
}
