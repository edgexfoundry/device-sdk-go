// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/config"
	"github.com/edgexfoundry/device-sdk-go/internal/registry"
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	consulapi "github.com/hashicorp/consul/api"
)

const clientCount int = 8

// InitDependencyClients triggers Service Client Initializer to establish connection to Metadata and Core Data Services
// through Metadata Client and Core Data Client.
// Service Client Initializer also needs to check the service status of Metadata and Core Data Services,
// because they are important dependencies of Device Service.
// The initialization process should be pending until Metadata Service and Core Data Service are both available.
func InitDependencyClients() error {
	if err := validateClientConfig(); err != nil {
		return err
	}

	initializeLoggingClient()

	if err := checkDependencyServices(); err != nil {
		return err
	}

	initializeClients()

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

func initializeLoggingClient() {
	var logTarget string
	config := common.CurrentConfig

	if config.Logging.EnableRemote {
		logTarget = config.Clients[common.ClientLogging].Url() + clients.ApiLoggingRoute
		fmt.Println("EnableRemote is true, using remote logging service")
	} else {
		logTarget = config.Logging.File
		fmt.Println("EnableRemote is false, using local log file")
	}

	common.LoggingClient = logger.NewClient(common.ServiceName, config.Logging.EnableRemote, logTarget, config.Logging.Level)
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
	for i := 0; i < 50; i++ {
		if common.UseRegistry {
			if checkServiceAvailableByConsul(common.CurrentConfig.Clients[serviceId].Name) == true {
				return nil
			}
		} else {
			if checkServiceAvailableByPing(serviceId) == nil {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
		common.LoggingClient.Debug(fmt.Sprintf("Checked %d times for %s availibility", i+1, serviceId))
	}

	errMsg := fmt.Sprintf("service dependency %s checking time out", serviceId)
	common.LoggingClient.Error(errMsg)
	return fmt.Errorf(errMsg)
}

func checkServiceAvailableByPing(serviceId string) error {
	common.LoggingClient.Info(fmt.Sprintf("Check %v service's status ...", serviceId))
	addr := common.CurrentConfig.Clients[serviceId].Url()
	timeout := int64(common.CurrentConfig.Clients[serviceId].Timeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + clients.ApiPingRoute)

	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Error getting ping: %v ", err))
	}
	return err
}

func checkServiceAvailableByConsul(serviceConsulId string) bool {
	common.LoggingClient.Info(fmt.Sprintf("Check %v service's status by Consul...", serviceConsulId))

	result := false

	isConsulUp := checkConsulAvailable()
	if !isConsulUp {
		return false
	}

	// Get a new client
	var host = common.CurrentConfig.Registry.Host
	var port = strconv.Itoa(common.CurrentConfig.Registry.Port)
	var consulAddr = common.BuildAddr(host, port)
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = consulAddr
	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		common.LoggingClient.Error(err.Error())
		return false
	}

	services, _, err := client.Catalog().Service(serviceConsulId, "", nil)
	if err != nil {
		common.LoggingClient.Error(err.Error())
		return false
	}
	if len(services) <= 0 {
		common.LoggingClient.Error(serviceConsulId + " service hasn't started...")
		return false
	}

	healthCheck, _, err := client.Health().Checks(serviceConsulId, nil)
	if err != nil {
		common.LoggingClient.Error(err.Error())
		return false
	}
	status := healthCheck.AggregatedStatus()
	if status == common.ServiceStatusPass {
		result = true
	} else {
		common.LoggingClient.Error(serviceConsulId + " service hasn't been available...")
		result = false
	}

	return result
}

func checkConsulAvailable() bool {
	addr := fmt.Sprintf("%v:%v", common.CurrentConfig.Registry.Host, common.CurrentConfig.Registry.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Consul cannot be reached, address: %v and error is \"%v\" ", addr, err.Error()))
		return false
	}
	conn.Close()
	return true
}

func initializeClients() {
	isRegistry := common.UseRegistry
	var waitGroup sync.WaitGroup
	waitGroup.Add(clientCount)

	consulEndpoint := &registry.ConsulEndpoint{RegistryClient: config.RegistryClient, WG: &waitGroup}

	metaAddr := common.CurrentConfig.Clients[common.ClientMetadata].Url()
	dataAddr := common.CurrentConfig.Clients[common.ClientData].Url()

	params := types.EndpointParams{
		UseRegistry: isRegistry,
		Interval:    15,
	}

	// initialize Core Metadata clients
	params.ServiceKey = common.CurrentConfig.Clients[common.ClientMetadata].Name

	params.Path = clients.ApiAddressableRoute
	params.Url = metaAddr + params.Path
	common.AddressableClient = metadata.NewAddressableClient(params, consulEndpoint)

	params.Path = clients.ApiDeviceRoute
	params.Url = metaAddr + params.Path
	common.DeviceClient = metadata.NewDeviceClient(params, consulEndpoint)

	params.Path = clients.ApiDeviceServiceRoute
	params.Url = metaAddr + params.Path
	common.DeviceServiceClient = metadata.NewDeviceServiceClient(params, consulEndpoint)

	params.Path = clients.ApiDeviceProfileRoute
	params.Url = metaAddr + params.Path
	common.DeviceProfileClient = metadata.NewDeviceProfileClient(params, consulEndpoint)

	params.Path = clients.ApiScheduleRoute
	params.Url = metaAddr + params.Path
	common.ScheduleClient = metadata.NewScheduleClient(params, consulEndpoint)

	params.Path = clients.ApiScheduleEventRoute
	params.Url = metaAddr + params.Path
	common.ScheduleEventClient = metadata.NewScheduleEventClient(params, consulEndpoint)

	// initialize Core Data clients
	params.ServiceKey = common.CurrentConfig.Clients[common.ClientData].Name

	params.Path = clients.ApiEventRoute
	params.Url = dataAddr + params.Path
	common.EventClient = coredata.NewEventClient(params, consulEndpoint)

	params.Path = common.APIValueDescriptorRoute
	params.Url = dataAddr + params.Path
	common.ValueDescriptorClient = coredata.NewValueDescriptorClient(params, consulEndpoint)

	if isRegistry {
		// wait for the first endpoint discovery to make sure all clients work
		waitGroup.Wait()
	}
}
