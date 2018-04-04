// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package cache

import (
	"github.com/tonyespy/gxds"

	"sync"
)

// Schedules is a local cache of schedules and scheduleevents,
// usually loaded into Core Metadata, however existing schedules
// scheduleevents can be used to seed this cache.
type Schedules struct {
	config *gxds.Config
}

var (
	scOnce    sync.Once
	schedules *Schedules
)

// Creates a singleton Schedules cache instance.
func NewSchedules(config *gxds.Config) *Schedules {

	scOnce.Do(func() {
		schedules = &Schedules{config: config}
	})

	return schedules
}
