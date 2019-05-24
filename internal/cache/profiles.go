// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"strings"

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
	CommandExists(profileName string, cmd string) (bool, error)
	ResourceOperations(profileName string, cmd string, method string) ([]contract.ResourceOperation, error)
	ResourceOperation(profileName string, object string, method string) (contract.ResourceOperation, error)
}

type profileCache struct {
	dpMap    map[string]contract.DeviceProfile // key is DeviceProfile name
	nameMap  map[string]string                 // key is id, and value is DeviceProfile name
	drMap    map[string]map[string]contract.DeviceResource
	dcMap    map[string]map[string][]contract.ResourceOperation
	setOpMap map[string]map[string][]contract.ResourceOperation
	ccMap    map[string]map[string]contract.Command
}

func (p *profileCache) ForName(name string) (contract.DeviceProfile, bool) {
	dp, ok := p.dpMap[name]
	return dp, ok
}

func (p *profileCache) ForId(id string) (contract.DeviceProfile, bool) {
	name, ok := p.nameMap[id]
	if !ok {
		return contract.DeviceProfile{}, ok
	}

	dp, ok := p.dpMap[name]
	return dp, ok
}

func (p *profileCache) All() []contract.DeviceProfile {
	ps := make([]contract.DeviceProfile, len(p.dpMap))
	i := 0
	for _, profile := range p.dpMap {
		ps[i] = profile
		i++
	}
	return ps
}

func (p *profileCache) Add(profile contract.DeviceProfile) error {
	if _, ok := p.dpMap[profile.Name]; ok {
		return fmt.Errorf("device profile %s has already existed in cache", profile.Name)
	}
	p.dpMap[profile.Name] = profile
	p.nameMap[profile.Id] = profile.Name
	p.drMap[profile.Name] = deviceResourceSliceToMap(profile.DeviceResources)
	p.dcMap[profile.Name], p.setOpMap[profile.Name] = profileResourceSliceToMaps(profile.DeviceCommands)
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
	if err := p.Remove(profile.Id); err != nil {
		return err
	}
	return p.Add(profile)
}

func (p *profileCache) Remove(id string) error {
	name, ok := p.nameMap[id]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", id)
	}

	return p.RemoveByName(name)
}

func (p *profileCache) RemoveByName(name string) error {
	profile, ok := p.dpMap[name]
	if !ok {
		return fmt.Errorf("device profile %s does not exist in cache", name)
	}

	delete(p.dpMap, name)
	delete(p.nameMap, profile.Id)
	delete(p.drMap, name)
	delete(p.dcMap, name)
	delete(p.setOpMap, name)
	delete(p.ccMap, name)
	return nil
}

func (p *profileCache) DeviceResource(profileName string, resourceName string) (contract.DeviceResource, bool) {
	drs, ok := p.drMap[profileName]
	if !ok {
		return contract.DeviceResource{}, ok
	}

	dr, ok := drs[resourceName]
	return dr, ok
}

// CommandExists returns a bool indicating whether the specified command exists for the
// specified (by name) device. If the specified device doesn't exist, an error is returned.
func (p *profileCache) CommandExists(profileName string, cmd string) (bool, error) {
	commands, ok := p.ccMap[profileName]
	if !ok {
		err := fmt.Errorf("specified profile: %s not found", profileName)
		return false, err
	}

	if _, ok := commands[cmd]; !ok {
		return false, nil
	}

	return true, nil
}

// Get ResourceOperations
func (p *profileCache) ResourceOperations(profileName string, cmd string, method string) ([]contract.ResourceOperation, error) {
	var resOps []contract.ResourceOperation
	var rosMap map[string][]contract.ResourceOperation
	var ok bool
	if strings.ToLower(method) == common.GetCmdMethod {
		if rosMap, ok = p.dcMap[profileName]; !ok {
			return nil, fmt.Errorf("specified profile: %s not found", profileName)
		}
	} else if strings.ToLower(method) == common.SetCmdMethod {
		if rosMap, ok = p.setOpMap[profileName]; !ok {
			return nil, fmt.Errorf("specified profile: %s not found", profileName)
		}
	}

	if resOps, ok = rosMap[cmd]; !ok {
		return nil, fmt.Errorf("specified cmd: %s not found", cmd)
	}
	return resOps, nil
}

// Return the first matched ResourceOperation
func (p *profileCache) ResourceOperation(profileName string, object string, method string) (contract.ResourceOperation, error) {
	var ro contract.ResourceOperation
	var rosMap map[string][]contract.ResourceOperation
	var ok bool
	if strings.ToLower(method) == common.GetCmdMethod {
		if rosMap, ok = p.dcMap[profileName]; !ok {
			return ro, fmt.Errorf("specified profile: %s not found", profileName)
		}
	} else if strings.ToLower(method) == common.SetCmdMethod {
		if rosMap, ok = p.setOpMap[profileName]; !ok {
			return ro, fmt.Errorf("specified profile: %s not found", profileName)
		}
	}

	if ro, ok = retrieveFirstRObyObject(rosMap, object); !ok {
		return ro, fmt.Errorf("specified ResourceOperation by object %s not found", object)
	}
	return ro, nil
}

func retrieveFirstRObyObject(rosMap map[string][]contract.ResourceOperation, object string) (contract.ResourceOperation, bool) {
	for _, ros := range rosMap {
		for _, ro := range ros {
			if ro.Object == object {
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
	getOpMap := make(map[string]map[string][]contract.ResourceOperation, defaultSize)
	setOpMap := make(map[string]map[string][]contract.ResourceOperation, defaultSize)
	cmdMap := make(map[string]map[string]contract.Command, defaultSize)
	for _, dp := range profiles {
		dpMap[dp.Name] = dp
		nameMap[dp.Id] = dp.Name
		drMap[dp.Name] = deviceResourceSliceToMap(dp.DeviceResources)
		getOpMap[dp.Name], setOpMap[dp.Name] = profileResourceSliceToMaps(dp.DeviceCommands)
		cmdMap[dp.Name] = commandSliceToMap(dp.CoreCommands)
	}
	pc = &profileCache{dpMap: dpMap, nameMap: nameMap, drMap: drMap, dcMap: getOpMap, setOpMap: setOpMap, ccMap: cmdMap}
	return pc
}

func Profiles() ProfileCache {
	if pc == nil {
		InitCache()
	}
	return pc
}
