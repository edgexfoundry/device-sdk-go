// TODO: Change Copyright to your company if open sourcing or remove header
//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces/mocks"
)

// This is an example of how to test the code that would typically be in the main() function use mocks
// Not to helpful for a simple main() , but can be if the main() has more complexity that should be unit tested
// TODO: add/update tests for your customized CreateAndRunService or remove for simple main()

func TestCreateAndRunService_Success(t *testing.T) {
	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return([]string{"Random-Boolean-Device, Random-Integer-Device"}, nil)
		mockAppService.On("SetFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("MakeItRun").Return(nil)

		return mockAppService, true
	}

	expected := 0
	actual := CreateAndRunService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_NewService_Failed(t *testing.T) {
	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		return nil, false
	}
	expected := -1
	actual := CreateAndRunService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_GetAppSettingStrings_Failed(t *testing.T) {
	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return(nil, fmt.Errorf("Failed"))

		return mockAppService, true
	}

	expected := -1
	actual := CreateAndRunService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_SetFunctionsPipeline_Failed(t *testing.T) {
	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return([]string{"Random-Boolean-Device, Random-Integer-Device"}, nil)
		mockAppService.On("SetFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(fmt.Errorf("Failed"))

		return mockAppService, true
	}

	expected := -1
	actual := CreateAndRunService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_MakeItRun_Failed(t *testing.T) {
	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return([]string{"Random-Boolean-Device, Random-Integer-Device"}, nil)
		mockAppService.On("SetFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("MakeItRun").Return(fmt.Errorf("Failed"))

		return mockAppService, true
	}

	expected := -1
	actual := CreateAndRunService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}
