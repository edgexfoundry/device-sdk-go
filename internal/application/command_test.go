//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models/mocks"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var testProtocols map[string]models.ProtocolProperties

var testDevice = models.Device{
	Name:           "test-device",
	AdminState:     models.Unlocked,
	OperatingState: models.Up,
	Protocols:      testProtocols,
	ServiceName:    "test-service",
	ProfileName:    "test-profile",
}

func mockDic() *di.Container {
	driverMock := &mocks.ProtocolDriver{}
	cr := sdkModels.CommandRequest{
		DeviceResourceName: "test-resource",
		Attributes:         nil,
		Type:               "String",
	}
	cv := &sdkModels.CommandValue{
		DeviceResourceName: "test-resource",
		Type:               "String",
		Value:              "test-value",
		Tags:               make(map[string]string),
	}
	objectRequest := sdkModels.CommandRequest{
		DeviceResourceName: "rw-object",
		Attributes:         nil,
		Type:               common.ValueTypeObject,
	}
	objectValue := &sdkModels.CommandValue{
		DeviceResourceName: "rw-object",
		Type:               common.ValueTypeObject,
		Value:              map[string]interface{}{"foo": "bar"},
		Tags:               make(map[string]string),
	}
	driverMock.On("HandleReadCommands", "test-device", testProtocols, []sdkModels.CommandRequest{cr}).Return(nil, nil)
	driverMock.On("HandleWriteCommands", "test-device", testProtocols, []sdkModels.CommandRequest{cr}, []*sdkModels.CommandValue{cv}).Return(nil)
	driverMock.On("HandleWriteCommands", "test-device", testProtocols, []sdkModels.CommandRequest{objectRequest}, []*sdkModels.CommandValue{objectValue}).Return(nil)

	devices := responses.MultiDevicesResponse{
		BaseWithTotalCountResponse: dtoCommon.BaseWithTotalCountResponse{},
		Devices: []dtos.Device{
			dtos.FromDeviceModelToDTO(testDevice),
		},
	}
	dcMock := &clientMocks.DeviceClient{}
	dcMock.On("DevicesByServiceName", context.Background(), "test-service", 0, -1).Return(devices, nil)

	profile := responses.DeviceProfileResponse{
		BaseResponse: dtoCommon.BaseResponse{},
		Profile: dtos.DeviceProfile{
			DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{Name: "test-profile"},
			DeviceResources: []dtos.DeviceResource{
				dtos.DeviceResource{
					Name: "test-resource",
					Properties: dtos.ResourceProperties{
						ValueType:    "String",
						ReadWrite:    "RW",
						DefaultValue: "test-value",
					},
				},
				dtos.DeviceResource{
					Name: "ro-resource",
					Properties: dtos.ResourceProperties{
						ValueType: "String",
						ReadWrite: "R",
					},
				},
				dtos.DeviceResource{
					Name: "wo-resource",
					Properties: dtos.ResourceProperties{
						ValueType: "String",
						ReadWrite: "W",
					},
				},
				dtos.DeviceResource{
					Name: "rw-object",
					Properties: dtos.ResourceProperties{
						ValueType: common.ValueTypeObject,
						ReadWrite: "RW",
					},
				},
			},
			DeviceCommands: []dtos.DeviceCommand{
				dtos.DeviceCommand{
					Name:               "test-command",
					IsHidden:           false,
					ReadWrite:          "RW",
					ResourceOperations: []dtos.ResourceOperation{{DeviceResource: "test-resource"}},
				},
				dtos.DeviceCommand{
					Name:               "ro-command",
					IsHidden:           false,
					ReadWrite:          "R",
					ResourceOperations: []dtos.ResourceOperation{{DeviceResource: "ro-resource"}},
				},
				dtos.DeviceCommand{
					Name:               "wo-command",
					IsHidden:           false,
					ReadWrite:          "W",
					ResourceOperations: []dtos.ResourceOperation{{DeviceResource: "wo-resource"}},
				},
				dtos.DeviceCommand{
					Name:               "exceed-command",
					IsHidden:           false,
					ReadWrite:          "R",
					ResourceOperations: []dtos.ResourceOperation{{DeviceResource: "test-resource"}, {DeviceResource: "ro-resource"}},
				},
			},
		},
	}
	dpcMock := &clientMocks.DeviceProfileClient{}
	dpcMock.On("DeviceProfileByName", context.Background(), "test-profile").Return(profile, nil)

	pwcMock := &clientMocks.ProvisionWatcherClient{}
	pwcMock.On("ProvisionWatchersByServiceName", context.Background(), "test-service", 0, -1).Return(responses.MultiProvisionWatchersResponse{}, nil)

	configuration := &config.ConfigurationStruct{
		Device: config.DeviceInfo{MaxCmdOps: 1},
	}

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ProtocolDriverName: func(get di.Get) interface{} {
			return driverMock
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
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	return dic
}

func TestCommandProcessor_ReadDeviceResource(t *testing.T) {
	dic := mockDic()
	err := cache.InitCache("test-service", dic)
	require.NoError(t, err)

	valid := NewCommandProcessor(testDevice, "test-resource", uuid.NewString(), nil, "", dic)
	invalidDeviceResource := NewCommandProcessor(testDevice, "invalid", uuid.NewString(), nil, "", dic)
	writeOnlyDeviceResource := NewCommandProcessor(testDevice, "wo-resource", uuid.NewString(), nil, "", dic)

	tests := []struct {
		name             string
		commandProcessor *CommandProcessor
		expectedErr      bool
	}{
		{"valid", valid, false},
		{"invalid - DeviceResource name not found", invalidDeviceResource, true},
		{"invalid - reading write-only DeviceResource", writeOnlyDeviceResource, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.commandProcessor.ReadDeviceResource()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCommandProcessor_ReadDeviceCommand(t *testing.T) {
	dic := mockDic()
	err := cache.InitCache("test-service", dic)
	require.NoError(t, err)

	valid := NewCommandProcessor(testDevice, "test-command", uuid.NewString(), nil, "", dic)
	invalidDeviceCommand := NewCommandProcessor(testDevice, "invalid", uuid.NewString(), nil, "", dic)
	writeOnlyDeviceCommand := NewCommandProcessor(testDevice, "wo-command", uuid.NewString(), nil, "", dic)
	outOfRangeResourceOperation := NewCommandProcessor(testDevice, "exceed-command", uuid.NewString(), nil, "", dic)

	tests := []struct {
		name             string
		commandProcessor *CommandProcessor
		expectedErr      bool
	}{
		{"valid", valid, false},
		{"invalid - DeviceCommand name not found", invalidDeviceCommand, true},
		{"invalid - reading write-only DeviceCommand", writeOnlyDeviceCommand, true},
		{"invalid - RO exceed MaxCmdOps count", outOfRangeResourceOperation, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.commandProcessor.ReadDeviceCommand()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCommandProcessor_WriteDeviceResource(t *testing.T) {
	dic := mockDic()
	err := cache.InitCache("test-service", dic)
	require.NoError(t, err)

	valid := NewCommandProcessor(testDevice, "test-resource", uuid.NewString(), map[string]interface{}{"test-resource": "test-value"}, "", dic)
	validObjectValue := NewCommandProcessor(testDevice, "rw-object", uuid.NewString(), map[string]interface{}{"rw-object": map[string]interface{}{"foo": "bar"}}, "", dic)
	invalidDeviceResource := NewCommandProcessor(testDevice, "invalid", uuid.NewString(), nil, "", dic)
	readOnlyDeviceResource := NewCommandProcessor(testDevice, "ro-resource", uuid.NewString(), nil, "", dic)
	noRequestBody := NewCommandProcessor(testDevice, "test-resource", uuid.NewString(), nil, "", dic)
	invalidRequestBody := NewCommandProcessor(testDevice, "test-resource", uuid.NewString(), map[string]interface{}{"wrong-resource": "wrong-value"}, "", dic)

	tests := []struct {
		name             string
		commandProcessor *CommandProcessor
		expectedErr      bool
	}{
		{"valid", valid, false},
		{"valid object value", validObjectValue, false},
		{"invalid - DeviceResource name not found", invalidDeviceResource, true},
		{"invalid - writing read-only DeviceResource", readOnlyDeviceResource, true},
		{"valid - no set parameter specified but default value exists", noRequestBody, false},
		{"valid - set parameter doesn't match requested command, using DefaultValue in DeviceResource.Properties", invalidRequestBody, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = tt.commandProcessor.WriteDeviceResource()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCommandProcessor_WriteDeviceCommand(t *testing.T) {
	dic := mockDic()
	err := cache.InitCache("test-service", dic)
	require.NoError(t, err)

	valid := NewCommandProcessor(testDevice, "test-command", uuid.NewString(), map[string]interface{}{"test-resource": "test-value"}, "", dic)
	invalidDeviceCommand := NewCommandProcessor(testDevice, "invalid", uuid.NewString(), nil, "", dic)
	readOnlyDeviceCommand := NewCommandProcessor(testDevice, "ro-command", uuid.NewString(), nil, "", dic)
	outOfRangeResourceOperation := NewCommandProcessor(testDevice, "exceed-command", uuid.NewString(), nil, "", dic)
	noRequestBody := NewCommandProcessor(testDevice, "test-command", uuid.NewString(), nil, "", dic)
	invalidRequestBody := NewCommandProcessor(testDevice, "test-command", uuid.NewString(), map[string]interface{}{"wrong-resource": "wrong-value"}, "", dic)

	tests := []struct {
		name             string
		commandProcessor *CommandProcessor
		expectedErr      bool
	}{
		{"valid", valid, false},
		{"invalid - DeviceCommand name not found", invalidDeviceCommand, true},
		{"invalid - writing read-only DeviceCommand", readOnlyDeviceCommand, true},
		{"invalid - RO exceed MaxCmdOps count", outOfRangeResourceOperation, true},
		{"valid - no set parameter specified but default value exist", noRequestBody, false},
		{"valid - parameter doesn't match requested command, using DefaultValue in DeviceResource.Properties", invalidRequestBody, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = tt.commandProcessor.WriteDeviceCommand()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
