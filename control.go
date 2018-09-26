// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func discoveryHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(os.Stdout, "service: discovery request")
	io.WriteString(w, "OK")
}

func transformHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	fmt.Fprintf(os.Stdout, "service: transform request: transformData: %s", vars["transformData"])
	io.WriteString(w, "OK")
}

func initControl() {
	svc.r.HandleFunc("/discovery", discoveryHandler).Methods("POST")
	svc.r.HandleFunc("/debug/transformData/{transformData}", transformHandler).Methods("GET")
}
