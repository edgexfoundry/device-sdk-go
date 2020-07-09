// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
)

var MetadataDeviceClientName = "MetadataDeviceClient"
var MetadataDeviceServiceClientName = "MetadataDeviceServiceClient"
var MetadataDeviceProfileClientName = "MetadataDeviceProfileClient"
var MetadataAddressableClientName = "MetadataAddressableClient"
var MetadataProvisionWatcherClientName = "MetadataProvisionWatcherClient"
var MetadataGeneralClientName = "MetadataGeneralClient"
var CoredataEventClientName = "CoredataEventClient"
var CoredataValueDescriptorName = "CoredataValueDescriptor"

func MetadataDeviceClientFrom(get di.Get) metadata.DeviceClient {
	return get(MetadataDeviceClientName).(metadata.DeviceClient)
}

func MetadataDeviceServiceClientFrom(get di.Get) metadata.DeviceServiceClient {
	return get(MetadataDeviceServiceClientName).(metadata.DeviceServiceClient)
}

func MetadataDeviceProfileClientFrom(get di.Get) metadata.DeviceProfileClient {
	return get(MetadataDeviceProfileClientName).(metadata.DeviceProfileClient)
}

func MetadataAddressableClientFrom(get di.Get) metadata.AddressableClient {
	return get(MetadataAddressableClientName).(metadata.AddressableClient)
}

func MetadataProvisionWatcherClientFrom(get di.Get) metadata.ProvisionWatcherClient {
	return get(MetadataProvisionWatcherClientName).(metadata.ProvisionWatcherClient)
}

func MetadataGeneralClientFrom(get di.Get) general.GeneralClient {
	return get(MetadataGeneralClientName).(general.GeneralClient)
}

func CoredataEventClientFrom(get di.Get) coredata.EventClient {
	return get(CoredataEventClientName).(coredata.EventClient)
}

func CoredataValueDescriptorClientFrom(get di.Get) coredata.ValueDescriptorClient {
	return get(CoredataValueDescriptorName).(coredata.ValueDescriptorClient)
}
