// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
)

const (
	ClientData     = "Data"
	ClientMetadata = "Metadata"

	EnvInstanceName = "EDGEX_INSTANCE_NAME"

	Colon      = ":"
	HttpScheme = "http://"
	HttpProto  = "HTTP"

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

	APIV2SecretRoute = v2.ApiBase + "/secret"

	IdVar        string = "id"
	NameVar      string = "name"
	CommandVar   string = "command"
	GetCmdMethod string = "get"
	SetCmdMethod string = "set"

	DeviceResourceReadOnly  string = "R"
	DeviceResourceWriteOnly string = "W"

	CorrelationHeader = clients.CorrelationHeader
	URLRawQuery       = "urlRawQuery"
	SDKReservedPrefix = "ds-"
)

// SDKVersion indicates the version of the SDK - will be overwritten by build
var SDKVersion string = "0.0.0"

// ServiceVersion indicates the version of the device service itself, not the SDK - will be overwritten by build
var ServiceVersion string = "0.0.0"
