// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	pwc *pwCache
)

type ProvisionWatcherCache interface {
	ForName(name string) (models.ProvisionWatcher, bool)
	ForId(id string) (models.ProvisionWatcher, bool)
	All() []models.ProvisionWatcher
	Add(pw models.ProvisionWatcher) error
	Update(pw models.ProvisionWatcher) error
	Remove(id string) error
	RemoveByName(name string) error
}

type pwCache struct {
	pwMap   map[string]*models.ProvisionWatcher // key is ProvisionWatcher name
	nameMap map[string]string                   // key is id, and value is ProvisionWatcher name
}

// ForName returns a ProvisionWatcher with the given name.
func (pw *pwCache) ForName(name string) (models.ProvisionWatcher, bool) {
	if watcher, ok := pw.pwMap[name]; ok {
		return *watcher, ok
	} else {
		return models.ProvisionWatcher{}, ok
	}
}

// ForId returns a ProvisionWatcher with the given ProvisionWatcher id.
func (pw *pwCache) ForId(id string) (models.ProvisionWatcher, bool) {
	name, ok := pw.nameMap[id]
	if !ok {
		return models.ProvisionWatcher{}, ok
	}

	if watcher, ok := pw.pwMap[name]; ok {
		return *watcher, ok
	} else {
		return models.ProvisionWatcher{}, ok
	}
}

// All() returns the current list of ProvisionWatchers in the cache.
func (pw *pwCache) All() []models.ProvisionWatcher {
	watchers := make([]models.ProvisionWatcher, len(pw.pwMap))
	i := 0
	for _, watcher := range pw.pwMap {
		watchers[i] = *watcher
		i++
	}
	return watchers
}

// Adds a new ProvisionWatcher to the cache
func (pw *pwCache) Add(watcher models.ProvisionWatcher) error {
	if _, ok := pw.pwMap[watcher.Name]; ok {
		return fmt.Errorf("ProvisionWatcher %s has already existed in cache", watcher.Name)
	}
	pw.pwMap[watcher.Name] = &watcher
	pw.nameMap[watcher.Id] = watcher.Name
	return nil
}

// Update updates the ProvisionWatcher in the cache
func (pw *pwCache) Update(watcher models.ProvisionWatcher) error {
	if err := pw.Remove(watcher.Id); err != nil {
		return err
	}
	return pw.Add(watcher)
}

// Remove removes the specified ProvisionWatcher by id from the cache.
func (pw *pwCache) Remove(id string) error {
	name, ok := pw.nameMap[id]
	if !ok {
		return fmt.Errorf("ProvisionWatcher %s does not exist in cache", id)
	}

	return pw.RemoveByName(name)
}

// RemoveByName removes the specified ProvisionWatcher by name from the cache.
func (pw *pwCache) RemoveByName(name string) error {
	watcher, ok := pw.pwMap[name]
	if !ok {
		return fmt.Errorf("ProvisionWatcher %s does not exist in cache", name)
	}

	delete(pw.nameMap, watcher.Id)
	delete(pw.pwMap, name)
	return nil
}

func newWatcherCache(watchers []models.ProvisionWatcher) ProvisionWatcherCache {
	defaultSize := len(watchers) * 2
	pwMap := make(map[string]*models.ProvisionWatcher, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	for i, pw := range watchers {
		pwMap[pw.Name] = &watchers[i]
		nameMap[pw.Id] = pw.Name
	}
	pwc = &pwCache{pwMap: pwMap, nameMap: nameMap}
	return pwc
}

func Watchers() ProvisionWatcherCache {
	if pwc == nil {
		InitCache()
	}
	return pwc
}
