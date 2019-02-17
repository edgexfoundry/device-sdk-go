// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/edgexfoundry/edgex-go/pkg/clients"
)

const (
	ClientData     = "Data"
	ClientMetadata = "Metadata"
	ClientLogging  = "Logging"

	APIv1Prefix       = "/api/v1"
	Colon             = ":"
	HttpScheme        = "http://"
	HttpProto         = "HTTP"
	StatusResponse    = "pong"
	ServiceStatusPass = "passing"
	ConfigFileName    = "configuration.toml"
	ConfigV2Stem      = "config/V2/"

	APICallbackRoute        = APIv1Prefix + "/callback"
	APIValueDescriptorRoute = APIv1Prefix + "/valuedescriptor"
	APIDiscoveryRoute       = APIv1Prefix + "/discovery"
	APIPingRoute            = APIv1Prefix + "/ping"

	SchedulerExecCMDPattern = APIv1Prefix + "/device/name/*/*"

	CorrelationHeader = clients.CorrelationHeader
)
