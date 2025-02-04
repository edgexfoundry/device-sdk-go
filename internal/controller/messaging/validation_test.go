//
// Copyright (C) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	messagingMocks "github.com/edgexfoundry/go-mod-messaging/v4/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces/mocks"
)

const (
	testDeviceName   = "testDevice"
	testServiceName  = "testService"
	testProfileName  = "testProfile"
	testProtocolName = "testProtocol"
)

func TestDeviceValidation(t *testing.T) {
	var wg sync.WaitGroup
	expectedRequestId := uuid.NewString()
	expectedCorrelationId := uuid.NewString()
	expectedRequestTopic := common.BuildTopic(common.DefaultBaseTopic, testServiceName, common.ValidateDeviceSubscribeTopic)
	expectedResponseTopic := common.BuildTopic(common.DefaultBaseTopic, common.ResponseTopic, testServiceName, expectedRequestId)
	expectedDevice := dtos.Device{
		Name:           testDeviceName,
		AdminState:     models.Locked,
		OperatingState: models.Up,
		ServiceName:    testServiceName,
		ProfileName:    testProfileName,
		Protocols: map[string]dtos.ProtocolProperties{
			testProtocolName: {"key": "value"},
		},
	}
	expectedAddDeviceRequest := requests.NewAddDeviceRequest(expectedDevice)
	expectedAddDeviceRequestBytes, err := json.Marshal(expectedAddDeviceRequest)
	require.NoError(t, err)
	validationFailedDevice := expectedDevice
	validationFailedDevice.Name = "validationFailedDevice"
	validationFailedDeviceRequest := requests.NewAddDeviceRequest(validationFailedDevice)
	validationFailedDeviceRequestBytes, err := json.Marshal(validationFailedDeviceRequest)
	require.NoError(t, err)

	mockLogger := &loggerMocks.LoggingClient{}
	mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Errorf", mock.Anything, mock.Anything).Return(nil)

	mockDriver := &mocks.ProtocolDriver{}
	mockDriver.On("ValidateDevice", dtos.ToDeviceModel(expectedDevice)).Return(nil)
	mockDriver.On("ValidateDevice", dtos.ToDeviceModel(validationFailedDevice)).Return(errors.New("validation failed"))

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) any {
			return &config.ConfigurationStruct{}
		},
		container.ProtocolDriverName: func(get di.Get) any {
			return mockDriver
		},
		container.DeviceServiceName: func(get di.Get) any {
			return &models.DeviceService{Name: testServiceName}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) any {
			return mockLogger
		},
	})

	tests := []struct {
		name          string
		request       any
		expectedError bool
		withEnv       bool
	}{
		{"valid with env - device validation succeed", expectedAddDeviceRequestBytes, false, true},
		{"valid with env - device validation failed", validationFailedDeviceRequestBytes, true, true},
		{"invalid with env - message payload is not AddDeviceRequest", []byte("invalid"), true, true},
		{"valid - device validation succeed", expectedAddDeviceRequest, false, false},
		{"valid - device validation failed", validationFailedDeviceRequest, true, false},
		{"invalid - message payload is not AddDeviceRequest", []byte("invalid"), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.withEnv {
				_ = os.Setenv("EDGEX_MSG_BASE64_PAYLOAD", common.ValueTrue)
				defer os.Setenv("EDGEX_MSG_BASE64_PAYLOAD", common.ValueFalse)
			}
			mockMessaging := &messagingMocks.MessageClient{}
			mockMessaging.On("Subscribe", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				topics := args.Get(0).([]types.TopicChannel)
				require.Len(t, topics, 1)
				require.Equal(t, expectedRequestTopic, topics[0].Topic)
				wg.Add(1)
				go func() {
					defer wg.Done()
					topics[0].Messages <- types.MessageEnvelope{
						RequestID:     expectedRequestId,
						CorrelationID: expectedCorrelationId,
						ReceivedTopic: expectedRequestTopic,
						ContentType:   common.ContentTypeJSON,
						Payload:       tt.request,
					}
					time.Sleep(time.Second * 1)
				}()
			}).Return(nil)
			mockMessaging.On("Publish", mock.Anything, expectedResponseTopic).Run(func(args mock.Arguments) {
				response := args.Get(0).(types.MessageEnvelope)
				assert.Equal(t, expectedRequestId, response.RequestID)
				if tt.expectedError {
					assert.Equal(t, 1, response.ErrorCode)
					assert.NotEmpty(t, response.Payload)
					assert.Equal(t, common.ContentTypeText, response.ContentType)
				} else {
					assert.Equal(t, expectedCorrelationId, response.CorrelationID)
					assert.Equal(t, 0, response.ErrorCode)
					assert.Empty(t, response.Payload)
					assert.Equal(t, common.ContentTypeJSON, response.ContentType)
				}
			}).Return(nil)

			dic.Update(di.ServiceConstructorMap{
				bootstrapContainer.MessagingClientName: func(get di.Get) any {
					return mockMessaging
				},
			})
			err := SubscribeDeviceValidation(context.Background(), dic)
			require.NoError(t, err)

			wg.Wait()
			mockMessaging.AssertExpectations(t)
		})
	}
}
