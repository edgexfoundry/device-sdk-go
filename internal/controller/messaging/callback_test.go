//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"testing"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	clientMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
)

const (
	testSysEvtBaseService = "testSysEvtService"
	testSysEvtWatcher     = "testSysEvtWatcher"
	testSysEvtProfile     = "testSysEvtProfile"
)

func TestMetadataSystemEventsCallback_provisionWatcherUpdateWithInstance(t *testing.T) {
	instanceName := testSysEvtBaseService + "_1"
	watcher := dtos.ProvisionWatcher{
		Name: testSysEvtWatcher, ServiceName: testSysEvtBaseService, AdminState: models.Unlocked,
		DiscoveredDevice: dtos.DiscoveredDevice{ProfileName: testSysEvtProfile},
	}

	dcMock := &clientMocks.DeviceClient{}
	dcMock.On("DevicesByServiceName", context.Background(), instanceName, 0, -1).
		Return(responses.MultiDevicesResponse{}, nil)
	dpcMock := &clientMocks.DeviceProfileClient{}
	dpcMock.On("DeviceProfileByName", context.Background(), testSysEvtProfile).
		Return(responses.DeviceProfileResponse{Profile: dtos.DeviceProfile{
			DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{Name: testSysEvtProfile},
		}}, nil)
	pwcMock := &clientMocks.ProvisionWatcherClient{}
	pwcMock.On("ProvisionWatchersByServiceName", context.Background(), testSysEvtBaseService, 0, -1).
		Return(responses.MultiProvisionWatchersResponse{ProvisionWatchers: []dtos.ProvisionWatcher{watcher}}, nil)

	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(_ di.Get) any { return &config.ConfigurationStruct{} },
		container.DeviceServiceName: func(_ di.Get) any {
			return &models.DeviceService{Name: instanceName}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(_ di.Get) any { return logger.NewMockClient() },
		bootstrapContainer.DeviceClientName:           func(_ di.Get) any { return dcMock },
		bootstrapContainer.DeviceProfileClientName:    func(_ di.Get) any { return dpcMock },
		bootstrapContainer.ProvisionWatcherClientName: func(_ di.Get) any { return pwcMock },
	})
	require.NoError(t, cache.InitCache(instanceName, testSysEvtBaseService, dic))
	probe := installMessagingMock(t, dic)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	require.NoError(t, MetadataSystemEventsCallback(ctx, testSysEvtBaseService, dic))

	watcher.AdminState = models.Locked
	systemEvent := dtos.NewSystemEvent(common.ProvisionWatcherSystemEventType, common.SystemEventActionUpdate,
		common.CoreMetaDataServiceKey, testSysEvtBaseService, nil, watcher)
	probe.sub() <- types.MessageEnvelope{Payload: systemEvent}

	require.Eventually(t, func() bool {
		pw, ok := cache.ProvisionWatchers().ForName(testSysEvtWatcher)
		return ok && pw.AdminState == models.Locked
	}, 3*time.Second, 5*time.Millisecond, "provision watcher should stay cached with the updated AdminState")
}
