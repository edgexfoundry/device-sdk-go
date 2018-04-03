// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package defines an interface used to build an EdgeX Foundry device
// service.  This interace provides an asbstraction layer for the device
// or protocol specific logic of a device service.
//
// TODO:
// * Determine if gxds still needs a ProtocolHandler, after implementing
//   basic command handling.
//
// * Investigate changing calling signatures to leverage std Go
//   interfaces, such as Reader/Writer, ...
//
package gxds

import "github.com/edgexfoundry/edgex-go/core/domain/models"

// ProtocolHandler is a high-level device-specific interface used by
// by other components of an EdgeX device service to interact with
// a specific class of devices.
type ProtocolHandler interface {
	// Initialize performs protocol-specific initialization
	// for the device service.
	Initialize()

	// DisconnectDevice...
	DisconnectDevice(device models.Device)

	// InitializeDevice...
	InitializeDevice(device models.Device)

	// Scan
	Scan()

	// CommandExists...
	// TODO: can CommandExists be handled by the devicestore?
	CommandExists(device models.Device, command string) bool

	// ExecuteCommand...
	ExecuteCommand(device models.Device, command string, args string) map[string]string
	// CompleteTransaction...
	CompleteTransaction(transactionId string, opId string, readings []models.Reading)
}
