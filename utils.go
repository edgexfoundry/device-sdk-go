// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018
// Canonical Ltd
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//
package device

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/edgexfoundry/device-sdk-go/registry"
)

const consulStatusPath = "/v1/agent/self"

var (
	// Need to set timeout because it hang until server close connection
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	netClient = &http.Client{Timeout: time.Second * 10}
)

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

func checkConsulUp(config *Config) error {
	consulUrl := buildAddr(config.Registry.Host, strconv.Itoa(config.Registry.Port))
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

func buildAddr(host string, port string) string {
	var buffer bytes.Buffer

	buffer.WriteString(httpScheme)
	buffer.WriteString(host)
	buffer.WriteString(colon)
	buffer.WriteString(port)

	return buffer.String()
}

func newConsulClient(serviceName string, config *Config) (*registry.ConsulClient, error) {
	consulClient := &registry.ConsulClient{}
	consulConfig := registry.Config{
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

func GetConsulClient(serviceName string, config *Config) (*registry.ConsulClient, error) {
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
