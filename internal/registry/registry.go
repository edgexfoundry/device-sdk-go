// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package registry

type Client interface {
	// Initialize Consul by connecting to the agent and registering the service/check
	Init(config RegistryConfig) error

	GetServiceEndpoint(serviceKey string) (ServiceEndpoint, error)

	// Look at the key/value pairs to update configuration
	CheckKeyValuePairs(configurationStruct interface{}, applicationName string, profiles []string) error
}

type ServiceEndpoint struct {
	Key     string
	Address string
	Port    int
}

type RegistryConfig struct {
	Address        string
	Port           int
	ServiceName    string
	ServiceAddress string
	ServicePort    int
	CheckAddress   string
	CheckInterval  string
}
