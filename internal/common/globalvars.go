// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	ServiceName           string
	ServiceVersion        string
	CurrentConfig         *Config
	CurrentDeviceService  contract.DeviceService
	UseRegistry           bool
	ServiceLocked         bool
	Driver                dsModels.ProtocolDriver
	EventClient           coredata.EventClient
	AddressableClient     metadata.AddressableClient
	DeviceClient          metadata.DeviceClient
	DeviceServiceClient   metadata.DeviceServiceClient
	DeviceProfileClient   metadata.DeviceProfileClient
	LoggingClient         logger.LoggingClient
	ValueDescriptorClient coredata.ValueDescriptorClient
)
