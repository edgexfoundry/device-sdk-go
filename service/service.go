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
// device-specific interfaces (e.g. ProtocolHandler).
//
package service

import (
	"bitbucket.org/tonyespy/gxds"
	"bitbucket.org/tonyespy/gxds/cache"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/core/clients/metadataclients"
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
	Version      string
	Config       *gxds.Config
	initAttempts int
	initialized  bool
	locked       bool
	ac           metadataclients.AddressableClient
	lc           logger.LoggingClient
	sc           metadataclients.ServiceClient
	ds           models.DeviceService
	r            *mux.Router
	cd           *cache.Devices
	co           *cache.Objects
	cp           *cache.Profiles
	cs           *cache.Schedules
	cw           *cache.Watchers
	proto        gxds.ProtocolHandler
}

func (s *Service) attemptInit(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	// TODO: service name should NOT be a configuration paramter
	// but instead must be hard-coded in the DS
	s.lc.Debug("Trying to find ds: " + s.Config.ServiceName)

	ds, err := s.sc.DeviceServiceForName(s.Config.ServiceName)
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
	if ds.Service.Name != s.Config.ServiceName {
		s.lc.Error(fmt.Sprintf("Failed to find ds: %s; attempts: %d",
			s.Config.ServiceName, s.initAttempts))

		// check for addressable
		fmt.Fprintf(os.Stderr, "Trying to find addressable for: %s", s.Config.ServiceName)
		addr, err := s.ac.AddressableForName(s.Config.ServiceName)
		if err != nil {
			s.lc.Error(fmt.Sprintf("AddressableForName: %s; failed: %v", s.Config.ServiceName, err))

			// don't quit, but instead try to create addressable & service
		}

		millis := time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)

		// TODO: same as above
		if addr.Name != s.Config.ServiceName {
			// TODO: does HTTPMethod need to be specified?
			addr = models.Addressable{
				BaseObject: models.BaseObject{
					Origin: millis,
				},
				Name:       s.Config.ServiceName,
				HTTPMethod: "POST",
				Protocol:   "HTTP",
				Address:    s.Config.ServiceHost,
				Port:       s.Config.ServicePort,
				Path:       "/api/v1/callback",
			}
			addr.Origin = millis

			// use s.clientService to register Addressable
			id, err := s.ac.Add(&addr)
			if err != nil {
				s.lc.Error(fmt.Sprintf("Add Addressable: %s; failed: %v",
					s.Config.ServiceName, err))
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
				Name:           s.Config.ServiceName,
				Labels:         s.Config.Labels,
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
			s.lc.Error(fmt.Sprintf("Add Deviceservice: %s; failed: %v", s.Config.ServiceName, err))
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
		s.lc.Debug(fmt.Sprintf("Found ds.Name: %s, s.Config.ServiceName: %s",
			ds.Service.Name, s.Config.ServiceName))
		s.initialized = true
		s.ds = ds
	}
}

// Initialize the Service
func (s *Service) Init(configFile *string, proto gxds.ProtocolHandler) (err error) {
	fmt.Fprintf(os.Stdout, "configuration file is: %s\n", *configFile)
	fmt.Fprintf(os.Stdout, "proto is: %v\n", proto)

	// TODO: check if proto is nil, and fail...

	s.Config, err = gxds.LoadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config file: %v\n", err)
		return err
	}

	var remoteLog bool = false
	var logTarget string

	if s.Config.LoggingRemoteURL == "" {
		logTarget = s.Config.LoggingFile
	} else {
		remoteLog = true
		logTarget = s.Config.LoggingRemoteURL
	}

	s.lc = logger.NewClient(s.Config.ServiceName, remoteLog, logTarget)

	done := make(chan struct{})

	s.r = mux.NewRouter()
	initCommand(s)
	initStatus(s.r)
	initService(s.r)
	initUpdate(s.r)

	s.proto = proto
	s.cp = cache.NewProfiles(s.Config)
	s.cw = cache.NewWatchers()
	s.co = cache.NewObjects()
	s.cd = cache.NewDevices(s.Config, proto)
	s.cs = cache.NewSchedules(s.Config)

	// set up clients
	metaPort := strconv.Itoa(s.Config.MetadataPort)
	s.ac = metadataclients.NewAddressableClient("http://" + s.Config.MetadataHost + metaPort + "/api/v1/addressable")
	s.sc = metadataclients.NewServiceClient("http://" + s.Config.MetadataHost + metaPort + "/api/v1/deviceservice")

	for s.initAttempts < s.Config.ConnectRetries && !s.initialized {
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
	// TODO: configure gorillamux

	return err
}

// Start the Service
func (s *Service) Start() {
}

// Stop shuts down the Service
func (s *Service) Stop() error {
	return nil
}

// New Service
// TODO: re-factor to make this a singleton
func New() (*Service, error) {
	return &Service{}, nil
}
