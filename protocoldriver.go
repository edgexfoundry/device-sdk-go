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
// TODO:
// * Determine if gxds should define separate 'Handler' and 'Driver'
//   protocol interfaces?  Can they be combined?
//
// * Investigate changing calling signatures to leverage std Go
//   interfaces, such as Reader/Writer, ...
//
package gxds

import (
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
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

	// Discover triggers protocol specific device discovery, which is
	// a synchronous operation which returns a list of new devices
	// which may be added to the device service based on service
	// configuration. This function may also optionally trigger sensor
	// discovery, which could result in dynamic device profile creation.
	//
	// TODO:
	//   * add models.ScanList (or define locally) for devices
	//   * add support for asynchronous discovery operation (i.e.
	//     support new devices added directly from driver code and/or
	//     new device(s) notification to DiscoveryLogic, which then adds
	//     the new devices).
	Discover() (devices *interface{}, err error)

	// Initialize performs protocol-specific initialization for the device
	// service.  If the DS supports asynchronous data pushed from devices/sensors,
	// then a valid receive' channel must be created and returned, otherwise nil
	// is returned.
	Initialize(lc logger.LoggingClient) (out <-chan struct{}, err error)

	// Process triggers an asynchronous protocol specific GET or SET operation
	// for the specified device. Device profile attributes are passed as part
	// of the *models.DeviceObject. The parameter 'value' must be provided for
	// a SET operation, otherwise it should be 'nil'.
	//
	// This function is always called in a new goroutine. The driver is responsible
	// for writing the command result to the send channel.
	//
	// TODO: once channel return values are working, re-factor so that a the value
	// written back is a struct with an error and string.
	//
	// NOTE - the Java-based device-virtual includes an additional parameter called
	// operations which is used to optimize how virtual resources are saved for SETs.
	//
	ProcessAsync(ro *models.ResourceOperation,
		device *models.Device,
		object *models.DeviceObject,
		value string,
		send chan<- *CommandResult)

	// ProcessCommand processes a command synchronously, and returns the result
	// instead of writing it to a send channel. This function is used when a SET
	// command is used on a valuedescriptor which specifes a mask, which requires
	// a read of the current value before the mask can be applied.
	//
	// NOTE - the Java-based device-virtual includes an additional parameter called
	// getOperations which is used to optimize how virtual resources are saved for SETs.
	//
	// TODO:
	//  * with a bit of re-work, it should be possible to get rid of this function
	//    and just use ProcessAsync (which could be re-named ProcessCommand()).
	ProcessCommand(operation string,
		device *models.Device,
		object *models.DeviceObject,
		value string) (result string, err error)
}
