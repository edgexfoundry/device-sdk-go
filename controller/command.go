// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func commandHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	fmt.Fprintf(os.Stdout, "command: device request: devid: %s cmd: %s", vars["deviceId"], vars["cmd"])
	io.WriteString(w, "OK")
}

func commandsHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	fmt.Fprintf(os.Stdout, "command: all devices request: cmd: %s", vars["cmd"])
	io.WriteString(w, "OK")
}

func initCommand(r *mux.Router) {
	s := r.PathPrefix("/api/v1/device").Subrouter()
	s.HandleFunc("/{deviceId}/{cmd}", commandHandler)
	s.HandleFunc("/all/{cmd}", commandsHandler)
}
