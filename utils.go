// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package device

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	consulapi "github.com/hashicorp/consul/api"
)

const consulStatusPath = "/v1/agent/self"

var (
	consul *consulapi.Client = nil
	// Need to set timeout because it hang until server close connection
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	netClient = &http.Client{Timeout: time.Second * 10}
)

// Return Consul client instance
func getConsulClient(config *Config) (*consulapi.Client, error) {
	consulUrl := config.Registry.Host + ":" + strconv.Itoa(config.Registry.Port)
	fails := 0
	for fails < config.Registry.FailLimit {
		// http.Get return error in case of wrong HTTP method or invalid URL
		// so we need to check for invalid status.
		response, err := netClient.Get(consulUrl + consulStatusPath)
		if err != nil {
			fmt.Println(err.Error())
			time.Sleep(time.Second * time.Duration(config.Registry.FailWaitTime))
			fails++
			continue
		}

		if response.StatusCode >= 200 && response.StatusCode < 300 {
			break
		} else {
			return nil, errors.New("Bad response from Consul service")
		}
	}
	if fails >= config.Registry.FailLimit {
		return nil, errors.New("Cannot get connection to Consul")
	}

	defaultConfig := consulapi.DefaultConfig()
	defaultConfig.Address = consulUrl

	consul, err := consulapi.NewClient(defaultConfig)
	if err != nil {
		return nil, err
	} else {
		return consul, nil
	}
}

// Register service in Consul and add health check
func registerDeviceService(consul *consulapi.Client, deviceServiceName string, config *Config) error {
	var err error

	// Register device service
	err = consul.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		Name:    deviceServiceName,
		Address: config.Service.Host,
		Port:    config.Service.Port,
	})
	if err != nil {
		return err
	}

	checkAddress := "http://" + config.Registry.Host + ":" + strconv.Itoa(config.Registry.Port) + config.Registry.CheckPath
	// Register the Health Check
	err = consul.Agent().CheckRegister(&consulapi.AgentCheckRegistration{
		Name:      "Health Check: " + deviceServiceName,
		Notes:     "Check the health of the API",
		ServiceID: deviceServiceName,
		AgentServiceCheck: consulapi.AgentServiceCheck{
			HTTP:     checkAddress,
			Interval: config.Registry.CheckInterval,
		},
	})

	return err
}

func connectToConsul(deviceServiceName string, config *Config) error {
	var err error

	consul, err = getConsulClient(config)
	if err != nil {
		return err
	}

	err = registerDeviceService(consul, deviceServiceName, config)
	if err != nil {
		return err
	}

	return err
}

// LoadConfig loads the local configuration file based upon the
// specified parameters and returns a pointer to the global Config
// struct which holds all of the local configuration settings for
// the DS.
func LoadConfig(profile string, configDir string) (config *Config, err error) {
	var name string

	if len(configDir) == 0 {
		configDir = "./res/"
	}

	if len(profile) > 0 {
		name = "configuration-" + profile + ".toml"
	} else {
		name = "configuration.toml"
	}

	path := configDir + name

	// As the toml package can panic if TOML is invalid,
	// or elements are found that don't match members of
	// the given struct, use a defered func to recover
	// from the panic and output a useful error.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("could not load configuration file; invalid TOML (%s)", path)
		}
	}()

	config = &Config{}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not load configuration file (%s): %v", path, err.Error())
	}

	// Decode the configuration from TOML
	//
	// TODO: invalid input can cause a SIGSEGV fatal error (INVESTIGATE)!!!
	//       - test missing keys, keys with wrong type, ...
	err = toml.Unmarshal(contents, config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse configuration file (%s): %v", path, err.Error())
	}

	return config, nil
}
