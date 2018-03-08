// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package gxds defines interfaces used to build an EdgeX Foundry device
// service.  These interaces provide and asbstraction layer for the device
// or protocol specific logic of a device service.
package gxds

import "github.com/edgexfoundry/core-domain-go/models"

// ProtocolHandler is a high-level device-specific interface used by
// by other components of an EdgeX device service to interact with
// a specific class of devices.
type ProtocolHandler interface {
	Initialize()
	DisconnectDevice(device models.Device)
	InitializeDevice(device models.Device)
	Scan()

	// TODO: can CommandExists be handled by the devicestore?
	CommandExists(device models.Device, command string) bool
	ExecuteCommand(device models.Device, command string, args string) map[string]string
	CompleteTransaction(transactionId string, opId string, readings []models.Reading)
}
