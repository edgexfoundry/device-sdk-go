//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/telemetry"
)

// Telemetry contains references to dependencies required by the Telemetry bootstrap implementation.
type Telemetry struct {
}

// New Telemetry create a new instance of Telemetry
func NewTelemetry() *Telemetry {
	return &Telemetry{}
}

// BootstrapHandler starts the telemetry collection
func (_ *Telemetry) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	_ startup.Timer,
	dic *di.Container) bool {

	logger := container.LoggingClientFrom(dic.Get)

	wg.Add(1)
	go telemetry.StartCpuUsageAverage(wg, ctx, logger)

	return true
}
