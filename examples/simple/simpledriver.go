// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a simple example implementation of
// a ProtocolDriver interface.
package simple

import (
	"fmt"

	device "github.com/edgexfoundry/device-sdk-go"
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
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
func (s *SimpleDriver) Initialize(svc *device.Service, lc logger.LoggingClient, asyncCh <-chan *device.CommandResult) error {
	s.lc = lc
	s.lc.Debug(fmt.Sprintf("SimpleHandler.Initialize called!"))
	return nil
}

// HandleCommand triggers an asynchronous protocol specific GET or SET operation
// for the specified device.
func (s *SimpleDriver) HandleCommands(d models.Device, reqs []device.CommandRequest,
	params string) (res []device.CommandResult, err error) {

	if len(reqs) != 1 {
		err = fmt.Errorf("SimpleDriver.HandleCommands; too many command requests; only one supported")
		return
	}

	s.lc.Debug(fmt.Sprintf("HandleCommand: dev: %s op: %v attrs: %v", d.Name, reqs[0].RO.Operation, reqs[0].DeviceObject.Attributes))

	res = make([]device.CommandResult, 1)

	// TODO: change CommandResult to get rid of pointer to RO
	res[0].RO = &reqs[0].RO
	res[0].Type = device.Bool
	res[0].BoolResult = true

	return
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *SimpleDriver) Stop(force bool) error {
	s.lc.Debug(fmt.Sprintf("Stop called: force=%v", force))
	return nil
}
