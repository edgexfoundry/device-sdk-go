//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapMocks "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces/mocks"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

const (
	TestDevice         = "testDevice"
	TestDeviceService  = "testDeviceService"
	TestProfile        = "testProfile"
	TestDeviceResource = "testResource"
	TestDeviceCommand  = "testCommand"
	TestValue          = "testValue"
	TestUnits          = "testUnits"
)

var TestProtocols map[string]dtos.ProtocolProperties

func NewMockDIC() *di.Container {
	cr := sdkModels.CommandRequest{
		DeviceResourceName: TestDeviceResource,
		Attributes:         nil,
		Type:               common.ValueTypeString,
	}
	cv := &sdkModels.CommandValue{
		DeviceResourceName: TestDeviceResource,
		Type:               common.ValueTypeString,
		Value:              TestValue,
		Tags:               make(map[string]string),
	}
	driverMock := &mocks.ProtocolDriver{}
	driverMock.On("HandleReadCommands", TestDevice, TestProtocols, []sdkModels.CommandRequest{cr}).Return(nil, nil)
	driverMock.On("HandleWriteCommands", TestDevice, TestProtocols, []sdkModels.CommandRequest{cr}, []*sdkModels.CommandValue{cv}).Return(nil)

	devices := responses.MultiDevicesResponse{
		BaseWithTotalCountResponse: dtoCommon.BaseWithTotalCountResponse{},
		Devices: []dtos.Device{
			{
				Name:           TestDevice,
				AdminState:     models.Unlocked,
				OperatingState: models.Up,
				Protocols:      TestProtocols,
				ServiceName:    TestDeviceService,
				ProfileName:    TestProfile,
			},
		},
	}
	dcMock := &clientMocks.DeviceClient{}
	dcMock.On("DevicesByServiceName", context.Background(), TestDeviceService, 0, -1).Return(devices, nil)

	profile := responses.DeviceProfileResponse{
		BaseResponse: dtoCommon.BaseResponse{},
		Profile: dtos.DeviceProfile{
			DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{Name: TestProfile},
			DeviceResources: []dtos.DeviceResource{
				{
					Name: TestDeviceResource,
					Properties: dtos.ResourceProperties{
						ValueType:    common.ValueTypeString,
						ReadWrite:    common.ReadWrite_RW,
						DefaultValue: TestValue,
						Units:        TestUnits,
					},
				},
			},
			DeviceCommands: []dtos.DeviceCommand{
				{
					Name:               TestDeviceCommand,
					IsHidden:           false,
					ReadWrite:          common.ReadWrite_RW,
					ResourceOperations: []dtos.ResourceOperation{{DeviceResource: TestDeviceResource}},
				},
			},
		},
	}
	dpcMock := &clientMocks.DeviceProfileClient{}
	dpcMock.On("DeviceProfileByName", context.Background(), TestProfile).Return(profile, nil)

	pwcMock := &clientMocks.ProvisionWatcherClient{}
	pwcMock.On("ProvisionWatchersByServiceName", context.Background(), TestDeviceService, 0, -1).Return(responses.MultiProvisionWatchersResponse{}, nil)

	configuration := &config.ConfigurationStruct{
		Device: config.DeviceInfo{MaxCmdOps: 1},
	}

	mockMetricsManager := &bootstrapMocks.MetricsManager{}
	mockMetricsManager.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMetricsManager.On("Unregister", mock.Anything)

	return di.NewContainer(di.ServiceConstructorMap{
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
		bootstrapContainer.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return mockMetricsManager
		},
	})
}
func Test_getUniqueOrigin(t *testing.T) {
	// nolint: gosec
	for i := 0; i < rand.Intn(1000); i++ {
		t.Run(fmt.Sprintf("TestCase%d", i), func(t *testing.T) {
			t.Parallel()
			o1 := getUniqueOrigin()
			o2 := getUniqueOrigin()
			assert.NotEqual(t, o1, o2)
		})
	}
}

func TestCommandValuesToEventDTO_ReadingUnits(t *testing.T) {
	dic := NewMockDIC()
	err := cache.InitCache(TestDeviceService, TestDeviceService, dic)
	require.NoError(t, err)

	stringCommandValue, e := sdkModels.NewCommandValue(TestDeviceResource, common.ValueTypeString, TestValue)
	require.NoError(t, e)

	tests := []struct {
		Name          string
		CommandValues []*sdkModels.CommandValue
		ReadingUnits  bool
	}{
		{"ReadingUnits is true, indicate Units in the Reading", []*sdkModels.CommandValue{stringCommandValue}, true},
		{"ReadingUnits is false, not to indicate Units in the Reading", []*sdkModels.CommandValue{stringCommandValue}, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			configuration := container.ConfigurationFrom(dic.Get)
			configuration.Writable.Reading.ReadingUnits = testCase.ReadingUnits
			dic.Update(di.ServiceConstructorMap{
				container.ConfigurationName: func(get di.Get) interface{} {
					return configuration
				},
			})
			event, err := CommandValuesToEventDTO(testCase.CommandValues, TestDevice, TestDeviceCommand, configuration.Device.DataTransform, dic)
			require.NoError(t, err)

			assert.Equal(t, TestDevice, event.DeviceName)
			assert.Equal(t, TestProfile, event.ProfileName)
			assert.Equal(t, TestDeviceCommand, event.SourceName)
			assert.Equal(t, TestDeviceResource, event.Readings[0].ResourceName)
			if testCase.ReadingUnits {
				assert.NotEmpty(t, event.Readings[0].Units, "units is not presented")
				assert.Equal(t, TestUnits, event.Readings[0].Units)
			} else {
				assert.Empty(t, event.Readings[0].Units, "units is presented")
			}
		})
	}
}

func TestCommandValuesToEventDTO_ReadingNilValue(t *testing.T) {
	dic := NewMockDIC()
	err := cache.InitCache(TestDeviceService, TestDeviceService, dic)
	require.NoError(t, err)
	cvs := []*sdkModels.CommandValue{
		testCommandNilValue(t, common.ValueTypeBool),
		testCommandNilValue(t, common.ValueTypeString),
		testCommandNilValue(t, common.ValueTypeUint8),
		testCommandNilValue(t, common.ValueTypeUint16),
		testCommandNilValue(t, common.ValueTypeUint32),
		testCommandNilValue(t, common.ValueTypeUint64),
		testCommandNilValue(t, common.ValueTypeInt8),
		testCommandNilValue(t, common.ValueTypeInt16),
		testCommandNilValue(t, common.ValueTypeInt32),
		testCommandNilValue(t, common.ValueTypeInt64),
		testCommandNilValue(t, common.ValueTypeFloat32),
		testCommandNilValue(t, common.ValueTypeFloat64),
		testCommandNilValue(t, common.ValueTypeBinary),
		testCommandNilValue(t, common.ValueTypeBoolArray),
		testCommandNilValue(t, common.ValueTypeStringArray),
		testCommandNilValue(t, common.ValueTypeUint8Array),
		testCommandNilValue(t, common.ValueTypeUint16Array),
		testCommandNilValue(t, common.ValueTypeUint32Array),
		testCommandNilValue(t, common.ValueTypeUint64Array),
		testCommandNilValue(t, common.ValueTypeInt8Array),
		testCommandNilValue(t, common.ValueTypeInt16Array),
		testCommandNilValue(t, common.ValueTypeInt32Array),
		testCommandNilValue(t, common.ValueTypeInt64Array),
		testCommandNilValue(t, common.ValueTypeFloat32Array),
		testCommandNilValue(t, common.ValueTypeFloat64Array),
		testCommandNilValue(t, common.ValueTypeObject),
		testCommandNilValue(t, common.ValueTypeObjectArray),
	}
	event, err := CommandValuesToEventDTO(cvs, TestDevice, TestDeviceCommand, true, dic)
	require.NoError(t, err)

	for _, r := range event.Readings {
		assert.Empty(t, r.Value)
		assert.Empty(t, r.BinaryValue)
		assert.Empty(t, r.ObjectValue)
		expectedNullReading := dtos.NewNullReading(TestProfile, TestDevice, TestDeviceResource, r.ValueType)
		expectedNullReading.Id = r.Id
		expectedNullReading.Origin = r.Origin
		assert.Equal(t, expectedNullReading, r)
	}
}
func testCommandNilValue(t *testing.T, valueType string) *sdkModels.CommandValue {
	cv, e := sdkModels.NewCommandValue(TestDeviceResource, valueType, nil)
	require.NoError(t, e)
	return cv
}
