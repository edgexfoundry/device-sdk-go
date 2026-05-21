//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapMocks "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	messagingMocks "github.com/edgexfoundry/go-mod-messaging/v4/messaging/mocks"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	driverMocks "github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces/mocks"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

const (
	testCmdServiceName  = "testCmdService"
	testCmdDeviceA      = "deviceA"
	testCmdDeviceB      = "deviceB"
	testCmdProfile      = "testCmdProfile"
	testCmdResourceName = "tempSensor"
)

// newCommandTestDIC builds a DIC and initializes the device/profile cache so that
// application.GetCommand can run through to driver.HandleReadCommands.
func newCommandTestDIC(t *testing.T, maxConcurrent int, driver *driverMocks.ProtocolDriver) *di.Container {
	t.Helper()

	cfg := &config.ConfigurationStruct{
		MaxConcurrentCommands: maxConcurrent,
		Device:                config.DeviceInfo{MaxCmdOps: 1},
	}
	deviceService := &models.DeviceService{Name: testCmdServiceName, AdminState: models.Unlocked}

	devices := responses.MultiDevicesResponse{
		Devices: []dtos.Device{
			{
				Name: testCmdDeviceA, ProfileName: testCmdProfile, ServiceName: testCmdServiceName,
				AdminState: models.Unlocked, OperatingState: models.Up,
			},
			{
				Name: testCmdDeviceB, ProfileName: testCmdProfile, ServiceName: testCmdServiceName,
				AdminState: models.Unlocked, OperatingState: models.Up,
			},
		},
	}
	profile := responses.DeviceProfileResponse{
		Profile: dtos.DeviceProfile{
			DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{Name: testCmdProfile},
			DeviceResources: []dtos.DeviceResource{
				{
					Name: testCmdResourceName,
					Properties: dtos.ResourceProperties{
						ValueType: common.ValueTypeString,
						ReadWrite: common.ReadWrite_RW,
					},
				},
			},
		},
	}

	dcMock := &clientMocks.DeviceClient{}
	dcMock.On("DevicesByServiceName", context.Background(), testCmdServiceName, 0, -1).Return(devices, nil)

	dpcMock := &clientMocks.DeviceProfileClient{}
	dpcMock.On("DeviceProfileByName", context.Background(), testCmdProfile).Return(profile, nil)

	pwcMock := &clientMocks.ProvisionWatcherClient{}
	pwcMock.On("ProvisionWatchersByServiceName", context.Background(), testCmdServiceName, 0, -1).
		Return(responses.MultiProvisionWatchersResponse{}, nil)

	mm := &bootstrapMocks.MetricsManager{}
	mm.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mm.On("Unregister", mock.Anything)

	failsTracker := container.NewAllowedFailuresTracker()

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName:                    func(_ di.Get) any { return cfg },
		container.DeviceServiceName:                    func(_ di.Get) any { return deviceService },
		container.ProtocolDriverName:                   func(_ di.Get) any { return driver },
		container.AllowedRequestFailuresTrackerName:    func(_ di.Get) any { return failsTracker },
		bootstrapContainer.LoggingClientInterfaceName:  func(_ di.Get) any { return logger.NewMockClient() },
		bootstrapContainer.DeviceClientName:            func(_ di.Get) any { return dcMock },
		bootstrapContainer.DeviceProfileClientName:     func(_ di.Get) any { return dpcMock },
		bootstrapContainer.ProvisionWatcherClientName:  func(_ di.Get) any { return pwcMock },
		bootstrapContainer.MetricsManagerInterfaceName: func(_ di.Get) any { return mm },
	})

	require.NoError(t, cache.InitCache(testCmdServiceName, testCmdServiceName, dic))
	return dic
}

func makeGetEnvelope(deviceName string) types.MessageEnvelope {
	return types.MessageEnvelope{
		RequestID:     uuid.NewString(),
		CorrelationID: uuid.NewString(),
		ContentType:   common.ContentTypeJSON,
		ReceivedTopic: "edgex/command/request/" + testCmdServiceName + "/" + deviceName + "/" + testCmdResourceName + "/get",
		// ds-returnevent=false avoids the test needing to produce a real *dtos.Event from the
		// (empty) driver response — we're testing dispatch concurrency, not event marshaling.
		QueryParams: map[string]string{common.ReturnEvent: common.ValueFalse},
	}
}

// messagingProbe captures the subscriber channel and counts publishes so tests can deterministically
// wait for worker goroutines to finish (avoiding cross-test races on package-global caches).
type messagingProbe struct {
	subReady   sync.WaitGroup
	subCh      chan types.MessageEnvelope
	totalCount atomic.Int32 // every Publish + PublishWithSizeLimit call
	busyCount  atomic.Int32 // ErrorCode == 1 publishes only (overload responses)
}

func installMessagingMock(t *testing.T, dic *di.Container) *messagingProbe {
	t.Helper()
	p := &messagingProbe{}
	p.subReady.Add(1)

	mockMessaging := &messagingMocks.MessageClient{}
	mockMessaging.On("Subscribe", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			topics := args.Get(0).([]types.TopicChannel)
			p.subCh = topics[0].Messages
			p.subReady.Done()
		}).Return(nil)
	mockMessaging.On("Publish", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			env := args.Get(0).(types.MessageEnvelope)
			if env.ErrorCode == 1 {
				p.busyCount.Add(1)
			}
			p.totalCount.Add(1)
		}).Return(nil)
	mockMessaging.On("PublishWithSizeLimit", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { p.totalCount.Add(1) }).
		Return(nil)

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.MessagingClientName: func(_ di.Get) any { return mockMessaging },
	})
	return p
}

func (p *messagingProbe) sub() chan types.MessageEnvelope {
	p.subReady.Wait()
	return p.subCh
}

// waitForPublishes blocks until totalCount >= n. Used at end-of-test to make sure all in-flight
// worker goroutines have finished before the test returns, so they don't race with the next
// test's cache.InitCache.
func (p *messagingProbe) waitForPublishes(t *testing.T, n int32) {
	t.Helper()
	require.Eventually(t, func() bool { return p.totalCount.Load() >= n },
		3*time.Second, 5*time.Millisecond, "expected %d publishes, got %d", n, p.totalCount.Load())
}

// Test_SubscribeCommands_secondRunsWhileFirstBlocked proves the fix: a command to device B can
// begin processing while a command to device A is still blocked inside the driver. With the
// pre-fix code this would deadlock — the consumer goroutine would be stuck dispatching the first
// command synchronously.
func Test_SubscribeCommands_secondRunsWhileFirstBlocked(t *testing.T) {
	unblockFirst := make(chan struct{})
	var phase atomic.Int32

	driver := &driverMocks.ProtocolDriver{}
	driver.On("HandleReadCommands", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) {
			if phase.Add(1) == 1 {
				<-unblockFirst
			}
		}).Return([]*sdkModels.CommandValue{}, nil)

	dic := newCommandTestDIC(t, 4, driver)
	probe := installMessagingMock(t, dic)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, SubscribeCommands(ctx, dic))

	sub := probe.sub()
	sub <- makeGetEnvelope(testCmdDeviceA)
	require.Eventually(t, func() bool { return phase.Load() >= 1 },
		2*time.Second, 5*time.Millisecond, "first driver call should have started")

	sub <- makeGetEnvelope(testCmdDeviceB)
	require.Eventually(t, func() bool { return phase.Load() >= 2 },
		2*time.Second, 5*time.Millisecond,
		"second driver call should run while first is blocked — HOL regression if it doesn't")

	close(unblockFirst)
	// Drain both worker goroutines before returning so they don't race with the next test's
	// cache.InitCache (the device cache is a package-global singleton).
	probe.waitForPublishes(t, 2)
}

// Test_SubscribeCommands_busyOnOverload proves the backpressure path: when the in-flight budget
// is exhausted, the next request is rejected inline with a busy envelope and the driver is NOT
// invoked again.
func Test_SubscribeCommands_busyOnOverload(t *testing.T) {
	blockAll := make(chan struct{})
	var entered atomic.Int32

	driver := &driverMocks.ProtocolDriver{}
	driver.On("HandleReadCommands", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) {
			entered.Add(1)
			<-blockAll
		}).Return([]*sdkModels.CommandValue{}, nil)

	dic := newCommandTestDIC(t, 1, driver) // capacity 1 forces the second message to overload
	probe := installMessagingMock(t, dic)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, SubscribeCommands(ctx, dic))

	sub := probe.sub()
	sub <- makeGetEnvelope(testCmdDeviceA)
	require.Eventually(t, func() bool { return entered.Load() == 1 },
		2*time.Second, 5*time.Millisecond, "first driver call should have entered")

	sub <- makeGetEnvelope(testCmdDeviceB)
	require.Eventually(t, func() bool { return probe.busyCount.Load() >= 1 },
		2*time.Second, 10*time.Millisecond, "busy envelope should be published for the rejected request")

	driver.AssertNumberOfCalls(t, "HandleReadCommands", 1)
	close(blockAll)
	// Drain the first worker (1 success publish) + already-counted busy publish = 2 total.
	probe.waitForPublishes(t, 2)
}
