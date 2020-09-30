// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	pc *profileCache
)

type ProfileCache interface {
	ForName(name string) (contract.DeviceProfile, bool)
	ForId(id string) (contract.DeviceProfile, bool)
	All() []contract.DeviceProfile
	Add(profile contract.DeviceProfile) error
	Update(profile contract.DeviceProfile) error
	Remove(id string) error
	RemoveByName(name string) error
	DeviceResource(profileName string, resourceName string) (contract.DeviceResource, bool)
	CommandExists(profileName string, cmd string, method string) (bool, error)
	ResourceOperations(profileName string, cmd string, method string) ([]contract.ResourceOperation, error)
	ResourceOperation(profileName string, deviceResource string, method string) (contract.ResourceOperation, error)
}

type profileCache struct {
	dpMap    map[string]contract.DeviceProfile // key is DeviceProfile name
	nameMap  map[string]string                 // key is id, and value is DeviceProfile name
	drMap    map[string]map[string]contract.DeviceResource
	getRoMap map[string]map[string][]contract.ResourceOperation
	setRoMap map[string]map[string][]contract.ResourceOperation
	ccMap    map[string]map[string]contract.Command
	mutex    sync.Mutex
}

func (p *profileCache) ForName(name string) (contract.DeviceProfile, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	dp, ok := p.dpMap[name]
	return dp, ok
}

func (p *profileCache) ForId(id string) (contract.DeviceProfile, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	name, ok := p.nameMap[id]
	if !ok {
		return contract.DeviceProfile{}, ok
	}

	dp, ok := p.dpMap[name]
	return dp, ok
}

func (p *profileCache) All() []contract.DeviceProfile {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	ps := make([]contract.DeviceProfile, len(p.dpMap))
	i := 0
	for _, profile := range p.dpMap {
		ps[i] = profile
		i++
	}
	return ps
}

func (p *profileCache) Add(profile contract.DeviceProfile) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.add(profile)
}

func (p *profileCache) add(profile contract.DeviceProfile) error {
	if _, ok := p.dpMap[profile.Name]; ok {
		return fmt.Errorf("device profile %s has already existed in cache", profile.Name)
	}
	p.dpMap[profile.Name] = profile
	p.nameMap[profile.Id] = profile.Name
	p.drMap[profile.Name] = deviceResourceSliceToMap(profile.DeviceResources)
	p.getRoMap[profile.Name], p.setRoMap[profile.Name] = profileResourceSliceToMaps(profile.DeviceCommands)
	p.ccMap[profile.Name] = commandSliceToMap(profile.CoreCommands)
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

func (p *profileCache) Update(profile contract.DeviceProfile) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.remove(profile.Id); err != nil {
		return err
	}
	return p.add(profile)
}

func (p *profileCache) Remove(id string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.remove(id)
}

func (p *profileCache) remove(id string) error {
	name, ok := p.nameMap[id]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", id)
	}

	return p.removeByName(name)
}

func (p *profileCache) RemoveByName(name string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removeByName(name)
}

func (p *profileCache) removeByName(name string) error {
	profile, ok := p.dpMap[name]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", name)
	}

	delete(p.dpMap, name)
	delete(p.nameMap, profile.Id)
	delete(p.drMap, name)
	delete(p.getRoMap, name)
	delete(p.setRoMap, name)
	delete(p.ccMap, name)
	return nil
}

func (p *profileCache) DeviceResource(profileName string, resourceName string) (contract.DeviceResource, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	drs, ok := p.drMap[profileName]
	if !ok {
		return contract.DeviceResource{}, ok
	}

	dr, ok := drs[resourceName]
	return dr, ok
}

// CommandExists returns a bool indicating whether the specified command exists for the
// specified (by name) device. If the specified device doesn't exist, an error is returned.
func (p *profileCache) CommandExists(profileName string, cmd string, method string) (bool, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	_, profileExist := p.dpMap[profileName]
	if !profileExist {
		err := fmt.Errorf("specified profile: %s not found", profileName)
		return false, err
	}
	// Check whether cmd exists in deviceCommands.
	var deviceCommands map[string][]contract.ResourceOperation
	if strings.ToLower(method) == common.GetCmdMethod {
		deviceCommands, _ = p.getRoMap[profileName]
	} else {
		deviceCommands, _ = p.setRoMap[profileName]
	}

	if _, dcExist := deviceCommands[cmd]; !dcExist {
		return false, nil
	}

	return true, nil
}

// Get ResourceOperations
func (p *profileCache) ResourceOperations(profileName string, cmd string, method string) ([]contract.ResourceOperation, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var resOps []contract.ResourceOperation
	var rosMap map[string][]contract.ResourceOperation
	var ok bool
	if strings.ToLower(method) == common.GetCmdMethod {
		if rosMap, ok = p.getRoMap[profileName]; !ok {
			return nil, fmt.Errorf("specified profile: %s not found", profileName)
		}
	} else if strings.ToLower(method) == common.SetCmdMethod {
		if rosMap, ok = p.setRoMap[profileName]; !ok {
			return nil, fmt.Errorf("specified profile: %s not found", profileName)
		}
	}

	if resOps, ok = rosMap[cmd]; !ok {
		return nil, fmt.Errorf("specified cmd: %s not found", cmd)
	}
	return resOps, nil
}

// Return the first matched ResourceOperation
func (p *profileCache) ResourceOperation(profileName string, deviceResource string, method string) (contract.ResourceOperation, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var ro contract.ResourceOperation
	var rosMap map[string][]contract.ResourceOperation
	var ok bool
	if strings.ToLower(method) == common.GetCmdMethod {
		if rosMap, ok = p.getRoMap[profileName]; !ok {
			return ro, fmt.Errorf("specified profile: %s not found", profileName)
		}
	} else if strings.ToLower(method) == common.SetCmdMethod {
		if rosMap, ok = p.setRoMap[profileName]; !ok {
			return ro, fmt.Errorf("specified profile: %s not found", profileName)
		}
	}

	if ro, ok = retrieveFirstRObyDeviceResource(rosMap, deviceResource); !ok {
		return ro, fmt.Errorf("specified ResourceOperation by deviceResource %s not found", deviceResource)
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
	dpMap := make(map[string]contract.DeviceProfile, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	drMap := make(map[string]map[string]contract.DeviceResource, defaultSize)
	getRoMap := make(map[string]map[string][]contract.ResourceOperation, defaultSize)
	setRoMap := make(map[string]map[string][]contract.ResourceOperation, defaultSize)
	cmdMap := make(map[string]map[string]contract.Command, defaultSize)
	for _, dp := range profiles {
		dpMap[dp.Name] = dp
		nameMap[dp.Id] = dp.Name
		drMap[dp.Name] = deviceResourceSliceToMap(dp.DeviceResources)
		getRoMap[dp.Name], setRoMap[dp.Name] = profileResourceSliceToMaps(dp.DeviceCommands)
		cmdMap[dp.Name] = commandSliceToMap(dp.CoreCommands)
	}
	pc = &profileCache{dpMap: dpMap, nameMap: nameMap, drMap: drMap, getRoMap: getRoMap, setRoMap: setRoMap, ccMap: cmdMap}
	return pc
}

func Profiles() ProfileCache {
	return pc
}
