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
	"os"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

type SimpleHandler struct{}

func (s *SimpleHandler) Initialize() {
	fmt.Println(os.Stdout, "SimpleHandler.Initialize called!")
}

func (s *SimpleHandler) DisconnectDevice(device models.Device) {
}

func (s *SimpleHandler) InitializeDevice(device models.Device) {
}

func (s *SimpleHandler) Scan() {
}

func (s *SimpleHandler) CommandExists(device models.Device, command string) bool {
	return false
}

func (s *SimpleHandler) ExecuteCommand(device models.Device, command string, args string) map[string]string {
	return nil
}

func (s *SimpleHandler) CompleteTransaction(transactionId string, opId string, readings []models.Reading) {
}
