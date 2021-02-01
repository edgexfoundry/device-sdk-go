// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/cache"
)

// AddProvisionWatcher adds a new Watcher to the cache and Core Metadata
// Returns new Watcher id or non-nil error.
func (s *DeviceService) AddProvisionWatcher(watcher models.ProvisionWatcher) (string, error) {
	if pw, ok := cache.ProvisionWatchers().ForName(watcher.Name); ok {
		return pw.Id, fmt.Errorf("name conflicted, ProvisionWatcher %s exists", watcher.Name)
	}

	_, ok := cache.Profiles().ForName(watcher.ProfileName)
	if !ok {
		errMsg := fmt.Sprintf("device profile %s doesn't exist for provision watcher %s", watcher.ProfileName, watcher.Name)
		s.LoggingClient.Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}

	s.LoggingClient.Debugf("Adding managed ProvisionWatcher %s", watcher.Name)
	req := requests.AddProvisionWatcherRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId: uuid.New().String(),
		},
		ProvisionWatcher: dtos.FromProvisionWatcherModelToDTO(watcher),
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	res, err := s.edgexClients.ProvisionWatcherClient.Add(ctx, []requests.AddProvisionWatcherRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("failed to add ProvisionWatcher to Core Metadata: %v", watcher.Name, err)
		return "", err
	}

	return res[0].Id, nil
}

// ProvisionWatchers return all managed Watchers from cache
func (s *DeviceService) ProvisionWatchers() []models.ProvisionWatcher {
	return cache.ProvisionWatchers().All()
}

// GetProvisionWatcherByName returns the Watcher by its name if it exists in the cache, or returns an error.
func (s *DeviceService) GetProvisionWatcherByName(name string) (models.ProvisionWatcher, error) {
	pw, ok := cache.ProvisionWatchers().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find ProvisionWatcher %s in cache", name)
		s.LoggingClient.Info(msg)
		return models.ProvisionWatcher{}, fmt.Errorf(msg)
	}
	return pw, nil
}

// RemoveProvisionWatcher removes the specified Watcher by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceService) RemoveProvisionWatcher(name string) error {
	pw, ok := cache.ProvisionWatchers().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find ProvisionWatcher %s in cache", name)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debugf("Removing managed ProvisionWatcher: %s", pw.Name)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	_, err := s.edgexClients.ProvisionWatcherClient.DeleteProvisionWatcherByName(ctx, name)
	if err != nil {
		s.LoggingClient.Errorf("failed to delete ProvisionWatcher %s in Core Metadata", name)
		return err
	}

	return nil
}

// UpdateProvisionWatcher updates the Watcher in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *DeviceService) UpdateProvisionWatcher(watcher models.ProvisionWatcher) error {
	_, ok := cache.ProvisionWatchers().ForName(watcher.Name)
	if !ok {
		msg := fmt.Sprintf("failed to find ProvisionWatcher %s in cache", watcher.Name)
		s.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}

	s.LoggingClient.Debugf("Updating managed ProvisionWatcher: %s", watcher.Name)
	req := requests.UpdateProvisionWatcherRequest{
		BaseRequest: commonDTO.BaseRequest{
			RequestId: uuid.New().String(),
		},
		ProvisionWatcher: dtos.FromProvisionWatcherModelToUpdateDTO(watcher),
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	_, err := s.edgexClients.ProvisionWatcherClient.Update(ctx, []requests.UpdateProvisionWatcherRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("failed to update ProvisionWatcher %s in Core Metadata: %v", watcher.Name, err)
		return err
	}

	return nil
}
