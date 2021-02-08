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

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	v2clients "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/http"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

// Clients contains references to dependencies required by the Clients bootstrap implementation.
type Clients struct {
}

// NewClients create a new instance of Clients
func NewClients() *Clients {
	return &Clients{}
}

func (_ *Clients) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {
	return InitDependencyClients(ctx, startupTimer, dic)
}

// InitDependencyClients triggers Service Client Initializer to establish connection to Metadata and Core Data Services
// through Metadata Client and Core Data Client.
// Service Client Initializer also needs to check the service status of Metadata and Core Data Services,
// because they are important dependencies of Device Service.
// The initialization process should be pending until Metadata Service and Core Data Service are both available.
func InitDependencyClients(ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	if err := validateClientConfig(container.ConfigurationFrom(dic.Get)); err != nil {
		lc.Error(err.Error())
		return false
	}

	if checkDependencyServices(ctx, startupTimer, dic) == false {
		return false
	}

	initializeClients(dic)

	lc.Info("Service clients initialize successful.")
	return true
}

func validateClientConfig(configuration *common.ConfigurationStruct) error {

	if len(configuration.Clients[common.ClientMetadata].Host) == 0 {
		return fmt.Errorf("fatal error; Host setting for Core Metadata client not configured")
	}

	if configuration.Clients[common.ClientMetadata].Port == 0 {
		return fmt.Errorf("fatal error; Port setting for Core Metadata client not configured")
	}

	if len(configuration.Clients[common.ClientData].Host) == 0 {
		return fmt.Errorf("fatal error; Host setting for Core Data client not configured")
	}

	if configuration.Clients[common.ClientData].Port == 0 {
		return fmt.Errorf("fatal error; Port setting for Core Ddata client not configured")
	}

	// TODO: validate other settings for sanity: maxcmdops, ...

	return nil
}

func checkDependencyServices(ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	var dependencyList = []string{common.ClientData, common.ClientMetadata}
	var waitGroup sync.WaitGroup
	checkingErr := true

	dependencyCount := len(dependencyList)
	waitGroup.Add(dependencyCount)

	for i := 0; i < dependencyCount; i++ {
		go func(wg *sync.WaitGroup, serviceName string) {
			defer wg.Done()
			if checkServiceAvailable(ctx, serviceName, startupTimer, dic) == false {
				checkingErr = false
			}
		}(&waitGroup, dependencyList[i])
	}
	waitGroup.Wait()

	return checkingErr
}

func checkServiceAvailable(ctx context.Context, serviceId string, startupTimer startup.Timer, dic *di.Container) bool {
	rc := bootstrapContainer.RegistryFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	for startupTimer.HasNotElapsed() {
		select {
		case <-ctx.Done():
			return false
		default:
			if rc != nil {
				if checkServiceAvailableViaRegistry(serviceId, rc, lc) == nil {
					return true
				}
			} else {
				configuration := container.ConfigurationFrom(dic.Get)
				if checkServiceAvailableByPing(serviceId, configuration, lc) == nil {
					return true
				}
			}
			startupTimer.SleepForInterval()
		}
	}

	lc.Error(fmt.Sprintf("dependency %s service checking time out", serviceId))
	return false
}

func checkServiceAvailableByPing(serviceId string, configuration *common.ConfigurationStruct, lc logger.LoggingClient) error {
	lc.Info(fmt.Sprintf("Check %v service's status by ping...", serviceId))
	addr := configuration.Clients[serviceId].Url()
	timeout := int64(configuration.Service.Timeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + clients.ApiPingRoute)
	if err != nil {
		lc.Error(err.Error())
	}

	return err
}

func checkServiceAvailableViaRegistry(serviceId string, rc registry.Client, lc logger.LoggingClient) error {
	lc.Info(fmt.Sprintf("Check %s service's status via Registry...", serviceId))

	if !rc.IsAlive() {
		errMsg := fmt.Sprintf("unable to check status of %s service: Registry not running", serviceId)
		lc.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	if serviceId == common.ClientData {
		serviceId = clients.CoreDataServiceKey
	} else {
		serviceId = clients.CoreMetaDataServiceKey
	}
	_, err := rc.IsServiceAvailable(serviceId)
	if err != nil {
		lc.Error(err.Error())
		return err
	}

	return nil
}

func initializeClients(dic *di.Container) {
	configuration := container.ConfigurationFrom(dic.Get)
	dc := v2clients.NewDeviceClient(configuration.Clients[common.ClientMetadata].Url())
	dsc := v2clients.NewDeviceServiceClient(configuration.Clients[common.ClientMetadata].Url())
	dpc := v2clients.NewDeviceProfileClient(configuration.Clients[common.ClientMetadata].Url())
	pwc := v2clients.NewProvisionWatcherClient(configuration.Clients[common.ClientMetadata].Url())
	ec := v2clients.NewEventClient(configuration.Clients[common.ClientData].Url())

	dic.Update(di.ServiceConstructorMap{
		container.MetadataDeviceClientName: func(get di.Get) interface{} {
			return dc
		},
		container.MetadataDeviceServiceClientName: func(get di.Get) interface{} {
			return dsc
		},
		container.MetadataDeviceProfileClientName: func(get di.Get) interface{} {
			return dpc
		},
		container.MetadataProvisionWatcherClientName: func(get di.Get) interface{} {
			return pwc
		},
		container.CoredataEventClientName: func(get di.Get) interface{} {
			return ec
		},
	})
}
