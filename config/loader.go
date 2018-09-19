// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package config

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/edgexfoundry/device-sdk-go/common"
	"github.com/edgexfoundry/device-sdk-go/registry"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)


const consulStatusPath = "/v1/agent/self"

var (
	// Need to set timeout because it hang until server close connection
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	netClient = &http.Client{Timeout: time.Second * 10}
	RegistryClient registry.Client
)

// LoadConfig loads the local configuration file based upon the
// specified parameters and returns a pointer to the global RegistryConfig
// struct which holds all of the local configuration settings for
// the DS.
func LoadConfig(useRegistry bool, profile string, confDir string) (config *common.Config, err error) {
	fmt.Fprintf(os.Stdout, "Init: useRegistry: %v profile: %s confDir: %s\n",
		useRegistry, profile, confDir)
	var confName string

	if len(confDir) == 0 {
		confDir = "./res/"
	}

	if len(profile) > 0 {
		confName = "configuration-" + profile + ".toml"
	} else {
		confName = "configuration.toml"
	}

	path := confDir + confName

	// As the toml package can panic if TOML is invalid,
	// or elements are found that don't match members of
	// the given struct, use a defered func to recover
	// from the panic and output a useful error.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("could not load configuration file; invalid TOML (%s)", path)
		}
	}()

	config = &common.Config{}
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

	var consulMsg string
	if useRegistry {
		consulMsg = "Register in consul..."
		RegistryClient, err = GetConsulClient(common.ServiceName, config)
		if err != nil {
			return nil, err
		}
	} else {
		consulMsg = "Bypassing registration in consul..."
	}
	fmt.Println(consulMsg)

	return config, nil
}

func GetConsulClient(serviceName string, config *common.Config) (*registry.ConsulClient, error) {
	err := checkConsulUp(config)
	if err != nil {
		return nil, err
	}

	consulClient, err := newConsulClient(serviceName, config)
	if err != nil {
		err = fmt.Errorf("connection to consul could not be made: %v", err.Error())
	}

	return consulClient, err
}

func checkConsulUp(config *common.Config) error {
	consulUrl := common.BuildAddr(config.Registry.Host, strconv.Itoa(config.Registry.Port))
	fmt.Println("Check consul is up...", consulUrl)
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
			return errors.New("bad response from Consul service")
		}
	}
	if fails >= config.Registry.FailLimit {
		return errors.New("can't get connection to Consul")
	}
	return nil
}

func newConsulClient(serviceName string, config *common.Config) (*registry.ConsulClient, error) {
	consulClient := &registry.ConsulClient{}
	consulConfig := registry.RegistryConfig{
		Address:        config.Registry.Host,
		Port:           config.Registry.Port,
		ServiceName:    serviceName,
		ServiceAddress: config.Service.Host,
		ServicePort:    config.Service.Port,
		CheckAddress:   fmt.Sprintf("http://%v:%v%v", config.Service.Host, config.Service.Port, config.Service.HealthCheck),
		CheckInterval:  config.Registry.CheckInterval,
	}
	err := consulClient.Init(consulConfig)
	return consulClient, err
}
