// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2020 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/urlclient/local"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
)

// InitDependencyClients triggers Service Client Initializer to establish connection to Metadata and Core Data Services
// through Metadata Client and Core Data Client.
// Service Client Initializer also needs to check the service status of Metadata and Core Data Services,
// because they are important dependencies of Device Service.
// The initialization process should be pending until Metadata Service and Core Data Service are both available.
func InitDependencyClients(ctx context.Context, waitGroup *sync.WaitGroup, startupTimer startup.Timer) bool {
	if err := validateClientConfig(); err != nil {
		common.LoggingClient.Error(err.Error())
		return false
	}

	if checkDependencyServices(ctx, startupTimer) == false {
		return false
	}

	initializeClients(ctx, waitGroup)

	common.LoggingClient.Info("Service clients initialize successful.")
	return true
}

func validateClientConfig() error {

	if len(common.CurrentConfig.Clients[common.ClientMetadata].Host) == 0 {
		return fmt.Errorf("fatal error; Host setting for Core Metadata client not configured")
	}

	if common.CurrentConfig.Clients[common.ClientMetadata].Port == 0 {
		return fmt.Errorf("fatal error; Port setting for Core Metadata client not configured")
	}

	if len(common.CurrentConfig.Clients[common.ClientData].Host) == 0 {
		return fmt.Errorf("fatal error; Host setting for Core Data client not configured")
	}

	if common.CurrentConfig.Clients[common.ClientData].Port == 0 {
		return fmt.Errorf("fatal error; Port setting for Core Ddata client not configured")
	}

	// TODO: validate other settings for sanity: maxcmdops, ...

	return nil
}

func checkDependencyServices(ctx context.Context, startupTimer startup.Timer) bool {
	var dependencyList = []string{common.ClientData, common.ClientMetadata}
	var waitGroup sync.WaitGroup
	checkingErr := true

	dependencyCount := len(dependencyList)
	waitGroup.Add(dependencyCount)

	for i := 0; i < dependencyCount; i++ {
		go func(wg *sync.WaitGroup, serviceName string) {
			defer wg.Done()
			if checkServiceAvailable(ctx, serviceName, startupTimer) == false {
				checkingErr = false
			}
		}(&waitGroup, dependencyList[i])
	}
	waitGroup.Wait()

	return checkingErr
}

func checkServiceAvailable(ctx context.Context, serviceId string, startupTimer startup.Timer) bool {
	for startupTimer.HasNotElapsed() {
		select {
		case <-ctx.Done():
			return false
		default:
			if common.RegistryClient != nil {
				if checkServiceAvailableViaRegistry(serviceId) == nil {
					return true
				}
			} else {
				if checkServiceAvailableByPing(serviceId) == nil {
					return true
				}
			}
			startupTimer.SleepForInterval()
		}
	}

	common.LoggingClient.Error(fmt.Sprintf("dependency %s service checking time out", serviceId))
	return false
}

func checkServiceAvailableByPing(serviceId string) error {
	common.LoggingClient.Info(fmt.Sprintf("Check %v service's status by ping...", serviceId))
	addr := common.CurrentConfig.Clients[serviceId].Url()
	timeout := int64(common.CurrentConfig.Service.BootTimeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + clients.ApiPingRoute)
	if err != nil {
		common.LoggingClient.Error(err.Error())
	}

	return err
}

func checkServiceAvailableViaRegistry(serviceId string) error {
	common.LoggingClient.Info(fmt.Sprintf("Check %s service's status via Registry...", serviceId))

	if !common.RegistryClient.IsAlive() {
		errMsg := fmt.Sprintf("unable to check status of %s service: Registry not running", serviceId)
		common.LoggingClient.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	if serviceId == common.ClientData {
		serviceId = clients.CoreDataServiceKey
	} else {
		serviceId = clients.CoreMetaDataServiceKey
	}
	_, err := common.RegistryClient.IsServiceAvailable(serviceId)
	if err != nil {
		common.LoggingClient.Error(err.Error())
		return err
	}

	return nil
}

func initializeClients(ctx context.Context, waitGroup *sync.WaitGroup) {
	// initialize Core Metadata clients
	common.AddressableClient = metadata.NewAddressableClient(local.New(common.CurrentConfig.Clients[common.ClientMetadata].Url() + clients.ApiAddressableRoute))
	common.DeviceClient = metadata.NewDeviceClient(local.New(common.CurrentConfig.Clients[common.ClientMetadata].Url() + clients.ApiDeviceRoute))
	common.DeviceServiceClient = metadata.NewDeviceServiceClient(local.New(common.CurrentConfig.Clients[common.ClientMetadata].Url() + clients.ApiDeviceServiceRoute))
	common.DeviceProfileClient = metadata.NewDeviceProfileClient(local.New(common.CurrentConfig.Clients[common.ClientMetadata].Url() + clients.ApiDeviceProfileRoute))
	common.MetadataGeneralClient = general.NewGeneralClient(local.New(common.CurrentConfig.Clients[common.ClientMetadata].Url()))
	common.ProvisionWatcherClient = metadata.NewProvisionWatcherClient(local.New(common.CurrentConfig.Clients[common.ClientMetadata].Url() + clients.ApiProvisionWatcherRoute))

	// initialize Core Data clients
	common.EventClient = coredata.NewEventClient(local.New(common.CurrentConfig.Clients[common.ClientData].Url() + clients.ApiEventRoute))
	common.ValueDescriptorClient = coredata.NewValueDescriptorClient(local.New(common.CurrentConfig.Clients[common.ClientData].Url() + common.APIValueDescriptorRoute))
}
