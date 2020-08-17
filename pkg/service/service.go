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
	"github.com/edgexfoundry/go-mod-registry/registry"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

var (
	sdk *DeviceServiceSDK
)

type DeviceServiceSDK struct {
	ServiceName    string
	LoggingClient  logger.LoggingClient
	registryClient registry.Client
	edgexClients   clients.EdgeXClient
	controller     controller.RestController
	config         *common.ConfigurationStruct
	svcInfo        *common.ServiceInfo
	deviceService  contract.DeviceService
	driver         dsModels.ProtocolDriver
	discovery      dsModels.ProtocolDiscovery
	asyncCh        chan *dsModels.AsyncValues
	deviceCh       chan []dsModels.DiscoveredDevice
	initialized    bool
}

// TODO: remove common.xxx = xxx assignment after refactor are done.
func (s *DeviceServiceSDK) Initialize(serviceName, serviceVersion string, proto interface{}) {
	if serviceName == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify device service name")
		os.Exit(1)
	}
	common.ServiceName = serviceName
	s.ServiceName = serviceName

	if serviceVersion == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify device service version")
		os.Exit(1)
	}
	common.ServiceVersion = serviceVersion

	if driver, ok := proto.(dsModels.ProtocolDriver); ok {
		common.Driver = driver
		s.driver = driver
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Please implement and specify the protocoldriver")
		os.Exit(1)
	}

	if discovery, ok := proto.(dsModels.ProtocolDiscovery); ok {
		common.Discovery = discovery
		s.discovery = discovery
	} else {
		s.discovery = nil
	}

	s.config = &common.ConfigurationStruct{}
}

func (s *DeviceServiceSDK) UpdateFromContainer(r *mux.Router, dic *di.Container) {
	s.LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)
	s.registryClient = bootstrapContainer.RegistryFrom(dic.Get)
	s.edgexClients.GeneralClient = container.MetadataGeneralClientFrom(dic.Get)
	s.edgexClients.DeviceClient = container.MetadataDeviceClientFrom(dic.Get)
	s.edgexClients.DeviceServiceClient = container.MetadataDeviceServiceClientFrom(dic.Get)
	s.edgexClients.DeviceProfileClient = container.MetadataDeviceProfileClientFrom(dic.Get)
	s.edgexClients.AddressableClient = container.MetadataAddressableClientFrom(dic.Get)
	s.edgexClients.ProvisionWatcherClient = container.MetadataProvisionWatcherClientFrom(dic.Get)
	s.edgexClients.EventClient = container.CoredataEventClientFrom(dic.Get)
	s.edgexClients.ValueDescriptorClient = container.CoredataValueDescriptorClientFrom(dic.Get)

	s.svcInfo = &container.ConfigurationFrom(dic.Get).Service
	s.controller = container.RestControllerFrom(dic.Get)
}

// Name returns the name of this Device Service
func (s *DeviceServiceSDK) Name() string {
	return s.ServiceName
}

// Version returns the version number of this Device Service
func (s *DeviceServiceSDK) Version() string {
	return common.ServiceVersion
}

// AsyncReadings returns a bool value to indicate whether the asynchronous reading is enabled.
func (s *DeviceServiceSDK) AsyncReadings() bool {
	return s.config.Service.EnableAsyncReadings
}

// AddRoute allows leveraging the existing internal web server to add routes specific to Device Service.
func (s *DeviceServiceSDK) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) error {
	return s.controller.AddRoute(route, handler, methods...)
}

// Stop shuts down the Service
func (s *DeviceServiceSDK) Stop(force bool) {
	if s.initialized {
		//_ = s.driver.Stop(false)
		_ = common.Driver.Stop(force)
	}
	autoevent.GetManager().StopAutoEvents()
}

// selfRegister register device service itself onto metadata.
func (s *DeviceServiceSDK) selfRegister() error {
	s.LoggingClient.Debug("Trying to find Device Service: " + common.ServiceName)

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	ds, err := s.edgexClients.DeviceServiceClient.DeviceServiceForName(ctx, common.ServiceName)

	if err != nil {
		if errsc, ok := err.(types.ErrServiceClient); ok && (errsc.StatusCode == http.StatusNotFound) {
			s.LoggingClient.Info(fmt.Sprintf("Device Service %s doesn't exist, creating a new one", s.ServiceName))
			ds, err = s.createNewDeviceService()
		} else {
			s.LoggingClient.Error(fmt.Sprintf("DeviceServicForName failed: %v", err))
			return err
		}
	} else {
		s.LoggingClient.Info(fmt.Sprintf("Device Service %s exists", ds.Name))
	}

	s.LoggingClient.Debug(fmt.Sprintf("Device Service in Core MetaData: %s", common.ServiceName))
	s.deviceService = ds
	// TODO: remove this after refactor are done.
	common.CurrentDeviceService = ds

	return nil
}

func (s *DeviceServiceSDK) createNewDeviceService() (contract.DeviceService, error) {
	addr, err := s.makeNewAddressable()
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("makeNewAddressable failed: %v", err))
		return contract.DeviceService{}, err
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	ds := contract.DeviceService{
		Name:           s.ServiceName,
		Labels:         s.svcInfo.Labels,
		OperatingState: contract.Enabled,
		Addressable:    *addr,
		AdminState:     contract.Unlocked,
	}
	ds.Origin = millis

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err := s.edgexClients.DeviceServiceClient.Add(ctx, &ds)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Add Deviceservice: %s; failed: %v", s.ServiceName, err))
		return contract.DeviceService{}, err
	}
	if err = common.VerifyIdFormat(id, "Device Service"); err != nil {
		return contract.DeviceService{}, err
	}

	// NOTE - this differs from Addressable and Device Resources,
	// neither of which require the '.Service'prefix
	ds.Id = id
	common.LoggingClient.Debug("New device service Id: " + ds.Id)

	return ds, nil
}

func (s *DeviceServiceSDK) makeNewAddressable() (*contract.Addressable, error) {
	// check whether there has been an existing addressable
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	addr, err := s.edgexClients.AddressableClient.AddressableForName(ctx, s.ServiceName)
	if err != nil {
		if errsc, ok := err.(types.ErrServiceClient); ok && (errsc.StatusCode == http.StatusNotFound) {
			s.LoggingClient.Info(fmt.Sprintf("Addressable %s doesn't exist, creating a new one", s.ServiceName))
			millis := time.Now().UnixNano() / int64(time.Millisecond)
			addr = contract.Addressable{
				Timestamps: contract.Timestamps{
					Origin: millis,
				},
				Name:       s.ServiceName,
				HTTPMethod: http.MethodPost,
				Protocol:   common.HttpProto,
				Address:    s.svcInfo.Host,
				Port:       s.svcInfo.Port,
				Path:       common.APICallbackRoute,
			}
			id, err := s.edgexClients.AddressableClient.Add(ctx, &addr)
			if err != nil {
				s.LoggingClient.Error(fmt.Sprintf("Add addressable failed %s, error: %v", addr.Name, err))
				return nil, err
			}
			if err = common.VerifyIdFormat(id, "Addressable"); err != nil {
				return nil, err
			}
			addr.Id = id
		} else {
			s.LoggingClient.Error(fmt.Sprintf("AddressableForName failed: %v", err))
			return nil, err
		}
	} else {
		s.LoggingClient.Info(fmt.Sprintf("Addressable %s exists", common.ServiceName))
	}

	return &addr, nil
}

// RunningService returns the Service instance which is running
func RunningService() *DeviceServiceSDK {
	return sdk
}

// DriverConfigs retrieves the driver specific configuration
func DriverConfigs() map[string]string {
	return sdk.config.Driver
}
