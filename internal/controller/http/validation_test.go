//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models/mocks"
)

var mockDevice = requests.AddDeviceRequest{
	BaseRequest: commonDTO.NewBaseRequest(),
	Device: dtos.Device{
		Name:           "test-device",
		Description:    "",
		AdminState:     models.Unlocked,
		OperatingState: models.Up,
		ServiceName:    "test-service",
		ProfileName:    "test-profile",
		Protocols:      nil,
	},
}

var validProtocols = map[string]models.ProtocolProperties{"valid": {}}
var invalidProtocols = map[string]models.ProtocolProperties{"invalid": {}}

func TestRestController_Validate(t *testing.T) {
	validDeviceRequest := mockDevice
	validDeviceRequest.Device.Protocols = dtos.FromProtocolModelsToDTOs(validProtocols)
	invalidDeviceRequest := mockDevice
	invalidDeviceRequest.Device.Protocols = dtos.FromProtocolModelsToDTOs(invalidProtocols)

	validatorMock := &mocks.DeviceValidator{}
	validatorMock.On("ValidateDevice", dtos.ToDeviceModel(validDeviceRequest.Device)).Return(nil)
	validatorMock.On("ValidateDevice", dtos.ToDeviceModel(invalidDeviceRequest.Device)).Return(errors.New("invalid"))

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.DeviceValidatorName: func(get di.Get) interface{} {
			return validatorMock
		},
	})

	tests := []struct {
		name               string
		deviceRequest      interface{}
		expectedStatusCode int
	}{
		{"Valid - validation succeed", validDeviceRequest, http.StatusOK},
		{"Invalid - validation failed", invalidDeviceRequest, http.StatusInternalServerError},
		{"Invalid - bad request body", "invalid", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.deviceRequest)
			require.NoError(t, err)
			reader := strings.NewReader(string(jsonData))

			req, err := http.NewRequest(http.MethodPost, common.ApiDeviceValidationRoute, reader)
			require.NoError(t, err)

			controller := NewRestController(mux.NewRouter(), dic, uuid.NewString())
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(controller.ValidateDevice)
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatusCode, recorder.Result().StatusCode, "Wrong status code")
		})
	}
}

func TestRestController_Validate_Not_Implemented(t *testing.T) {
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.DeviceValidatorName: func(get di.Get) interface{} {
			return nil
		},
	})

	validDevice := mockDevice
	validDevice.Device.Protocols = dtos.FromProtocolModelsToDTOs(validProtocols)

	jsonData, err := json.Marshal(validProtocols)
	require.NoError(t, err)
	reader := strings.NewReader(string(jsonData))

	req, err := http.NewRequest(http.MethodPost, common.ApiDeviceValidationRoute, reader)
	require.NoError(t, err)

	controller := NewRestController(mux.NewRouter(), dic, uuid.NewString())
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.ValidateDevice)
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode, "Wrong status code")
}
