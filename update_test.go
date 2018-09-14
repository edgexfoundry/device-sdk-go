// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package device

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/gorilla/mux"
)

// Test update REST calls
func TestUpdate(t *testing.T) {
	var tests = []struct {
		name   string
		method string
		body   string
		code   int
	}{
		{"Empty body", http.MethodPut, "", http.StatusBadRequest},
		{"Empty json", http.MethodPut, "{}", http.StatusBadRequest},
		{"Invalid type", http.MethodPut, `{"id":"5b9a4f9a64562a2f966fdb0b","type":"INVALID"}`, http.StatusBadRequest},
		{"Invalid method", http.MethodPost, `{"id":"5b9a4f9a64562a2f966fdb0b","type":"DEVICE"}`, http.StatusBadRequest},
		{"Invalid id", http.MethodPut, `{"id":"5b9a4f9a64562a2f966fdb0b","type":"DEVICE"}`, http.StatusInternalServerError},
	}

	lc := logger.NewClient("update_test", false, "")
	r := mux.NewRouter().PathPrefix(apiV1).Subrouter()
	svc = &Service{Name: "update-test", lc: lc, r: r, locked: true}
	initUpdate()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonStr = []byte(tt.body)
			req := httptest.NewRequest(tt.method, v1Callback, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			svc.r.ServeHTTP(rr, req)
			if status := rr.Code; status != tt.code {
				t.Errorf("CallbackHandler: handler returned wrong status code: got %v want %v",
					status, http.StatusLocked)
			}
		})
	}
}
