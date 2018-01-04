// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package simple

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/core-domain-go/models"
)

type SimpleHandler struct{}

func (s *SimpleHandler) Initialize() {
	fmt.Println(os.Stdout, "SimpleHandler.Initialize called!")
}

func (s *SimpleHandler) DisconnectDevice(device models.Device) {
}

func (s *SimpleHandler)	InitializeDevice(device models.Device) {
}


func (s *SimpleHandler) Scan() {
}

func (s *SimpleHandler)	CommandExists(device models.Device, command string) bool {
	return false
}

func (s *SimpleHandler)	ExecuteCommand(device models.Device, command string, args string) map[string]string {
	return nil
}

func (s *SimpleHandler) CompleteTransaction(transactionId string, opId string, readings []models.Reading) {
}
