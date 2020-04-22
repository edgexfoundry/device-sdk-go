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
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/controller"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

var (
	svc *Service
)

// A Service listens for requests and routes them to the right command
type Service struct {
	svcInfo     *common.ServiceInfo
	asyncCh     chan *dsModels.AsyncValues
	deviceCh    chan []dsModels.DiscoveredDevice
	startTime   time.Time
	controller  controller.RestController
	initiazlied bool
}

// Name returns the name of this Device Service
func (s *Service) Name() string {
	return common.ServiceName
}

// Version returns the version number of this Device Service
func (s *Service) Version() string {
	return common.ServiceVersion
}

// AsyncReadings returns a bool value to indicate whether the asynchronous reading is enabled.
func (s *Service) AsyncReadings() bool {
	return common.CurrentConfig.Service.EnableAsyncReadings
}

// AddRoute allows leveraging the existing internal web server to add routes specific to Device Service.
func (s *Service) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) error {
	return s.controller.AddRoute(route, handler, methods...)
}

// Stop shuts down the Service
func (s *Service) Stop(force bool) {
	if s.initiazlied {
		_ = common.Driver.Stop(force)
	}
	autoevent.GetManager().StopAutoEvents()
}

// selfRegister register device service itself onto metadata.
func selfRegister() error {
	common.LoggingClient.Debug("Trying to find Device Service: " + common.ServiceName)

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	ds, err := common.DeviceServiceClient.DeviceServiceForName(ctx, common.ServiceName)

	if err != nil {
		if errsc, ok := err.(types.ErrServiceClient); ok && (errsc.StatusCode == http.StatusNotFound) {
			common.LoggingClient.Info(fmt.Sprintf("Device Service %s doesn't exist, creating a new one", common.ServiceName))
			ds, err = createNewDeviceService()
		} else {
			common.LoggingClient.Error(fmt.Sprintf("DeviceServicForName failed: %v", err))
			return err
		}
	} else {
		common.LoggingClient.Info(fmt.Sprintf("Device Service %s exists", ds.Name))
	}

	common.LoggingClient.Debug(fmt.Sprintf("Device Service in Core MetaData: %s", common.ServiceName))
	common.CurrentDeviceService = ds

	return nil
}

func createNewDeviceService() (contract.DeviceService, error) {
	addr, err := makeNewAddressable()
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("makeNewAddressable failed: %v", err))
		return contract.DeviceService{}, err
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	ds := contract.DeviceService{
		Name:           common.ServiceName,
		Labels:         svc.svcInfo.Labels,
		OperatingState: "ENABLED",
		Addressable:    *addr,
		AdminState:     "UNLOCKED",
	}
	ds.Origin = millis

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err := common.DeviceServiceClient.Add(ctx, &ds)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Add Deviceservice: %s; failed: %v", common.ServiceName, err))
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

func makeNewAddressable() (*contract.Addressable, error) {
	// check whether there has been an existing addressable
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	addr, err := common.AddressableClient.AddressableForName(ctx, common.ServiceName)
	if err != nil {
		if errsc, ok := err.(types.ErrServiceClient); ok && (errsc.StatusCode == http.StatusNotFound) {
			common.LoggingClient.Info(fmt.Sprintf("Addressable %s doesn't exist, creating a new one", common.ServiceName))
			millis := time.Now().UnixNano() / int64(time.Millisecond)
			addr = contract.Addressable{
				Timestamps: contract.Timestamps{
					Origin: millis,
				},
				Name:       common.ServiceName,
				HTTPMethod: http.MethodPost,
				Protocol:   common.HttpProto,
				Address:    svc.svcInfo.Host,
				Port:       svc.svcInfo.Port,
				Path:       common.APICallbackRoute,
			}
			id, err := common.AddressableClient.Add(ctx, &addr)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("Add addressable failed %s, error: %v", addr.Name, err))
				return nil, err
			}
			if err = common.VerifyIdFormat(id, "Addressable"); err != nil {
				return nil, err
			}
			addr.Id = id
		} else {
			common.LoggingClient.Error(fmt.Sprintf("AddressableForName failed: %v", err))
			return nil, err
		}
	} else {
		common.LoggingClient.Info(fmt.Sprintf("Addressable %s exists", common.ServiceName))
	}

	return &addr, nil
}

func newService(dic *di.Container) *Service {
	svc = &Service{}
	svc.startTime = time.Now()
	svc.svcInfo = &container.ConfigurationFrom(dic.Get).Service
	svc.controller = container.RestControllerFrom(dic.Get)
	svc.initiazlied = false

	return svc
}

// RunningService returns the Service instance which is running
func RunningService() *Service {
	return svc
}

// DriverConfigs retrieves the driver specific configuration
func DriverConfigs() map[string]string {
	return common.CurrentConfig.Driver
}
