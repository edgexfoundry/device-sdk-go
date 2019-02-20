// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

// processAsyncResults processes readings that are pushed from
// a DS implementation. Each is reading is optionally transformed
// before being pushed to Core Data.
func processAsyncResults() {
	for !svc.stopped {
		acv := <-svc.asyncCh
		readings := make([]models.Reading, 0, len(acv.CommandValues))

		device, ok := cache.Devices().ForName(acv.DeviceName)
		if !ok {
			common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - recieved Device %s not found in cache", acv.DeviceName))
			continue
		}

		for _, cv := range acv.CommandValues {
			// get the device resource associated with the rsp.RO
			dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, cv.RO.Object)
			if !ok {
				common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - Device Resource %s not found in Device %s", cv.RO.Object, acv.DeviceName))
				continue
			}

			if common.CurrentConfig.Device.DataTransform {
				err := transformer.TransformReadResult(cv, dr.Properties.Value)
				if err != nil {
					common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - CommandValue (%s) transformed failed: %v", cv.String(), err))
					cv = ds_models.NewStringValue(cv.RO, cv.Origin, fmt.Sprintf("Transformation failed for device resource, with value: %s, property value: %v, and error: %v", cv.String(), dr.Properties.Value, err))
				}
			}

			err := transformer.CheckAssertion(cv, dr.Properties.Value.Assertion, &device)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - Assertion failed for device resource: %s, with value: %s and assertion: %s, %v", cv.RO.Object, cv.String(), dr.Properties.Value.Assertion, err))
				cv = ds_models.NewStringValue(cv.RO, cv.Origin, fmt.Sprintf("Assertion failed for device resource, with value: %s and assertion: %s", cv.String(), dr.Properties.Value.Assertion))
			}

			if len(cv.RO.Mappings) > 0 {
				newCV, ok := transformer.MapCommandValue(cv)
				if ok {
					cv = newCV
				} else {
					common.LoggingClient.Warn(fmt.Sprintf("processAsyncResults - Mapping failed for Device Resource Operation: %v, with value: %s, %v", cv.RO, cv.String(), err))
				}
			}

			reading := common.CommandValueToReading(cv, device.Name)
			readings = append(readings, *reading)
		}

		// push to Core Data
		event := &models.Event{Device: acv.DeviceName, Readings: readings}
		ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
		_, err := common.EventClient.Add(event, ctx)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - Failed to push event %v: %v", event, err))
		}
	}
}
