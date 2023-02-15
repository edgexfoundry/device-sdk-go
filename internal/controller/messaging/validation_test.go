//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	loggerMocks "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	messagingMocks "github.com/edgexfoundry/go-mod-messaging/v3/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/models/mocks"
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
	validationFailedDevice := expectedDevice
	validationFailedDevice.Name = "validationFailedDevice"
	expectedAddDeviceRequestBytes, err := json.Marshal(requests.NewAddDeviceRequest(expectedDevice))
	require.NoError(t, err)
	validationFailedDeviceBytes, err := json.Marshal(requests.NewAddDeviceRequest(validationFailedDevice))
	require.NoError(t, err)

	mockLogger := &loggerMocks.LoggingClient{}
	mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Errorf", mock.Anything, mock.Anything).Return(nil)

	mockValidator := &mocks.DeviceValidator{}
	mockValidator.On("ValidateDevice", dtos.ToDeviceModel(expectedDevice)).Return(nil)
	mockValidator.On("ValidateDevice", dtos.ToDeviceModel(validationFailedDevice)).Return(errors.New("validation failed"))

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) any {
			return &config.ConfigurationStruct{}
		},
		container.DeviceValidatorName: func(get di.Get) any {
			return mockValidator
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
		requestBytes  []byte
		expectedError bool
	}{
		{"valid - device validation succeed", expectedAddDeviceRequestBytes, false},
		{"valid - device validation failed", validationFailedDeviceBytes, true},
		{"invalid - message payload is not AddDeviceRequest", []byte("invalid"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
						Payload:       tt.requestBytes,
					}
					time.Sleep(time.Second * 1)
				}()
			}).Return(nil)
			mockMessaging.On("Publish", mock.Anything, expectedResponseTopic).Run(func(args mock.Arguments) {
				response := args.Get(0).(types.MessageEnvelope)
				assert.Equal(t, expectedRequestId, response.RequestID)
				if tt.expectedError {
					assert.Equal(t, response.ErrorCode, 1)
					assert.NotEmpty(t, response.Payload)
					assert.Equal(t, response.ContentType, types.ContentTypeText)
				} else {
					assert.Equal(t, expectedCorrelationId, response.CorrelationID)
					assert.Equal(t, response.ErrorCode, 0)
					assert.Empty(t, response.Payload)
					assert.Equal(t, response.ContentType, types.ContentTypeJSON)
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
