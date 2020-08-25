// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/general"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

const (
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

func LoadProfiles(path string, dic *di.Container) error {
	if path == "" {
		return nil
	}
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	absPath, err := filepath.Abs(path)
	if err != nil {
		lc.Error(fmt.Sprintf("profiles: couldn't create absolute path for: %s; %v", path, err))
		return err
	}
	lc.Debug(fmt.Sprintf("created absolute path for loading pre-defined Device Profiles: %s", absPath))

	dpc := container.MetadataDeviceProfileClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	profiles, err := dpc.DeviceProfiles(ctx)
	if err != nil {
		lc.Error(fmt.Sprintf("couldn't read Device Profile from Core Metadata: %v", err))
		return err
	}
	pMap := profileSliceToMap(profiles)

	fileInfo, err := ioutil.ReadDir(absPath)
	if err != nil {
		lc.Error(fmt.Sprintf("profiles: couldn't read directory: %s; %v", absPath, err))
		return err
	}

	for _, file := range fileInfo {
		var profile contract.DeviceProfile

		fName := file.Name()
		lfName := strings.ToLower(fName)
		if strings.HasSuffix(lfName, yamlExt) || strings.HasSuffix(lfName, ymlExt) {
			fullPath := absPath + "/" + fName
			yamlFile, err := ioutil.ReadFile(fullPath)
			if err != nil {
				lc.Error(fmt.Sprintf("profiles: couldn't read file: %s; %v", fullPath, err))
				continue
			}

			err = yaml.Unmarshal(yamlFile, &profile)
			if err != nil {
				lc.Error(fmt.Sprintf("invalid Device Profile: %s; %v", fullPath, err))
				continue
			}

			// TODO: this section will be removed after the deprecated fields are truly removed
			handleDeprecatedFields(&profile)

			// if profile already exists in metadata, skip it
			if p, ok := pMap[profile.Name]; ok {
				_ = cache.Profiles().Add(p)
				continue
			}

			// add profile to metadata
			ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
			id, err := dpc.Add(ctx, &profile)
			if err != nil {
				lc.Error(fmt.Sprintf("Add Device Profile %s to Core Metadata failed: %v", fullPath, err))
				continue
			}
			if err = common.VerifyIdFormat(id, "Device Profile"); err != nil {
				return err
			}

			profile.Id = id
			cache.Profiles().Add(profile)
			CreateDescriptorsFromProfile(
				&profile,
				lc,
				container.GeneralClientFrom(dic.Get),
				container.CoredataValueDescriptorClientFrom(dic.Get))
		}
	}
	return nil
}

func handleDeprecatedFields(profile *contract.DeviceProfile) {
	for _, pr := range profile.DeviceCommands {
		for i, ro := range pr.Get {
			pr.Get[i] = handleRODeprecatedFields(ro)
		}
		for i, ro := range pr.Set {
			pr.Set[i] = handleRODeprecatedFields(ro)
		}
	}
}

func handleRODeprecatedFields(ro contract.ResourceOperation) contract.ResourceOperation {
	if ro.DeviceResource != "" {
		ro.Object = ro.DeviceResource
	} else if ro.Object != "" {
		ro.DeviceResource = ro.Object
	}
	if ro.DeviceCommand != "" {
		ro.Resource = ro.DeviceCommand
	} else if ro.Resource != "" {
		ro.DeviceCommand = ro.Resource
	}
	return ro
}

func profileSliceToMap(profiles []contract.DeviceProfile) map[string]contract.DeviceProfile {
	result := make(map[string]contract.DeviceProfile, len(profiles))
	for _, dp := range profiles {
		result[dp.Name] = dp
	}
	return result
}

func CreateDescriptorsFromProfile(
	profile *contract.DeviceProfile,
	lc logger.LoggingClient,
	gc general.GeneralClient,
	vdc coredata.ValueDescriptorClient) {
	if isValueDescriptorManagedByMetadata(lc, gc) {
		lc.Debug("Value Descriptor is now managed by Core Metadata")
		return
	}

	dcs := profile.DeviceCommands
	for _, dc := range dcs {
		for _, op := range dc.Get {
			createDescriptorFromResourceOperation(profile.Name, op, lc, vdc)
		}
		for _, op := range dc.Set {
			createDescriptorFromResourceOperation(profile.Name, op, lc, vdc)
		}
	}
}

// This is a temporary solution and will move the whole
// Value Descriptor management logic to Core Metadata in Geneva
func isValueDescriptorManagedByMetadata(lc logger.LoggingClient, gc general.GeneralClient) bool {
	lc.Debug("Getting EnableValueDescriptorManagement configuration value from Core Metadata")
	correlation := uuid.New().String()
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, correlation)

	configString, err := gc.FetchConfiguration(ctx)
	if err != nil {
		lc.Error(fmt.Sprintf("Error when getting configuration from Core Metadata: %v ", err))
		return false
	}

	var metadataConfig map[string]interface{}
	err = json.Unmarshal([]byte(configString), &metadataConfig)
	if err != nil {
		lc.Error(fmt.Sprintf("Error when parsing configuration from Core Metadata: %v ", err))
		return false
	}

	writable, ok := metadataConfig["Writable"].(map[string]interface{})
	if !ok {
		lc.Error(fmt.Sprintf("Error when retrieving Writable configuration from Core Metadata: %v", metadataConfig))
		return false
	}
	enableValueDescriptorManagement, ok := writable["EnableValueDescriptorManagement"].(bool)
	if !ok {
		lc.Error(fmt.Sprintf("Error when retrieving EnableValueDescriptorManagement configuration from Core Metadata: %v", writable))
		return false
	}

	return enableValueDescriptorManagement
}

func createDescriptorFromResourceOperation(
	profileName string,
	op contract.ResourceOperation,
	lc logger.LoggingClient,
	vdc coredata.ValueDescriptorClient) {
	if _, ok := cache.ValueDescriptors().ForName(op.DeviceResource); ok {
		// Value Descriptor has been created
		return
	} else {
		dr, ok := cache.Profiles().DeviceResource(profileName, op.DeviceResource)
		if !ok {
			lc.Error(fmt.Sprintf("can't find Device Resource %s to match Device Command (Resource Operation) %v in Device Profile %s", op.DeviceResource, op, profileName))
		}
		desc, err := createDescriptor(op.DeviceResource, dr, lc, vdc)
		if err != nil {
			lc.Error(fmt.Sprintf("created Value Descriptor %v failed: %v", desc, err))
		} else {
			_ = cache.ValueDescriptors().Add(*desc)
		}
	}
}

func createDescriptor(
	name string,
	dr contract.DeviceResource,
	lc logger.LoggingClient,
	vdc coredata.ValueDescriptorClient) (*contract.ValueDescriptor, error) {
	value := dr.Properties.Value
	units := dr.Properties.Units

	lc.Debug(fmt.Sprintf("createing ValueDescriptor: %s, value: %v, units: %v", name, value, units))

	desc := &contract.ValueDescriptor{
		Name:          name,
		Min:           value.Minimum,
		Max:           value.Maximum,
		Type:          value.Type,
		UomLabel:      units.DefaultValue,
		DefaultValue:  value.DefaultValue,
		Formatting:    "%s",
		Description:   dr.Description,
		FloatEncoding: value.FloatEncoding,
		MediaType:     value.MediaType,
	}

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err := vdc.Add(ctx, desc)
	if err != nil {
		return nil, err
	}

	if err = common.VerifyIdFormat(id, "Value Descriptor"); err != nil {
		return nil, err
	}
	desc.Id = id

	return desc, nil
}
