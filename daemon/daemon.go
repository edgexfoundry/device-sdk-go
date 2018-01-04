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

package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"bitbucket.org/tonyespy/gxds/controller"
	"bitbucket.org/tonyespy/gxds/data"
	"bitbucket.org/tonyespy/gxds"
	"github.com/edgexfoundry/core-clients-go/metadataclients"
	"github.com/edgexfoundry/core-domain-go/models"
	"gopkg.in/mgo.v2/bson"
)

var (
	// TODO: grab settings from daemon-config.json OR Consul
	metaPort           string = ":48081"
	metaHost           string = "localhost"
	metaAddressableUrl string = "http://" + metaHost + metaPort + "/api/v1/addressable"
	metaServiceUrl     string = "http://" + metaHost + metaPort + "/api/v1/deviceservice"
)

type configFile struct {
	Name            string
	Host            string
	Port            int
	Labels          []string
	Timeout         int
	OpenMessage     string
	ConnectRetries  int
	ConnectWait     int
	ConnectInterval int
	MaxLimit        int
	HeartBeatTime   int
	DataTransform   bool
	MetaHost        string
	MetaPort        int
	CoreHost        string
	CorePort        int
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
	sc            metadataclients.ServiceClient
	ds            models.DeviceService
	mux           *controller.Mux
	dst           *data.DeviceStore
	ost           *data.ObjectStore
	pst           *data.ProfileStore
	wst           *data.WatcherStore
	proto         gxds.ProtocolHandler
}

func (d *Daemon) attemptInit(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	fmt.Fprintf(os.Stderr, "Trying to find ds: %s\n", d.config.Name)
	ds, err := d.sc.DeviceServiceForName(d.config.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DeviceServicForName failed: %v\n", err)

		// TODO: restore if/when the issue with detecting 'not-found'
	        // is resolved.  Otherwise, just log errors and move on.
		//
	        // https://github.com/edgexfoundry/core-clients-go/issues/5
		// return
	}

	fmt.Fprintf(os.Stderr, "DeviceServiceForName returned: %s\n", ds.Service.Name)
	fmt.Fprintf(os.Stderr, "DeviceServiceId is: %s\n", ds.Service.Id)
	
	// TODO: this checks if names are equal, not if the resulting ds is a valid instance
	if ds.Service.Name != d.config.Name {
		fmt.Fprintf(os.Stderr, "Failed to find ds: %s; attempts: %d\n", d.config.Name, d.initAttempts)

		// check for addressable
		fmt.Fprintf(os.Stderr, "Trying to find addressable for: %s\n", d.config.Name)
		addr, err := d.ac.AddressableForName(d.config.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "AddressableForName: %s; failed: %v\n", d.config.Name, err)

			// don't quit, but instead try to create addressable & service
		}

		millis := time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Microsecond)

		// TODO: same as above
		if addr.Name != d.config.Name {
			// TODO: does HTTPMethod need to be specified?
			addr = models.Addressable{
				BaseObject: models.BaseObject{
					Origin:     millis,
				},
				Name:       d.config.Name,
				HTTPMethod: "POST",
				Protocol:   "HTTP",
				Address:    d.config.Host,
				Port:       d.config.Port,
				Path:       "/api/v1/callback",
			}
			addr.Origin = millis

			// use d.clientService to register Addressable
			id, err := d.ac.Add(&addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Add Addressable: %s; failed: %v\n", d.config.Name, err)
				return
			}

			// TODO: add back length check in from non-public metadata-clients logic
			//
			// if len(bodyBytes) != 24 || !bson.IsObjectIdHex(bodyString) {
			//
			if !bson.IsObjectIdHex(id) {
				fmt.Fprintf(os.Stderr, "Add addressable returned invalid Id: %s\n", id)
				return
			}

			addr.Id = bson.ObjectIdHex(id)
			fmt.Fprintf(os.Stdout, "New addressable Id: %s\n", addr.Id.Hex())
		}

		// setup the service
		ds = models.DeviceService{
			Service: models.Service{
				Name:           d.config.Name,
				Labels:         d.config.Labels,
				OperatingState: "enabled",
				Addressable:    addr,
			},
			AdminState:     "unlocked",				
		}

		ds.Service.Origin = millis

		fmt.Fprintf(os.Stdout, "Adding new deviceservice: %s\n", ds.Service.Name)
		fmt.Fprintf(os.Stdout, "New deviceservice: %v\n", ds)
		
		// use d.clientService to register the deviceservice
		id, err := d.sc.Add(&ds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Add Deviceservice: %s; failed: %v\n", d.config.Name, err)
			return
		}

		// TODO: add back length check in from non-public metadata-clients logic
		//
		// if len(bodyBytes) != 24 || !bson.IsObjectIdHex(bodyString) {
		//
		if !bson.IsObjectIdHex(id) {
			fmt.Fprintf(os.Stderr, "Add deviceservice returned invalid Id: %s\n", id)
			return
		}

		// NOTE - this differs from Addressable and Device objects,
		// neither of which require the '.Service'prefix
		ds.Service.Id = bson.ObjectIdHex(id)
		fmt.Fprintf(os.Stdout, "New deviceservice Id: %s\n", ds.Service.Id.Hex())

		d.initialized = true
		d.ds = ds
	} else {
		fmt.Fprintf(os.Stderr, "Found ds.Name: %s, d.config.Name: %s\n", ds.Service.Name, d.config.Name)
		d.initialized = true
		d.ds = ds
	}
}

func (d *Daemon) loadConfig(configPath *string) error {
	f, err := os.Open(*configPath)
	defer f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config file: %s; open failed: %v\n", *configPath, err)
		return err
	}

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

	done := make(chan struct{})

	d.mux, err = controller.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading starting controller: %v\n", err)
		return err
	}

	d.proto = proto
	d.pst = data.NewProfileStore()
	d.wst = data.NewWatcherStore()
	d.ost = data.NewObjectStore()
	d.dst = data.NewDeviceStore(proto)

	// TODO: host, ports & urls are hard-coded in metadataclients
	d.ac = metadataclients.NewAddressableClient(metaAddressableUrl)
	d.sc = metadataclients.NewServiceClient(metaServiceUrl)

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
	d.dst.Init(d.ds.Service.Id.Hex())

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
