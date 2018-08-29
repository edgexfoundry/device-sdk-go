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
	// service.  If the DS supports asynchronous data pushed from devices/sensors,
	// then a valid receive' channel will be given, otherwise the channel is nil
	// and must not be used.
	Initialize(lc logger.LoggingClient, asyncCh <-chan *CommandResult) error

	// HandleOperation triggers an asynchronous protocol specific GET or SET operation
	// for the specified device. Device profile attributes are passed as part
	// of the *models.DeviceObject. The parameter 'value' must be provided for
	// a SET operation, otherwise it should be 'nil'.
	//
	// This function is always called in a new goroutine. The driver is responsible
	// for writing the CommandResults to the send channel. The driver is also
	// responsible for closing send channel if/when Stop is called.
	//
	// NOTE - the Java-based device-virtual includes an additional parameter called
	// operations which is used to optimize how virtual resources are saved for SETs.
	//
	HandleOperation(ro *models.ResourceOperation,
		device *models.Device,
		object *models.DeviceObject,
		desc *models.ValueDescriptor,
		value string,
		send chan<- *CommandResult)

	// Stop instructs the protocol-specific DS code to shutdown gracefully, or
	// if the force parameter is 'true', immediately. The driver is responsible
	// for closing any in-use channels, including the channel used to send async
	// readings (if supported).
	Stop(force bool) error
}
