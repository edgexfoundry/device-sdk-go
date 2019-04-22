// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"sync"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type Watchers struct {
	devices map[string]contract.Device
}

var (
	wcOnce   sync.Once
	watchers *Watchers
)

// Create a singleton WatcherCache instance
func newWatchers() *Watchers {

	wcOnce.Do(func() {
		watchers = &Watchers{}
	})

	return watchers
}
