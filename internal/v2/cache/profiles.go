// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

var (
	pc *profileCache
)

type ProfileCache interface {
	ForName(name string) (contract.DeviceProfile, bool)
	ForId(id string) (contract.DeviceProfile, bool)
	All() []contract.DeviceProfile
	Add(profile contract.DeviceProfile) edgexErr.EdgeX
	Update(profile contract.DeviceProfile) edgexErr.EdgeX
	Remove(id string) edgexErr.EdgeX
	RemoveByName(name string) edgexErr.EdgeX
	DeviceResource(profileName string, resourceName string) (contract.DeviceResource, bool)
	CommandExists(profileName string, cmd string, method string) (bool, edgexErr.EdgeX)
	ResourceOperations(profileName string, cmd string, method string) ([]contract.ResourceOperation, edgexErr.EdgeX)
	ResourceOperation(profileName string, deviceResource string, method string) (contract.ResourceOperation, edgexErr.EdgeX)
}

type profileCache struct {
	deviceProfileMap         map[string]*contract.DeviceProfile // key is DeviceProfile name
	nameMap                  map[string]string                  // key is id, and value is DeviceProfile name
	deviceResourceMap        map[string]map[string]contract.DeviceResource
	getResourceOperationsMap map[string]map[string][]contract.ResourceOperation
	setResourceOperationsMap map[string]map[string][]contract.ResourceOperation
	commandsMap              map[string]map[string]contract.Command
	mutex                    sync.Mutex
}

// ForName returns a profile with the given profile name.
func (p *profileCache) ForName(name string) (contract.DeviceProfile, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	profile, ok := p.deviceProfileMap[name]
	return *profile, ok
}

// ForName returns a profile with the given profile id.
func (p *profileCache) ForId(id string) (contract.DeviceProfile, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	name, ok := p.nameMap[id]
	if !ok {
		return contract.DeviceProfile{}, ok
	}

	profile, ok := p.deviceProfileMap[name]
	return *profile, ok
}

// All returns the current list of profiles in the cache.
func (p *profileCache) All() []contract.DeviceProfile {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	i := 0
	ps := make([]contract.DeviceProfile, len(p.deviceProfileMap))
	for _, profile := range p.deviceProfileMap {
		ps[i] = *profile
		i++
	}
	return ps
}

// Add adds a new profile to the cache. This method is used to populate the
// profile cache with pre-existing or recently-added profiles from Core Metadata.
func (p *profileCache) Add(profile contract.DeviceProfile) edgexErr.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.add(profile)
}

func (p *profileCache) add(profile contract.DeviceProfile) edgexErr.EdgeX {
	if _, ok := p.deviceProfileMap[profile.Name]; ok {
		errMsg := fmt.Sprintf("device %s has already existed in cache", profile.Name)
		return edgexErr.NewCommonEdgeX(edgexErr.KindDuplicateName, errMsg, nil)
	}

	p.deviceProfileMap[profile.Name] = &profile
	p.nameMap[profile.Id] = profile.Name
	p.deviceResourceMap[profile.Name] = deviceResourceSliceToMap(profile.DeviceResources)
	p.getResourceOperationsMap[profile.Name], p.setResourceOperationsMap[profile.Name] = profileResourceSliceToMaps(profile.DeviceCommands)
	p.commandsMap[profile.Name] = commandSliceToMap(profile.CoreCommands)
	return nil
}

func deviceResourceSliceToMap(deviceResources []contract.DeviceResource) map[string]contract.DeviceResource {
	result := make(map[string]contract.DeviceResource, len(deviceResources))
	for _, dr := range deviceResources {
		result[dr.Name] = dr
	}

	return result
}

func profileResourceSliceToMaps(profileResources []contract.ProfileResource) (map[string][]contract.ResourceOperation, map[string][]contract.ResourceOperation) {
	getResult := make(map[string][]contract.ResourceOperation, len(profileResources))
	setResult := make(map[string][]contract.ResourceOperation, len(profileResources))
	for _, pr := range profileResources {
		if len(pr.Get) > 0 {
			getResult[pr.Name] = pr.Get
		}
		if len(pr.Set) > 0 {
			setResult[pr.Name] = pr.Set
		}
	}

	return getResult, setResult
}

func commandSliceToMap(commands []contract.Command) map[string]contract.Command {
	result := make(map[string]contract.Command, len(commands))
	for _, cmd := range commands {
		result[cmd.Name] = cmd
	}

	return result
}

// Update updates the profile in the cache
func (p *profileCache) Update(profile contract.DeviceProfile) edgexErr.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.remove(profile.Id); err != nil {
		return err
	}

	return p.add(profile)
}

// Remove removes the specified profile by id from the cache.
func (p *profileCache) Remove(id string) edgexErr.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.remove(id)
}

func (p *profileCache) remove(id string) edgexErr.EdgeX {
	name, ok := p.nameMap[id]
	if !ok {
		errMsg := fmt.Sprintf("failed to find profile with given id %s in cache", id)
		return edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	return p.removeByName(name)
}

// RemoveByName removes the specified profile by name from the cache.
func (p *profileCache) RemoveByName(name string) edgexErr.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removeByName(name)
}

func (p *profileCache) removeByName(name string) edgexErr.EdgeX {
	profile, ok := p.deviceProfileMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find profile %s in cache", name)
		return edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	delete(p.deviceProfileMap, name)
	delete(p.nameMap, profile.Id)
	delete(p.deviceResourceMap, name)
	delete(p.getResourceOperationsMap, name)
	delete(p.setResourceOperationsMap, name)
	delete(p.commandsMap, name)
	return nil
}

// DeviceResource returns the DeviceResource with given profileName and resourceName
func (p *profileCache) DeviceResource(profileName string, resourceName string) (contract.DeviceResource, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	drs, ok := p.deviceResourceMap[profileName]
	if !ok {
		return contract.DeviceResource{}, ok
	}

	dr, ok := drs[resourceName]
	return dr, ok
}

// CommandExists returns a bool indicating whether the specified command exists for the
// specified (by name) device. If the specified device doesn't exist, an error is returned.
func (p *profileCache) CommandExists(profileName string, cmd string, method string) (bool, edgexErr.EdgeX) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	_, ok := p.deviceProfileMap[profileName]
	if !ok {
		errMsg := fmt.Sprintf("failed to find profile %s in cache", profileName)
		return false, edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}
	// Check whether cmd exists in deviceCommands.
	var deviceCommands map[string][]contract.ResourceOperation
	if strings.ToLower(method) == common.GetCmdMethod {
		deviceCommands, _ = p.getResourceOperationsMap[profileName]
	} else {
		deviceCommands, _ = p.setResourceOperationsMap[profileName]
	}

	if _, ok := deviceCommands[cmd]; !ok {
		errMsg := fmt.Sprintf("failed to find %s command %s in profile %s", method, cmd, profileName)
		return false, edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	return true, nil
}

// ResourceOperations returns the ResourceOperations with given command and method.
func (p *profileCache) ResourceOperations(profileName string, cmd string, method string) ([]contract.ResourceOperation, edgexErr.EdgeX) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var ok bool
	var ros []contract.ResourceOperation
	var rosMap map[string][]contract.ResourceOperation

	if _, ok = p.deviceProfileMap[profileName]; !ok {
		errMsg := fmt.Sprintf("failed to find profile %s in cache", profileName)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	if strings.ToLower(method) == common.GetCmdMethod {
		rosMap, ok = p.getResourceOperationsMap[profileName]
	} else if strings.ToLower(method) == common.SetCmdMethod {
		rosMap, ok = p.setResourceOperationsMap[profileName]
	}

	if !ok {
		errMsg := fmt.Sprintf("failed to find %s ResourceOperations in profile %s", method, profileName)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	if ros, ok = rosMap[cmd]; !ok {
		errMsg := fmt.Sprintf("failed to find %s command in profile %s", cmd, profileName)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	return ros, nil
}

// ResourceOperation returns the first matched ResourceOperation with given deviceResource and method
func (p *profileCache) ResourceOperation(profileName string, deviceResource string, method string) (contract.ResourceOperation, edgexErr.EdgeX) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var ok bool
	var ro contract.ResourceOperation
	var rosMap map[string][]contract.ResourceOperation

	if _, ok = p.deviceProfileMap[profileName]; !ok {
		errMsg := fmt.Sprintf("failed to find profile %s in cache", profileName)
		return ro, edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	if strings.ToLower(method) == common.GetCmdMethod {
		rosMap, ok = p.getResourceOperationsMap[profileName]
	} else if strings.ToLower(method) == common.SetCmdMethod {
		rosMap, ok = p.setResourceOperationsMap[profileName]
	}

	if !ok {
		errMsg := fmt.Sprintf("failed to find %s ResourceOperations in profile %s", method, profileName)
		return ro, edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}

	if ro, ok = retrieveFirstRObyDeviceResource(rosMap, deviceResource); !ok {
		errMsg := fmt.Sprintf("failed to find %s ResourceOpertaion with DeviceResource %s in profile %s", method, deviceResource, profileName)
		return ro, edgexErr.NewCommonEdgeX(edgexErr.KindInvalidId, errMsg, nil)
	}
	return ro, nil
}

func retrieveFirstRObyDeviceResource(rosMap map[string][]contract.ResourceOperation, deviceResource string) (contract.ResourceOperation, bool) {
	for _, ros := range rosMap {
		for _, ro := range ros {
			if ro.DeviceResource == deviceResource {
				return ro, true
			}
		}
	}

	return contract.ResourceOperation{}, false
}

func newProfileCache(profiles []contract.DeviceProfile) ProfileCache {
	defaultSize := len(profiles) * 2
	dpMap := make(map[string]*contract.DeviceProfile, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	drMap := make(map[string]map[string]contract.DeviceResource, defaultSize)
	getRoMap := make(map[string]map[string][]contract.ResourceOperation, defaultSize)
	setRoMap := make(map[string]map[string][]contract.ResourceOperation, defaultSize)
	cmdMap := make(map[string]map[string]contract.Command, defaultSize)
	for _, dp := range profiles {
		dpMap[dp.Name] = &dp
		nameMap[dp.Id] = dp.Name
		drMap[dp.Name] = deviceResourceSliceToMap(dp.DeviceResources)
		getRoMap[dp.Name], setRoMap[dp.Name] = profileResourceSliceToMaps(dp.DeviceCommands)
		cmdMap[dp.Name] = commandSliceToMap(dp.CoreCommands)
	}
	pc = &profileCache{
		deviceProfileMap:         dpMap,
		nameMap:                  nameMap,
		deviceResourceMap:        drMap,
		getResourceOperationsMap: getRoMap,
		setResourceOperationsMap: setRoMap,
		commandsMap:              cmdMap}
	return pc
}

func Profiles() ProfileCache {
	return pc
}
