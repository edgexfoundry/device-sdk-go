// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package common

const (
	ClientData     = "Data"
	ClientMetadata = "Metadata"

	ServiceName    = "device-simple"
	ServiceVersion = "0.1"

	APIPrefix  = "/api/v1"
	Colon      = ":"
	HttpScheme = "http://"
	HttpProto  = "HTTP"

	V1Addressable = APIPrefix + "/addressable"
	V1Callback    = APIPrefix + "/callback"
	V1Device      = APIPrefix + "/device"
	V1DevService  = APIPrefix + "/deviceservice"
	V1Event       = APIPrefix + "/event"
)