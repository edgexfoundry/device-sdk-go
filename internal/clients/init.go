// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
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

	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"
)

func BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {
	return InitDependencyClients(ctx, wg, startupTimer, dic)
}

// InitDependencyClients triggers Service Client Initializer to establish connection to Metadata and Core Data Services
// through Metadata Client and Core Data Client.
// Service Client Initializer also needs to check the service status of Metadata and Core Data Services,
// because they are important dependencies of Device Service.
// The initialization process should be pending until Metadata Service and Core Data Service are both available.
func InitDependencyClients(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	// Remove the core-data client when using the MessageBus
	if config.Device.UseMessageBus {
		delete(config.Clients, common.CoreDataServiceKey)
	}
	err := validateClientConfig(container.ConfigurationFrom(dic.Get))
	if err != nil {
		lc.Error(err.Error())
		return false
	}

	if !checkDependencyServices(ctx, startupTimer, dic) {
		return false
	}

	lc.Info("Service clients initialize successful.")
	return true
}

func validateClientConfig(configuration *config.ConfigurationStruct) errors.EdgeX {
	for serviceKey, serviceInfo := range configuration.GetBootstrap().Clients {
		if len(serviceInfo.Host) == 0 {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("fatal error; Host setting for %s client not configured", serviceKey), nil)
		}
		if serviceInfo.Port == 0 {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("fatal error; Port setting for %s client not configured", serviceKey), nil)
		}
	}

	// TODO: validate other settings for sanity: maxcmdops, ...

	return nil
}

func checkDependencyServices(ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	var dependencyList = []string{common.CoreMetaDataServiceKey}
	var waitGroup sync.WaitGroup
	checkingErr := true

	dependencyCount := len(dependencyList)
	waitGroup.Add(dependencyCount)

	for i := 0; i < dependencyCount; i++ {
		go func(wg *sync.WaitGroup, serviceKey string) {
			defer wg.Done()
			if !checkServiceAvailable(ctx, serviceKey, startupTimer, dic) {
				checkingErr = false
			}
		}(&waitGroup, dependencyList[i])
	}
	waitGroup.Wait()

	return checkingErr
}

func checkServiceAvailable(ctx context.Context, serviceKey string, startupTimer startup.Timer, dic *di.Container) bool {
	rc := bootstrapContainer.RegistryFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	timeout, err := time.ParseDuration(configuration.Service.RequestTimeout)
	if err != nil {
		lc.Errorf("Unable to parse RequestTimeout value of '%s' duration: %w", configuration.Service.RequestTimeout, err)
		return false
	}

	for startupTimer.HasNotElapsed() {
		select {
		case <-ctx.Done():
			return false
		default:
			if rc != nil {
				if checkServiceAvailableViaRegistry(serviceKey, rc, lc) {
					return true
				}
			} else {
				if checkServiceAvailableByPing(serviceKey, timeout, configuration, lc) {
					return true
				}
			}
			startupTimer.SleepForInterval()
		}
	}

	lc.Errorf("dependency %s service checking time out", serviceKey)
	return false
}

func checkServiceAvailableByPing(serviceKey string, timeout time.Duration, configuration *config.ConfigurationStruct, lc logger.LoggingClient) bool {
	lc.Infof("Check %v service's status by ping...", serviceKey)
	addr := configuration.Clients[serviceKey].Url()

	client := http.Client{
		Timeout: timeout,
	}

	_, err := client.Get(addr + common.ApiPingRoute)
	if err != nil {
		lc.Error(err.Error())
		return false
	}

	return true
}

func checkServiceAvailableViaRegistry(serviceKey string, rc registry.Client, lc logger.LoggingClient) bool {
	lc.Infof("Check %s service's status via Registry...", serviceKey)

	if !rc.IsAlive() {
		lc.Warnf("unable to check status of %s service: Registry not running", serviceKey)
		return false
	}

	res, err := rc.IsServiceAvailable(serviceKey)
	if err != nil {
		lc.Error(err.Error())
	}

	return res
}
