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
	id := vars["id"]
	cmd := vars["command"]

	s.lc.Debug(fmt.Sprintf("commandFunc: dev: %s cmd: %s", id, cmd))

	// TODO - models.Device isn't thread safe currently
	d := s.cd.DeviceById(id)
	if d == nil {
		// TODO: standardize error message format (use of prefix)
		msg := fmt.Sprintf("dev: %s not found; %s %s", id, r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	if d.AdminState == "LOCKED" {
		msg := fmt.Sprintf("dev: %s locked; %s %s", id, r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusLocked)
		return
	}

	// TODO: need to mark device when operation in progress, so it can't be removed till completed

	// NOTE: as currently implemented, CommandExists checks the existence of a deviceprofile
	// *resource* name, not a *command* name! A deviceprofile's command section is only used
	// to trigger valuedescriptor creation.
	exists, err := s.cp.CommandExists(d.Name, cmd)

	// TODO: once cache locking has been implemented, this should never happen
	if err != nil {
		msg := fmt.Sprintf("command: internal error; dev: %s not found in cache; %s %s", id, r.Method, r.URL)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusExpectationFailed)
		return
	}

	if !exists {
		msg := fmt.Sprintf("command: %s for dev: %s not found; %s %s", cmd, id, r.Method, r.URL)
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

	executeCommand(s, w, d, cmd, r.Method, string(body))
}

func commandAllFunc(s *Service, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	s.lc.Debug(fmt.Sprintf("cmd: dev: all cmd: %s", vars["command"]))
	w.WriteHeader(200)
	io.WriteString(w, "OK")

	// pseudo-logic
	// loop thru all existing devices:
	// if devices.deviceBy(id).locked --> return http.StatusLocked; cache access needs to be sync'd
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

func executeCommand(s *Service, w http.ResponseWriter, d *models.Device, cmd string, method string, args string) {
	var count int
	var responses = make([]string, 16, 64)

	// TODO: add support for PUT/SET commands
	var value = ""

	// make ResourceOperations
	ops, err := s.cp.GetResourceOperations(d.Name, cmd, method)
	if err != nil {
		s.lc.Error(err.Error())

		// TODO: review as this doesn't match the RAML
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// TODO: this should be documented in "Device Profile Guide"; might be too generous?
	if len(ops) > 64 {
		msg := fmt.Sprintf("resourceoperation limit (64) execeeded for dev: %s cmd: %s method: %s", d.Name, cmd, method)
		s.lc.Error(msg)

		// TODO: review as this doesn't match the RAML
		http.Error(w, msg, http.StatusExpectationFailed)
		return
	}

	rChan := make(chan string)
	devObjs := s.cp.GetDeviceObjects(d.Name)
	if devObjs == nil {
		msg := fmt.Sprintf("command: internal error; no devObjs for dev: %s; %s %s", d.Name, cmd, method)
		s.lc.Error(msg)
		http.Error(w, msg, http.StatusExpectationFailed)
		return
	}

	for _, op := range ops {

		// TODO: figure out how to get real deviceObject from profiles...
		objName := op.Object
		s.lc.Debug(fmt.Sprintf("deviceObject: %s", objName))

		devObj, ok := devObjs[objName]
		if !ok {
			msg := fmt.Sprintf("no devobject: %s for dev: %s cmd: %s method: %s", objName, d.Name, cmd, method)
			// TODO: review as this doesn't match the RAML
			http.Error(w, msg, http.StatusExpectationFailed)
			return
		}

		go s.proto.ProcessAsync(&op, d, &devObj, value, rChan)
		count++
	}

	// wait for responses
	for count != 0 {
		rsp := <-rChan
		responses = append(responses, rsp)
		count--
	}

	// TODO:
	s.lc.Debug(fmt.Sprintf("protocoldriver results for dev: %s cmd: %s method: %s", d.Name, cmd, method))

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
	sr.Handle("/{id}/{command}", ch).Methods(http.MethodGet, http.MethodPut)

	// TODO: RAML specifies GET, PUT, and POST, with no apparent difference between
	// PUT and POST! This code limits to just GET/PUT. Discuss and update in device
	// services requirements document.

	ch = &commandHandler{fn: commandAllFunc, s: s}
	sr.Handle("/all/{command}", ch).Methods(http.MethodGet, http.MethodPut)
}
