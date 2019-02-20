// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/mitchellh/consulstructure"
)

type ConsulClient struct {
	Consul *consulapi.Client
	config RegistryConfig
}

func (c *ConsulClient) Init(config RegistryConfig) error {
	var err error // Declare error to be used throughout function
	c.config = config

	// Connect to the Consul Agent
	defaultConfig := &consulapi.Config{}
	defaultConfig.Address = config.Address + ":" + strconv.Itoa(config.Port)
	c.Consul, err = consulapi.NewClient(defaultConfig)
	if err != nil {
		return err
	}

	// Register the Service
	_, _ = fmt.Fprintf(os.Stdout, "Register the Service ...\n")
	err = c.Consul.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		Name:    config.ServiceName,
		Address: config.ServiceAddress,
		Port:    config.ServicePort,
	})
	if err != nil {
		return err
	}

	// Register the Health Check
	_, _ = fmt.Fprintf(os.Stdout, "Register the Health Check ...\n")
	err = c.Consul.Agent().CheckRegister(&consulapi.AgentCheckRegistration{
		Name:      "Health Check: " + config.ServiceName,
		Notes:     "Check the health of the API",
		ServiceID: config.ServiceName,
		AgentServiceCheck: consulapi.AgentServiceCheck{
			HTTP:     config.CheckAddress,
			Interval: config.CheckInterval,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *ConsulClient) GetServiceEndpoint(serviceKey string) (ServiceEndpoint, error) {
	if c.Consul == nil {
		return ServiceEndpoint{}, fmt.Errorf("consul client hasn't been initialized")
	}

	services, err := c.Consul.Agent().Services()
	if err != nil {
		return ServiceEndpoint{}, err
	}

	endpoint := ServiceEndpoint{}
	for key, service := range services {
		if key == serviceKey {
			endpoint.Port = service.Port
			endpoint.Key = key
			endpoint.Address = service.Address
		}
	}
	return endpoint, nil
}

func (c *ConsulClient) CheckConfigExistence() bool {
	if c.Consul == nil {
		_, _ = fmt.Fprintf(os.Stdout, "consul client hasn't been initialized\n")
		return false
	}

	stem := common.ConfigStem + c.config.ServiceName
	if stemKeys, _, err := c.Consul.KV().Keys(stem, "", nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Retrieving KV from Consul failed...\n")
		return false
	} else if len(stemKeys) == 0 {
		return false
	} else {
		return true
	}
}

func (c *ConsulClient) PopulateConfig(config common.Config) error {
	if c.Consul == nil {
		return fmt.Errorf("consul client hasn't been initialized")
	}

	_, _ = fmt.Fprintf(os.Stdout, "populating config from file to Consul\n")
	stem := fmt.Sprintf("%s%s", common.ConfigStem, c.config.ServiceName)
	err := populateValue(stem, reflect.ValueOf(config), c.Consul)

	return err
}

func (c *ConsulClient) LoadConfig(config *common.Config) (*common.Config, error) {
	_, _ = fmt.Fprintf(os.Stdout, "Look at the key/value pairs to update configuration from registry ...\n")
	var err error

	cfg := &consulapi.Config{}
	cfg.Address = c.config.Address + ":" + strconv.Itoa(c.config.Port)
	updateCh := make(chan interface{})
	errCh := make(chan error)
	dec := &consulstructure.Decoder{
		Consul:   cfg,
		Target:   &common.Config{},
		Prefix:   common.ConfigStem + common.ServiceName,
		UpdateCh: updateCh,
		ErrCh:    errCh,
	}

	defer dec.Close()
	defer close(updateCh)
	defer close(errCh)
	go dec.Run()

	select {
	case <-time.After(2 * time.Second):
		err = errors.New("timeout loading config from registry")
	case ex := <-errCh:
		err = errors.New(ex.Error())
	case raw := <-updateCh:
		actual, ok := raw.(*common.Config)
		if !ok {
			return config, errors.New("type check failed")
		}
		config = actual
		//Check that information was successfully read from Consul
		if config.Service.Port == 0 {
			return nil, errors.New("error reading from Consul")
			//} else {
			//Handle List in special way
			//config.MapValueToList()
		}
	}

	return config, err
}
