// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

var (
	initOnce sync.Once
)

// Init basic state for cache
func InitCache(
	serviceName string,
	lc logger.LoggingClient,
	vdc coredata.ValueDescriptorClient,
	dc metadata.DeviceClient,
	pwc metadata.ProvisionWatcherClient) {
	initOnce.Do(func() {
		ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())

		vds, err := vdc.ValueDescriptors(ctx)
		if err != nil {
			lc.Error(fmt.Sprintf("Value Descriptor cache initialization failed: %v", err))
			vds = make([]contract.ValueDescriptor, 0)
		}
		newValueDescriptorCache(vds)

		ds, err := dc.DevicesForServiceByName(ctx, serviceName)
		if err != nil {
			lc.Error(fmt.Sprintf("Device cache initialization failed: %v", err))
			ds = make([]contract.Device, 0)
		}
		newDeviceCache(ds)

		pws, err := pwc.ProvisionWatchersForServiceByName(ctx, serviceName)
		if err != nil {
			lc.Error(fmt.Sprintf("Provision Watcher cache initialization failed %v", err))
			pws = make([]contract.ProvisionWatcher, 0)
		}
		newProvisionWatcherCache(pws)

		dps := make([]contract.DeviceProfile, len(ds))
		for i, d := range ds {
			dps[i] = d.Profile
		}
		newProfileCache(dps)
	})
}
