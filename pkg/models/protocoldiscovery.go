// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

// ProtocolDiscovery is a low-level device-specific interface implemented
// by device services that support dynamic device discovery.
type ProtocolDiscovery interface {
	// Discover triggers protocol specific device discovery, which is
	// a synchronous operation which returns a list of new devices
	// which may be added to the device service based on service
	// config. This function may also optionally trigger sensor
	// discovery, which could result in dynamic device profile creation.
	//
	// TODO: add models.ScanList (or define locally) for devices
	Discover() (devices *interface{}, err error)
}
