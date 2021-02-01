// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/google/uuid"
)

var (
	previousOrigin int64
	originMutex    sync.Mutex
)

func GetUniqueOrigin() int64 {
	originMutex.Lock()
	defer originMutex.Unlock()
	now := time.Now().UnixNano()
	if now <= previousOrigin {
		now = previousOrigin + 1
	}
	previousOrigin = now
	return now
}

func UpdateLastConnected(name string, lc logger.LoggingClient, dc interfaces.DeviceClient) {
	t := time.Now().UnixNano() / int64(time.Millisecond)
	req := requests.UpdateDeviceRequest{
		BaseRequest: common.BaseRequest{
			RequestId: uuid.New().String(),
		},
		Device: dtos.UpdateDevice{
			Name:          &name,
			LastConnected: &t,
		},
	}

	_, err := dc.Update(context.Background(), []requests.UpdateDeviceRequest{req})
	if err != nil {
		lc.Errorf("failed to update LastConnected for device %s in Core Metadata: %v", name, err)
	}
}

func UpdateOperatingState(name string, state string, lc logger.LoggingClient, dc interfaces.DeviceClient) {
	req := requests.UpdateDeviceRequest{
		BaseRequest: common.BaseRequest{
			RequestId: uuid.New().String(),
		},
		Device: dtos.UpdateDevice{
			Name:           &name,
			OperatingState: &state,
		},
	}

	_, err := dc.Update(context.Background(), []requests.UpdateDeviceRequest{req})
	if err != nil {
		lc.Errorf("failed to update OperatingState for device %s in Core Metadata: %v", name, err)
	}
}

func SendEvent(event dtos.Event, correlationID string, lc logger.LoggingClient, ec interfaces.EventClient) {
	ctx := context.WithValue(context.Background(), CorrelationHeader, correlationID)
	ctx = context.WithValue(ctx, clients.ContentType, clients.ContentTypeJSON)
	req := requests.AddEventRequest{
		BaseRequest: common.BaseRequest{
			RequestId: uuid.New().String(),
		},
		Event: event,
	}

	res, err := ec.Add(ctx, req)
	if err != nil {
		lc.Errorf("failed to push event to core-data: %s", err)
	} else {
		lc.Infof("pushed event to core-data: %s", res.Id)
	}
}
