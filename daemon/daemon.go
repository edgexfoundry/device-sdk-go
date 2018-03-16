// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package daemon(service?) implements the core logic of a device service,
// which include loading configuration, handling service registration,
// creation of object caches, REST APIs, and basic daemon functionality.
// Clients of this package must provide concrete implementations of the
// device-specific interfaces (e.g. ProtocolHandler).
package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/tonyespy/gxds/controller"
	"bitbucket.org/tonyespy/gxds/cache"
	"bitbucket.org/tonyespy/gxds"
	"github.com/edgexfoundry/core-clients-go/metadataclients"
	"github.com/edgexfoundry/core-domain-go/models"
	logger "github.com/edgexfoundry/edgex-go/support/logging-client"
	"gopkg.in/mgo.v2/bson"
)

type configFile struct {
	ServiceName      string
	ServiceHost      string
	ServicePort      int
	Labels           []string
	Timeout          int
	OpenMessage      string
	ConnectRetries   int
	ConnectWait      int
	ConnectInterval  int
	MaxLimit         int
	HeartBeatTime    int
	DataTransform    bool
	MetadbHost       string
	MetadbPort       int
	CoreHost         string
	CorePort         int
	LoggingFile      string
	LoggingRemoteURL string
}

// TODO:
//  * add consul registration support
//  * design REST API framework
//  * design Protocol framework
//  * re-name?  daemon --> baseservice

// A Daemon listens for requests and routes them to the right command
type Daemon struct {
	Version       string
	config        configFile
	initAttempts  int
	initialized   bool
	ac            metadataclients.AddressableClient
	lc            logger.LoggingClient
	sc            metadataclients.ServiceClient
	ds            models.DeviceService
	mux           *controller.Mux
	cd            *cache.Devices
	co            *cache.Objects
	cp            *cache.Profiles
	cw            *cache.Watchers
	proto         gxds.ProtocolHandler
}

func (d *Daemon) attemptInit(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	// TODO: service name should NOT be a configuration paramter
	// but instead must be hard-coded in the DS
	d.lc.Debug("Trying to find ds: " + d.config.ServiceName)

	ds, err := d.sc.DeviceServiceForName(d.config.ServiceName)
	if err != nil {
		d.lc.Error(fmt.Sprintf("DeviceServicForName failed: %v", err))

		// TODO: restore if/when the issue with detecting 'not-found'
	        // is resolved.  Otherwise, just log errors and move on.
		//
	        // https://github.com/edgexfoundry/core-clients-go/issues/5
		// return
	}

	d.lc.Debug("DeviceServiceForName returned: " + ds.Service.Name)
	d.lc.Debug(fmt.Sprintf("DeviceServiceId is: %s", ds.Service.Id))

	// TODO: this checks if names are equal, not if the resulting ds is a valid instance
	if ds.Service.Name != d.config.ServiceName {
		d.lc.Error(fmt.Sprintf("Failed to find ds: %s; attempts: %d",
			d.config.ServiceName, d.initAttempts))

		// check for addressable
		fmt.Fprintf(os.Stderr, "Trying to find addressable for: %s", d.config.ServiceName)
		addr, err := d.ac.AddressableForName(d.config.ServiceName)
		if err != nil {
			d.lc.Error(fmt.Sprintf("AddressableForName: %s; failed: %v", d.config.ServiceName, err))

			// don't quit, but instead try to create addressable & service
		}

		millis := time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)

		// TODO: same as above
		if addr.Name != d.config.ServiceName {
			// TODO: does HTTPMethod need to be specified?
			addr = models.Addressable{
				BaseObject: models.BaseObject{
					Origin:     millis,
				},
				Name:       d.config.ServiceName,
				HTTPMethod: "POST",
				Protocol:   "HTTP",
				Address:    d.config.ServiceHost,
				Port:       d.config.ServicePort,
				Path:       "/api/v1/callback",
			}
			addr.Origin = millis

			// use d.clientService to register Addressable
			id, err := d.ac.Add(&addr)
			if err != nil {
				d.lc.Error(fmt.Sprintf("Add Addressable: %s; failed: %v",
					d.config.ServiceName, err))
				return
			}

			// TODO: add back length check in from non-public metadata-clients logic
			//
			// if len(bodyBytes) != 24 || !bson.IsObjectIdHex(bodyString) {
			//
			if !bson.IsObjectIdHex(id) {
				d.lc.Error("Add addressable returned invalid Id: " + id)
				return
			}

			addr.Id = bson.ObjectIdHex(id)
			d.lc.Error("New addressable Id: " + addr.Id.Hex())
		}

		// setup the service
		ds = models.DeviceService{
			Service: models.Service{
				Name:           d.config.ServiceName,
				Labels:         d.config.Labels,
				OperatingState: "ENABLED",
				Addressable:    addr,
			},
			AdminState:     "UNLOCKED",
		}

		ds.Service.Origin = millis

		d.lc.Debug("Adding new deviceservice: " + ds.Service.Name)
		d.lc.Debug(fmt.Sprintf("New deviceservice: %v", ds))
		
		// use d.clientService to register the deviceservice
		id, err := d.sc.Add(&ds)
		if err != nil {
			d.lc.Error(fmt.Sprintf("Add Deviceservice: %s; failed: %v", d.config.ServiceName, err))
			return
		}

		// TODO: add back length check in from non-public metadata-clients logic
		//
		// if len(bodyBytes) != 24 || !bson.IsObjectIdHex(bodyString) {
		//
		if !bson.IsObjectIdHex(id) {
			d.lc.Error("Add deviceservice returned invalid Id: %s", id)
			return
		}

		// NOTE - this differs from Addressable and Device objects,
		// neither of which require the '.Service'prefix
		ds.Service.Id = bson.ObjectIdHex(id)
		d.lc.Debug("New deviceservice Id: " + ds.Service.Id.Hex())

		d.initialized = true
		d.ds = ds
	} else {
		d.lc.Debug(fmt.Sprintf("Found ds.Name: %s, d.config.ServiceName: %s",
			ds.Service.Name, d.config.ServiceName))
		d.initialized = true
		d.ds = ds
	}
}

func (d *Daemon) loadConfig(configPath *string) error {
	f, err := os.Open(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config file: %s; open failed: %v\n", *configPath, err)
		return err
	}
	defer f.Close()

	fmt.Fprintf(os.Stdout, "config file opened: %s\n", *configPath)

	jsonParser := json.NewDecoder(f)
	err = jsonParser.Decode(&(d.config))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	fmt.Fprintf(os.Stdout, "name: %v\n", d.config)

	return nil
}

// Initialize the Daemon
func (d *Daemon) Init(configFile *string, proto gxds.ProtocolHandler) error {
	fmt.Fprintf(os.Stdout, "configuration file is: %s\n", *configFile)
	fmt.Fprintf(os.Stdout, "proto is: %v\n", proto)

	// TODO: check if proto is nil, and fail...

	err := d.loadConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config file: %v\n", err)
		return err
	}

	var remoteLog bool = false
	var logTarget string

	if d.config.LoggingRemoteURL == "" {
	        logTarget = d.config.LoggingFile
	} else {
		remoteLog = true
		logTarget = d.config.LoggingRemoteURL
	}

	d.lc = logger.NewClient(d.config.ServiceName, remoteLog, logTarget)

	done := make(chan struct{})

	d.mux, err = controller.New()
	if err != nil {
		d.lc.Error(fmt.Sprintf("error loading starting controller: %v", err))
		return err
	}

	d.proto = proto
	d.cp = cache.NewProfiles()
	d.cw = cache.NewWatchers()
	d.co = cache.NewObjects()
	d.cd = cache.NewDevices(proto)

	// set up clients
	metaPort := strconv.Itoa(d.config.MetadbPort)
	d.ac = metadataclients.NewAddressableClient("http://" + d.config.MetadbHost + metaPort + "/api/v1/addressable")
	d.sc = metadataclients.NewServiceClient("http://" + d.config.MetadbHost + metaPort + "/api/v1/deviceservice")

	for d.initAttempts < d.config.ConnectRetries && !d.initialized {
		d.initAttempts++

		if d.initAttempts > 1 {
			time.Sleep(30 * time.Second)
		}

		go d.attemptInit(done)
		<-done // wait for background attempt to finish
	}

	if !d.initialized {
		err = fmt.Errorf("Couldn't register to metadata service; MaxLimit reached.")
		return err
	}

	// initialize devicestore
	// TODO: add method to Service to return this...
	d.cd.Init(d.ds.Service.Id.Hex())

	// TODO: initialize scheduler
	// TODO: configure gorillamux

	return err
}

// Start the Daemon
func (d *Daemon) Start() {
}

// Stop shuts down the Daemon
func (d *Daemon) Stop() error {
	return nil
}

// New Daemon
// TODO: re-factor to make this a singleton
func New() (*Daemon, error) {
	return &Daemon{}, nil
}
