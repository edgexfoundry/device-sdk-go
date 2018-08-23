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
package device

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/bson"
)

const (
	apiV1      = "/api/v1"
	colon      = ":"
	httpScheme = "http://"
	httpProto  = "HTTP"

	coreDataServiceKey     = "edgex-core-data"
	coreMetadataServiceKey = "edgex-core-metadata"

	v1Addressable = "/api/v1/addressable"
	v1Callback    = "/api/v1/callback"
	v1Device      = "/api/v1/device"
	v1DevService  = "/api/v1/deviceservice"
	v1Event       = "/api/v1/event"
)

var (
	svc *Service
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
	ds            models.DeviceService
	r             *mux.Router
	cs            *Schedules
	cw            *Watchers
	proto         ProtocolDriver
	asyncCh       <-chan *CommandResult
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
			addr = models.Addressable{
				BaseObject: models.BaseObject{
					Origin: millis,
				},
				Name:       s.Name,
				HTTPMethod: http.MethodPost,
				Protocol:   httpProto,
				Address:    s.c.Service.Host,
				Port:       s.c.Service.Port,
				Path:       v1Callback,
			}
			addr.Origin = millis

			// use s.clientService to register Addressable
			id, err := s.ac.Add(&addr)
			if err != nil {
				s.lc.Error(fmt.Sprintf("Add Addressable: %s; failed: %v", s.Name, err))
				return
			}

			if len(id) != 24 || !bson.IsObjectIdHex(id) {
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
		id, err := s.sc.Add(&ds)
		if err != nil {
			s.lc.Error(fmt.Sprintf("Add Deviceservice: %s; failed: %v", s.Name, err))
			return
		}

		if len(id) != 24 || !bson.IsObjectIdHex(id) {
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

	if len(s.c.Clients[ClientMetadata].Host) == 0 {
		return fmt.Errorf("Fatal error; Host setting for Core Metadata client not configured")
	}

	if s.c.Clients[ClientMetadata].Port == 0 {
		return fmt.Errorf("Fatal error; Port setting for Core Metadata client not configured")
	}

	if len(s.c.Clients[ClientData].Host) == 0 {
		return fmt.Errorf("Fatal error; Host setting for Core Data client not configured")
	}

	if s.c.Clients[ClientData].Port == 0 {
		return fmt.Errorf("Fatal error; Port setting for Core Ddata client not configured")
	}

	// TODO: validate other settings for sanity: maxcmdops, ...

	return nil
}

func buildAddr(host string, port string) string {
	var buffer bytes.Buffer

	buffer.WriteString(httpScheme)
	buffer.WriteString(host)
	buffer.WriteString(colon)
	buffer.WriteString(port)

	return buffer.String()
}

// Start the device service. The bool useRegisty indicates whether the registry
// should be used to read initial configuration settings. This also controls
// whether the service registers itself the registry. The profile and confDir
// are used to locate the local TOML configuration file.
func (s *Service) Start(useRegistry bool, profile string, confDir string) (err error) {
	fmt.Fprintf(os.Stdout, "Init: useRegistry: %v profile: %s confDir: %s\n",
		useRegistry, profile, confDir)

	s.c, err = LoadConfig(profile, confDir)
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

	s.cw = newWatchers()
	s.cs = newSchedules(s.c)

	// initialize Core Metadata clients
	metaPort := strconv.Itoa(s.c.Clients[ClientMetadata].Port)
	metaHost := s.c.Clients[ClientMetadata].Host
	metaAddr := buildAddr(metaHost, metaPort)
	metaPath := v1Addressable
	metaURL := metaAddr + metaPath

	params := types.EndpointParams{
		// TODO: Can't use edgex-go internal constants!
		//ServiceKey:internal.CoreMetaDataServiceKey,
		ServiceKey:  coreMetadataServiceKey,
		Path:        metaPath,
		UseRegistry: s.useRegistry,
		Url:         metaURL}

	s.ac = metadata.NewAddressableClient(params, types.Endpoint{})

	params.Path = v1Device
	params.Url = metaAddr + params.Path
	s.dc = metadata.NewDeviceClient(params, types.Endpoint{})

	params.Path = v1DevService
	params.Url = metaAddr + params.Path
	s.sc = metadata.NewDeviceServiceClient(params, types.Endpoint{})

	params.Path = v1Deviceprofile
	params.Url = metaAddr + params.Path
	s.dpc = metadata.NewDeviceProfileClient(params, types.Endpoint{})

	// initialize Core Data clients
	dataPort := strconv.Itoa(s.c.Clients[ClientData].Port)
	dataHost := s.c.Clients[ClientData].Host
	dataAddr := buildAddr(dataHost, dataPort)
	dataPath := v1Event
	dataURL := dataAddr + dataPath

	params.ServiceKey = coreDataServiceKey
	params.Path = dataPath
	params.UseRegistry = s.useRegistry
	params.Url = dataURL

	s.ec = coredata.NewEventClient(params, types.Endpoint{})

	params.Path = v1Valuedescriptor
	params.Url = dataAddr + dataPath
	s.vdc = coredata.NewValueDescriptorClient(params, types.Endpoint{})

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

	// initialize devices, objects & profiles
	newProfileCache()
	newDeviceCache(s.ds.Service.Id.Hex())

	// TODO: initialize scheduler

	// initialize driver
	if s.AsyncReadings {
		// TODO: make channel buffer size a setting
		s.asyncCh = make(<-chan *CommandResult, 16)

		go s.processAsyncResults()
	}

	err = s.proto.Initialize(s.lc, s.asyncCh)
	if err != nil {
		s.lc.Error(fmt.Sprintf("ProtocolDriver.Initialize failure: %v; exiting.", err))
		return err
	}

	// Setup REST API
	s.r = mux.NewRouter().PathPrefix(apiV1).Subrouter()
	initStatus(s)
	initCommand(s)
	initService(s)
	initUpdate(s)

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

// NewService create a new device service instance with the given
// name, version and ProtocolDriver, which cannot be nil.
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
