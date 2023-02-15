// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2022 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a basic EdgeX Foundry device service implementation
// meant to be embedded in an application, similar in approach to the builtin
// net/http package.
package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/clients"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v3/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
	restController "github.com/edgexfoundry/device-sdk-go/v3/internal/controller/http"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/config"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	bootstrapTypes "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/edgexfoundry/go-mod-registry/v3/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	ds *DeviceService
)

// UpdatableConfig interface allows services to have custom configuration populated from configuration stored
// in the Configuration Provider (aka Consul). Services using custom configuration must implement this interface
// on their custom configuration, even if they do not use Configuration Provider. If they do not use the
// Configuration Provider they can have a dummy implementation of this interface.
// This wraps the actual interface from go-mod-bootstrap so device service code doesn't have to have the additional
// direct import of go-mod-bootstrap.
type UpdatableConfig interface {
	interfaces.UpdatableConfig
}

type DeviceService struct {
	ServiceName     string
	LoggingClient   logger.LoggingClient
	RegistryClient  registry.Client
	SecretProvider  interfaces.SecretProvider
	MetricsManager  interfaces.MetricsManager
	edgexClients    clients.EdgeXClients
	controller      *restController.RestController
	config          *config.ConfigurationStruct
	deviceService   *models.DeviceService
	driver          sdkModels.ProtocolDriver
	discovery       sdkModels.ProtocolDiscovery
	validator       sdkModels.DeviceValidator
	manager         sdkModels.AutoEventManager
	asyncCh         chan *sdkModels.AsyncValues
	deviceCh        chan []sdkModels.DiscoveredDevice
	initialized     bool
	dic             *di.Container
	flags           flags.Common
	configProcessor *bootstrapConfig.Processor
	ctx             context.Context
	wg              *sync.WaitGroup
}

func (s *DeviceService) Initialize(serviceName, serviceVersion string, proto interface{}) {
	if serviceName == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify device service name")
		os.Exit(1)
	}
	s.ServiceName = serviceName

	if serviceVersion == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify device service version")
		os.Exit(1)
	}
	sdkCommon.ServiceVersion = serviceVersion

	if driver, ok := proto.(sdkModels.ProtocolDriver); ok {
		s.driver = driver
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Please implement and specify the protocoldriver")
		os.Exit(1)
	}

	if discovery, ok := proto.(sdkModels.ProtocolDiscovery); ok {
		s.discovery = discovery
	} else {
		s.discovery = nil
	}

	if validator, ok := proto.(sdkModels.DeviceValidator); ok {
		s.validator = validator
	} else {
		s.validator = nil
	}

	s.deviceService = &models.DeviceService{
		Name: serviceName,
	}

	s.config = &config.ConfigurationStruct{}
}

func (s *DeviceService) UpdateFromContainer(r *mux.Router, dic *di.Container) {
	s.LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)
	s.RegistryClient = bootstrapContainer.RegistryFrom(dic.Get)
	s.SecretProvider = bootstrapContainer.SecretProviderFrom(dic.Get)
	s.MetricsManager = bootstrapContainer.MetricsManagerFrom(dic.Get)
	s.edgexClients.DeviceClient = bootstrapContainer.DeviceClientFrom(dic.Get)
	s.edgexClients.DeviceServiceClient = bootstrapContainer.DeviceServiceClientFrom(dic.Get)
	s.edgexClients.DeviceProfileClient = bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	s.edgexClients.ProvisionWatcherClient = bootstrapContainer.ProvisionWatcherClientFrom(dic.Get)
	s.edgexClients.EventClient = bootstrapContainer.EventClientFrom(dic.Get)
	s.config = container.ConfigurationFrom(dic.Get)
	s.manager = container.ManagerFrom(dic.Get)
	s.controller = restController.NewRestController(r, dic, s.ServiceName)
}

// Name returns the name of this Device Service
func (s *DeviceService) Name() string {
	return s.ServiceName
}

// Version returns the version number of this Device Service
func (s *DeviceService) Version() string {
	return sdkCommon.ServiceVersion
}

// GetSecretProvider returns the SecretProvider
func (s *DeviceService) GetSecretProvider() interfaces.SecretProvider {
	return s.SecretProvider
}

// GetMetricsManager returns the Metrics Manager used to register counter, gauge, gaugeFloat64 or timer metric types from
// github.com/rcrowley/go-metrics
func (s *DeviceService) GetMetricsManager() interfaces.MetricsManager {
	return s.MetricsManager
}

// GetLoggingClient returns the logger.LoggingClient
func (s *DeviceService) GetLoggingClient() logger.LoggingClient {
	return s.LoggingClient
}

// AsyncReadings returns a bool value to indicate whether the asynchronous reading is enabled.
func (s *DeviceService) AsyncReadings() bool {
	return s.config.Device.EnableAsyncReadings
}

func (s *DeviceService) DeviceDiscovery() bool {
	return s.config.Device.Discovery.Enabled
}

// AddRoute allows leveraging the existing internal web server to add routes specific to Device Service.
func (s *DeviceService) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) error {
	return s.controller.AddRoute(route, handler, methods...)
}

// Stop shuts down the Service
func (s *DeviceService) Stop(force bool) {
	if s.initialized {
		err := s.driver.Stop(force)
		if err != nil {
			s.LoggingClient.Error(err.Error())
		}
	}
}

// LoadCustomConfig uses the Config Processor from go-mod-bootstrap to attempt to load service's
// custom configuration. It uses the same command line flags to process the custom config in the same manner
// as the standard configuration.
func (s *DeviceService) LoadCustomConfig(customConfig UpdatableConfig, sectionName string) error {
	if s.configProcessor == nil {
		s.configProcessor = bootstrapConfig.NewProcessorForCustomConfig(s.flags, s.ctx, s.wg, s.dic)
	}

	if err := s.configProcessor.LoadCustomConfigSection(customConfig, sectionName); err != nil {
		return err
	}

	s.controller.SetCustomConfigInfo(customConfig)

	return nil
}

// ListenForCustomConfigChanges uses the Config Processor from go-mod-bootstrap to attempt to listen for
// changes to the specified custom configuration section. LoadCustomConfig must be called previously so that
// the instance of svc.configProcessor has already been set.
func (s *DeviceService) ListenForCustomConfigChanges(
	configToWatch interface{},
	sectionName string,
	changedCallback func(interface{})) error {
	if s.configProcessor == nil {
		return fmt.Errorf(
			"custom configuration must be loaded before '%s' section can be watched for changes",
			sectionName)
	}

	s.configProcessor.ListenForCustomConfigChanges(configToWatch, sectionName, changedCallback)
	return nil
}

// selfRegister register device service itself onto metadata.
func (s *DeviceService) selfRegister() errors.EdgeX {
	localDeviceService := models.DeviceService{
		Name:        s.ServiceName,
		Labels:      s.config.Device.Labels,
		BaseAddress: bootstrapTypes.DefaultHttpProtocol + "://" + s.config.Service.Host + ":" + strconv.FormatInt(int64(s.config.Service.Port), 10),
		AdminState:  models.Unlocked,
	}
	*s.deviceService = localDeviceService
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck

	s.LoggingClient.Debugf("trying to find device service %s", localDeviceService.Name)
	res, err := s.edgexClients.DeviceServiceClient.DeviceServiceByName(ctx, localDeviceService.Name)
	if err != nil {
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			s.LoggingClient.Infof("device service %s doesn't exist, creating a new one", localDeviceService.Name)
			req := requests.NewAddDeviceServiceRequest(dtos.FromDeviceServiceModelToDTO(localDeviceService))
			idRes, err := s.edgexClients.DeviceServiceClient.Add(ctx, []requests.AddDeviceServiceRequest{req})
			if err != nil {
				s.LoggingClient.Errorf("failed to add device service %s: %v", localDeviceService.Name, err)
				return err
			}
			s.deviceService.Id = idRes[0].Id
			s.LoggingClient.Debugf("new device service id: %s", localDeviceService.Id)
		} else {
			s.LoggingClient.Errorf("failed to find device service %s", localDeviceService.Name)
			return err
		}
	} else {
		s.LoggingClient.Infof("device service %s exists, updating it", s.ServiceName)
		req := requests.NewUpdateDeviceServiceRequest(dtos.FromDeviceServiceModelToUpdateDTO(localDeviceService))
		req.Service.Id = nil
		_, err = s.edgexClients.DeviceServiceClient.Update(ctx, []requests.UpdateDeviceServiceRequest{req})
		if err != nil {
			s.LoggingClient.Errorf("failed to update device service %s with local config: %v", localDeviceService.Name, err)
			oldDeviceService := dtos.ToDeviceServiceModel(res.Service)
			*s.deviceService = oldDeviceService
		}
	}

	return nil
}

// DriverConfigs retrieves the driver specific configuration
func (s *DeviceService) DriverConfigs() map[string]string {
	return s.config.Driver
}

// SetDeviceOpState sets the operating state of device
func (s *DeviceService) SetDeviceOpState(name string, state models.OperatingState) error {
	d, err := s.GetDeviceByName(name)
	if err != nil {
		return err
	}

	d.OperatingState = state
	return s.UpdateDevice(d)
}

// RunningService returns the Service instance which is running
func RunningService() *DeviceService {
	return ds
}

// DriverConfigs retrieves the driver specific configuration
// TODO remove this in EDGEX3.0
func DriverConfigs() map[string]string {
	return ds.config.Driver
}
