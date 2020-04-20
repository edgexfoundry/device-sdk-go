// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// processAsyncResults processes readings that are pushed from
// a DS implementation. Each is reading is optionally transformed
// before being pushed to Core Data.
func processAsyncResults(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer func() {
		close(svc.asyncCh)
		wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case acv := <-svc.asyncCh:
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
					common.LoggingClient.Debug(fmt.Sprintf("processAsyncResults - getting resource operation failed: %s", err.Error()))
				} else if len(ro.Mappings) > 0 {
					newCV, ok := transformer.MapCommandValue(cv, ro.Mappings)
					if ok {
						cv = newCV
					} else {
						common.LoggingClient.Warn(fmt.Sprintf("processAsyncResults - Mapping failed for Device Resource Operation: %s, with value: %s, %v", ro.DeviceCommand, cv.String(), err))
					}
				}

				reading := common.CommandValueToReading(cv, device.Name, dr.Properties.Value.MediaType, dr.Properties.Value.FloatEncoding)
				readings = append(readings, *reading)
			}

			// push to Core Data
			cevent := contract.Event{Device: device.Name, Readings: readings}
			event := &dsModels.Event{Event: cevent}
			event.Origin = common.GetUniqueOrigin()
			common.SendEvent(event)

		}
	}
}

// processAsyncFilterAndAdd filter and add devices discovered by
// device service protocol discovery.
func processAsyncFilterAndAdd(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer func() {
		close(svc.deviceCh)
		wg.Done()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case devices := <-svc.deviceCh:
			id := handler.ReleaseLock()
			pws := cache.ProvisionWatchers().All()
			ctx := context.WithValue(context.Background(), common.CorrelationHeader, id)
			for _, d := range devices {
				for _, pw := range pws {
					if !whitelistPass(d, pw) {
						break
					}
					if !blacklistPass(d, pw) {
						break
					}

					if _, ok := cache.Devices().ForName(d.Name); ok {
						common.LoggingClient.Debug(fmt.Sprintf("Candidate discovered device %s already existed", d.Name))
						break
					}

					common.LoggingClient.Info(fmt.Sprintf("Updating discovered device %s to Edgex", d.Name))
					millis := time.Now().UnixNano() / int64(time.Millisecond)
					device := &contract.Device{
						Name:           d.Name,
						Profile:        pw.Profile,
						Protocols:      d.Protocols,
						Labels:         d.Labels,
						Service:        pw.Service,
						AdminState:     pw.AdminState,
						OperatingState: contract.Enabled,
						AutoEvents:     nil,
					}
					device.Origin = millis
					device.Description = d.Description
					_, err := common.DeviceClient.Add(ctx, device)
					if err != nil {
						common.LoggingClient.Error(fmt.Sprintf("Created discovered device %s failed: %v", device.Name, err))
					}
				}
			}
			common.LoggingClient.Debug("Filtered device addition finished")
		}
	}
}

func whitelistPass(d dsModels.DiscoveredDevice, pw contract.ProvisionWatcher) bool {
	// a candidate device should pass all identifiers
	for name, regex := range pw.Identifiers {
		// ignore the device protocol properties name
		for _, protocol := range d.Protocols {
			if value, ok := protocol[name]; ok {
				matched, err := regexp.MatchString(regex, value)
				if !matched || err != nil {
					common.LoggingClient.Debug(fmt.Sprintf("Device %s's %s value %s did not match PW identifier: %s", d.Name, name, value, regex))
					return false
				}
			} else {
				common.LoggingClient.Debug(fmt.Sprintf("Identifier field: %s, did not exist in discovered device %s", name, d.Name))
				return false
			}
		}
	}
	return true
}

func blacklistPass(d dsModels.DiscoveredDevice, pw contract.ProvisionWatcher) bool {
	// a candidate should match none of the blocking identifiers
	for name, blacklist := range pw.BlockingIdentifiers {
		// ignore the device protocol properties name
		for _, protocol := range d.Protocols {
			if value, ok := protocol[name]; ok {
				for _, v := range blacklist {
					if value == v {
						common.LoggingClient.Debug(fmt.Sprintf("Discovered Device %s's %s should not be %s", d.Name, name, value))
						return false
					}
				}
			}
		}
	}
	return true
}
