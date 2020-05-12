// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
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

	EnvInstanceName = "EDGEX_INSTANCE_NAME"

	Colon      = ":"
	HttpScheme = "http://"
	HttpProto  = "HTTP"

	BootTimeoutSecondsDefault = 30
	BootRetrySecondsDefault   = 1

	ConfigStemDevice   = "edgex/devices/"
	ConfigMajorVersion = "1.0/"

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
