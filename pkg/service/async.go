// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/transformer"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/cache"
	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

// processAsyncResults processes readings that are pushed from
// a DS implementation. Each is reading is optionally transformed
// before being pushed to Core Data.
// In this function, AsyncBufferSize is used to create a buffer for
// processing AsyncValues concurrently, so that events may arrive
// out-of-order in core-data / app service when AsyncBufferSize value
// is greater than or equal to two. Alternatively, we can process
// AsyncValues one by one in the same order by changing the AsyncBufferSize
// value to one.
func (s *DeviceService) processAsyncResults(ctx context.Context, wg *sync.WaitGroup, dic *di.Container) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	working := make(chan bool, s.config.Service.AsyncBufferSize)
	for {
		select {
		case <-ctx.Done():
			return
		case acv := <-s.asyncCh:
			go s.sendAsyncValues(acv, working, dic)
		}
	}
}

// sendAsyncValues convert AsyncValues to event and send the event to CoreData
func (s *DeviceService) sendAsyncValues(acv *dsModels.AsyncValues, working chan bool, dic *di.Container) {
	working <- true
	defer func() {
		<-working
	}()

	event, err := transformer.CommandValuesToEventDTO(acv.CommandValues, acv.DeviceName, dic)
	if err != nil {
		s.LoggingClient.Errorf("failed to transform CommandValues to Event: %v", err)
		return
	}

	common.SendEvent(event, "", s.LoggingClient, s.edgexClients.EventClient)
}

// processAsyncFilterAndAdd filter and add devices discovered by
// device service protocol discovery.
func (s *DeviceService) processAsyncFilterAndAdd(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case devices := <-s.deviceCh:
			ctx := context.Background()
			pws := cache.ProvisionWatchers().All()
			for _, d := range devices {
				for _, pw := range pws {
					if whitelistPass(d, pw, s.LoggingClient) && blacklistPass(d, pw, s.LoggingClient) {
						if _, ok := cache.Devices().ForName(d.Name); ok {
							s.LoggingClient.Debugf("Candidate discovered device %s already existed", d.Name)
							break
						}

						s.LoggingClient.Infof("Adding discovered device %s to Edgex", d.Name)
						device := contract.Device{
							Name:           d.Name,
							Description:    d.Description,
							ProfileName:    pw.ProfileName,
							Protocols:      d.Protocols,
							Labels:         d.Labels,
							ServiceName:    pw.ProfileName,
							AdminState:     pw.AdminState,
							OperatingState: contract.Up,
							AutoEvents:     nil,
						}

						req := requests.NewAddDeviceRequest(dtos.FromDeviceModelToDTO(device))
						_, err := s.edgexClients.DeviceClient.Add(ctx, []requests.AddDeviceRequest{req})
						if err != nil {
							s.LoggingClient.Errorf("failed to create discovered device %s: %v", device.Name, err)
						} else {
							break
						}
					}
				}
			}
			s.LoggingClient.Debug("Filtered device addition finished")
		}
	}
}

func whitelistPass(d dsModels.DiscoveredDevice, pw contract.ProvisionWatcher, lc logger.LoggingClient) bool {
	// ignore the device protocol properties name
	for _, protocol := range d.Protocols {
		matchedCount := 0
		for name, regex := range pw.Identifiers {
			if value, ok := protocol[name]; ok {
				matched, err := regexp.MatchString(regex, value)
				if !matched || err != nil {
					lc.Debug(fmt.Sprintf("Device %s's %s value %s did not match PW identifier: %s", d.Name, name, value, regex))
					break
				}
				matchedCount += 1
			}
		}
		// match succeed on all identifiers
		if matchedCount == len(pw.Identifiers) {
			return true
		}
	}
	return false
}

func blacklistPass(d dsModels.DiscoveredDevice, pw contract.ProvisionWatcher, lc logger.LoggingClient) bool {
	// a candidate should match none of the blocking identifiers
	for name, blacklist := range pw.BlockingIdentifiers {
		// ignore the device protocol properties name
		for _, protocol := range d.Protocols {
			if value, ok := protocol[name]; ok {
				for _, v := range blacklist {
					if value == v {
						lc.Debug(fmt.Sprintf("Discovered Device %s's %s should not be %s", d.Name, name, value))
						return false
					}
				}
			}
		}
	}
	return true
}
