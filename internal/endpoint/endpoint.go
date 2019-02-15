// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package endpoint

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-registry"
	"os"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
)

type Endpoint struct {
	RegistryClient registry.Client
	passFirstRun   bool
	WG             *sync.WaitGroup
}

func (endpoint Endpoint) Monitor(params types.EndpointParams, ch chan string) {
	for {
		data, err := endpoint.RegistryClient.GetServiceEndpoint(params.ServiceKey)
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
		}
		url := fmt.Sprintf("http://%s:%v%s", data.Host, data.Port, params.Path)
		ch <- url

		// After the first run, the client can be indicated initialized
		if !endpoint.passFirstRun {
			endpoint.WG.Done()
			endpoint.passFirstRun = true
		}

		time.Sleep(time.Second * time.Duration(params.Interval))
	}
}
