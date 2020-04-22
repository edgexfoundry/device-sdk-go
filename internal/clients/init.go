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

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/urlclient"
)

// InitDependencyClients triggers Service Client Initializer to establish connection to Metadata and Core Data Services
// through Metadata Client and Core Data Client.
// Service Client Initializer also needs to check the service status of Metadata and Core Data Services,
// because they are important dependencies of Device Service.
// The initialization process should be pending until Metadata Service and Core Data Service are both available.
func InitDependencyClients(ctx context.Context, waitGroup *sync.WaitGroup) error {
	if err := validateClientConfig(); err != nil {
		return err
	}

	if err := checkDependencyServices(); err != nil {
		return err
	}

	initializeClients(ctx, waitGroup)

	common.LoggingClient.Info("Service clients initialize successful.")
	return nil
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

func checkDependencyServices() error {
	var dependencyList = []string{common.ClientData, common.ClientMetadata}

	var waitGroup sync.WaitGroup
	dependencyCount := len(dependencyList)
	waitGroup.Add(dependencyCount)
	checkingErrs := make(chan<- error, dependencyCount)

	for i := 0; i < dependencyCount; i++ {
		go func(wg *sync.WaitGroup, serviceName string) {
			defer wg.Done()
			if err := checkServiceAvailable(serviceName); err != nil {
				checkingErrs <- err
			}
		}(&waitGroup, dependencyList[i])
	}

	waitGroup.Wait()
	close(checkingErrs)

	if len(checkingErrs) > 0 {
		return fmt.Errorf("checking required dependencied services failed ")
	} else {
		return nil
	}
}

func checkServiceAvailable(serviceId string) error {
	for i := 0; i < common.CurrentConfig.Service.ConnectRetries; i++ {
		if common.RegistryClient != nil {
			if checkServiceAvailableViaRegistry(serviceId) == true {
				return nil
			}
		} else {
			if checkServiceAvailableByPing(serviceId) == nil {
				return nil
			}
		}
		time.Sleep(time.Duration(common.CurrentConfig.Service.Timeout) * time.Millisecond)
		common.LoggingClient.Debug(fmt.Sprintf("Checked %d times for %s availibility", i+1, serviceId))
	}

	errMsg := fmt.Sprintf("service dependency %s checking time out", serviceId)
	common.LoggingClient.Error(errMsg)
	return fmt.Errorf(errMsg)
}

func checkServiceAvailableByPing(serviceId string) error {
	common.LoggingClient.Info(fmt.Sprintf("Check %v service's status ...", serviceId))
	addr := common.CurrentConfig.Clients[serviceId].Url()
	timeout := int64(common.CurrentConfig.Service.BootTimeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + clients.ApiPingRoute)

	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Error getting ping: %v ", err))
	}
	return err
}

func checkServiceAvailableViaRegistry(serviceId string) bool {
	common.LoggingClient.Info(fmt.Sprintf("Check %s service's status via Registry...", serviceId))

	if !common.RegistryClient.IsAlive() {
		common.LoggingClient.Error("unable to check status of %s service: Registry not running")

		return false
	}

	if serviceId == common.ClientData {
		serviceId = clients.CoreDataServiceKey
	} else {
		serviceId = clients.CoreMetaDataServiceKey
	}
	_, err := common.RegistryClient.IsServiceAvailable(serviceId)
	if err != nil {
		common.LoggingClient.Error(err.Error())
		return false
	}

	return true
}

func initializeClients(ctx context.Context, waitGroup *sync.WaitGroup) {
	// initialize Core Metadata clients
	common.AddressableClient = metadata.NewAddressableClient(
		urlclient.NewMetadata(
			ctx,
			common.RegistryClient,
			waitGroup,
			clients.ApiAddressableRoute,
		),
	)

	common.DeviceClient = metadata.NewDeviceClient(
		urlclient.NewMetadata(
			ctx,
			common.RegistryClient,
			waitGroup,
			clients.ApiDeviceRoute,
		),
	)

	common.DeviceServiceClient = metadata.NewDeviceServiceClient(
		urlclient.NewMetadata(
			ctx,
			common.RegistryClient,
			waitGroup,
			clients.ApiDeviceServiceRoute,
		),
	)

	common.DeviceProfileClient = metadata.NewDeviceProfileClient(
		urlclient.NewMetadata(
			ctx,
			common.RegistryClient,
			waitGroup,
			clients.ApiDeviceProfileRoute,
		),
	)

	common.MetadataGeneralClient = general.NewGeneralClient(
		urlclient.NewMetadata(
			ctx,
			common.RegistryClient,
			waitGroup,
			"",
		),
	)

	common.ProvisionWatcherClient = metadata.NewProvisionWatcherClient(
		urlclient.NewMetadata(
			ctx,
			common.RegistryClient,
			waitGroup,
			clients.ApiProvisionWatcherRoute,
		),
	)

	// initialize Core Data clients
	common.EventClient = coredata.NewEventClient(
		urlclient.NewData(
			ctx,
			common.RegistryClient,
			waitGroup,
			clients.ApiEventRoute,
		),
	)

	common.ValueDescriptorClient = coredata.NewValueDescriptorClient(
		urlclient.NewData(
			ctx,
			common.RegistryClient,
			waitGroup,
			common.APIValueDescriptorRoute,
		),
	)
}
