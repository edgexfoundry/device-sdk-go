// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package provides a simple example implementation of
// a ProtocolHandler interface.
//
package simple

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
)

type SimpleDriver struct {
	lc logger.LoggingClient
}

func (s *SimpleDriver) DisconnectDevice(address *models.Addressable) error {
	return nil
}

func (s *SimpleDriver) Discover() (devices *interface{}, err error) {
	return nil, nil
}

// TODO: pass a logger to ProtocolDriver!
func (s *SimpleDriver) Initialize(lc logger.LoggingClient) (<-chan struct{}, error) {
	s.lc = lc
	s.lc.Debug(fmt.Sprintf("SimpleHandler.Initialize called!"))
	return nil, nil
}

func (s *SimpleDriver) ProcessAsync(operation *models.ResourceOperation,
	device *models.Device,
	object *models.DeviceObject,
	value string,
	send chan<- string) error {
	return nil
}

func (s *SimpleDriver) ProcessCommand(operation string,
	device *models.Device,
	object *models.DeviceObject,
	value string) (result string, err error) {
	return "", nil
}
