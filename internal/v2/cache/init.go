// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

var (
	initOnce sync.Once
)

// Init basic state for cache
func InitV2Cache() {
	initOnce.Do(func() {
		// TODO: retrieve data from metadata when v2 core-contracts is ready.
		var ds []models.Device
		newDeviceCache(ds)

		var dps []models.DeviceProfile
		newProfileCache(dps)
	})
}
