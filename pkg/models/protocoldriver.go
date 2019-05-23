// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package defines interfaces and structs used to build an EdgeX Foundry Device
// Service.  The interfaces provide an asbstraction layer for the device
// or protocol specific logic of a Device Service, and the structs represents request
// and response data format used by the protocol driver.
//
package models

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// ProtocolDriver is a low-level device-specific interface used by
// by other components of an EdgeX Device Service to interact with
// a specific class of devices.
type ProtocolDriver interface {
	// Initialize performs protocol-specific initialization for the device
	// service. The given *AsyncValues channel can be used to push asynchronous
	// events and readings to Core Data.
	Initialize(lc logger.LoggingClient, asyncCh chan<- *AsyncValues) error

	// HandleReadCommands passes a slice of CommandRequest struct each representing
	// a ResourceOperation for a specific device resource.
	HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []CommandRequest) ([]*CommandValue, error)

	// HandleWriteCommands passes a slice of CommandRequest struct each representing
	// a ResourceOperation for a specific device resource.
	// Since the commands are actuation commands, params provide parameters for the individual
	// command.
	HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []CommandRequest, params []*CommandValue) error

	// Stop instructs the protocol-specific DS code to shutdown gracefully, or
	// if the force parameter is 'true', immediately. The driver is responsible
	// for closing any in-use channels, including the channel used to send async
	// readings (if supported).
	Stop(force bool) error
}
