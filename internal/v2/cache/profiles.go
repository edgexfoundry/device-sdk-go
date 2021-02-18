// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
)

var (
	pc *profileCache
)

type ProfileCache interface {
	ForName(name string) (models.DeviceProfile, bool)
	All() []models.DeviceProfile
	Add(profile models.DeviceProfile) errors.EdgeX
	Update(profile models.DeviceProfile) errors.EdgeX
	RemoveByName(name string) errors.EdgeX
	DeviceResource(profileName string, resourceName string) (models.DeviceResource, bool)
	CommandExists(profileName string, cmd string, method string) (bool, errors.EdgeX)
	ResourceOperations(profileName string, cmd string, method string) ([]models.ResourceOperation, errors.EdgeX)
	ResourceOperation(profileName string, deviceResource string, method string) (models.ResourceOperation, errors.EdgeX)
}

type profileCache struct {
	deviceProfileMap         map[string]*models.DeviceProfile // key is DeviceProfile name
	deviceResourceMap        map[string]map[string]models.DeviceResource
	getResourceOperationsMap map[string]map[string][]models.ResourceOperation
	setResourceOperationsMap map[string]map[string][]models.ResourceOperation
	commandsMap              map[string]map[string]models.Command
	mutex                    sync.RWMutex
}

func newProfileCache(profiles []models.DeviceProfile) ProfileCache {
	defaultSize := len(profiles)
	dpMap := make(map[string]*models.DeviceProfile, defaultSize)
	drMap := make(map[string]map[string]models.DeviceResource, defaultSize)
	getRoMap := make(map[string]map[string][]models.ResourceOperation, defaultSize)
	setRoMap := make(map[string]map[string][]models.ResourceOperation, defaultSize)
	cmdMap := make(map[string]map[string]models.Command, defaultSize)
	for _, dp := range profiles {
		dpMap[dp.Name] = &dp
		drMap[dp.Name] = deviceResourceSliceToMap(dp.DeviceResources)
		getRoMap[dp.Name], setRoMap[dp.Name] = deviceCommandSliceToMap(dp.DeviceCommands)
		cmdMap[dp.Name] = commandSliceToMap(dp.CoreCommands)
	}

	pc = &profileCache{
		deviceProfileMap:         dpMap,
		deviceResourceMap:        drMap,
		getResourceOperationsMap: getRoMap,
		setResourceOperationsMap: setRoMap,
		commandsMap:              cmdMap}
	return pc
}

// ForName returns a profile with the given profile name.
func (p *profileCache) ForName(name string) (models.DeviceProfile, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	profile, ok := p.deviceProfileMap[name]
	if !ok {
		return models.DeviceProfile{}, false
	}
	return *profile, ok
}

// All returns the current list of profiles in the cache.
func (p *profileCache) All() []models.DeviceProfile {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	i := 0
	ps := make([]models.DeviceProfile, len(p.deviceProfileMap))
	for _, profile := range p.deviceProfileMap {
		ps[i] = *profile
		i++
	}
	return ps
}

// Add adds a new profile to the cache. This method is used to populate the
// profile cache with pre-existing or recently-added profiles from Core Metadata.
func (p *profileCache) Add(profile models.DeviceProfile) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.add(profile)
}

func (p *profileCache) add(profile models.DeviceProfile) errors.EdgeX {
	if _, ok := p.deviceProfileMap[profile.Name]; ok {
		errMsg := fmt.Sprintf("Profile %s has already existed in cache", profile.Name)
		return errors.NewCommonEdgeX(errors.KindDuplicateName, errMsg, nil)
	}

	p.deviceProfileMap[profile.Name] = &profile
	p.deviceResourceMap[profile.Name] = deviceResourceSliceToMap(profile.DeviceResources)
	p.getResourceOperationsMap[profile.Name], p.setResourceOperationsMap[profile.Name] = deviceCommandSliceToMap(profile.DeviceCommands)
	p.commandsMap[profile.Name] = commandSliceToMap(profile.CoreCommands)
	return nil
}

func deviceResourceSliceToMap(deviceResources []models.DeviceResource) map[string]models.DeviceResource {
	result := make(map[string]models.DeviceResource, len(deviceResources))
	for _, dr := range deviceResources {
		result[dr.Name] = dr
	}

	return result
}

func deviceCommandSliceToMap(deviceCommands []models.DeviceCommand) (map[string][]models.ResourceOperation, map[string][]models.ResourceOperation) {
	getResult := make(map[string][]models.ResourceOperation, len(deviceCommands))
	setResult := make(map[string][]models.ResourceOperation, len(deviceCommands))
	for _, deviceCommand := range deviceCommands {
		if len(deviceCommand.Get) > 0 {
			getResult[deviceCommand.Name] = deviceCommand.Get
		}
		if len(deviceCommand.Set) > 0 {
			setResult[deviceCommand.Name] = deviceCommand.Set
		}
	}

	return getResult, setResult
}

func commandSliceToMap(commands []models.Command) map[string]models.Command {
	result := make(map[string]models.Command, len(commands))
	for _, cmd := range commands {
		result[cmd.Name] = cmd
	}

	return result
}

// Update updates the profile in the cache
func (p *profileCache) Update(profile models.DeviceProfile) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.removeByName(profile.Name); err != nil {
		return err
	}

	return p.add(profile)
}

// RemoveByName removes the specified profile by name from the cache.
func (p *profileCache) RemoveByName(name string) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removeByName(name)
}

func (p *profileCache) removeByName(name string) errors.EdgeX {
	_, ok := p.deviceProfileMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find Profile %s in cache", name)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	delete(p.deviceProfileMap, name)
	delete(p.deviceResourceMap, name)
	delete(p.getResourceOperationsMap, name)
	delete(p.setResourceOperationsMap, name)
	delete(p.commandsMap, name)
	return nil
}

// DeviceResource returns the DeviceResource with given profileName and resourceName
func (p *profileCache) DeviceResource(profileName string, resourceName string) (models.DeviceResource, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	drs, ok := p.deviceResourceMap[profileName]
	if !ok {
		return models.DeviceResource{}, ok
	}

	dr, ok := drs[resourceName]
	return dr, ok
}

// CommandExists returns a bool indicating whether the specified command exists for the
// specified (by name) profile. If the specified profile doesn't exist, an error is returned.
func (p *profileCache) CommandExists(profileName string, cmd string, method string) (bool, errors.EdgeX) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	_, ok := p.deviceProfileMap[profileName]
	if !ok {
		errMsg := fmt.Sprintf("failed to find Profile %s in cache", profileName)
		return false, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// Check whether cmd exists in deviceCommands.
	var deviceCommands map[string][]models.ResourceOperation
	if strings.ToLower(method) == common.GetCmdMethod {
		deviceCommands, _ = p.getResourceOperationsMap[profileName]
	} else {
		deviceCommands, _ = p.setResourceOperationsMap[profileName]
	}

	if _, ok := deviceCommands[cmd]; !ok {
		return false, nil
	}

	return true, nil
}

// ResourceOperations returns the ResourceOperations with given command and method.
func (p *profileCache) ResourceOperations(profileName string, cmd string, method string) ([]models.ResourceOperation, errors.EdgeX) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if err := p.verifyProfileExists(profileName); err != nil {
		return nil, err
	}

	rosMap, err := p.verifyResourceOperationsExists(method, profileName)
	if err != nil {
		return nil, err
	}

	var ok bool
	var ros []models.ResourceOperation
	if ros, ok = rosMap[cmd]; !ok {
		errMsg := fmt.Sprintf("failed to find DeviceCommand %s in Profile %s", cmd, profileName)
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	return ros, nil
}

// ResourceOperation returns the first matched ResourceOperation with given deviceResource and method
func (p *profileCache) ResourceOperation(profileName string, deviceResource string, method string) (models.ResourceOperation, errors.EdgeX) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if err := p.verifyProfileExists(profileName); err != nil {
		return models.ResourceOperation{}, err
	}

	rosMap, err := p.verifyResourceOperationsExists(method, profileName)
	if err != nil {
		return models.ResourceOperation{}, err
	}

	for _, ros := range rosMap {
		for _, ro := range ros {
			if ro.DeviceResource == deviceResource {
				return ro, nil
			}
		}
	}

	errMsg := fmt.Sprintf("failed to find ResourceOpertaion with DeviceResource %s in Profile %s", deviceResource, profileName)
	return models.ResourceOperation{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
}

func (p *profileCache) verifyProfileExists(profileName string) errors.EdgeX {
	if _, ok := p.deviceProfileMap[profileName]; !ok {
		errMsg := fmt.Sprintf("failed to find Profile %s in cache", profileName)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	return nil
}

func (p *profileCache) verifyResourceOperationsExists(method string, profileName string) (map[string][]models.ResourceOperation, errors.EdgeX) {
	var ok bool
	var rosMap map[string][]models.ResourceOperation

	if strings.ToLower(method) == common.GetCmdMethod {
		rosMap, ok = p.getResourceOperationsMap[profileName]
	} else if strings.ToLower(method) == common.SetCmdMethod {
		rosMap, ok = p.setResourceOperationsMap[profileName]
	}

	if !ok {
		errMsg := fmt.Sprintf("failed to find %s ResourceOperations in Profile %s", method, profileName)
		return rosMap, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	return rosMap, nil
}

func Profiles() ProfileCache {
	return pc
}
