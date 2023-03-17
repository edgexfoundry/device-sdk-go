// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
)

const (
	jsonExt = ".json"
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

func LoadProfiles(path string, dic *di.Container) errors.EdgeX {
	if path == "" {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path", err)
	}

	files, err := os.ReadDir(absPath)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory", err)
	}

	if len(files) == 0 {
		return nil
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	lc.Infof("Loading pre-defined profiles from %s(%d files found)", absPath, len(files))

	var addProfilesReq []requests.DeviceProfileRequest
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	for _, file := range files {
		var profile dtos.DeviceProfile
		fullPath := filepath.Join(absPath, file.Name())
		if strings.HasSuffix(fullPath, yamlExt) || strings.HasSuffix(fullPath, ymlExt) {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}

			err = yaml.Unmarshal(content, &profile)
			if err != nil {
				lc.Errorf("Failed to YAML decode profile %s: %v", file.Name(), err)
				continue
			}
		} else if strings.HasSuffix(fullPath, jsonExt) {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}

			err = json.Unmarshal(content, &profile)
			if err != nil {
				lc.Errorf("Failed to JSON decode profile %s: %v", file.Name(), err)
				continue
			}
		} else {
			continue
		}

		res, err := dpc.DeviceProfileByName(context.Background(), profile.Name)
		if err == nil {
			lc.Infof("Profile %s exists, using the existing one", profile.Name)
			_, exist := cache.Profiles().ForName(profile.Name)
			if !exist {
				err = cache.Profiles().Add(dtos.ToDeviceProfileModel(res.Profile))
				if err != nil {
					return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to cache the profile %s", res.Profile.Name), err)
				}
			}
		} else {
			lc.Infof("Profile %s not found in Metadata, adding it ...", profile.Name)
			req := requests.NewDeviceProfileRequest(profile)
			addProfilesReq = append(addProfilesReq, req)
		}
	}

	if len(addProfilesReq) == 0 {
		return nil
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, edgexErr := dpc.Add(ctx, addProfilesReq)
	return edgexErr
}
