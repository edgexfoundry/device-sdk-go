// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	"github.com/edgexfoundry/device-sdk-go/internal/handler/callback"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

const (
	statusOK string = "OK"
)

type ConfigRespMap struct {
	Configuration map[string]interface{}
}

func statusFunc(w http.ResponseWriter, req *http.Request) {
	result := handler.StatusHandler()
	io.WriteString(w, result)
}

func discoveryFunc(w http.ResponseWriter, req *http.Request) {
	if checkServiceLocked(w, req) {
		return
	}

	vars := mux.Vars(req)
	go handler.DiscoveryHandler(vars)
	io.WriteString(w, statusOK)
	w.WriteHeader(http.StatusAccepted)
}

func transformFunc(w http.ResponseWriter, req *http.Request) {
	if checkServiceLocked(w, req) {
		return
	}

	vars := mux.Vars(req)
	_, appErr := handler.TransformHandler(vars)
	if appErr != nil {
		w.WriteHeader(appErr.Code())
	} else {
		io.WriteString(w, statusOK)
	}
}

func callbackFunc(w http.ResponseWriter, req *http.Request) {
	if checkServiceLocked(w, req) {
		return
	}

	defer req.Body.Close()
	dec := json.NewDecoder(req.Body)
	cbAlert := contract.CallbackAlert{}

	err := dec.Decode(&cbAlert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		common.LoggingClient.Error(fmt.Sprintf("Invalid callback request: %v", err))
		return
	}

	appErr := callback.CallbackHandler(cbAlert, req.Method)
	if appErr != nil {
		http.Error(w, appErr.Message(), appErr.Code())
	} else {
		io.WriteString(w, statusOK)
	}
}

func commandFunc(w http.ResponseWriter, req *http.Request) {
	if checkServiceLocked(w, req) {
		return
	}
	vars := mux.Vars(req)

	body, ok := readBodyAsString(w, req)
	if !ok {
		return
	}

	event, appErr := handler.CommandHandler(vars, body, req.Method)

	if appErr != nil {
		http.Error(w, fmt.Sprintf("%s %s", appErr.Message(), req.URL.Path), appErr.Code())
	} else if event != nil {
		if event.HasBinaryValue() {
			// TODO: Add conditional toggle in case caller of command does not require this response.
			// Encode response as application/CBOR.
			if len(event.EncodedEvent) <= 0 {
				var err error
				event.EncodedEvent, err = common.EventClient.MarshalEvent(event.Event)
				if err != nil {
					common.LoggingClient.Error("DeviceCommand: Error encoding event", "device", event.Device, "error", err)
				} else {
					common.LoggingClient.Trace("DeviceCommand: EventClient.MarshalEvent encoded event", "device", event.Device, "event", event)
				}
			} else {
				common.LoggingClient.Trace("DeviceCommand: EventClient.MarshalEvent passed through encoded event", "device", event.Device, "event", event)
			}
			// TODO: Resolve why this header is not included in response from Core-Command to originating caller (while the written body is).
			w.Header().Set(clients.ContentType, clients.ContentTypeCBOR)
			w.Write(event.EncodedEvent)
		} else {
			w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
			json.NewEncoder(w).Encode(event)
		}
		// push to Core Data
		go common.SendEvent(event)
	}
}

func commandAllFunc(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	common.LoggingClient.Debug(fmt.Sprintf("Controller - Command: execute the Get command %s from all operational devices", vars[common.CommandVar]))

	if checkServiceLocked(w, req) {
		return
	}

	body, ok := readBodyAsString(w, req)
	if !ok {
		return
	}

	events, appErr := handler.CommandAllHandler(vars[common.CommandVar], body, req.Method)
	if appErr != nil {
		http.Error(w, appErr.Message(), appErr.Code())
	} else if len(events) > 0 {
		// push to Core Data
		for _, event := range events {
			if event != nil {
				go common.SendEvent(event)
			}
		}
		w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
		json.NewEncoder(w).Encode(events)
	}
}

func checkServiceLocked(w http.ResponseWriter, req *http.Request) bool {
	if common.ServiceLocked {
		msg := fmt.Sprintf("%s is locked; %s %s", common.ServiceName, req.Method, req.URL)
		common.LoggingClient.Error(msg)
		http.Error(w, msg, http.StatusLocked) // status=423
		return true
	}
	return false
}

func readBodyAsString(w http.ResponseWriter, req *http.Request) (string, bool) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		msg := fmt.Sprintf("commandFunc: error reading request body for: %s %s", req.Method, req.URL)
		common.LoggingClient.Error(msg)
		return "", false
	}

	if len(body) == 0 && req.Method == http.MethodPut {
		msg := fmt.Sprintf("no request body provided; %s %s", req.Method, req.URL)
		common.LoggingClient.Error(msg)
		http.Error(w, msg, http.StatusBadRequest) // status=400
		return "", false
	}

	return string(body), true
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	var t common.Telemetry

	// The device service is to be considered the System Of Record (SOR) for accurate information.
	// (Here, we fetch metrics for a given device service that's been generated by device-sdk-go.)
	var rtm runtime.MemStats

	// Read full memory stats
	runtime.ReadMemStats(&rtm)

	// Miscellaneous memory stats
	t.Alloc = rtm.Alloc
	t.TotalAlloc = rtm.TotalAlloc
	t.Sys = rtm.Sys
	t.Mallocs = rtm.Mallocs
	t.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	t.LiveObjects = t.Mallocs - t.Frees

	encode(t, w)

	return
}

// TODO: Identify the appropriate configuration (for requested device service).
// TODO: 'Placeholder' (for configuration) is common.CurrentConfig
func configHandler(w http.ResponseWriter, _ *http.Request) {
	encode(common.CurrentConfig, w)
}

// Helper function for encoding the response when servicing a REST call.
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add(clients.ContentType, clients.ContentTypeJSON)
	enc := json.NewEncoder(w)
	err := enc.Encode(i)

	if err != nil {
		common.LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
