// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// processAsyncResults processes readings that are pushed from
// a DS implementation. Each is reading is optionally transformed
// before being pushed to Core Data.
func processAsyncResults() {
	for !svc.stopped {
		acv := <-svc.asyncCh
		readings := make([]contract.Reading, 0, len(acv.CommandValues))

		device, ok := cache.Devices().ForName(acv.DeviceName)
		if !ok {
			common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - recieved Device %s not found in cache", acv.DeviceName))
			continue
		}

		for _, cv := range acv.CommandValues {
			// get the device resource associated with the rsp.RO
			dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, cv.DeviceResourceName)
			if !ok {
				common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - Device Resource %s not found in Device %s", cv.DeviceResourceName, acv.DeviceName))
				continue
			}

			if common.CurrentConfig.Device.DataTransform {
				err := transformer.TransformReadResult(cv, dr.Properties.Value)
				if err != nil {
					common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - CommandValue (%s) transformed failed: %v", cv.String(), err))
					cv = dsModels.NewStringValue(cv.DeviceResourceName, cv.Origin, fmt.Sprintf("Transformation failed for device resource, with value: %s, property value: %v, and error: %v", cv.String(), dr.Properties.Value, err))
				}
			}

			err := transformer.CheckAssertion(cv, dr.Properties.Value.Assertion, &device)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - Assertion failed for device resource: %s, with value: %s and assertion: %s, %v", cv.DeviceResourceName, cv.String(), dr.Properties.Value.Assertion, err))
				cv = dsModels.NewStringValue(cv.DeviceResourceName, cv.Origin, fmt.Sprintf("Assertion failed for device resource, with value: %s and assertion: %s", cv.String(), dr.Properties.Value.Assertion))
			}

			ro, err := cache.Profiles().ResourceOperation(device.Profile.Name, cv.DeviceResourceName, common.GetCmdMethod)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("processAsyncResults - getting resource operation failed: %s", err.Error()))
				continue
			}

			if len(ro.Mappings) > 0 {
				newCV, ok := transformer.MapCommandValue(cv, ro.Mappings)
				if ok {
					cv = newCV
				} else {
					common.LoggingClient.Warn(fmt.Sprintf("processAsyncResults - Mapping failed for Device Resource Operation: %s, with value: %s, %v", ro.Resource, cv.String(), err))
				}
			}

			reading := common.CommandValueToReading(cv, device.Name, dr.Properties.Value.FloatEncoding)
			readings = append(readings, *reading)
		}

		// push to Core Data
		cevent := contract.Event{Device: device.Name, Readings: readings}
		event := &dsModels.Event{Event: cevent}
		event.Origin = time.Now().UnixNano()
		common.SendEvent(event)
	}
}
