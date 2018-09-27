// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"io"
	"net/http"
)

func statusHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "pong")
}

func initStatus() {
	svc.r.HandleFunc("/ping", statusHandler)
}
