// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2020 Dell Inc.
//
// SPDX-License-Identifier: Apache-2.0

package endpoint

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/interfaces"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

type Endpoint struct {
	ctx            context.Context
	wg             *sync.WaitGroup
	registryClient registry.Client
	serviceKey     string // The key of the service as found in the service registry (e.g. Consul)
	path           string // The path to the service's endpoint following port number in the URL
	interval       int    // The interval in milliseconds governing how often the client polls to keep the endpoint current

}

func New(
	ctx context.Context,
	wg *sync.WaitGroup,
	registryClient registry.Client,
	serviceKey string,
	path string,
	interval int) *Endpoint {

	return &Endpoint{
		ctx:            ctx,
		wg:             wg,
		registryClient: registryClient,
		serviceKey:     serviceKey,
		path:           path,
		interval:       interval,
	}
}

func (e Endpoint) Monitor() chan interfaces.URLStream {
	ch := make(chan interfaces.URLStream)
	ticker := time.NewTicker(time.Second * time.Duration(e.interval))
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()

		// run fetchURL once before looping so we get the first check before the first timer interval
		e.fetchURL(ch)
		for {
			select {
			case <-e.ctx.Done():
				ticker.Stop()
				return

			case <-ticker.C:
				e.fetchURL(ch)
			}
		}
	}()

	return ch
}

func (e Endpoint) fetchURL(ch chan interfaces.URLStream) {
	endpoint, err := (e.registryClient).GetServiceEndpoint(e.serviceKey)
	if err != nil {
		_, _ = fmt.Println(fmt.Errorf("unable to get service endpoint for %s: %s", e.serviceKey, err.Error()))
		return
	}

	ch <- interfaces.URLStream(fmt.Sprintf("http://%s:%v%s", endpoint.Host, endpoint.Port, e.path))
}
