// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package defines an interface used to build an EdgeX Foundry device
// service.  This interace provides an asbstraction layer for the device
// or protocol specific logic of a device service.
//
package device

import (
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// ProtocolDriver is a low-level device-specific interface used by
// by other components of an EdgeX device service to interact with
// a specific class of devices.
type ProtocolDriver interface {

	// DisconnectDevice is when a device is removed from the device
	// service. This function allows for protocol specific disconnection
	// logic to be performed.  Device services which don't require this
	// function should just return 'nil'.
	//
	// TODO: the Java code uses this signature, with the addressable
	// appearing to be that of the device service itself. I'm not sure
	// how this gets tied by the driver code to an actual device. Maybe
	// this should be *models.Device?
	//
	DisconnectDevice(address *models.Addressable) error

	// Initialize performs protocol-specific initialization for the device
	// service. The given *CommandResult channel can be used to push asynchronous
	// events and readings to Core Data.
	Initialize(s *Service, lc logger.LoggingClient, asyncCh <-chan *CommandResult) error

	// HandleCommands passes a slice of CommandRequest structs each representing
	// a ResourceOperation for a specific device resource (aka DeviceObject).
	// If commands are actuation commands, then params may be used to provide
	// an optional JSON encoded string specifying paramters for the individual
	// commands.
	//
	// TODO: add param to CommandRequest and have command endpoint parse the params.
	HandleCommands(d models.Device, reqs []CommandRequest, params string) ([]CommandResult, error)

	// Stop instructs the protocol-specific DS code to shutdown gracefully, or
	// if the force parameter is 'true', immediately. The driver is responsible
	// for closing any in-use channels, including the channel used to send async
	// readings (if supported).
	Stop(force bool) error
}
