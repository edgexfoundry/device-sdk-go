// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

var (
	initOnce sync.Once
)

// Init basic state for cache
func InitV2Cache(
	serviceName string,
	lc logger.LoggingClient,
	vdc coredata.ValueDescriptorClient,
	dc metadata.DeviceClient,
	pwc metadata.ProvisionWatcherClient) {
	initOnce.Do(func() {
		// TODO: uncomment when v2 core-contracts is ready.
		//ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
		//ds, err := dc.DevicesForServiceByName(ctx, serviceName)
		//if err != nil {
		//	lc.Error(fmt.Sprintf("Device cache initialization failed: %v", err))
		//	ds = make([]contract.Device, 0)
		//}

		//dps := make([]contract.DeviceProfile, len(ds))
		//for i, d := range ds {
		//	dps[i] = d.Profile
		//}

		var ds []contract.Device
		newDeviceCache(ds)

		var dps []contract.DeviceProfile
		newProfileCache(dps)
	})
}
