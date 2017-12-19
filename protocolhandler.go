// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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
package gxds

import "github.com/edgexfoundry/core-domain-go/models"

// A ProtocolHandler implements low-level protocol functions
type ProtocolHandler interface {
	Initialize()
	DisconnectDevice(device models.Device)
	InitializeDevice(device models.Device)
	Scan()
	CommandExists(device models.Device, command string) map[string]string
	ExecuteCommand(device models.Device, command string, args string)
	SendTransaction(deviceName string, readings []models.Reading)
	CompleteTransaction(transactionId string, opId string, readings []models.Reading)
}
