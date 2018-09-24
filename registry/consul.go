//
// Copyright (c) 2018
// IOTech
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"fmt"
	consulapi "github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	Consul *consulapi.Client
}

func (c *ConsulClient) Init(config Config) error {
	var err error // Declare error to be used throughout function

	// Connect to the Consul Agent
	defaultConfig := &consulapi.Config{}
	defaultConfig.Address = config.Address + ":" + strconv.Itoa(config.Port)
	c.Consul, err = consulapi.NewClient(defaultConfig)
	if err != nil {
		return err
	}

	// Register the Service
	fmt.Println("Register the Service ...")
	err = c.Consul.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		Name:    config.ServiceName,
		Address: config.ServiceAddress,
		Port:    config.ServicePort,
	})
	if err != nil {
		return err
	}

	// Register the Health Check
	fmt.Println("Register the Health Check ...")
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

func (c *ConsulClient) CheckKeyValuePairs(configuration interface{}, applicationName string, profiles []string) error {
	fmt.Println("Look at the key/value pairs to update configuration from registry ...")
	// Consul wasn't initialized
	if c.Consul == nil {
		err := errors.New("Consul wasn't initialized, can't check key/value pairs")
		return err
	}

	kv := c.Consul.KV()

	// Reflection to get the field names (These will be part of the key names)
	configValue := reflect.ValueOf(configuration)
	// Loop through the fields
	for i := 0; i < configValue.Elem().NumField(); i++ {
		fieldName := configValue.Elem().Type().Field(i).Name
		fieldValue := configValue.Elem().Field(i)
		keyPath := "config/" + applicationName + ";" + strings.Join(profiles, ";") + "/" + fieldName
		var byteValue []byte // Byte array that will be passed to Consul

		// Switch off of the value type
		switch fieldValue.Kind() {
		case reflect.Bool:
			byteValue = []byte(strconv.FormatBool(fieldValue.Bool()))

			// Check if the key is already there
			pair, _, err := kv.Get(keyPath, nil)
			if err != nil {
				return err
			}
			// Pair doesn't exist, create it
			if pair == nil {
				pair = &consulapi.KVPair{
					Key:   keyPath,
					Value: byteValue,
				}
				_, err = kv.Put(pair, nil)
				if err != nil {
					return err
				}
			} else { // Pair does exist, get the new value
				pair, _, err = kv.Get(keyPath, nil)
				if err != nil {
					return err
				}

				newValue, err := strconv.ParseBool(string(pair.Value))
				if err != nil {
					return err
				}

				fieldValue.SetBool(newValue) // Set the new value
			}
			break
		case reflect.String:
			byteValue = []byte(fieldValue.String())

			// Check if the key is already there
			pair, _, err := kv.Get(keyPath, nil)
			if err != nil {
				return err
			}
			// Pair doesn't exist, create it
			if pair == nil {
				pair = &consulapi.KVPair{
					Key:   keyPath,
					Value: byteValue,
				}
				_, err = kv.Put(pair, nil)
				if err != nil {
					return err
				}
			} else { // Pair does exist, get the new value
				pair, _, err = kv.Get(keyPath, nil)
				if err != nil {
					return err
				}

				newValue := string(pair.Value)

				fieldValue.SetString(newValue) // Set the new value
			}
			break
		case reflect.Int:
			byteValue = []byte(strconv.FormatInt(fieldValue.Int(), 10))

			// Check if the key is already there
			pair, _, err := kv.Get(keyPath, nil)
			if err != nil {
				return err
			}
			// Pair doesn't exist, create it
			if pair == nil {
				pair = &consulapi.KVPair{
					Key:   keyPath,
					Value: byteValue,
				}
				_, err = kv.Put(pair, nil)
				if err != nil {
					return err
				}
			} else { // Pair does exist, get the new value
				pair, _, err = kv.Get(keyPath, nil)
				if err != nil {
					return err
				}

				newValue, err := strconv.ParseInt(string(pair.Value), 10, 64)
				if err != nil {
					return err
				}

				fieldValue.SetInt(newValue) // Set the new value
			}
			break

		default:
			// TODO Can't parse struct
			fmt.Println("Unexpected fieldValue: ", fieldValue)
			//err := errors.New("Can't get the type of field: " + keyPath)
			//return err

		}
	}

	return nil
}
