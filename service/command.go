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
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/gorilla/mux"
)

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.

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

	// TODO - models.Device isn't thread safe currently
	device := s.cd.DeviceById(deviceId)
	if device == nil {
		msg := fmt.Sprintf("device: %s not found; %s %s", deviceId, r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	if device.AdminState == "LOCKED" {
		msg := fmt.Sprintf("device: %s locked; %s %s", deviceId, r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusLocked)
		return
	}

	// TODO: need to mark device when operation in progress, so it can't be removed till completed...

	// TODO: implement CommandExists; current failure point as it alwys returns exists=false
	exists, err := s.cp.CommandExists(device.Name, cmd)

	// TODO: ASSERT if err != nil

	if !exists {
		msg := fmt.Sprintf("command: %s for device: %s not found; %s %s", cmd, deviceId, r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("error reading request body for: %s %s", r.Method, r.URL)
		s.lc.Error(msg)
	}

	// TODO: RAML doesn't mention StatusBadRequest (400)
	if len(body) == 0 && r.Method == "PUT" {
		msg := fmt.Sprintf("no command args provided; %s %s", r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	executeCommand(s, w, device, cmd, string(body))
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

func executeCommand(s *Service, w http.ResponseWriter, device *models.Device, cmd string, args string) {
	w.WriteHeader(200)
	io.WriteString(w, "OK")
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

func initCommand(s *Service) {
	s.lc.Debug("initCommand called")

	sr := s.r.PathPrefix("/device").Subrouter()
	ch := &commandHandler{fn: commandFunc, s: s}
	sr.Handle("/{deviceId}/{cmd}", ch).Methods(http.MethodGet, http.MethodPut)

	// TODO: RAML specifies GET, PUT, and POST, with no apparent difference between
	// PUT and POST! This code limits to just GET/PUT. Discuss and update in device
	// services requirements document.

	ch = &commandHandler{fn: commandAllFunc, s: s}
	sr.Handle("/all/{cmd}", ch).Methods(http.MethodGet, http.MethodPut)
}
