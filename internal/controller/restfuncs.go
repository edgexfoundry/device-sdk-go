// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/edgexfoundry/device-sdk-go/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	"github.com/edgexfoundry/device-sdk-go/internal/handler/callback"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

const (
	statusOK             string = "OK"
	statusNotImplemented string = "Discovery not implemented"
	statusUnavailable    string = "Discovery disabled by configuration"
	statusLocked         string = "OperatingState disabled"
)

type ConfigRespMap struct {
	Configuration map[string]interface{}
}

func (c *RestController) statusFunc(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(clients.ContentType, clients.ContentTypeText)
	w.Write([]byte("pong"))
}

func (c *RestController) versionFunc(w http.ResponseWriter, _ *http.Request) {
	res := struct {
		Version string `json:"version"`
	}{handler.VersionHandler()}
	w.Header().Add(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	c.encode(res, w)
}

func (c *RestController) discoveryFunc(w http.ResponseWriter, req *http.Request) {
	ds := container.DeviceServiceFrom(c.dic.Get)
	if c.checkServiceLocked(w, req, ds.AdminState) {
		return
	}

	if ds.OperatingState == contract.Disabled {
		http.Error(w, statusLocked, http.StatusLocked) // status=423
		return
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	if !configuration.Device.Discovery.Enabled {
		http.Error(w, statusUnavailable, http.StatusServiceUnavailable) // status=503
		return
	}

	discovery := container.ProtocolDiscoveryFrom(c.dic.Get)
	if discovery == nil {
		http.Error(w, statusNotImplemented, http.StatusNotImplemented) // status=501
		return
	}

	go autodiscovery.DiscoveryWrapper(discovery, c.LoggingClient)
	w.WriteHeader(http.StatusAccepted) //status=202
}

func (c *RestController) transformFunc(w http.ResponseWriter, req *http.Request) {
	if c.checkServiceLocked(w, req, container.DeviceServiceFrom(c.dic.Get).AdminState) {
		return
	}

	vars := mux.Vars(req)
	_, appErr := handler.TransformHandler(vars, c.LoggingClient)
	if appErr != nil {
		w.WriteHeader(appErr.Code())
	} else {
		w.Write([]byte(statusOK))
	}
}

func (c *RestController) callbackFunc(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	dec := json.NewDecoder(req.Body)
	cbAlert := contract.CallbackAlert{}

	err := dec.Decode(&cbAlert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		c.LoggingClient.Error(fmt.Sprintf("Invalid callback request: %v", err))
		return
	}

	appErr := callback.CallbackHandler(cbAlert, req.Method, c.dic)
	if appErr != nil {
		http.Error(w, appErr.Message(), appErr.Code())
	} else {
		w.Write([]byte(statusOK))
	}
}

func (c *RestController) commandFunc(w http.ResponseWriter, req *http.Request) {
	if c.checkServiceLocked(w, req, container.DeviceServiceFrom(c.dic.Get).AdminState) {
		return
	}
	vars := mux.Vars(req)

	body, ok := c.readBodyAsString(w, req)
	if !ok {
		return
	}

	event, appErr := handler.CommandHandler(vars, body, req.Method, req.URL.RawQuery, c.dic)

	if appErr != nil {
		http.Error(w, fmt.Sprintf("%s %s", appErr.Message(), req.URL.Path), appErr.Code())
	} else if event != nil {
		ec := container.CoredataEventClientFrom(c.dic.Get)
		if event.HasBinaryValue() {
			// TODO: Add conditional toggle in case caller of command does not require this response.
			// Encode response as application/CBOR.
			if len(event.EncodedEvent) <= 0 {
				var err error
				event.EncodedEvent, err = ec.MarshalEvent(event.Event)
				if err != nil {
					c.LoggingClient.Error("DeviceCommand: Error encoding event", "device", event.Device, "error", err)
				} else {
					c.LoggingClient.Trace("DeviceCommand: EventClient.MarshalEvent encoded event", "device", event.Device, "event", event)
				}
			} else {
				c.LoggingClient.Trace("DeviceCommand: EventClient.MarshalEvent passed through encoded event", "device", event.Device, "event", event)
			}
			// TODO: Resolve why this header is not included in response from Core-Command to originating caller (while the written body is).
			w.Header().Set(clients.ContentType, clients.ContentTypeCBOR)
			w.Write(event.EncodedEvent)
		} else {
			w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
			json.NewEncoder(w).Encode(event)
		}
		// push to Core Data
		go common.SendEvent(event, c.LoggingClient, container.CoredataEventClientFrom(c.dic.Get))
	}
}

func (c *RestController) commandAllFunc(w http.ResponseWriter, req *http.Request) {
	if c.checkServiceLocked(w, req, container.DeviceServiceFrom(c.dic.Get).AdminState) {
		return
	}

	vars := mux.Vars(req)
	c.LoggingClient.Debug(fmt.Sprintf("execute the Get command %s from all operational devices", vars[common.CommandVar]))

	body, ok := c.readBodyAsString(w, req)
	if !ok {
		return
	}

	events, appErr := handler.CommandAllHandler(vars[common.CommandVar], body, req.Method, req.URL.RawQuery, c.dic)
	if appErr != nil {
		http.Error(w, appErr.Message(), appErr.Code())
	} else if len(events) > 0 {
		// push to Core Data
		for _, event := range events {
			if event != nil {
				go common.SendEvent(event, c.LoggingClient, container.CoredataEventClientFrom(c.dic.Get))
			}
		}
		w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
		json.NewEncoder(w).Encode(events)
	}
}

func (c *RestController) checkServiceLocked(w http.ResponseWriter, req *http.Request, locked contract.AdminState) bool {
	if locked == contract.Locked {
		msg := fmt.Sprintf("Service is locked; %s %s", req.Method, req.URL)
		c.LoggingClient.Error(msg)
		http.Error(w, msg, http.StatusLocked) // status=423
		return true
	}
	return false
}

func (c *RestController) readBodyAsString(w http.ResponseWriter, req *http.Request) (string, bool) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		msg := fmt.Sprintf("error reading request body for: %s %s", req.Method, req.URL)
		c.LoggingClient.Error(msg)
		return "", false
	}

	if len(body) == 0 && req.Method == http.MethodPut {
		msg := fmt.Sprintf("no request body provided; %s %s", req.Method, req.URL)
		c.LoggingClient.Error(msg)
		http.Error(w, msg, http.StatusBadRequest) // status=400
		return "", false
	}

	return string(body), true
}

func (c *RestController) metricsFunc(w http.ResponseWriter, _ *http.Request) {
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

	c.encode(t, w)

	return
}

func (c *RestController) configFunc(w http.ResponseWriter, _ *http.Request) {
	configuration := container.ConfigurationFrom(c.dic.Get)
	c.encode(configuration, w)
}

// Helper function for encoding the response when servicing a REST call.
func (c *RestController) encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add(clients.ContentType, clients.ContentTypeJSON)
	enc := json.NewEncoder(w)
	err := enc.Encode(i)

	if err != nil {
		c.LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
