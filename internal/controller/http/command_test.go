//
// Copyright (C) 2022-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapMocks "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	messagingMocks "github.com/edgexfoundry/go-mod-messaging/v4/messaging/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces/mocks"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"

	"github.com/labstack/echo/v4"
)

const (
	testService = "test-service"
	testProfile = "test-profile"

	testDevice        = "test-device"
	lockedDevice      = "locked-device"
	downedDevice      = "down-device"
	driverErrorDevice = "driver-device"

	testCommand      = "test-command"
	readOnlyCommand  = "ro-command"
	writeOnlyCommand = "wo-command"
	exceedCommand    = "exceed-command"

	testResource      = "test-resource"
	readOnlyResource  = "ro-resource"
	writeOnlyResource = "wo-resource"
	objectResource    = "object-resource"

	testRegexResource = "^t.+-resource"
)

func mockDic() *di.Container {
	devices := []dtos.Device{
		dtos.Device{
			Name:           testDevice,
			AdminState:     models.Unlocked,
			OperatingState: models.Up,
			ServiceName:    testService,
			ProfileName:    testProfile,
		},
		dtos.Device{
			Name:           lockedDevice,
			AdminState:     models.Locked,
			OperatingState: models.Up,
			ServiceName:    testService,
			ProfileName:    testProfile,
		},
		dtos.Device{
			Name:           downedDevice,
			AdminState:     models.Unlocked,
			OperatingState: models.Down,
			ServiceName:    testService,
			ProfileName:    testProfile,
		},
		dtos.Device{
			Name:           driverErrorDevice,
			AdminState:     models.Unlocked,
			OperatingState: models.Unlocked,
			ServiceName:    testService,
			ProfileName:    testProfile,
		},
	}
	deviceResponse := responses.NewMultiDevicesResponse("", "", http.StatusOK, 4, devices)
	profile := dtos.DeviceProfile{
		DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{
			Name: testProfile,
		},
		DeviceResources: []dtos.DeviceResource{
			dtos.DeviceResource{
				Name: testResource,
				Properties: dtos.ResourceProperties{
					ValueType:    common.ValueTypeString,
					ReadWrite:    common.ReadWrite_RW,
					DefaultValue: "default",
				},
			},
			dtos.DeviceResource{
				Name: readOnlyResource,
				Properties: dtos.ResourceProperties{
					ValueType: common.ValueTypeString,
					ReadWrite: common.ReadWrite_R,
				},
			},
			dtos.DeviceResource{
				Name: writeOnlyResource,
				Properties: dtos.ResourceProperties{
					ValueType: common.ValueTypeString,
					ReadWrite: common.ReadWrite_W,
				},
			},
			dtos.DeviceResource{
				Name: objectResource,
				Properties: dtos.ResourceProperties{
					ValueType: common.ValueTypeObject,
					ReadWrite: common.ReadWrite_RW,
				},
			},
		},
		DeviceCommands: []dtos.DeviceCommand{
			dtos.DeviceCommand{
				Name:               testCommand,
				ReadWrite:          common.ReadWrite_RW,
				ResourceOperations: []dtos.ResourceOperation{{DeviceResource: testResource, DefaultValue: "default"}},
			},
			dtos.DeviceCommand{
				Name:               readOnlyCommand,
				ReadWrite:          common.ReadWrite_R,
				ResourceOperations: []dtos.ResourceOperation{{DeviceResource: readOnlyResource}},
			},
			dtos.DeviceCommand{
				Name:               writeOnlyCommand,
				ReadWrite:          common.ReadWrite_W,
				ResourceOperations: []dtos.ResourceOperation{{DeviceResource: writeOnlyResource}},
			},
			dtos.DeviceCommand{
				Name:               exceedCommand,
				ReadWrite:          common.ReadWrite_RW,
				ResourceOperations: []dtos.ResourceOperation{{DeviceResource: testResource}, {DeviceResource: testResource}},
			},
		},
	}
	profileResponse := responses.NewDeviceProfileResponse("", "", http.StatusOK, profile)
	provisionWatcherResponse := responses.NewMultiProvisionWatchersResponse("", "", http.StatusOK, 0, nil)
	commandValue := &sdkModels.CommandValue{
		DeviceResourceName: testResource,
		Type:               common.ValueTypeString,
		Value:              "test",
	}

	mockDeviceClient := &clientMocks.DeviceClient{}
	mockDeviceClient.On("DevicesByServiceName", context.Background(), testService, 0, -1).Return(deviceResponse, nil)
	mockDeviceProfileClient := &clientMocks.DeviceProfileClient{}
	mockDeviceProfileClient.On("DeviceProfileByName", context.Background(), testProfile).Return(profileResponse, nil)
	mockProvisionWatcherClient := &clientMocks.ProvisionWatcherClient{}
	mockProvisionWatcherClient.On("ProvisionWatchersByServiceName", context.Background(), testService, 0, -1).Return(provisionWatcherResponse, nil)
	mockDriver := &mocks.ProtocolDriver{}
	mockDriver.On("HandleReadCommands", testDevice, mock.Anything, mock.Anything).Return([]*sdkModels.CommandValue{commandValue}, nil)
	mockDriver.On("HandleReadCommands", driverErrorDevice, mock.Anything, mock.Anything).Return(nil, errors.New("ProtocolDriver returned error"))
	mockDriver.On("HandleWriteCommands", testDevice, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockDriver.On("HandleWriteCommands", driverErrorDevice, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("ProtocolDriver returned error"))

	mockMetricsManager := &bootstrapMocks.MetricsManager{}
	mockMetricsManager.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMetricsManager.On("Unregister", mock.Anything)

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) any {
			return &config.ConfigurationStruct{
				Device: config.DeviceInfo{
					MaxCmdOps: 1,
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) any {
			return logger.NewMockClient()
		},
		bootstrapContainer.DeviceClientName: func(get di.Get) any {
			return mockDeviceClient
		},
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) any {
			return mockDeviceProfileClient
		},
		bootstrapContainer.ProvisionWatcherClientName: func(get di.Get) any {
			return mockProvisionWatcherClient
		},
		container.ProtocolDriverName: func(get di.Get) any {
			return mockDriver
		},
		container.DeviceServiceName: func(get di.Get) any {
			return &models.DeviceService{
				Name:       testService,
				AdminState: models.Unlocked,
			}
		},
		bootstrapContainer.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return mockMetricsManager
		},
		container.AllowedRequestFailuresTrackerName: func(get di.Get) any {
			return container.NewAllowedFailuresTracker()
		},
	})

	return dic
}

func TestRestController_GetCommand(t *testing.T) {
	e := echo.New()
	dic := mockDic()

	edgexErr := cache.InitCache(testService, testService, dic)
	require.NoError(t, edgexErr)

	controller := NewRestController(e, dic, testService)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceName         string
		commandName        string
		expectedStatusCode int
	}{
		{"valid - read device resource", testDevice, testResource, http.StatusOK},
		{"valid - read device command", testDevice, testCommand, http.StatusOK},
		{"valid - read regex device resource", testDevice, testRegexResource, http.StatusOK},
		{"invalid - device name parameter is empty", "", testResource, http.StatusBadRequest},
		{"invalid - command is empty", testDevice, "", http.StatusBadRequest},
		{"invalid - device name not found", "notFound", testCommand, http.StatusNotFound},
		{"invalid - command name not found", testDevice, "notFound", http.StatusNotFound},
		{"invalid - device is LOCKED", lockedDevice, testResource, http.StatusLocked},
		{"invalid - device OperatingState is DOWN", downedDevice, testResource, http.StatusLocked},
		{"invalid - device resource is write-only", testDevice, writeOnlyResource, http.StatusMethodNotAllowed},
		{"invalid - device command is write-only", testDevice, writeOnlyCommand, http.StatusMethodNotAllowed},
		{"invalid - device command resource operations exceed MaxCmdOps", testDevice, exceedCommand, http.StatusInternalServerError},
		{"invalid - error in ProtocolDriver implementation", driverErrorDevice, testResource, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, common.ApiDeviceNameCommandNameRoute, http.NoBody)

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name, common.Command)
			c.SetParamValues(testCase.deviceName, testCase.commandName)

			err := controller.GetCommand(c)
			assert.NoError(t, err)

			var res responses.EventResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
			if testCase.expectedStatusCode == http.StatusOK {
				assert.Empty(t, res.Message, "Message should be empty when it is successful")
			} else {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}

func TestRestController_GetCommand_ServiceLocked(t *testing.T) {
	e := echo.New()
	dic := mockDic()
	dic.Update(di.ServiceConstructorMap{
		container.DeviceServiceName: func(get di.Get) any {
			return &models.DeviceService{
				Name:       testService,
				AdminState: models.Locked,
			}
		},
	})

	edgexErr := cache.InitCache(testService, testService, dic)
	require.NoError(t, edgexErr)

	controller := NewRestController(e, dic, testService)
	assert.NotNil(t, controller)

	req := httptest.NewRequest(http.MethodGet, common.ApiDeviceNameCommandNameRoute, http.NoBody)

	// Act
	recorder := httptest.NewRecorder()
	c := e.NewContext(req, recorder)
	c.SetParamNames(common.Name, common.Command)
	c.SetParamValues(testDevice, testResource)
	err := controller.GetCommand(c)
	assert.NoError(t, err)

	var res responses.EventResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusLocked, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, http.StatusLocked, res.StatusCode, "Response status code not as expected")
	assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
}

func TestRestController_GetCommand_ReturnEvent(t *testing.T) {
	e := echo.New()
	dic := mockDic()

	edgexErr := cache.InitCache(testService, testService, dic)
	require.NoError(t, edgexErr)

	controller := NewRestController(e, dic, testService)
	assert.NotNil(t, controller)

	req := httptest.NewRequest(http.MethodGet, common.ApiDeviceNameCommandNameRoute, http.NoBody)

	query := req.URL.Query()
	query.Add("ds-returnevent", common.ValueFalse)
	req.URL.RawQuery = query.Encode()
	// Act
	recorder := httptest.NewRecorder()
	c := e.NewContext(req, recorder)
	c.SetParamNames(common.Name, common.Command)
	c.SetParamValues(testDevice, testResource)
	err := controller.GetCommand(c)
	assert.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Empty(t, recorder.Body.Bytes())
}

func TestRestController_SetCommand(t *testing.T) {
	e := echo.New()
	validRequest := map[string]any{testResource: "value", writeOnlyResource: "value"}
	invalidRequest := map[string]any{"invalid": "test"}
	emptyValueRequest := map[string]any{objectResource: ""}

	dic := mockDic()

	sdkCommon.InitializeSentMetrics(logger.NewMockClient(), dic)
	err := cache.InitCache(testService, testService, dic)
	require.NoError(t, err)

	controller := NewRestController(e, dic, testService)
	assert.NotNil(t, controller)

	tests := []struct {
		name               string
		deviceName         string
		commandName        string
		request            map[string]any
		expectedStatusCode int
	}{
		{"valid - device resource", testDevice, testResource, validRequest, http.StatusOK},
		{"valid - write-only device resource", testDevice, writeOnlyResource, validRequest, http.StatusOK},
		{"valid - device resource not specified in request body but default value provided", testDevice, testResource, invalidRequest, http.StatusOK},
		{"valid - device command", testDevice, testCommand, validRequest, http.StatusOK},
		{"valid - write-only device command", testDevice, writeOnlyCommand, validRequest, http.StatusOK},
		{"valid - device command not specified in request body but default value provided", testDevice, testCommand, invalidRequest, http.StatusOK},
		{"invalid - device name parameter is empty", "", testResource, validRequest, http.StatusBadRequest},
		{"invalid - command is empty", testDevice, "", validRequest, http.StatusBadRequest},
		{"invalid - device name not found", "notFound", testResource, validRequest, http.StatusNotFound},
		{"invalid - command name not found", testDevice, "notFound", validRequest, http.StatusNotFound},
		{"invalid - device is LOCKED", lockedDevice, testResource, validRequest, http.StatusLocked},
		{"invalid - device OperatingState is DOWN", downedDevice, testResource, validRequest, http.StatusLocked},
		{"invalid - device resource is read-only", testDevice, readOnlyResource, validRequest, http.StatusMethodNotAllowed},
		{"invalid - device command is read-only", testDevice, readOnlyCommand, validRequest, http.StatusMethodNotAllowed},
		{"invalid - device command resource operations exceed MaxCmdOps", testDevice, exceedCommand, validRequest, http.StatusInternalServerError},
		{"invalid - write empty string to non string device resource", testDevice, objectResource, emptyValueRequest, http.StatusBadRequest},
		{"invalid - error in ProtocolDriver implementation", driverErrorDevice, testResource, validRequest, http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.request)
			require.NoError(t, err)

			reader := strings.NewReader(string(jsonData))
			req := httptest.NewRequest(http.MethodPut, common.ApiDeviceNameCommandNameRoute, reader)

			var wg sync.WaitGroup
			if testCase.commandName != writeOnlyCommand && testCase.commandName != writeOnlyResource {
				wg.Add(1)
			}
			messagingClientMock := &messagingMocks.MessageClient{}
			messagingClientMock.On("PublishWithSizeLimit", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				go func() {
					defer wg.Done()
				}()
			}).Return(nil)
			dic.Update(di.ServiceConstructorMap{
				bootstrapContainer.MessagingClientName: func(get di.Get) any {
					return messagingClientMock
				},
			})

			// Act
			recorder := httptest.NewRecorder()
			c := e.NewContext(req, recorder)
			c.SetParamNames(common.Name, common.Command)
			c.SetParamValues(testCase.deviceName, testCase.commandName)
			err = controller.SetCommand(c)
			assert.NoError(t, err)

			var res commonDTO.BaseResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			require.NoError(t, err)

			// Assert
			assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
			assert.Equal(t, testCase.expectedStatusCode, recorder.Result().StatusCode, "HTTP status code not as expected")
			assert.Equal(t, testCase.expectedStatusCode, res.StatusCode, "Response status code not as expected")
			if testCase.expectedStatusCode == http.StatusOK {
				assert.Empty(t, res.Message, "Message should be empty when it is successful")

				wg.Wait()
				if testCase.commandName != writeOnlyCommand && testCase.commandName != writeOnlyResource {
					messagingClientMock.AssertNumberOfCalls(t, "PublishWithSizeLimit", 1)
				}
			} else {
				assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
			}
		})
	}
}

func TestRestController_SetCommand_ServiceLocked(t *testing.T) {
	e := echo.New()
	dic := mockDic()
	dic.Update(di.ServiceConstructorMap{
		container.DeviceServiceName: func(get di.Get) any {
			return &models.DeviceService{
				Name:       testService,
				AdminState: models.Locked,
			}
		},
	})
	validRequest := map[string]any{testResource: "value"}

	edgexErr := cache.InitCache(testService, testService, dic)
	require.NoError(t, edgexErr)

	controller := NewRestController(e, dic, testService)
	assert.NotNil(t, controller)

	jsonData, err := json.Marshal(validRequest)
	require.NoError(t, err)

	reader := strings.NewReader(string(jsonData))
	req := httptest.NewRequest(http.MethodPut, common.ApiDeviceNameCommandNameRoute, reader)

	// Act
	recorder := httptest.NewRecorder()
	c := e.NewContext(req, recorder)
	c.SetParamNames(common.Name, common.Command)
	c.SetParamValues(testDevice, testResource)
	err = controller.SetCommand(c)
	assert.NoError(t, err)

	var res commonDTO.BaseResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, common.ApiVersion, res.ApiVersion, "API Version not as expected")
	assert.Equal(t, http.StatusLocked, recorder.Result().StatusCode, "HTTP status code not as expected")
	assert.Equal(t, http.StatusLocked, res.StatusCode, "Response status code not as expected")
	assert.NotEmpty(t, res.Message, "Response message doesn't contain the error message")
}
