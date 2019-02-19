// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

const (
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

func LoadProfiles(path string) error {
	if path == "" {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("profiles: couldn't create absolute path for: %s; %v", path, err))
		return err
	}
	common.LoggingClient.Debug(fmt.Sprintf("profiles: created absolute path for loading pre-defined Device Profiles: %s", absPath))

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	profiles, err := common.DeviceProfileClient.DeviceProfiles(ctx)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("profiles: couldn't read Device Profile from Core Metadata: %v", err))
		return err
	}
	pMap := profileSliceToMap(profiles)

	fileInfo, err := ioutil.ReadDir(absPath)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("profiles: couldn't read directory: %s; %v", absPath, err))
		return err
	}

	for _, file := range fileInfo {
		var profile models.DeviceProfile

		fName := file.Name()
		lfName := strings.ToLower(fName)
		if strings.HasSuffix(lfName, yamlExt) || strings.HasSuffix(lfName, ymlExt) {
			fullPath := absPath + "/" + fName
			yamlFile, err := ioutil.ReadFile(fullPath)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("profiles: couldn't read file: %s; %v", fullPath, err))
				continue
			}

			err = yaml.Unmarshal(yamlFile, &profile)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("profiles: invalid Device Profile: %s; %v", fullPath, err))
				continue
			}

			// if profile already exists in metadata, skip it
			if p, ok := pMap[profile.Name]; ok {
				cache.Profiles().Add(p)
				continue
			}

			// add profile to metadata
			ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
			id, err := common.DeviceProfileClient.Add(&profile, ctx)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("profiles: Add Device Profile: %s to Core Metadata failed: %v", fullPath, err))
				continue
			}
			if err = common.VerifyIdFormat(id, "Device Profile"); err != nil {
				return err
			}

			profile.Id = id
			cache.Profiles().Add(profile)
			CreateDescriptorsFromProfile(&profile)
		}
	}
	return nil
}

func profileSliceToMap(profiles []models.DeviceProfile) map[string]models.DeviceProfile {
	result := make(map[string]models.DeviceProfile, len(profiles))
	for _, dp := range profiles {
		result[dp.Name] = dp
	}
	return result
}

func CreateDescriptorsFromProfile(profile *models.DeviceProfile) {
	prs := profile.Resources
	for _, pr := range prs {
		for _, op := range pr.Get {
			createDescriptorFromResourceOperation(profile.Name, op)
		}
		for _, op := range pr.Set {
			createDescriptorFromResourceOperation(profile.Name, op)
		}
	}

}

func createDescriptorFromResourceOperation(profileName string, op models.ResourceOperation) {
	if _, ok := cache.ValueDescriptors().ForName(op.Parameter); ok {
		// Value Descriptor has been created
		return
	} else {
		dr, ok := cache.Profiles().DeviceResource(profileName, op.Object)
		if !ok {
			common.LoggingClient.Error(fmt.Sprintf("can't find Device Object %s to match Resource Operation %v in Device Profile %s", op.Object, op, profileName))
		}
		desc, err := createDescriptor(op.Parameter, dr)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("createing Value Descriptor %v failed: %v", desc, err))
		} else {
			cache.ValueDescriptors().Add(*desc)
		}
	}
}

func createDescriptor(name string, dr models.DeviceResource) (*models.ValueDescriptor, error) {
	value := dr.Properties.Value
	units := dr.Properties.Units

	common.LoggingClient.Debug(fmt.Sprintf("ps: createDescriptor: %s, value: %v, units: %v", name, value, units))

	desc := &models.ValueDescriptor{
		Name:         name,
		Min:          value.Minimum,
		Max:          value.Maximum,
		Type:         value.Type,
		UomLabel:     units.DefaultValue,
		DefaultValue: value.DefaultValue,
		Formatting:   "%s",
		Description:  dr.Description,
	}

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	id, err := common.ValueDescriptorClient.Add(desc, ctx)
	if err != nil {
		return nil, err
	}

	if err = common.VerifyIdFormat(id, "Value Descriptor"); err != nil {
		return nil, err
	}

	desc.Id = id
	common.LoggingClient.Debug(fmt.Sprintf("profiles: created Value Descriptor id: %s", id))

	return desc, nil
}
