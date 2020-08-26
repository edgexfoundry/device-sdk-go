// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

// AddProvisionWatcher adds a new Watcher to the cache and Core Metadata
// Returns new Watcher id or non-nil error.
func (s *DeviceService) AddProvisionWatcher(watcher contract.ProvisionWatcher) (id string, err error) {
	if pw, ok := cache.ProvisionWatchers().ForName(watcher.Name); ok {
		return pw.Id, fmt.Errorf("name conflicted, watcher %s exists", watcher.Name)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Adding managed watcher: %s\n", watcher.Name))

	prf, ok := cache.Profiles().ForName(watcher.Profile.Name)
	if !ok {
		errMsg := fmt.Sprintf("Device Profile %s doesn't exist for Watcher %s", watcher.Profile.Name, watcher.Name)
		s.LoggingClient.Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}

	millis := time.Now().UnixNano() / int64(time.Millisecond)
	watcher.Origin = millis
	watcher.Service = s.deviceService
	watcher.Profile = prf

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err = s.edgexClients.ProvisionWatcherClient.Add(ctx, &watcher)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Add Watcher failed %s, error: %v", watcher.Name, err))
		return "", err
	}
	if err = common.VerifyIdFormat(id, "Watcher"); err != nil {
		return "", err
	}
	watcher.Id = id
	_ = cache.ProvisionWatchers().Add(watcher)

	return id, nil
}

// ProvisionWatchers return all managed Watchers from cache
func (s *DeviceService) ProvisionWatchers() []contract.ProvisionWatcher {
	return cache.ProvisionWatchers().All()
}

// GetProvisionWatcherByName returns the Watcher by its name if it exists in the cache, or returns an error.
func (s *DeviceService) GetProvisionWatcherByName(name string) (contract.ProvisionWatcher, error) {
	pw, ok := cache.ProvisionWatchers().ForName(name)
	if !ok {
		msg := fmt.Sprintf("Watcher %s cannot be found in cache", name)
		s.LoggingClient.Info(msg)
		return contract.ProvisionWatcher{}, fmt.Errorf(msg)
	}
	return pw, nil
}

// RemoveProvisionWatcher removes the specified Watcher by id from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceService) RemoveProvisionWatcher(id string) error {
	pw, ok := cache.ProvisionWatchers().ForId(id)
	if !ok {
		msg := fmt.Sprintf("ProvisionWatcher %s cannot be found in cache", id)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Removing managed ProvisionWatcher: %s", pw.Name))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := s.edgexClients.ProvisionWatcherClient.Delete(ctx, id)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Delete ProvisionWatcher %s from Core Metadata failed", id))
		return err
	}

	return nil
}

// UpdateProvisionWatcher updates the Watcher in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *DeviceService) UpdateProvisionWatcher(watcher contract.ProvisionWatcher) error {
	_, ok := cache.ProvisionWatchers().ForId(watcher.Id)
	if !ok {
		msg := fmt.Sprintf("provisionwatcher %s cannot be found in cache", watcher.Id)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debug(fmt.Sprintf("Updating managed ProvisionWatcher: %s", watcher.Name))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	err := s.edgexClients.ProvisionWatcherClient.Update(ctx, watcher)
	if err != nil {
		s.LoggingClient.Error(fmt.Sprintf("Update ProvisionWatcher %s from Core Metadata failed: %v", watcher.Name, err))
		return err
	}

	return nil
}
