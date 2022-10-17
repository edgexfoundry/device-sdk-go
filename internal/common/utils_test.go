//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	mocks2 "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	msgMocks "github.com/edgexfoundry/go-mod-messaging/v2/messaging/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

const (
	TestDevice         = "testDevice"
	TestProfile        = "testProfile"
	TestDeviceResource = "testResource"
	TestDeviceCommand  = "testCommand"
	testUUIDString     = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"
)

var TestProtocols map[string]dtos.ProtocolProperties

func buildEvent() dtos.Event {
	event := dtos.NewEvent(TestProfile, TestDevice, TestDeviceCommand)
	value := string(make([]byte, 1000))
	_ = event.AddSimpleReading(TestDeviceResource, common.ValueTypeString, value)
	event.Id = testUUIDString
	event.Readings[0].Id = testUUIDString
	return event
}

func NewMockDIC() *di.Container {
	configuration := &config.ConfigurationStruct{
		Device: config.DeviceInfo{MaxCmdOps: 1},
	}

	return di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})
}

func TestSendEvent(t *testing.T) {

	event := buildEvent()
	req := requests.NewAddEventRequest(event)
	bytes, _, err := req.Encode()
	require.NoError(t, err)
	eventSize := len(bytes) / 1024 // to kilobyte

	tests := []struct {
		name          string
		event         *dtos.Event
		maxEventSize  int64
		useMessageBus bool
		eventTooLarge bool
	}{
		{"Valid, unlimited max event size", &event, 0, false, false},
		{"Valid, publish to message bus", &event, int64(eventSize + 1), true, false},
		{"Valid, push to core data ", &event, int64(eventSize + 1), false, false},
		{"Invalid, over max event size", &event, int64(eventSize - 1), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dic := NewMockDIC()
			ecMock := &clientMocks.EventClient{}
			ecMock.On("Add", mock.Anything, mock.Anything).Return(dtoCommon.BaseWithIdResponse{}, nil)
			mcMock := &msgMocks.MessageClient{}
			mcMock.On("Publish", mock.Anything, mock.Anything).Return(nil)

			dic.Update(di.ServiceConstructorMap{
				container.ConfigurationName: func(get di.Get) interface{} {
					return &config.ConfigurationStruct{
						MaxEventSize: tt.maxEventSize,
						Device: config.DeviceInfo{
							UseMessageBus: tt.useMessageBus,
						},
					}
				},
				bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
					return mcMock
				},
				bootstrapContainer.EventClientName: func(get di.Get) interface{} {
					return ecMock
				},
			})

			SendEvent(tt.event, testUUIDString, dic)
			if tt.eventTooLarge {
				ecMock.AssertNumberOfCalls(t, "Add", 0)
				mcMock.AssertNumberOfCalls(t, "Publish", 0)
			} else if tt.useMessageBus {
				ecMock.AssertNumberOfCalls(t, "Add", 0)
				mcMock.AssertNumberOfCalls(t, "Publish", 1)
			} else {
				ecMock.AssertNumberOfCalls(t, "Add", 1)
				mcMock.AssertNumberOfCalls(t, "Publish", 0)
			}
		})
	}
}

func Test_HandleEventReadingMetrics(t *testing.T) {
	event1Reading := dtos.NewEvent("Test", "Test", "Test")
	err := event1Reading.AddSimpleReading("Test", common.ValueTypeInt32, int32(123))
	require.NoError(t, err)
	event3Readings := dtos.NewEvent("Test", "Test", "Test")
	err = event3Readings.AddSimpleReading("Test1", common.ValueTypeInt32, int32(123))
	require.NoError(t, err)
	err = event3Readings.AddSimpleReading("Test2", common.ValueTypeInt32, int32(123))
	require.NoError(t, err)
	err = event3Readings.AddSimpleReading("Test3", common.ValueTypeInt32, int32(123))
	require.NoError(t, err)

	tests := []struct {
		Name                    string
		TimesToCall             int
		Event                   dtos.Event
		MetricsManagerAvailable bool
		RegisterError           bool
		ExpectedEventCount      int64
		ExpectedReadingCount    int64
	}{
		{"First Pass", 1, event1Reading, true, false, 1, 1},
		{"3 Passes with 1 reading", 3, event1Reading, true, false, 3, 3},
		{"3 Passes with 3 readings", 3, event3Readings, true, false, 3, 9},
		{"No MetricsManager", 1, event1Reading, false, false, 1, 1},
		{"Error Registering", 1, event1Reading, true, true, 1, 1},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			eventsSent = nil
			readingsSent = nil

			mockManager := &mocks.MetricsManager{}
			mockLogger := &mocks2.LoggingClient{}

			if test.RegisterError {
				mockManager.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed"))
			} else {
				mockManager.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)

			}

			mockLogger.On("Warn", mock.Anything).Return()
			mockLogger.On("Debugf", mock.Anything, mock.Anything).Return()
			mockLogger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return()

			dic := di.NewContainer(di.ServiceConstructorMap{
				bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
					return mockLogger
				},
				bootstrapContainer.MetricsManagerInterfaceName: func(get di.Get) interface{} {
					if test.MetricsManagerAvailable {
						return mockManager
					}

					return nil
				},
			})

			n := 0
			for n < test.TimesToCall {
				handleEventReadingMetrics(&test.Event, mockLogger, dic)
				n++
			}

			if test.MetricsManagerAvailable {
				mockManager.AssertNumberOfCalls(t, "Register", 2)
				mockLogger.AssertNumberOfCalls(t, "Warn", 0)

				if test.RegisterError {
					mockLogger.AssertNumberOfCalls(t, "Errorf", 2)
					mockLogger.AssertNumberOfCalls(t, "Debugf", 0)

				} else {
					mockLogger.AssertNumberOfCalls(t, "Errorf", 0)
					mockLogger.AssertNumberOfCalls(t, "Debugf", 2)
				}

			} else {
				mockManager.AssertNumberOfCalls(t, "Register", 0)
				mockLogger.AssertNumberOfCalls(t, "Warn", 1)
				mockLogger.AssertNumberOfCalls(t, "Debugf", 0)
			}

			require.NotNil(t, eventsSent)
			require.NotNil(t, readingsSent)

			assert.Equal(t, test.ExpectedEventCount, eventsSent.Count())
			assert.Equal(t, test.ExpectedReadingCount, readingsSent.Count())
		})
	}
}
