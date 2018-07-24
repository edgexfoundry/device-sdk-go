// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package gxds

import (
	"encoding/json"
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

	// TODO - models.Device isn't thread safe currently
	d := dc.DeviceById(id)
	if d == nil {
		// TODO: standardize error message format (use of prefix)
		msg := fmt.Sprintf("dev: %s not found; %s %s", id, r.Method, r.URL)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusNotFound) // status=404
		return
	}

	if d.AdminState == "LOCKED" {
		msg := fmt.Sprintf("%s is locked; %s %s", id, r.Method, r.URL)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusLocked) // status=423
		return
	}

	// TODO: need to mark device when operation in progress, so it can't be removed till completed

	// NOTE: as currently implemented, CommandExists checks the existence of a deviceprofile
	// *resource* name, not a *command* name! A deviceprofile's command section is only used
	// to trigger valuedescriptor creation.
	exists, err := pc.CommandExists(d.Name, cmd)

	// TODO: once cache locking has been implemented, this should never happen
	if err != nil {
		msg := fmt.Sprintf("internal error; dev: %s not found in cache; %s %s", id, r.Method, r.URL)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError) // status=500
		return
	}

	if !exists {
		msg := fmt.Sprintf("%s for dev: %s not found; %s %s", cmd, id, r.Method, r.URL)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusNotFound) // status=404
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("commandFunc: error reading request body for: %s %s", r.Method, r.URL)
		svc.lc.Error(msg)
	}

	if len(body) == 0 && r.Method == http.MethodPut {
		msg := fmt.Sprintf("no request body provided; %s %s", r.Method, r.URL)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusBadRequest) // status=400
		return
	}

	executeCommand(s, w, d, cmd, r.Method, string(body))
}

func commandAllFunc(s *Service, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	svc.lc.Debug(fmt.Sprintf("cmd: dev: all cmd: %s", vars["command"]))
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
	readings := make([]models.Reading, 0, s.c.Device.MaxCmdOps)

	// TODO: add support for PUT/SET commands
	var value = ""

	// make ResourceOperations
	ops, err := pc.GetResourceOperations(d.Name, cmd, method)
	if err != nil {
		svc.lc.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound) // status=404
		return
	}

	if len(ops) > s.c.Device.MaxCmdOps {
		msg := fmt.Sprintf("MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: %s",
			s.c.Device.MaxCmdOps, d.Name, cmd, method)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError) // status=500
		return
	}

	rChan := make(chan *CommandResult)
	devObjs := pc.getDeviceObjects(d.Name)
	if devObjs == nil {
		msg := fmt.Sprintf("internal error; no devObjs for dev: %s; %s %s", d.Name, cmd, method)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError) // status=500
		return
	}

	for _, op := range ops {
		objName := op.Object
		svc.lc.Debug(fmt.Sprintf("deviceObject: %s", objName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		devObj, ok := devObjs[objName]

		svc.lc.Debug(fmt.Sprintf("deviceObject: %v", devObj))
		if !ok {
			msg := fmt.Sprintf("no devobject: %s for dev: %s cmd: %s method: %s", objName, d.Name, cmd, method)
			http.Error(w, msg, http.StatusInternalServerError) // status=500
			return
		}

		// TODO: lookup valuedescriptor & pass to HandleOperation

		go s.proto.HandleOperation(&op, d, &devObj, nil, value, rChan)
		count++
	}

	var transformsOK bool = true

	// wait for fixed number of results
	for count != 0 {
		svc.lc.Debug(fmt.Sprintf("command: waiting for protcol response; count: %d", count))
		cr := <-rChan

		// get the device resource associated with the rsp.RO
		do := pc.getDeviceObject(d, cr.RO)

		ok := cr.TransformResult(do.Properties.Value)
		if !ok {
			transformsOK = false
		}

		// TODO: handle Mappings (part of RO)

		// TODO: the Java SDK supports a RO secondary device resource(object).
		// If defined, then a RO result will generate a reading for the
		// secondary object. As this use case isn't defined and/or used in
		// any of the existing Java device services, this concept hasn't
		// been implemened in gxds. TBD at the devices f2f whether this
		// be killed completely.

		reading := cr.Reading(d.Name, do.Name)
		readings = append(readings, *reading)

		svc.lc.Debug(fmt.Sprintf("dev: %s RO: %v reading: %v", d.Name, cr.RO, reading))

		count--
	}

	close(rChan)

	// push to Core Data
	event := &models.Event{Device: d.Name, Readings: readings}
	_, err = s.ec.Add(event)
	if err != nil {
		msg := fmt.Sprintf("internal error; failed to push event for dev: %s cmd: %s to CoreData: %s", d.Name, cmd, err)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError) // status=500
		return
	}

	// TODO: the 'all' form of the endpoint returns 200 if a transform
	// overflow or assertion trips...
	if !transformsOK {
		msg := fmt.Sprintf("Transform failed for dev: %s cmd: %s method: %s", d.Name, cmd, method)
		http.Error(w, msg, http.StatusInternalServerError) // status=500
	}

	// TODO: enforce config.MaxCmdValueLen; need to include overhead for
	// the rest of the Reading JSON + Event JSON length?  Should there be
	// a separate JSON body max limit for retvals & command parameters?

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (c *commandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	svc.lc.Debug(fmt.Sprintf("*commandHandler: ServeHTTP %s url: %v vars: %v", r.Method, r.URL, vars))
	// TODO: use for all endpoints vs. having a StatusHandler, UpdateHandler, ...
	if c.s.locked {
		msg := fmt.Sprintf("%s is locked; %s %s", c.s.Name, r.Method, r.URL)
		svc.lc.Error(msg)
		http.Error(w, msg, http.StatusLocked) // status=423
		return
	}

	c.fn(c.s, w, r)
}

func initCommand(s *Service) {
	svc.lc.Debug("initCommand called")

	sr := s.r.PathPrefix("/device").Subrouter()
	ch := &commandHandler{fn: commandFunc, s: s}
	sr.Handle("/{id}/{command}", ch).Methods(http.MethodGet, http.MethodPut)

	ch = &commandHandler{fn: commandAllFunc, s: s}
	sr.Handle("/all/{command}", ch).Methods(http.MethodGet, http.MethodPut)
}
