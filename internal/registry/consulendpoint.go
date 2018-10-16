// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

type ConsulEndpoint struct {
	RegistryClient Client
	passFirstRun   bool
	WG             *sync.WaitGroup
}

func (consulEndpoint ConsulEndpoint) Monitor(params types.EndpointParams, ch chan string) {
	for {
		data, err := consulEndpoint.RegistryClient.GetServiceEndpoint(params.ServiceKey)
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
		}
		url := fmt.Sprintf("http://%s:%v%s", data.Address, data.Port, params.Path)
		ch <- url

		// After the first run, the client can be indicated initialized
		if !consulEndpoint.passFirstRun {
			consulEndpoint.WG.Done()
			consulEndpoint.passFirstRun = true
		}

		time.Sleep(time.Second * time.Duration(params.Interval))
	}
}
