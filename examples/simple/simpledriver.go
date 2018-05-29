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

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/tonyespy/gxds"
)

type SimpleDriver struct {
	lc logger.LoggingClient
}

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (s *SimpleDriver) DisconnectDevice(address *models.Addressable) error {
	return nil
}

// Discover triggers protocol specific device discovery, which is
// a synchronous operation which returns a list of new devices
// which may be added to the device service based on service
// configuration. This function may also optionally trigger sensor
// discovery, which could result in dynamic device profile creation.
func (s *SimpleDriver) Discover() (devices *interface{}, err error) {
	return nil, nil
}

// Initialize performs protocol-specific initialization for the device
// service.  If the DS supports asynchronous data pushed from devices/sensors,
// then a valid receive' channel must be created and returned, otherwise nil
// is returned.
func (s *SimpleDriver) Initialize(lc logger.LoggingClient) (<-chan struct{}, error) {
	s.lc = lc
	s.lc.Debug(fmt.Sprintf("SimpleHandler.Initialize called!"))
	return nil, nil
}

// HandleOperation triggers an asynchronous protocol specific GET or SET operation
// for the specified device. Device profile attributes are passed as part
// of the *models.DeviceObject. The parameter 'value' must be provided for
// a SET operation, otherwise it should be 'nil'.
//
// This function is always called in a new goroutine. The driver is responsible
// for writing the command result to the send channel.
//
// Note - DeviceObject represents a deviceResource defined in deviceprofile.
//
func (s *SimpleDriver) HandleOperation(ro *models.ResourceOperation,
	d *models.Device, do *models.DeviceObject, desc *models.ValueDescriptor,
	value string, send chan<- *gxds.CommandResult) {

	s.lc.Debug(fmt.Sprintf("HandleCommand: dev: %s op: %v attrs: %v", d.Name, ro.Operation, do.Attributes))

	cr := &gxds.CommandResult{RO: ro, Type: gxds.Bool, BoolResult: true}

	send <- cr
}
