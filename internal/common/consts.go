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
	v2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
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

	APIV2DeviceCallbackRoute    = v2.ApiBase + "/callback/device"
	APIV2DeviceCallbackIdRoute  = v2.ApiBase + "/callback/device/id/{id}"
	APIV2ProfileCallbackRoute   = v2.ApiBase + "/callback/profile"
	APIV2ProfileCallbackIdRoute = v2.ApiBase + "/callback/profile/id/{id}"
	APIV2WatcherCallbackRoute   = v2.ApiBase + "/callback/watcher"
	APIV2WatcherCallbackIdRoute = v2.ApiBase + "/callback/watcher/id/{id}"
	APIV2DiscoveryRoute         = v2.ApiBase + "/discovery"
	APIV2IdCommandRoute         = v2.ApiBase + "/device/{id}/{command}"
	APIV2NameCommandRoute       = v2.ApiBase + "/device/name/{name}/{command}"

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
