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
	"log"
	"net/http"

	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/common"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/runtime"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Trigger implements ITrigger to support Triggers
type Trigger struct {
	Configuration common.ConfigurationStruct
	Runtime       runtime.GolangRuntime
	outputData    string
	logging       logger.LoggingClient
}

// Initialize ...
func (h *Trigger) Initialize(logger logger.LoggingClient) error {
	h.logging = logger
	http.HandleFunc("/", h.requestHandler)   // set router - just a GET for now
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	return nil
}
func (h *Trigger) requestHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	// event := event.Event{Data: "DATA FROM HTTP"}
	edgexContext := excontext.Context{Configuration: h.Configuration,
		Trigger:       h,
		LoggingClient: h.logging,
	}
	var event models.Event
	decoder.Decode(&event)

	h.Runtime.ProcessEvent(edgexContext, event)
	// bytes, _ := getBytes(h.outputData)
	w.Write(([]byte)(h.outputData))

	h.outputData = ""
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

// GetConfiguration gets the config
func (h *Trigger) GetConfiguration() common.ConfigurationStruct {
	//
	return h.Configuration
}

// GetData This function might return data
func (h *Trigger) GetData() interface{} {
	return "data"
}

// Complete ...
func (h *Trigger) Complete(outputData string) {
	//
	h.outputData = outputData

}
