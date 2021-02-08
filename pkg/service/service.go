// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
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

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/clients"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/controller"
	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

var (
	ds *DeviceService
)

type DeviceService struct {
	ServiceName    string
	LoggingClient  logger.LoggingClient
	RegistryClient registry.Client
	SecretProvider interfaces.SecretProvider
	edgexClients   clients.EdgeXClients
	controller     *controller.RestController
	config         *common.ConfigurationStruct
	deviceService  *models.DeviceService
	driver         dsModels.ProtocolDriver
	discovery      dsModels.ProtocolDiscovery
	manager        dsModels.AutoEventManager
	asyncCh        chan *dsModels.AsyncValues
	deviceCh       chan []dsModels.DiscoveredDevice
	initialized    bool
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
	common.ServiceVersion = serviceVersion

	if driver, ok := proto.(dsModels.ProtocolDriver); ok {
		s.driver = driver
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Please implement and specify the protocoldriver")
		os.Exit(1)
	}

	if discovery, ok := proto.(dsModels.ProtocolDiscovery); ok {
		s.discovery = discovery
	} else {
		s.discovery = nil
	}

	s.config = &common.ConfigurationStruct{}
}

func (s *DeviceService) UpdateFromContainer(r *mux.Router, dic *di.Container) {
	s.LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)
	s.RegistryClient = bootstrapContainer.RegistryFrom(dic.Get)
	s.SecretProvider = bootstrapContainer.SecretProviderFrom(dic.Get)
	s.edgexClients.DeviceClient = container.MetadataDeviceClientFrom(dic.Get)
	s.edgexClients.DeviceServiceClient = container.MetadataDeviceServiceClientFrom(dic.Get)
	s.edgexClients.DeviceProfileClient = container.MetadataDeviceProfileClientFrom(dic.Get)
	s.edgexClients.ProvisionWatcherClient = container.MetadataProvisionWatcherClientFrom(dic.Get)
	s.edgexClients.EventClient = container.CoredataEventClientFrom(dic.Get)
	s.config = container.ConfigurationFrom(dic.Get)
	s.manager = container.ManagerFrom(dic.Get)
	s.controller = controller.NewRestController(r, dic)
}

// Name returns the name of this Device Service
func (s *DeviceService) Name() string {
	return s.ServiceName
}

// Version returns the version number of this Device Service
func (s *DeviceService) Version() string {
	return common.ServiceVersion
}

// AsyncReadings returns a bool value to indicate whether the asynchronous reading is enabled.
func (s *DeviceService) AsyncReadings() bool {
	return s.config.Service.EnableAsyncReadings
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
		_ = s.driver.Stop(false)
	}
	s.manager.StopAutoEvents()
}

// selfRegister register device service itself onto metadata.
func (s *DeviceService) selfRegister() error {
	newDeviceService := models.DeviceService{
		Name:        s.ServiceName,
		Labels:      s.config.Service.Labels,
		BaseAddress: s.config.Service.Protocol + "://" + s.config.Service.Host + ":" + strconv.FormatInt(int64(s.config.Service.Port), 10),
		AdminState:  models.Unlocked,
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())

	s.LoggingClient.Debugf("trying to find device service %s", newDeviceService.Name)
	res, err := s.edgexClients.DeviceServiceClient.DeviceServiceByName(ctx, newDeviceService.Name)
	if err != nil {
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			s.LoggingClient.Infof("device service %s doesn't exist, creating a new one", newDeviceService.Name)
			req := requests.NewAddDeviceServiceRequest(dtos.FromDeviceServiceModelToDTO(newDeviceService))
			idRes, err := s.edgexClients.DeviceServiceClient.Add(ctx, []requests.AddDeviceServiceRequest{req})
			if err != nil {
				s.LoggingClient.Errorf("failed to add device service %s: %v", newDeviceService.Name, err)
				return err
			}
			newDeviceService.Id = idRes[0].Id
			s.LoggingClient.Debugf("new device service id: %s", newDeviceService.Id)
		} else {
			s.LoggingClient.Errorf("failed to find device service %s", newDeviceService.Name)
			return err
		}
	} else {
		s.LoggingClient.Infof("device service %s exists, updating it", s.ServiceName)
		req := requests.NewUpdateDeviceServiceRequest(dtos.FromDeviceServiceModelToUpdateDTO(newDeviceService))
		_, err = s.edgexClients.DeviceServiceClient.Update(ctx, []requests.UpdateDeviceServiceRequest{req})
		if err != nil {
			s.LoggingClient.Errorf("failed to update device service %s with local config: %v", newDeviceService.Name, err)
			newDeviceService = dtos.ToDeviceServiceModel(res.Service)
		}
	}

	s.deviceService = &newDeviceService
	return nil
}

// RunningService returns the Service instance which is running
func RunningService() *DeviceService {
	return ds
}

// DriverConfigs retrieves the driver specific configuration
func DriverConfigs() map[string]string {
	return ds.config.Driver
}
