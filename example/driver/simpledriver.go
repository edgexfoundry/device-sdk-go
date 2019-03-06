// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a simple example implementation of
// a ProtocolDriver interface.
//
package driver

import (
	"fmt"
	"time"

	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
)

type SimpleDriver struct {
	lc           logger.LoggingClient
	asyncCh      chan<- *ds_models.AsyncValues
	switchButton bool
}

// DisconnectDevice handles protocol-specific cleanup when a device
// is removed.
func (s *SimpleDriver) DisconnectDevice(deviceName string, protocols map[string]map[string]string) error {
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *SimpleDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *ds_models.AsyncValues) error {
	s.lc = lc
	s.asyncCh = asyncCh
	return nil
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *SimpleDriver) HandleReadCommands(deviceName string, protocols map[string]map[string]string, reqs []ds_models.CommandRequest) (res []*ds_models.CommandValue, err error) {

	if len(reqs) != 1 {
		err = fmt.Errorf("SimpleDriver.HandleReadCommands; too many command requests; only one supported")
		return
	}

	s.lc.Debug(fmt.Sprintf("SimpleDriver.HandleReadCommands: protocols: %v operation: %v attributes: %v", protocols, reqs[0].RO.Operation, reqs[0].DeviceResource.Attributes))

	res = make([]*ds_models.CommandValue, 1)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	cv, _ := ds_models.NewBoolValue(&reqs[0].RO, now, s.switchButton)
	res[0] = cv

	return
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *SimpleDriver) HandleWriteCommands(deviceName string, protocols map[string]map[string]string, reqs []ds_models.CommandRequest,
	params []*ds_models.CommandValue) error {

	if len(reqs) != 1 {
		err := fmt.Errorf("SimpleDriver.HandleWriteCommands; too many command requests; only one supported")
		return err
	}
	if len(params) != 1 {
		err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the number of parameter is not correct; only one supported")
		return err
	}

	s.lc.Debug(fmt.Sprintf("SimpleDriver.HandleWriteCommands: protocols: %v, operation: %v, parameters: %v", protocols, reqs[0].RO.Operation, params))
	var err error
	if s.switchButton, err = params[0].BoolValue(); err != nil {
		err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the data type of parameter should be Boolean, parameter: %s", params[0].String())
		return err
	}

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *SimpleDriver) Stop(force bool) error {
	s.lc.Debug(fmt.Sprintf("SimpleDriver.Stop called: force=%v", force))
	return nil
}
