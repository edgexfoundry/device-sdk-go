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

	"bitbucket.org/clientcto/go-core-clients/metadataclients"
	"bitbucket.org/clientcto/go-core-domain/models"
	"github.com/tonyespy/go-xds/controller"
	"github.com/tonyespy/go-xds/data"
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
}

func (d *Daemon) attemptInit(done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	fmt.Fprintf(os.Stderr, "Trying to find ds: %s\n", d.config.Name)
	ds, err := d.sc.DeviceServiceForName(d.config.Name)

//	ds, err := d.sc.DeviceServiceForName("edgex-device-virtual")
	if err != nil {
		fmt.Fprintf(os.Stderr, "DeviceServicForName failed: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "DeviceServiceForName returned: %s\n", ds.Service.Name)
	fmt.Fprintf(os.Stderr, "DeviceServiceId is: %s\n", ds.Service.Id)
	
	// TODO: this checks if names are equal, not if the resulting ds is a valid instance
	if ds.Service.Name != d.config.Name {
		fmt.Fprintf(os.Stderr, "Failed to find ds: %s; attempts: %d\n", d.config.Name, d.initAttempts)

		// get time for Origin timestamps
		now := time.Now()
		nanos := now.UnixNano()
		millis := nanos / 1000000

		// check for addressable
		fmt.Fprintf(os.Stderr, "Trying to find addressable for: %s\n", d.config.Name)
		addr, err := d.ac.AddressableForName(d.config.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "AddressableForName: %s; failed: %v\n", d.config.Name, err)

			// don't quit, but instead try to create addressable & service
		}

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

			fmt.Fprintf(os.Stderr, "New addressable created: %v\n", addr)

			// use d.clientService to register Addressable
			err = d.ac.Add(addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Add Addressable: %s; failed: %v\n", d.config.Name, err)
				return
			}
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

		fmt.Fprintf(os.Stderr, "Adding new deviceservice: %s\n", ds.Service.Name)
		fmt.Fprintf(os.Stderr, "New deviceservice created: %v\n", ds)
		
		// use d.clientService to register the deviceservice
		err = d.sc.Add(ds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Add Deviceservice: %s; failed: %v\n", d.config.Name, err)
			return
		}

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
func (d *Daemon) Init(configFile *string) error {
	fmt.Fprintf(os.Stdout, "configuration file is: %s\n", *configFile)

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

	d.dst, err = data.NewDeviceStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating DeviceStore: %v\n", err)
		return err
	}

	// TODO: host, ports & urls are hard-coded in metadataclients
	d.ac = metadataclients.NewAddressableClient()
	d.sc = metadataclients.NewServiceClient()

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

	// TODO: add method to Service to return this...
	d.dst.Init(d.ds.Service.Id.Hex())

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
