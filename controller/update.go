// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017-2018 Canonical Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package controller

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
