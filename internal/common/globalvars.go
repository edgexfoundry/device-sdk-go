// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

var (
	ServiceName            string
	ServiceVersion         string
	CurrentConfig          *ConfigurationStruct
	CurrentDeviceService   contract.DeviceService
	ServiceLocked          bool
	Driver                 dsModels.ProtocolDriver
	RegistryClient         registry.Client
	Discovery              dsModels.ProtocolDiscovery
	EventClient            coredata.EventClient
	AddressableClient      metadata.AddressableClient
	DeviceClient           metadata.DeviceClient
	DeviceServiceClient    metadata.DeviceServiceClient
	DeviceProfileClient    metadata.DeviceProfileClient
	LoggingClient          logger.LoggingClient
	ValueDescriptorClient  coredata.ValueDescriptorClient
	MetadataGeneralClient  general.GeneralClient
	ProvisionWatcherClient metadata.ProvisionWatcherClient
)
