// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"net/http"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	v2clients "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/http"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
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
	configuration := container.ConfigurationFrom(dic.Get)

	err := validateClientConfig(container.ConfigurationFrom(dic.Get))
	if err != nil {
		lc.Error(err.Error())
		return false
	}

	if checkDependencyServices(ctx, startupTimer, dic) == false {
		return false
	}
	initCoreServiceClients(dic)

	if configuration.MessageQueue.Enabled && initMessagingClient(ctx, wg, startupTimer, dic) == false {
		return false
	}

	lc.Info("Service clients initialize successful.")
	return true
}

func validateClientConfig(configuration *config.ConfigurationStruct) errors.EdgeX {
	if len(configuration.Clients[clients.CoreMetaDataServiceKey].Host) == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "fatal error; Host setting for Core Metadata client not configured", nil)
	}

	if configuration.Clients[clients.CoreMetaDataServiceKey].Port == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "fatal error; Port setting for Core Metadata client not configured", nil)
	}

	if len(configuration.Clients[clients.CoreDataServiceKey].Host) == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "fatal error; Host setting for Core Data client not configured", nil)
	}

	if configuration.Clients[clients.CoreDataServiceKey].Port == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "fatal error; Port setting for Core Data client not configured", nil)
	}

	// TODO: validate other settings for sanity: maxcmdops, ...

	return nil
}

func checkDependencyServices(ctx context.Context, startupTimer startup.Timer, dic *di.Container) bool {
	var dependencyList = []string{clients.CoreDataServiceKey, clients.CoreMetaDataServiceKey}
	var waitGroup sync.WaitGroup
	checkingErr := true

	dependencyCount := len(dependencyList)
	waitGroup.Add(dependencyCount)

	for i := 0; i < dependencyCount; i++ {
		go func(wg *sync.WaitGroup, serviceKey string) {
			defer wg.Done()
			if checkServiceAvailable(ctx, serviceKey, startupTimer, dic) == false {
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
				configuration := container.ConfigurationFrom(dic.Get)
				if checkServiceAvailableByPing(serviceKey, configuration, lc) {
					return true
				}
			}
			startupTimer.SleepForInterval()
		}
	}

	lc.Errorf("dependency %s service checking time out", serviceKey)
	return false
}

func checkServiceAvailableByPing(serviceKey string, configuration *config.ConfigurationStruct, lc logger.LoggingClient) bool {
	lc.Infof("Check %v service's status by ping...", serviceKey)
	addr := configuration.Clients[serviceKey].Url()
	timeout := int64(configuration.Service.Timeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + clients.ApiPingRoute)
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

func initCoreServiceClients(dic *di.Container) {
	configuration := container.ConfigurationFrom(dic.Get)
	dc := v2clients.NewDeviceClient(configuration.Clients[clients.CoreMetaDataServiceKey].Url())
	dsc := v2clients.NewDeviceServiceClient(configuration.Clients[clients.CoreMetaDataServiceKey].Url())
	dpc := v2clients.NewDeviceProfileClient(configuration.Clients[clients.CoreMetaDataServiceKey].Url())
	pwc := v2clients.NewProvisionWatcherClient(configuration.Clients[clients.CoreMetaDataServiceKey].Url())
	ec := v2clients.NewEventClient(configuration.Clients[clients.CoreDataServiceKey].Url())

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

func initMessagingClient(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	msgClient, err := messaging.NewMessageClient(
		types.MessageBusConfig{
			PublishHost: types.HostInfo{
				Host:     configuration.MessageQueue.Host,
				Port:     configuration.MessageQueue.Port,
				Protocol: configuration.MessageQueue.Protocol,
			},
			Type:     configuration.MessageQueue.Type,
			Optional: configuration.MessageQueue.Optional,
		})
	if err != nil {
		lc.Errorf("Failed to create MessageClient: %v", err)
		return false
	}

	for startupTimer.HasNotElapsed() {
		select {
		case <-ctx.Done():
			return false
		default:
			err = msgClient.Connect()
			if err != nil {
				lc.Warnf("Unable to connect MessageBus: %v", err)
			} else {
				wg.Add(1)
				go func() {
					defer wg.Done()
					select {
					case <-ctx.Done():
						msgClient.Disconnect()
						lc.Infof("Disconnecting from MessageBus")
					}
				}()
				dic.Update(di.ServiceConstructorMap{
					container.MessagingClientName: func(get di.Get) interface{} {
						return msgClient
					},
				})
				return true
			}
			startupTimer.SleepForInterval()
		}
	}

	lc.Error("Connecting to MessageBus time out")
	return false
}
