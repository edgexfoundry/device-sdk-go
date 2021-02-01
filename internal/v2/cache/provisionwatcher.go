//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

var (
	pwc *provisionWatcherCache
)

type ProvisionWatcherCache interface {
	ForName(name string) (models.ProvisionWatcher, bool)
	ForId(id string) (models.ProvisionWatcher, bool)
	All() []models.ProvisionWatcher
	Add(device models.ProvisionWatcher) error
	Update(device models.ProvisionWatcher) error
	RemoveById(id string) error
	RemoveByName(name string) error
	UpdateAdminState(id string, state models.AdminState) error
}

type provisionWatcherCache struct {
	pwMap   map[string]*models.ProvisionWatcher // key is ProvisionWatcher name
	nameMap map[string]string                   // key is id, and value is ProvisionWatcher name
	mutex   sync.Mutex
}

// ForName returns a provision watcher with the given name.
func (p *provisionWatcherCache) ForName(name string) (models.ProvisionWatcher, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	watcher, ok := p.pwMap[name]
	if !ok {
		return models.ProvisionWatcher{}, false
	}
	return *watcher, ok
}

// ForId returns a provision watcher with the given id.
func (p *provisionWatcherCache) ForId(id string) (models.ProvisionWatcher, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	name, ok := p.nameMap[id]
	if !ok {
		return models.ProvisionWatcher{}, ok
	}

	watcher, ok := p.pwMap[name]
	return *watcher, ok
}

// All returns the current list of provision watchers in the cache.
func (p *provisionWatcherCache) All() []models.ProvisionWatcher {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	i := 0
	watchers := make([]models.ProvisionWatcher, len(p.pwMap))
	for _, watcher := range p.pwMap {
		watchers[i] = *watcher
		i++
	}
	return watchers
}

// Add adds a new provision watcher to the cache.
func (p *provisionWatcherCache) Add(watcher models.ProvisionWatcher) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.add(watcher)
}

func (p *provisionWatcherCache) add(watcher models.ProvisionWatcher) error {
	if _, ok := p.pwMap[watcher.Name]; ok {
		errMsg := fmt.Sprintf("provision watcher %s has already existed in cache", watcher.Name)
		return errors.NewCommonEdgeX(errors.KindDuplicateName, errMsg, nil)
	}

	p.pwMap[watcher.Name] = &watcher
	p.nameMap[watcher.Id] = watcher.Name
	return nil
}

// Update updates the provision watcher in the cache
func (p *provisionWatcherCache) Update(watcher models.ProvisionWatcher) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.removeById(watcher.Id); err != nil {
		return err
	}
	return p.add(watcher)
}

// RemoveById removes the specified provision watcher by id from the cache.
func (p *provisionWatcherCache) RemoveById(id string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removeById(id)
}

func (p *provisionWatcherCache) removeById(id string) error {
	name, ok := p.nameMap[id]
	if !ok {
		errMsg := fmt.Sprintf("failed to find provisionwatcher with given id %s in cache", id)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	}

	return p.removeByName(name)
}

// RemoveByName removes the specified provision watcher by name from the cache.
func (p *provisionWatcherCache) RemoveByName(name string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removeByName(name)
}

func (p *provisionWatcherCache) removeByName(name string) error {
	watcher, ok := p.pwMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find provisionwatcher %s in cache", name)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	}

	delete(p.pwMap, name)
	delete(p.nameMap, watcher.Id)
	return nil
}

// UpdateAdminState updates the provision watcher admin state in cache by id.
func (p *provisionWatcherCache) UpdateAdminState(id string, state models.AdminState) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	name, ok := p.nameMap[id]
	if !ok {
		errMsg := fmt.Sprintf("failed to find provisionwatcher with given id %s in cache", id)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	}

	p.pwMap[name].AdminState = state
	return nil
}

func newProvisionWatcherCache(pws []models.ProvisionWatcher) ProvisionWatcherCache {
	defaultSize := len(pws)
	pwMap := make(map[string]*models.ProvisionWatcher, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	for i, pw := range pws {
		pwMap[pw.Name] = &pws[i]
		nameMap[pw.Id] = pw.Name
	}
	pwc = &provisionWatcherCache{pwMap: pwMap, nameMap: nameMap}
	return pwc
}

func ProvisionWatchers() ProvisionWatcherCache {
	return pwc
}
