// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// Package service(service?) implements the core logic of a device service,
// which include loading configuration, handling service registration,
// creation of object caches, REST APIs, and basic service functionality.
// Clients of this package must provide concrete implementations of the
// device-specific interfaces (e.g. ProtocolDriver).
//
package service

import (
	"github.com/tonyespy/gxds"
	"github.com/tonyespy/gxds/cache"

	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/core/clients/metadata"
	"github.com/edgexfoundry/edgex-go/core/clients/types"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

// TODO:
//  * add consul registration support
//  * design REST API framework
//  * design Protocol framework
//  * re-name?  service --> baseservice

// A Service listens for requests and routes them to the right command
type Service struct {
	Name         string
	Version      string
	c            *gxds.Config
	initAttempts int
	initialized  bool
	locked       bool
	useRegistry  bool
	ac           metadata.AddressableClient
	lc           logger.LoggingClient
	sc           metadata.DeviceServiceClient
	ds           models.DeviceService
	r            *mux.Router
	cd           *cache.Devices
	co           *cache.Objects
	cp           *cache.Profiles
	cs           *cache.Schedules
	cw           *cache.Watchers
	proto        gxds.ProtocolDriver
}

func (s *Service) attemptInit(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	s.lc.Debug("Trying to find ds: " + s.Name)

	ds, err := s.sc.DeviceServiceForName(s.Name)
	if err != nil {
		s.lc.Error(fmt.Sprintf("DeviceServicForName failed: %v", err))

		// TODO: restore if/when the issue with detecting 'not-found'
		// is resolves.  Otherwise, just log errors and move on.
		//
		// https://github.com/edgexfoundry/core-clients-go/issues/5
		// return
	}

	s.lc.Debug("DeviceServiceForName returned: " + ds.Service.Name)
	s.lc.Debug(fmt.Sprintf("DeviceServiceId is: %s", ds.Service.Id))

	// TODO: this checks if names are equal, not if the resulting ds is a valid instance
	if ds.Service.Name != s.Name {
		s.lc.Error(fmt.Sprintf("Failed to find ds: %s; attempts: %d", s.Name, s.initAttempts))

		// check for addressable
		s.lc.Error(fmt.Sprintf("Trying to find addressable for: %s", s.Name))
		addr, err := s.ac.AddressableForName(s.Name)
		if err != nil {
			s.lc.Error(fmt.Sprintf("AddressableForName: %s; failed: %v", s.Name, err))

			// don't quit, but instead try to create addressable & service
		}

		millis := time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)

		// TODO: same as above
		if addr.Name != s.Name {
			// TODO: does HTTPMethod need to be specified?
			addr = models.Addressable{
				BaseObject: models.BaseObject{
					Origin: millis,
				},
				Name:       s.Name,
				HTTPMethod: "POST",
				Protocol:   "HTTP",
				Address:    s.c.Service.Host,
				Port:       s.c.Service.Port,
				Path:       "/api/v1/callback",
			}
			addr.Origin = millis

			// use s.clientService to register Addressable
			id, err := s.ac.Add(&addr)
			if err != nil {
				s.lc.Error(fmt.Sprintf("Add Addressable: %s; failed: %v", s.Name, err))
				return
			}

			// TODO: add back length check in from non-public metadata-clients logic
			//
			// if len(bodyBytes) != 24 || !bson.IsObjectIdHex(bodyString) {
			//
			if !bson.IsObjectIdHex(id) {
				s.lc.Error("Add addressable returned invalid Id: " + id)
				return
			}

			addr.Id = bson.ObjectIdHex(id)
			s.lc.Error("New addressable Id: " + addr.Id.Hex())
		}

		// setup the service
		ds = models.DeviceService{
			Service: models.Service{
				Name:           s.Name,
				Labels:         s.c.Service.Labels,
				OperatingState: "ENABLED",
				Addressable:    addr,
			},
			AdminState: "UNLOCKED",
		}

		ds.Service.Origin = millis

		s.lc.Debug("Adding new deviceservice: " + ds.Service.Name)
		s.lc.Debug(fmt.Sprintf("New deviceservice: %v", ds))

		// use s.clientService to register the deviceservice
		id, err := s.sc.Add(&ds)
		if err != nil {
			s.lc.Error(fmt.Sprintf("Add Deviceservice: %s; failed: %v", s.Name, err))
			return
		}

		// TODO: add back length check in from non-public metadata-clients logic
		//
		// if len(bodyBytes) != 24 || !bson.IsObjectIdHex(bodyString) {
		//
		if !bson.IsObjectIdHex(id) {
			s.lc.Error("Add deviceservice returned invalid Id: %s", id)
			return
		}

		// NOTE - this differs from Addressable and Device objects,
		// neither of which require the '.Service'prefix
		ds.Service.Id = bson.ObjectIdHex(id)
		s.lc.Debug("New deviceservice Id: " + ds.Service.Id.Hex())

		s.initialized = true
		s.ds = ds
	} else {
		s.lc.Debug(fmt.Sprintf("Found ds.Name: %s, s.Name: %s", ds.Service.Name, s.Name))
		s.initialized = true
		s.ds = ds
	}
}

func (s *Service) validateClientConfig() error {

	if len(s.c.Clients["Metadata"].Host) == 0 {
		return fmt.Errorf("Fatal error; Host setting for Core Metadata client not configured")
	}

	if s.c.Clients["Metadata"].Port == 0 {
		return fmt.Errorf("Fatal error; Port setting for Core Metadata client not configured")
	}

	if len(s.c.Clients["Data"].Host) == 0 {
		return fmt.Errorf("Fatal error; Host setting for Core Data client not configured")
	}

	if s.c.Clients["Data"].Port == 0 {
		return fmt.Errorf("Fatal error; Port setting for Core Ddata client not configured")
	}

	// TODO: validate other settings for sanity: maxcmdops, ...

	return nil
}

// Initialize the Service
func (s *Service) Init(useRegistry bool, profile string, confDir string, proto gxds.ProtocolDriver) (err error) {
	fmt.Fprintf(os.Stdout, "Init: useRegistry: %v profile: %s confDir: %s proto is: %v\n",
		useRegistry, profile, confDir, proto)

	// TODO: check if proto is nil, and fail...

	s.c, err = gxds.LoadConfig(profile, confDir)
	if err != nil {
		s.lc.Error(fmt.Sprintf("error loading config file: %v\n", err))
		return err
	}

	// TODO: add useRegistry logic

	// TODO: validate that metadata and core config settings are set
	err = s.validateClientConfig()
	if err != nil {
		return err
	}

	var remoteLog bool = false
	var logTarget string

	if s.c.Logging.RemoteURL == "" {
		logTarget = s.c.Logging.File
	} else {
		remoteLog = true
		logTarget = s.c.Logging.RemoteURL
	}

	s.lc = logger.NewClient(s.Name, remoteLog, logTarget)

	done := make(chan struct{})

	s.proto = proto
	s.cp = cache.NewProfiles(s.c, s.lc, s.useRegistry)
	s.cw = cache.NewWatchers()
	s.co = cache.NewObjects(s.c, s.lc)
	s.cd = cache.NewDevices(s.c, s.lc, s.useRegistry)
	s.cs = cache.NewSchedules(s.c)

	// set up clients
	metaPort := strconv.Itoa(s.c.Clients["Metadata"].Port)
	metaHost := s.c.Clients["Metadata"].Host
	metaAddr := "http://" + metaHost + ":" + metaPort
	metaPath := "/api/v1/addressable"
	metaURL := metaAddr + metaPath

	// TODO: edgex-go - endpoint paths shouldn't be in config files!!!

	// Create metadata clients
	params := types.EndpointParams{
		// TODO: Can't use edgex-go internal constants!
		//ServiceKey:internal.CoreMetaDataServiceKey,
		ServiceKey:  "edgex-core-metadata",
		Path:        metaPath,
		UseRegistry: s.useRegistry,
		Url:         metaURL}

	s.ac = metadata.NewAddressableClient(params, types.Endpoint{})

	params.Path = "/api/v1/deviceservice"
	params.Url = metaAddr + params.Path
	s.sc = metadata.NewDeviceServiceClient(params, types.Endpoint{})

	for s.initAttempts < s.c.Service.ConnectRetries && !s.initialized {
		s.initAttempts++

		if s.initAttempts > 1 {
			time.Sleep(30 * time.Second)
		}

		go s.attemptInit(done)
		<-done // wait for background attempt to finish
	}

	if !s.initialized {
		err = fmt.Errorf("Couldn't register to metadata service; MaxLimit reaches.")
		return err
	}

	// initialize devicestore
	// TODO: add method to Service to return this...
	s.cd.Init(s.ds.Service.Id.Hex())

	// TODO: initialize scheduler

	// initialize driver
	s.proto.Initialize(s.lc)

	// Setup REST API
	s.r = mux.NewRouter().PathPrefix("/api/v1").Subrouter()
	initStatus(s)
	initCommand(s)
	initService(s)
	initUpdate(s)

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(s.c.Service.Timeout), "Request timed out")

	return err
}

// Start the Service
func (s *Service) Start() {
	s.lc.Info("*Service Start() called")
	s.lc.Error(http.ListenAndServe(":"+strconv.Itoa(s.c.Service.Port), s.r).Error())
	s.lc.Debug("*Service Start() exit")
}

// Stop shuts down the Service
func (s *Service) Stop() error {
	return nil
}

// New Service
// TODO: re-factor to make this a singleton
func New(name string) (*Service, error) {
	return &Service{Name: name}, nil
}
