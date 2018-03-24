// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package controller

import (
	"github.com/gorilla/mux"
)

// TODO: need to add support for graceful shutdown

// A Daemon listens for requests and routes them to the right command
type Mux struct {
	initialized   bool
	router        *mux.Router
}

func (m *Mux) Init() {
	m.router = mux.NewRouter()
	initCommand(m.router)
	initStatus(m.router)
	initService(m.router)
	initUpdate(m.router)
}

// New Mux
// TODO: re-factor to make this a singleton
func New() (*Mux, error) {
	return &Mux{}, nil
}
