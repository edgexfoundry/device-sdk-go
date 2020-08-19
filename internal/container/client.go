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

var GeneralClientName = di.TypeInstanceToName((*general.GeneralClient)(nil))
var MetadataDeviceClientName = di.TypeInstanceToName((*metadata.DeviceClient)(nil))
var MetadataDeviceServiceClientName = di.TypeInstanceToName((*metadata.DeviceServiceClient)(nil))
var MetadataDeviceProfileClientName = di.TypeInstanceToName((*metadata.DeviceProfileClient)(nil))
var MetadataAddressableClientName = di.TypeInstanceToName((*metadata.AddressableClient)(nil))
var MetadataProvisionWatcherClientName = di.TypeInstanceToName((*metadata.ProvisionWatcherClient)(nil))
var CoredataEventClientName = di.TypeInstanceToName((*coredata.EventClient)(nil))
var CoredataValueDescriptorClientName = di.TypeInstanceToName((*coredata.ValueDescriptorClient)(nil))

func GeneralClientFrom(get di.Get) general.GeneralClient {
	return get(GeneralClientName).(general.GeneralClient)
}

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

func CoredataEventClientFrom(get di.Get) coredata.EventClient {
	return get(CoredataEventClientName).(coredata.EventClient)
}

func CoredataValueDescriptorClientFrom(get di.Get) coredata.ValueDescriptorClient {
	return get(CoredataValueDescriptorClientName).(coredata.ValueDescriptorClient)
}
