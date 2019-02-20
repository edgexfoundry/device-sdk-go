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
	"strings"
)

const (
	getOpsStr string = "get"
	setOpsStr string = "set"
)

var (
	pc *profileCache
)

type ProfileCache interface {
	ForName(name string) (models.DeviceProfile, bool)
	ForId(id string) (models.DeviceProfile, bool)
	All() []models.DeviceProfile
	Add(profile models.DeviceProfile) error
	Update(profile models.DeviceProfile) error
	Remove(id string) error
	RemoveByName(name string) error
	DeviceResource(profileName string, resourceName string) (models.DeviceResource, bool)
	CommandExists(profileName string, cmd string) (bool, error)
	ResourceOperations(profileName string, cmd string, method string) ([]models.ResourceOperation, error)
	ResourceOperation(profileName string, object string, method string) (models.ResourceOperation, error)
}

type profileCache struct {
	dpMap    map[string]models.DeviceProfile // key is DeviceProfile name
	nameMap  map[string]string               // key is id, and value is DeviceProfile name
	drMap    map[string]map[string]models.DeviceResource
	getOpMap map[string]map[string][]models.ResourceOperation
	setOpMap map[string]map[string][]models.ResourceOperation
	cmdMap   map[string]map[string]models.Command
}

func (p *profileCache) ForName(name string) (models.DeviceProfile, bool) {
	dp, ok := p.dpMap[name]
	return dp, ok
}

func (p *profileCache) ForId(id string) (models.DeviceProfile, bool) {
	name, ok := p.nameMap[id]
	if !ok {
		return models.DeviceProfile{}, ok
	}

	dp, ok := p.dpMap[name]
	return dp, ok
}

func (p *profileCache) All() []models.DeviceProfile {
	ps := make([]models.DeviceProfile, len(p.dpMap))
	i := 0
	for _, profile := range p.dpMap {
		ps[i] = profile
		i++
	}
	return ps
}

func (p *profileCache) Add(profile models.DeviceProfile) error {
	if _, ok := p.dpMap[profile.Name]; ok {
		return fmt.Errorf("device profile %s has already existed in cache", profile.Name)
	}
	p.dpMap[profile.Name] = profile
	p.nameMap[profile.Id] = profile.Name
	p.drMap[profile.Name] = deviceResourceSliceToMap(profile.DeviceResources)
	p.getOpMap[profile.Name], p.setOpMap[profile.Name] = profileResourceSliceToMaps(profile.Resources)
	p.cmdMap[profile.Name] = commandSliceToMap(profile.Commands)
	return nil
}

func deviceResourceSliceToMap(deviceResources []models.DeviceResource) map[string]models.DeviceResource {
	result := make(map[string]models.DeviceResource, len(deviceResources))
	for _, dr := range deviceResources {
		result[dr.Name] = dr
	}
	return result
}

func profileResourceSliceToMaps(profileResources []models.ProfileResource) (map[string][]models.ResourceOperation, map[string][]models.ResourceOperation) {
	getResult := make(map[string][]models.ResourceOperation, len(profileResources))
	setResult := make(map[string][]models.ResourceOperation, len(profileResources))
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

func commandSliceToMap(commands []models.Command) map[string]models.Command {
	result := make(map[string]models.Command, len(commands))
	for _, cmd := range commands {
		result[cmd.Name] = cmd
	}
	return result
}

func (p *profileCache) Update(profile models.DeviceProfile) error {
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
	delete(p.getOpMap, name)
	delete(p.setOpMap, name)
	delete(p.cmdMap, name)
	return nil
}

func (p *profileCache) DeviceResource(profileName string, resourceName string) (models.DeviceResource, bool) {
	drs, ok := p.drMap[profileName]
	if !ok {
		return models.DeviceResource{}, ok
	}

	dr, ok := drs[resourceName]
	return dr, ok
}

// CommandExists returns a bool indicating whether the specified command exists for the
// specified (by name) device. If the specified device doesn't exist, an error is returned.
func (p *profileCache) CommandExists(profileName string, cmd string) (bool, error) {
	commands, ok := p.cmdMap[profileName]
	if !ok {
		err := fmt.Errorf("profiles: CommandExists: specified profile: %s not found", profileName)
		return false, err
	}

	if _, ok := commands[cmd]; !ok {
		return false, nil
	}

	return true, nil
}

// Get ResourceOperations
func (p *profileCache) ResourceOperations(profileName string, cmd string, method string) ([]models.ResourceOperation, error) {
	var resOps []models.ResourceOperation
	var rosMap map[string][]models.ResourceOperation
	var ok bool
	if strings.ToLower(method) == getOpsStr {
		if rosMap, ok = p.getOpMap[profileName]; !ok {
			return nil, fmt.Errorf("profiles: ResourceOperations: specified profile: %s not found", profileName)
		}
	} else if strings.ToLower(method) == setOpsStr {
		if rosMap, ok = p.setOpMap[profileName]; !ok {
			return nil, fmt.Errorf("profiles: ResourceOperations: specified profile: %s not found", profileName)
		}
	}

	if resOps, ok = rosMap[cmd]; !ok {
		return nil, fmt.Errorf("profiles: ResourceOperations: specified cmd: %s not found", cmd)
	}
	return resOps, nil
}

// Return the first matched ResourceOperation
func (p *profileCache) ResourceOperation(profileName string, object string, method string) (models.ResourceOperation, error) {
	var ro models.ResourceOperation
	var rosMap map[string][]models.ResourceOperation
	var ok bool
	if strings.ToLower(method) == getOpsStr {
		if rosMap, ok = p.getOpMap[profileName]; !ok {
			return ro, fmt.Errorf("profiles: ResourceOperation: specified profile: %s not found", profileName)
		}
	} else if strings.ToLower(method) == setOpsStr {
		if rosMap, ok = p.setOpMap[profileName]; !ok {
			return ro, fmt.Errorf("profiles: ResourceOperations: specified profile: %s not found", profileName)
		}
	}

	if ro, ok = retrieveFirstRObyObject(rosMap, object); !ok {
		return ro, fmt.Errorf("profiles: specified ResourceOperation by object %s not found", object)
	}
	return ro, nil
}

func retrieveFirstRObyObject(rosMap map[string][]models.ResourceOperation, object string) (models.ResourceOperation, bool) {
	for _, ros := range rosMap {
		for _, ro := range ros {
			if ro.Object == object {
				return ro, true
			}
		}
	}
	return models.ResourceOperation{}, false
}

func newProfileCache(profiles []models.DeviceProfile) ProfileCache {
	defaultSize := len(profiles) * 2
	dpMap := make(map[string]models.DeviceProfile, defaultSize)
	nameMap := make(map[string]string, defaultSize)
	drMap := make(map[string]map[string]models.DeviceResource, defaultSize)
	getOpMap := make(map[string]map[string][]models.ResourceOperation, defaultSize)
	setOpMap := make(map[string]map[string][]models.ResourceOperation, defaultSize)
	cmdMap := make(map[string]map[string]models.Command, defaultSize)
	for _, dp := range profiles {
		dpMap[dp.Name] = dp
		nameMap[dp.Id] = dp.Name
		drMap[dp.Name] = deviceResourceSliceToMap(dp.DeviceResources)
		getOpMap[dp.Name], setOpMap[dp.Name] = profileResourceSliceToMaps(dp.Resources)
		cmdMap[dp.Name] = commandSliceToMap(dp.Commands)
	}
	pc = &profileCache{dpMap: dpMap, nameMap: nameMap, drMap: drMap, getOpMap: getOpMap, setOpMap: setOpMap, cmdMap: cmdMap}
	return pc
}

func Profiles() ProfileCache {
	if pc == nil {
		InitCache()
	}
	return pc
}
