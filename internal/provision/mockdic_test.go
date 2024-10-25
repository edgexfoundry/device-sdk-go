// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
// # Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
	"context"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapMocks "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/mock"
)

const (
	TestDeviceService             = "testDeviceService"
	TestDeviceWithTags            = "testDeviceWithTags"
	TestDeviceWithoutTags         = "testDeviceWithoutTags"
	TestProfile                   = "testProfile"
	TestDeviceResourceWithTags    = "testResourceWithTags"
	TestDeviceResourceWithoutTags = "testResourceWithoutTags"
	TestDeviceCommandWithTags     = "testCommandWithTags"
	TestDeviceCommandWithoutTags  = "testCommandWithoutTags"
	TestResourceTagName           = "testResourceTagName"
	TestResourceTagValue          = "testResourceTagValue"
	TestCommandTagName            = "testCommandTagName"
	TestCommandTagValue           = "testCommandTagValue"
	TestDeviceTagName             = "testDeviceTagName"
	TestDeviceTagValue            = "testDeviceTagValue"
	TestDuplicateTagName          = "testDuplicateTagName"
)

var profile = responses.DeviceProfileResponse{
	Profile: dtos.DeviceProfile{
		DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{Name: TestProfile},
		DeviceResources: []dtos.DeviceResource{
			{
				Name: TestDeviceResourceWithTags,
				Tags: dtos.Tags{
					TestResourceTagName: TestResourceTagValue,
				},
			},
			{
				Name: TestDeviceResourceWithoutTags,
			},
		},
		DeviceCommands: []dtos.DeviceCommand{
			{
				Name: TestDeviceCommandWithTags,
				Tags: dtos.Tags{
					TestCommandTagName:   TestCommandTagValue,
					TestDuplicateTagName: TestCommandTagValue,
				},
			},
			{
				Name: TestDeviceCommandWithoutTags,
			},
		},
	},
}

func NewMockDIC() (*di.Container, *clientMocks.DeviceProfileClient) {
	configuration := &config.ConfigurationStruct{
		Device: config.DeviceInfo{MaxCmdOps: 1},
	}
	deviceService := &models.DeviceService{Name: TestDeviceService}

	devices := responses.MultiDevicesResponse{
		Devices: []dtos.Device{
			{
				Name:        TestDeviceWithTags,
				ProfileName: TestProfile,
				Tags: dtos.Tags{
					TestDeviceTagName:    TestDeviceTagValue,
					TestDuplicateTagName: TestDeviceTagValue,
				},
			},
			{
				Name:        TestDeviceWithoutTags,
				ProfileName: TestProfile,
			},
		},
	}
	dcMock := &clientMocks.DeviceClient{}
	dcMock.On("DevicesByServiceName", context.Background(), TestDeviceService, 0, -1).Return(devices, nil)

	dpcMock := &clientMocks.DeviceProfileClient{}
	dpcMock.On("DeviceProfileByName", context.Background(), TestProfile).Return(profile, nil)

	pwcMock := &clientMocks.ProvisionWatcherClient{}
	pwcMock.On("ProvisionWatchersByServiceName", context.Background(), TestDeviceService, 0, -1).Return(responses.MultiProvisionWatchersResponse{}, nil)

	mockMetricsManager := &bootstrapMocks.MetricsManager{}
	mockMetricsManager.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMetricsManager.On("Unregister", mock.Anything)

	return di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		container.DeviceServiceName: func(get di.Get) any {
			return deviceService
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return dcMock
		},
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return dpcMock
		},
		bootstrapContainer.ProvisionWatcherClientName: func(get di.Get) interface{} {
			return pwcMock
		},
		bootstrapContainer.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return mockMetricsManager
		},
	}), dpcMock
}
