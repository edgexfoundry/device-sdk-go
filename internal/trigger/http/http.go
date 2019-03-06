//
// Copyright (c) 2019 Intel Corporation
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
//

package http

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/http"

	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/webserver"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Trigger implements ITrigger to support Triggers
type Trigger struct {
	Configuration common.ConfigurationStruct
	Runtime       runtime.GolangRuntime
	outputData    string
	logging       logger.LoggingClient
	Webserver     *webserver.WebServer
}

// Initialize ...
func (trigger *Trigger) Initialize(logger logger.LoggingClient) error {
	trigger.logging = logger
	trigger.logging.Info("Initializing HTTP Trigger")
	trigger.Webserver.SetupTriggerRoute(trigger.requestHandler)
	trigger.logging.Info("HTTP Trigger Initialized")

	return nil
}
func (trigger *Trigger) requestHandler(writer http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	// event := event.Event{Data: "DATA FROM HTTP"}
	edgexContext := excontext.Context{Configuration: trigger.Configuration,
		Trigger:       trigger,
		LoggingClient: trigger.logging,
	}
	var event models.Event
	decoder.Decode(&event)

	trigger.Runtime.ProcessEvent(edgexContext, event)
	writer.Write(([]byte)(trigger.outputData))

	trigger.outputData = ""
}
func getBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Complete ...
func (trigger *Trigger) Complete(outputData string) {
	//
	trigger.outputData = outputData

}
