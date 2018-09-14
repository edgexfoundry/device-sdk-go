//
// Copyright (c) 2018
// IOTech
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"os"
	"time"
)

type ConsulEndpoint struct {
	RegistryClient Client
}

func (consulEndpoint ConsulEndpoint) Monitor(params types.EndpointParams, ch chan string) {
	for {
		data, err := consulEndpoint.RegistryClient.GetServiceEndpoint(params.ServiceKey)
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
		}
		url := fmt.Sprintf("http://%s:%v%s", data.Address, data.Port, params.Path)
		ch <- url
		time.Sleep(time.Second * time.Duration(15))
	}
}
