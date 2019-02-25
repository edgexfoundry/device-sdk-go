// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type ProvisionWatcherClientMock struct {
}

// Get the provision watcher by id
func (ProvisionWatcherClientMock) ProvisionWatcher(id string, ctx context.Context) (models.ProvisionWatcher, error) {
	panic("implement me")
}

// Get a list of all provision watchers
func (ProvisionWatcherClientMock) ProvisionWatchers(ctx context.Context) ([]models.ProvisionWatcher, error) {
	panic("implement me")
}

// Get the provision watcher by name
func (ProvisionWatcherClientMock) ProvisionWatcherForName(name string, ctx context.Context) (models.ProvisionWatcher, error) {
	panic("implement me")
}

// Get the provision watchers that are on a service
func (ProvisionWatcherClientMock) ProvisionWatchersForService(serviceId string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	panic("implement me")
}

// Get the provision watchers that are on a service(by name)
func (ProvisionWatcherClientMock) ProvisionWatchersForServiceByName(serviceName string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	return []models.ProvisionWatcher{}, nil
}

// Get the provision watchers for a profile
func (ProvisionWatcherClientMock) ProvisionWatchersForProfile(profileId string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	panic("implement me")
}

// Get the provision watchers for a profile (by name)
func (ProvisionWatcherClientMock) ProvisionWatchersForProfileByName(profileName string, ctx context.Context) ([]models.ProvisionWatcher, error) {
	panic("implement me")
}

// Add a provision watcher - handle error codes
func (ProvisionWatcherClientMock) Add(dev *models.ProvisionWatcher, ctx context.Context) (string, error) {
	panic("implement me")
}

// Update a provision watcher - handle error codes
func (ProvisionWatcherClientMock) Update(dev models.ProvisionWatcher, ctx context.Context) error {
	panic("implement me")
}

// Delete a provision watcher (specified by id)
func (ProvisionWatcherClientMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}
