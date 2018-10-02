// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a basic EdgeX Foundry device service implementation
// meant to be embedded in an application, similar in approach to the builtin
// net/http package.
package device

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/edgexfoundry/device-sdk-go/registry"
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

const (
	apiV1      = "/api/v1"
	colon      = ":"
	httpScheme = "http://"
	httpProto  = "HTTP"

	v1Addressable = "/api/v1/addressable"
	v1Callback    = "/api/v1/callback"
	v1Device      = "/api/v1/device"
	v1DevService  = "/api/v1/deviceservice"
	v1Event       = "/api/v1/event"
)

var (
	svc            *Service
	registryClient registry.Client
)

// A Service listens for requests and routes them to the right command
type Service struct {
	Name          string
	Version       string
	Discovery     ProtocolDiscovery
	AsyncReadings bool
	c             *Config
	initAttempts  int
	initialized   bool
	locked        bool
	useRegistry   bool
	stopped       bool
	ec            coredata.EventClient
	ac            metadata.AddressableClient
	dc            metadata.DeviceClient
	sc            metadata.DeviceServiceClient
	dpc           metadata.DeviceProfileClient
	lc            logger.LoggingClient
	vdc           coredata.ValueDescriptorClient
	scc           metadata.ScheduleClient
	scec          metadata.ScheduleEventClient
	ds            models.DeviceService
	r             *mux.Router
	scca          ScheduleCacheInterface
	cw            *Watchers
	proto         ProtocolDriver
	asyncCh       <-chan *CommandResult
}

func attemptInit(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	svc.lc.Debug("Trying to find ds: " + svc.Name)

	ds, err := svc.sc.DeviceServiceForName(svc.Name)
	if err != nil {
		svc.lc.Error(fmt.Sprintf("DeviceServicForName failed: %v", err))

		// TODO: restore if/when the issue with detecting 'not-found'
		// is resolves.  Otherwise, just log errors and move on.
		//
		// https://github.com/edgexfoundry/core-clients-go/issues/5
		// return
	}

	svc.lc.Debug("DeviceServiceForName returned: " + ds.Service.Name)
	svc.lc.Debug(fmt.Sprintf("DeviceServiceId is: %s", ds.Service.Id))

	// TODO: this checks if names are equal, not if the resulting ds is a valid instance
	if ds.Service.Name != svc.Name {
		svc.lc.Error(fmt.Sprintf("Failed to find ds: %s; attempts: %d", svc.Name, svc.initAttempts))

		// check for addressable
		svc.lc.Error(fmt.Sprintf("Trying to find addressable for: %s", svc.Name))
		addr, err := svc.ac.AddressableForName(svc.Name)
		if err != nil {
			svc.lc.Error(fmt.Sprintf("AddressableForName: %s; failed: %v", svc.Name, err))

			// don't quit, but instead try to create addressable & service
		}

		millis := time.Now().UnixNano() / int64(time.Millisecond)

		// TODO: same as above
		if addr.Name != svc.Name {
			addr = models.Addressable{
				BaseObject: models.BaseObject{
					Origin: millis,
				},
				Name:       svc.Name,
				HTTPMethod: http.MethodPost,
				Protocol:   httpProto,
				Address:    svc.c.Service.Host,
				Port:       svc.c.Service.Port,
				Path:       v1Callback,
			}
			addr.Origin = millis

			id, err := svc.ac.Add(&addr)
			if err != nil {
				svc.lc.Error(fmt.Sprintf("Add Addressable: %s; failed: %v", svc.Name, err))
				return
			}

			if len(id) != 24 || !bson.IsObjectIdHex(id) {
				svc.lc.Error("Add addressable returned invalid Id: " + id)
				return
			}

			addr.Id = bson.ObjectIdHex(id)
			svc.lc.Error("New addressable Id: " + addr.Id.Hex())
		}

		// setup the service
		ds = models.DeviceService{
			Service: models.Service{
				Name:           svc.Name,
				Labels:         svc.c.Service.Labels,
				OperatingState: "ENABLED",
				Addressable:    addr,
			},
			AdminState: "UNLOCKED",
		}

		ds.Service.Origin = millis
		id, err := svc.sc.Add(&ds)
		if err != nil {
			svc.lc.Error(fmt.Sprintf("Add Deviceservice: %s; failed: %v", svc.Name, err))
			return
		}

		if len(id) != 24 || !bson.IsObjectIdHex(id) {
			svc.lc.Error("Add deviceservice returned invalid Id: %s", id)
			return
		}

		// NOTE - this differs from Addressable and Device objects,
		// neither of which require the '.Service'prefix
		ds.Service.Id = bson.ObjectIdHex(id)
		svc.lc.Debug("New deviceservice Id: " + ds.Service.Id.Hex())

		svc.initialized = true
		svc.ds = ds
	} else {
		svc.lc.Debug(fmt.Sprintf("Found ds.Name: %s, svc.Name: %s", ds.Service.Name, svc.Name))
		svc.initialized = true
		svc.ds = ds
	}
}

func validateClientConfig() error {

	if len(svc.c.Clients[ClientMetadata].Host) == 0 {
		return fmt.Errorf("Fatal error; Host setting for Core Metadata client not configured")
	}

	if svc.c.Clients[ClientMetadata].Port == 0 {
		return fmt.Errorf("Fatal error; Port setting for Core Metadata client not configured")
	}

	if len(svc.c.Clients[ClientData].Host) == 0 {
		return fmt.Errorf("Fatal error; Host setting for Core Data client not configured")
	}

	if svc.c.Clients[ClientData].Port == 0 {
		return fmt.Errorf("Fatal error; Port setting for Core Ddata client not configured")
	}

	// TODO: validate other settings for sanity: maxcmdops, ...

	return nil
}

// Start the device service. The bool useRegisty indicates whether the registry
// should be used to read initial configuration settings. This also controls
// whether the service registers itself the registry. The profile and confDir
// are used to locate the local TOML configuration file.
func (s *Service) Start(useRegistry bool, profile string, confDir string) (err error) {
	fmt.Fprintf(os.Stdout, "Init: useRegistry: %v profile: %s confDir: %s\n",
		useRegistry, profile, confDir)
	s.useRegistry = useRegistry
	s.c, err = LoadConfig(profile, confDir)
	if err != nil {
		fmt.Printf("error loading config file: %v \n", err)
		return err
	}

	var consulMsg string
	if useRegistry {
		consulMsg = "Register in consul..."
		registryClient, err = GetConsulClient(s.Name, s.c)
		if err != nil {
			return err
		}
	} else {
		consulMsg = "Bypassing registration in consul..."
	}
	fmt.Println(consulMsg)

	// TODO: validate that metadata and core config settings are set
	err = validateClientConfig()
	if err != nil {
		return err
	}

	initDependencyClients()

	done := make(chan struct{})

	s.cw = newWatchers()
	s.scca = getScheduleCache(s.c)

	for s.initAttempts < s.c.Service.ConnectRetries && !s.initialized {
		s.initAttempts++

		if s.initAttempts > 1 {
			time.Sleep(30 * time.Second)
		}

		go attemptInit(done)
		<-done // wait for background attempt to finish
	}

	if !s.initialized {
		err = fmt.Errorf("Couldn't register to metadata service; MaxLimit reaches.")
		return err
	}

	// initialize devices, objects & profiles
	newProfileCache()
	newDeviceCache(s.ds.Service.Id.Hex())

	// TODO: initialize scheduler

	// initialize driver
	if s.AsyncReadings {
		// TODO: make channel buffer size a setting
		s.asyncCh = make(<-chan *CommandResult, 16)

		go processAsyncResults()
	}

	err = s.proto.Initialize(s, s.lc, s.asyncCh)
	if err != nil {
		s.lc.Error(fmt.Sprintf("ProtocolDriver.Initialize failure: %v; exiting.", err))
		return err
	}

	// Setup REST API
	s.r = mux.NewRouter().PathPrefix(apiV1).Subrouter()
	initStatus()
	initCommand()
	initControl()
	initUpdate()

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(s.c.Service.Timeout), "Request timed out")

	// TODO: call ListenAndServe in a goroutine

	s.lc.Info("*Service Start() called")
	s.lc.Error(http.ListenAndServe(colon+strconv.Itoa(s.c.Service.Port), s.r).Error())
	s.lc.Debug("*Service Start() exit")

	return err
}

// Stop shuts down the Service
func (s *Service) Stop(force bool) error {

	s.stopped = true
	s.proto.Stop(force)
	return nil
}

// AddDevice adds a new device to the device service.
func (s *Service) AddDevice(dev models.Device) error {
	return dc.Add(&dev)
}

// NewService create a new device service instance with the given
// name, version and ProtocolDriver, which cannot be nil.
// Note - this function is a singleton, if called more than once,
// it will alwayd return an error.
func NewService(name string, version string, proto ProtocolDriver) (*Service, error) {

	if svc != nil {
		err := fmt.Errorf("NewService: service already exists!\n")
		return nil, err
	}

	if len(name) == 0 {
		err := fmt.Errorf("NewService: empty name specified\n")
		return nil, err
	}

	if proto == nil {
		err := fmt.Errorf("NewService: no ProtocolDriver specified\n")
		return nil, err
	}

	svc = &Service{Name: name, Version: version, proto: proto}

	return svc, nil
}
