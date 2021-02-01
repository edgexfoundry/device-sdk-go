// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/interfaces"
)

var MetadataDeviceClientName = di.TypeInstanceToName((*interfaces.DeviceClient)(nil))
var MetadataDeviceServiceClientName = di.TypeInstanceToName((*interfaces.DeviceServiceClient)(nil))
var MetadataDeviceProfileClientName = di.TypeInstanceToName((*interfaces.DeviceProfileClient)(nil))
var MetadataProvisionWatcherClientName = di.TypeInstanceToName((*interfaces.ProvisionWatcherClient)(nil))
var CoredataEventClientName = di.TypeInstanceToName((*interfaces.EventClient)(nil))

func MetadataDeviceClientFrom(get di.Get) interfaces.DeviceClient {
	return get(MetadataDeviceClientName).(interfaces.DeviceClient)
}

func MetadataDeviceServiceClientFrom(get di.Get) interfaces.DeviceServiceClient {
	return get(MetadataDeviceServiceClientName).(interfaces.DeviceServiceClient)
}

func MetadataDeviceProfileClientFrom(get di.Get) interfaces.DeviceProfileClient {
	return get(MetadataDeviceProfileClientName).(interfaces.DeviceProfileClient)
}

func MetadataProvisionWatcherClientFrom(get di.Get) interfaces.ProvisionWatcherClient {
	return get(MetadataProvisionWatcherClientName).(interfaces.ProvisionWatcherClient)
}

func CoredataEventClientFrom(get di.Get) interfaces.EventClient {
	return get(CoredataEventClientName).(interfaces.EventClient)
}
