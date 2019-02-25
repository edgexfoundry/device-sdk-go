// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func LoadWatchers(watcherList []common.WatcherInfo) error {
	for _, pw := range watcherList {
		if _, ok := cache.Watchers().ForName(pw.Name); ok {
			common.LoggingClient.Debug(fmt.Sprintf("Watcher %s exists, using the existing one", pw.Name))
			continue
		} else {
			common.LoggingClient.Debug(fmt.Sprintf("Watcher %s doesn't exist, creating a new one", pw.Name))
			err := createWatcher(pw)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("Creating Watcher from config failed: %v", pw))
				return err
			}
		}
	}
	return nil
}

func createWatcher(wi common.WatcherInfo) error {
	prf, ok := cache.Profiles().ForName(wi.Profile)
	if !ok {
		errMsg := fmt.Sprintf("Device Profile %s doesn't exist for Watcher %s", wi.Profile, wi.Name)
		common.LoggingClient.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	pw := &models.ProvisionWatcher{
		Name:           wi.Name,
		Profile:        prf,
		Service:        common.CurrentDeviceService,
		OperatingState: models.Enabled,
	}
	pw.Origin = millis
	pw.Identifiers = make(map[string]string, 1)
	pw.Identifiers[wi.Key] = wi.MatchString
	common.LoggingClient.Debug(fmt.Sprintf("Adding Watcher: %v", pw))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err := common.ProvisionWatcherClient.Add(pw, ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Add Watcher failed %v, error: %v", pw, err))
		return err
	}
	if err = common.VerifyIdFormat(id, "Watcher"); err != nil {
		return err
	}
	pw.Id = id
	cache.Watchers().Add(*pw)

	return nil
}
