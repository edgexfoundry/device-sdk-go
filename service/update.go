// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/edgexfoundry/core-domain-go/models"
	"github.com/gorilla/mux"
)

func callbackHandler(w http.ResponseWriter, req *http.Request) {
	// use req.Method vs. method

	dec := json.NewDecoder(req.Body)
	cbAlert := models.CallbackAlert{}

	err := dec.Decode(&cbAlert)
	if err != nil {
		// TODO: handle error properly
		fmt.Fprintf(os.Stderr, "service: callbackHandler invalid request: %v\n", err)
	}

	action := cbAlert.ActionType
	id := cbAlert.Id

	fmt.Fprintf(os.Stderr, "service: callbackHandler action: %v id: %s\n", action, id)

	io.WriteString(w, "OK")
}

func initUpdate(r *mux.Router) {
	r.HandleFunc("/api/v1/callback", callbackHandler)
}
