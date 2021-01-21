// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"

	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

var (
	ValidBooleanWatcher         = contract.ProvisionWatcher{}
	ValidIntegerWatcher         = contract.ProvisionWatcher{}
	ValidUnsignedIntegerWatcher = contract.ProvisionWatcher{}
	ValidFloatWatcher           = contract.ProvisionWatcher{}
	DuplicateFloatWatcher       = contract.ProvisionWatcher{}
	NewProvisionWatcher         = contract.ProvisionWatcher{}
)

type ProvisionWatcherClientMock struct {
}

// Get the provision watcher by id
func (ProvisionWatcherClientMock) ProvisionWatcher(_ context.Context, _ string) (contract.ProvisionWatcher, error) {
	panic("implement me")
}

// Get a list of all provision watchers
func (ProvisionWatcherClientMock) ProvisionWatchers(_ context.Context) ([]contract.ProvisionWatcher, error) {
	panic("implement me")
}

// Get the provision watcher by name
func (ProvisionWatcherClientMock) ProvisionWatcherForName(_ context.Context, _ string) (contract.ProvisionWatcher, error) {
	panic("implement me")
}

// Get the provision watchers that are on a service
func (ProvisionWatcherClientMock) ProvisionWatchersForService(_ context.Context, _ string) ([]contract.ProvisionWatcher, error) {
	panic("implement me")
}

// Get the provision watchers that are on a service(by name)
func (ProvisionWatcherClientMock) ProvisionWatchersForServiceByName(_ context.Context, _ string) ([]contract.ProvisionWatcher, error) {
	err := populateProvisionWatcherMock()
	if err != nil {
		return nil, err
	}
	return []contract.ProvisionWatcher{
		ValidBooleanWatcher,
		ValidIntegerWatcher,
		ValidUnsignedIntegerWatcher,
		ValidFloatWatcher,
	}, nil
}

// Get the provision watchers for a profile
func (ProvisionWatcherClientMock) ProvisionWatchersForProfile(_ context.Context, _ string) ([]contract.ProvisionWatcher, error) {
	panic("implement me")
}

// Get the provision watchers for a profile (by name)
func (ProvisionWatcherClientMock) ProvisionWatchersForProfileByName(_ context.Context, _ string) ([]contract.ProvisionWatcher, error) {
	panic("implement me")
}

// Add a provision watcher - handle error codes
func (ProvisionWatcherClientMock) Add(_ context.Context, _ *contract.ProvisionWatcher) (string, error) {
	panic("implement me")
}

// Update a provision watcher - handle error codes
func (ProvisionWatcherClientMock) Update(_ context.Context, _ contract.ProvisionWatcher) error {
	panic("implement me")
}

// Delete a provision watcher (specified by id)
func (ProvisionWatcherClientMock) Delete(_ context.Context, _ string) error {
	panic("implement me")
}

func populateProvisionWatcherMock() error {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	watchers, err := loadData(basepath + "/data/provisionwatcher")
	if err != nil {
		return err
	}

	_ = json.Unmarshal(watchers[WatcherBool], &ValidBooleanWatcher)
	_ = json.Unmarshal(watchers[WatcherInt], &ValidIntegerWatcher)
	_ = json.Unmarshal(watchers[WatcherUint], &ValidUnsignedIntegerWatcher)
	_ = json.Unmarshal(watchers[WatcherFloat], &ValidFloatWatcher)
	_ = json.Unmarshal(watchers[WatcherFloat], &DuplicateFloatWatcher)
	_ = json.Unmarshal(watchers[WatcherNew], &NewProvisionWatcher)

	return nil
}
