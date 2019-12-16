// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package endpoint

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

type Endpoint struct {
	RegistryClient registry.Client
	passFirstRun   bool
	WG             *sync.WaitGroup
}

func (e Endpoint) Monitor(params types.EndpointParams) chan string {
	ch := make(chan string, 1)
	ch <- e.fetch(params)
	go func() {
		for {
			url, err := e.buildURL(params)
			if err != nil {
				fmt.Fprintln(os.Stdout, err.Error())
			}
			ch <- url
			time.Sleep(time.Second * time.Duration(params.Interval))
		}
	}()
	return ch
}

func (e Endpoint) fetch(params types.EndpointParams) string {
	url, err := e.buildURL(params)
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
	}
	return url
}

func (e Endpoint) buildURL(params types.EndpointParams) (string, error) {
	if e.RegistryClient != nil {
		endpoint, err := (e.RegistryClient).GetServiceEndpoint(params.ServiceKey)
		if err != nil {
			return "", fmt.Errorf("unable to get Service endpoint for %s: %s", params.ServiceKey, err.Error())
		}
		return fmt.Sprintf("http://%s:%v%s", endpoint.Host, endpoint.Port, params.Path), nil
	} else {
		return "", fmt.Errorf("unable to get Service endpoint for %s: Registry client is nil", params.ServiceKey)
	}
}
