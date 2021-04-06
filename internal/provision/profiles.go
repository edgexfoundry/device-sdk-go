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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
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

	fileInfo, err := ioutil.ReadDir(absPath)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory", err)
	}

	var addProfilesReq []requests.DeviceProfileRequest
	dpc := container.MetadataDeviceProfileClientFrom(dic.Get)
	lc.Infof("Loading pre-defined profiles from %s", absPath)
	for _, file := range fileInfo {
		var profile dtos.DeviceProfile

		fName := file.Name()
		lfName := strings.ToLower(fName)
		if strings.HasSuffix(lfName, yamlExt) || strings.HasSuffix(lfName, ymlExt) {
			fullPath := filepath.Join(absPath, fName)
			yamlFile, err := ioutil.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}

			err = yaml.Unmarshal(yamlFile, &profile)
			if err != nil {
				lc.Errorf("Failed to decode profile %s: %v", fullPath, err)
				continue
			}

			_, err = dpc.DeviceProfileByName(context.Background(), profile.Name)
			if err == nil {
				lc.Infof("Profile %s exists, using the existing one", profile.Name)
			} else {
				lc.Infof("Profile %s not found in Metadata, adding it ...", profile.Name)
				req := requests.NewDeviceProfileRequest(profile)
				addProfilesReq = append(addProfilesReq, req)
			}
		}
	}
	if len(addProfilesReq) == 0 {
		return nil
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, edgexErr := dpc.Add(ctx, addProfilesReq)
	return edgexErr
}
