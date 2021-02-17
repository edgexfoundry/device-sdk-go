// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/cache"
)

const (
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

func LoadProfiles(path string, dic *di.Container) errors.EdgeX {
	if path == "" {
		return nil
	}
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path", err)
	}
	lc.Debugf("created absolute path for loading pre-defined device profiles: %s", absPath)

	dpc := container.MetadataDeviceProfileClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	res, edgexErr := dpc.AllDeviceProfiles(ctx, nil, 0, -1)
	if edgexErr != nil {
		return edgexErr
	}
	profileMap := profileDTOSliceToMap(res.Profiles)

	fileInfo, err := ioutil.ReadDir(absPath)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory", err)
	}

	for _, file := range fileInfo {
		var profile models.DeviceProfile

		fName := file.Name()
		lfName := strings.ToLower(fName)
		if strings.HasSuffix(lfName, yamlExt) || strings.HasSuffix(lfName, ymlExt) {
			fullPath := filepath.Join(absPath, fName)
			yamlFile, err := ioutil.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("failed to read %s: %v", fullPath, err)
				continue
			}

			err = yaml.Unmarshal(yamlFile, &profile)
			if err != nil {
				lc.Errorf("filed to decode profile %s: %v", fullPath, err)
				continue
			}
			// if profile already exists in metadata, skip it
			if p, ok := profileMap[profile.Name]; ok {
				_ = cache.Profiles().Add(p)
				continue
			}

			// add profile to metadata
			ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
			_, err = dpc.AddByYaml(ctx, fullPath)
			if err != nil {
				lc.Errorf("failed to add profile %s to metadata: %v", fullPath, err)
				continue
			}
		}
	}
	return nil
}

func profileDTOSliceToMap(profiles []dtos.DeviceProfile) map[string]models.DeviceProfile {
	result := make(map[string]models.DeviceProfile, len(profiles))
	for _, p := range profiles {
		result[p.Name] = dtos.ToDeviceProfileModel(p)
	}

	return result
}
