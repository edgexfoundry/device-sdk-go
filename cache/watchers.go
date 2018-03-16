// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package cache

import (
	"sync"

	"github.com/edgexfoundry/core-clients-go/metadataclients"
	"github.com/edgexfoundry/core-domain-go/models"
)

type Watchers struct {
	devices map[string]models.Device
	ac      metadataclients.AddressableClient
	dc      metadataclients.DeviceClient
}

var (
	wcOnce   sync.Once
	watchers *Watchers
)

// Create a singleton WatcherCache instance
func NewWatchers() *Watchers {

	wcOnce.Do(func() {
		watchers = &Watchers{}
	})

	return watchers
}
