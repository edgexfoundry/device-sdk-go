// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
)

type EdgeXClients struct {
	GeneralClient          general.GeneralClient
	DeviceClient           metadata.DeviceClient
	DeviceServiceClient    metadata.DeviceServiceClient
	DeviceProfileClient    metadata.DeviceProfileClient
	AddressableClient      metadata.AddressableClient
	ProvisionWatcherClient metadata.ProvisionWatcherClient
	EventClient            coredata.EventClient
	ValueDescriptorClient  coredata.ValueDescriptorClient
}
