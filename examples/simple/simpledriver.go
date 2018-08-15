// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package provides a simple example implementation of
// a ProtocolDriver interface.
//
package simple

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	device "github.com/edgexfoundry/device-sdk-go"
)

type SimpleDriver struct {
	lc logger.LoggingClient
}

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (s *SimpleDriver) DisconnectDevice(address *models.Addressable) error {
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.  If the DS supports asynchronous data pushed from devices/sensors,
// then a valid receive' channel must be created and returned, otherwise nil
// is returned.
func (s *SimpleDriver) Initialize(lc logger.LoggingClient, asyncCh <-chan *device.CommandResult) error {
	s.lc = lc
	s.lc.Debug(fmt.Sprintf("SimpleHandler.Initialize called!"))
	return nil
}

// HandleOperation triggers an asynchronous protocol specific GET or SET operation
// for the specified device. Device profile attributes are passed as part
// of the *models.DeviceObject. The parameter 'value' must be provided for
// a SET operation, otherwise it should be 'nil'.
//
// This function is always called in a new goroutine. The driver is responsible
// for writing the CommandResults to the send channel.
//
// Note - DeviceObject represents a deviceResource defined in deviceprofile.
//
func (s *SimpleDriver) HandleOperation(ro *models.ResourceOperation,
	d *models.Device, do *models.DeviceObject, desc *models.ValueDescriptor,
	value string, send chan<- *device.CommandResult) {

	s.lc.Debug(fmt.Sprintf("HandleCommand: dev: %s op: %v attrs: %v", d.Name, ro.Operation, do.Attributes))

	cr := &device.CommandResult{RO: ro, Type: device.Bool, BoolResult: true}

	send <- cr
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *SimpleDriver) Stop(force bool) error {
	s.lc.Debug(fmt.Sprintf("Stop called: force=%v", force))
	return nil
}
