//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	msgMocks "github.com/edgexfoundry/go-mod-messaging/v2/messaging/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
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
				container.MessagingClientName: func(get di.Get) interface{} {
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
