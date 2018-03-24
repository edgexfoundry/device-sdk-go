// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package controller

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func statusHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "pong");
}

func initStatus(r *mux.Router) {
	s := r.PathPrefix("/api/v1").Subrouter()
	s.HandleFunc("/ping", statusHandler)
}
