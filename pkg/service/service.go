// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
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
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/internal/clients"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/controller"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/go-mod-registry/registry"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	ds *DeviceService
)

type DeviceService struct {
	ServiceName    string
	LoggingClient  logger.LoggingClient
	RegistryClient registry.Client
	edgexClients   clients.EdgeXClients
	controller     *controller.RestController
	config         *common.ConfigurationStruct
	deviceService  contract.DeviceService
	driver         dsModels.ProtocolDriver
	discovery      dsModels.ProtocolDiscovery
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
	s.edgexClients.GeneralClient = container.GeneralClientFrom(dic.Get)
	s.edgexClients.DeviceClient = container.MetadataDeviceClientFrom(dic.Get)
	s.edgexClients.DeviceServiceClient = container.MetadataDeviceServiceClientFrom(dic.Get)
	s.edgexClients.DeviceProfileClient = container.MetadataDeviceProfileClientFrom(dic.Get)
	s.edgexClients.AddressableClient = container.MetadataAddressableClientFrom(dic.Get)
	s.edgexClients.ProvisionWatcherClient = container.MetadataProvisionWatcherClientFrom(dic.Get)
	s.edgexClients.EventClient = container.CoredataEventClientFrom(dic.Get)
	s.edgexClients.ValueDescriptorClient = container.CoredataValueDescriptorClientFrom(dic.Get)
	s.config = container.ConfigurationFrom(dic.Get)
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
	autoevent.GetManager().StopAutoEvents()
}

// selfRegister register device service itself onto metadata.
func (s *DeviceService) selfRegister() error {
	addr, err := s.createAndUpdateAddressable()
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("createAndUpdateAddressable failed: %v", err))
		return err
	}

	newDeviceService := contract.DeviceService{
		Name:           s.ServiceName,
		Labels:         s.config.Service.Labels,
		OperatingState: contract.Enabled,
		Addressable:    *addr,
		AdminState:     contract.Unlocked,
	}
	newDeviceService.Origin = time.Now().UnixNano() / int64(time.Millisecond)

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	s.LoggingClient.Debug("Trying to find DeviceService: " + s.ServiceName)
	ds, err := s.edgexClients.DeviceServiceClient.DeviceServiceForName(ctx, s.ServiceName)
	if err != nil {
		if errsc, ok := err.(types.ErrServiceClient); ok && (errsc.StatusCode == http.StatusNotFound) {
			s.LoggingClient.Info(fmt.Sprintf("DeviceService %s doesn't exist, creating a new one", s.ServiceName))
			id, err := s.edgexClients.DeviceServiceClient.Add(ctx, &newDeviceService)
			if err != nil {
				s.LoggingClient.Error(fmt.Sprintf("Failed to add Deviceservice %s: %v", s.ServiceName, err))
				return err
			}
			if err = common.VerifyIdFormat(id, "Device Service"); err != nil {
				return err
			}
			// NOTE - this differs from Addressable and Device Resources,
			// neither of which require the '.Service'prefix
			newDeviceService.Id = id
			s.LoggingClient.Debug("New DeviceService Id: " + newDeviceService.Id)
		} else {
			s.LoggingClient.Error(fmt.Sprintf("DeviceServicForName failed: %v", err))
			return err
		}
	} else {
		s.LoggingClient.Info(fmt.Sprintf("DeviceService %s exists, updating it", ds.Name))
		err = s.edgexClients.DeviceServiceClient.Update(ctx, newDeviceService)
		if err != nil {
			s.LoggingClient.Error(fmt.Sprintf("Failed to update DeviceService %s: %v", newDeviceService.Name, err))
			// use the existed one to at least make sure config is in sync with metadata.
			newDeviceService = ds
		}
		newDeviceService.Id = ds.Id
	}

	s.deviceService = newDeviceService
	return nil
}

// TODO: Addressable will be removed in v2.
func (s *DeviceService) createAndUpdateAddressable() (*contract.Addressable, error) {
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	newAddr := contract.Addressable{
		Timestamps: contract.Timestamps{
			Origin: time.Now().UnixNano() / int64(time.Millisecond),
		},
		Name:       s.ServiceName,
		HTTPMethod: http.MethodPost,
		Protocol:   common.HttpProto,
		Address:    s.config.Service.Host,
		Port:       s.config.Service.Port,
		Path:       common.APICallbackRoute,
	}

	addr, err := s.edgexClients.AddressableClient.AddressableForName(ctx, s.ServiceName)
	if err != nil {
		if errsc, ok := err.(types.ErrServiceClient); ok && (errsc.StatusCode == http.StatusNotFound) {
			s.LoggingClient.Info(fmt.Sprintf("Addressable %s doesn't exist, creating a new one", s.ServiceName))
			id, err := s.edgexClients.AddressableClient.Add(ctx, &newAddr)
			if err != nil {
				s.LoggingClient.Error(fmt.Sprintf("Failed to add Addressable %s: %v", newAddr.Name, err))
				return nil, err
			}
			if err = common.VerifyIdFormat(id, "Addressable"); err != nil {
				return nil, err
			}
			newAddr.Id = id
		} else {
			s.LoggingClient.Error(fmt.Sprintf("AddressableForName failed: %v", err))
			return nil, err
		}
	} else {
		s.LoggingClient.Info(fmt.Sprintf("Addressable %s exists, updating it", s.ServiceName))
		err = s.edgexClients.AddressableClient.Update(ctx, newAddr)
		if err != nil {
			s.LoggingClient.Error(fmt.Sprintf("Failed to update Addressable %s: %v", s.ServiceName, err))
			// use the existed one to at least make sure config is in sync with metadata.
			newAddr = addr
		}
		newAddr.Id = addr.Id
	}

	return &newAddr, nil
}

// RunningService returns the Service instance which is running
func RunningService() *DeviceService {
	return ds
}

// DriverConfigs retrieves the driver specific configuration
func DriverConfigs() map[string]string {
	return ds.config.Driver
}
