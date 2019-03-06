// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	ServiceName           string
	ServiceVersion        string
	CurrentConfig         *Config
	CurrentDeviceService  models.DeviceService
	UseRegistry           bool
	ServiceLocked         bool
	Driver                ds_models.ProtocolDriver
	EventClient           coredata.EventClient
	AddressableClient     metadata.AddressableClient
	DeviceClient          metadata.DeviceClient
	DeviceServiceClient   metadata.DeviceServiceClient
	DeviceProfileClient   metadata.DeviceProfileClient
	LoggingClient         logger.LoggingClient
	ValueDescriptorClient coredata.ValueDescriptorClient
)
