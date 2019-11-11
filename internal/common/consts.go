// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

const (
	ClientData     = "Data"
	ClientMetadata = "Metadata"
	ClientLogging  = "Logging"

	Colon      = ":"
	HttpScheme = "http://"
	HttpProto  = "HTTP"

	RegistryDefault    = "LOAD_FROM_FILE"
	ConfigDirectory    = "./res"
	ConfigFileName     = "configuration.toml"
	ConfigRegistryStem = "edgex/devices/1.0/"
	WritableKey        = "/Writable"
	RegistryFailLimit  = 3

	APICallbackRoute        = clients.ApiCallbackRoute
	APIValueDescriptorRoute = clients.ApiValueDescriptorRoute
	APIPingRoute            = clients.ApiPingRoute
	APIVersionRoute         = clients.ApiVersionRoute
	APIMetricsRoute         = clients.ApiMetricsRoute
	APIConfigRoute          = clients.ApiConfigRoute
	APIAllCommandRoute      = clients.ApiDeviceRoute + "/all/{command}"
	APIIdCommandRoute       = clients.ApiDeviceRoute + "/{id}/{command}"
	APINameCommandRoute     = clients.ApiDeviceRoute + "/name/{name}/{command}"
	APIDiscoveryRoute       = clients.ApiBase + "/discovery"
	APITransformRoute       = clients.ApiBase + "/debug/transformData/{transformData}"

	IdVar        string = "id"
	NameVar      string = "name"
	CommandVar   string = "command"
	GetCmdMethod string = "get"
	SetCmdMethod string = "set"

	CorrelationHeader = clients.CorrelationHeader
	URLRawQuery       = "urlRawQuery"
	SDKReservedPrefix = "ds-"
)
