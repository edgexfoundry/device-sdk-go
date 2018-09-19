// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package common

import (
	"bytes"
)

func BuildAddr(host string, port string) string {
	var buffer bytes.Buffer

	buffer.WriteString(HttpScheme)
	buffer.WriteString(host)
	buffer.WriteString(Colon)
	buffer.WriteString(port)

	return buffer.String()
}

