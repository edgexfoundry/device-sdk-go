// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package device

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func callbackHandler(w http.ResponseWriter, req *http.Request) {
	// use req.Method vs. method

	dec := json.NewDecoder(req.Body)
	cbAlert := models.CallbackAlert{}

	err := dec.Decode(&cbAlert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		svc.lc.Error(fmt.Sprintf("callbackHandler invalid request: %v", err))
		return
	}

	action := req.Method
	actionType := cbAlert.ActionType
	id := cbAlert.Id

	switch actionType {
	case models.DEVICE:
		if action == http.MethodPost {
			err = dc.AddById(id)
			if err == nil {
				svc.lc.Info(fmt.Sprintf("Added device %s", id))
			}
		} else if action == http.MethodPut {
			err = dc.Update(id)
			if err == nil {
				svc.lc.Info(fmt.Sprintf("Updated device %s", id))
			}
		} else if action == http.MethodDelete {
			err = dc.RemoveById(id)
			if err == nil {
				svc.lc.Info(fmt.Sprintf("Removed device %s", id))
			}
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			svc.lc.Error(fmt.Sprintf("callbackHandler invalid device action: %s", action))
			return
		}
		break
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
		svc.lc.Error(fmt.Sprintf("callbackHandler invalid action type: %s", actionType))
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		svc.lc.Error(err.Error())
		return
	}

	io.WriteString(w, "OK")
}

func initUpdate() {
	svc.r.HandleFunc("/callback", callbackHandler)
}
