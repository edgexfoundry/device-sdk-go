// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	pwc *provisionWatcherCache
)

type ProvisionWatcherCache interface {
	ForName(name string) (contract.ProvisionWatcher, bool)
	ForId(id string) (contract.ProvisionWatcher, bool)
	All() []contract.ProvisionWatcher
	Add(device contract.ProvisionWatcher) error
	Update(device contract.ProvisionWatcher) error
	Remove(id string) error
	RemoveByName(name string) error
	UpdateAdminState(id string, state contract.AdminState) error
}

type provisionWatcherCache struct {
	pwMap   map[string]*contract.ProvisionWatcher // key is ProvisionWatcher name
	nameMap map[string]string                     // key is id, and value is ProvisionWatcher name
	mutex   sync.Mutex
}

// ForName returns a provisionwatcher with the given name.
func (p *provisionWatcherCache) ForName(name string) (contract.ProvisionWatcher, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if watcher, ok := p.pwMap[name]; ok {
		return *watcher, ok
	} else {
		return contract.ProvisionWatcher{}, ok
	}
}

// ForId returns a provisionwatcher with the given id.
func (p *provisionWatcherCache) ForId(id string) (contract.ProvisionWatcher, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	name, ok := p.nameMap[id]
	if !ok {
		return contract.ProvisionWatcher{}, ok
	}

	if watcher, ok := p.pwMap[name]; ok {
		return *watcher, ok
	} else {
		return contract.ProvisionWatcher{}, ok
	}
}

// All() returns the current list of provisionwatchers in the cache.
func (p *provisionWatcherCache) All() []contract.ProvisionWatcher {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	watchers := make([]contract.ProvisionWatcher, len(p.pwMap))
	i := 0
	for _, watcher := range p.pwMap {
		watchers[i] = *watcher
		i++
	}
	return watchers
}

// Adds a new provisionwatcher to the cache.
func (p *provisionWatcherCache) Add(watcher contract.ProvisionWatcher) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.add(watcher)
}

func (p *provisionWatcherCache) add(watcher contract.ProvisionWatcher) error {
	if _, ok := p.pwMap[watcher.Name]; ok {
		return fmt.Errorf("watcher %s has already existed in cache", watcher.Name)
	}
	p.pwMap[watcher.Name] = &watcher
	p.nameMap[watcher.Id] = watcher.Name
	return nil
}

// Update updates the provisionwatcher in the cache
func (p *provisionWatcherCache) Update(watcher contract.ProvisionWatcher) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.remove(watcher.Id); err != nil {
		return err
	}
	return p.add(watcher)
}

// Remove removes the specified provisionwatcher by id from the cache.
func (p *provisionWatcherCache) Remove(id string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.remove(id)
}

func (p *provisionWatcherCache) remove(id string) error {
	name, ok := p.nameMap[id]
	if !ok {
		return fmt.Errorf("watcher %s does not exist in cache", id)
	}
	return p.removeByName(name)
}

// RemoveByName removes the specified provisionwatcher by name from the cache.
func (p *provisionWatcherCache) RemoveByName(name string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removeByName(name)
}

func (p *provisionWatcherCache) removeByName(name string) error {
	watcher, ok := p.pwMap[name]
	if !ok {
		return fmt.Errorf("watcher %s does not exist in cache", name)
	}

	delete(p.pwMap, name)
	delete(p.nameMap, watcher.Id)
	return nil
}

func (p *provisionWatcherCache) UpdateAdminState(id string, state contract.AdminState) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	name, ok := p.nameMap[id]
	if !ok {
		return fmt.Errorf("watcher %s cannot be found in cache", id)
	}

	p.pwMap[name].AdminState = state
	return nil
}

func newProvisionWatcherCache(watchers []contract.ProvisionWatcher) ProvisionWatcherCache {
	defaultSize := len(watchers) * 2
	pwMap := make(map[string]*contract.ProvisionWatcher, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	for i, w := range watchers {
		pwMap[w.Name] = &watchers[i]
		nameMap[w.Id] = w.Name
	}
	pwc = &provisionWatcherCache{pwMap: pwMap, nameMap: nameMap}
	return pwc
}

func ProvisionWatchers() ProvisionWatcherCache {
	return pwc
}
