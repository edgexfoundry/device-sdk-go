// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package service

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.

const commandPath = "/{deviceId}/{cmd}"
const commandAllPath = "/all/{cmd}"
const deviceV1 = "/api/v1/device"

// common for all REST APIs?
type handlerFunc func(s *Service, w http.ResponseWriter, r *http.Request)

type commandHandler struct {
	fn handlerFunc
	s  *Service
}

func commandFunc(s *Service, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceId := vars["deviceId"]
	cmd := vars["cmd"]

	s.lc.Debug(fmt.Sprintf("commandFunc: deviceId: %s cmd: %s", deviceId, cmd))

	exists, locked := s.cd.IsDeviceLocked(deviceId)
	if exists == false {
		msg := fmt.Sprintf("device: %s not found; %s %s", deviceId, r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	if locked {
		msg := fmt.Sprintf("device: %s locked; %s %s", deviceId, r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusLocked)
		return
	}

	// pseudo-logic
	// if devices.deviceBy(deviceId).locked --> return http.StatusLocked; cache access needs to be sync'd
	// TODO: add check for device-not-found; Java code doesn't check this
	// TODO: need to mark device when operation in progress, so it can't be removed till completed...
	// if commandExists == false --> return http.StatusNotFound (404);
	//    (in Java, <proto>Handler implements commandExists, which delegates to the ProfileStore
	//    executeCommand
	//      (also from <proto>Handler:
	//      - creates new transaction
	//      - eventually calls <proto>Driver.process
	//      - waits on transaction to complete
	//      - formats reading(s) into an event, sends to core-data, return result

	w.WriteHeader(200)
	io.WriteString(w, "OK")
}

func commandAllFunc(s *Service, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	s.lc.Debug(fmt.Sprintf("command: device: all cmd: %s", vars["cmd"]))
	w.WriteHeader(200)
	io.WriteString(w, "OK")

	// pseudo-logic
	// loop thru all existing devices:
	// if devices.deviceBy(deviceId).locked --> return http.StatusLocked; cache access needs to be sync'd
	// TODO: add check for device-not-found; Java code doesn't check this
	// TODO: need to mark device when operation in progress, so it can't be removed till completed...
	// if commandExists == false --> return http.StatusNotFound (404);
	//    (in Java, <proto>Handler implements commandExists, which delegates to the ProfileStore
	//    executeCommand
	//      (also from <proto>Handler:
	//      - creates new transaction
	//      - eventually calls <proto>Driver.process
	//      - waits on transaction to complete
	//      - formats reading(s) into an event, sends to core-data, return result
}

func (c *commandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	c.s.lc.Debug(fmt.Sprintf("*commandHandler: ServeHTTP %s url: %v vars: %v", r.Method, r.URL, vars))
	// TODO: use for all endpoints vs. having a StatusHandler, UpdateHandler, ...
	if c.s.locked {
		msg := fmt.Sprintf("%s is locked; %s %s", c.s.Config.ServiceName, r.Method, r.URL)
		c.s.lc.Error(msg)
		http.Error(w, msg, http.StatusLocked)
		return
	}

	c.fn(c.s, w, r)
}

func initCommand(s *Service) *mux.Router {
	s.lc.Debug("initCommand called")

	sr := s.r.PathPrefix(deviceV1).Subrouter()

	ch := &commandHandler{fn: commandFunc, s: s}
	sr.Handle(deviceV1+commandPath, ch)

	ch = &commandHandler{fn: commandAllFunc, s: s}
	sr.Handle(deviceV1+commandAllPath, ch)

	return sr
}
