// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autodiscovery

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/net/context"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/controller/http/correlation"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/utils"
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
)

type discoveryLocker struct {
	busy bool
	mux  sync.Mutex
}

var locker discoveryLocker

func DiscoveryWrapper(driver interfaces.ProtocolDriver, ctx context.Context, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	locker.mux.Lock()
	if locker.busy {
		lc.Info("another device discovery process is currently running")
		locker.mux.Unlock()
		return
	}
	locker.busy = true
	locker.mux.Unlock()

	requestId := correlation.IdFromContext(ctx)
	if len(requestId) == 0 {
		requestId = uuid.NewString()
		ctx = context.WithValue(ctx, common.CorrelationHeader, requestId) // nolint: staticcheck
		lc.Debugf("device discovery correlation id is empty, set it to %s", requestId)
	}
	dic.Update(di.ServiceConstructorMap{
		container.DiscoveryRequestIdName: func(get di.Get) any {
			return requestId
		},
	})

	utils.PublishDeviceDiscoveryProgressSystemEvent(requestId, 0, 0, "", ctx, dic)
	lc.Debugf("protocol discovery triggered with correlation id: %s", requestId)
	err := driver.Discover()
	if err != nil {
		errMsg := fmt.Sprintf("failed to trigger protocol discovery with correlation id: %s, err: %s", requestId, err.Error())
		utils.PublishDeviceDiscoveryProgressSystemEvent(requestId, -1, 0, errMsg, ctx, dic)
		lc.Error(errMsg)
	} else {
		utils.PublishDeviceDiscoveryProgressSystemEvent(requestId, 100, 0, "", ctx, dic)
	}

	// ReleaseLock
	locker.mux.Lock()
	locker.busy = false
	dic.Update(di.ServiceConstructorMap{
		container.DiscoveryRequestIdName: func(get di.Get) any {
			return ""
		},
	})
	locker.mux.Unlock()
}
