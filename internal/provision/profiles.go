// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/file"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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
	var addProfilesReq []requests.DeviceProfileRequest
	if path == "" {
		return nil
	}
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	parsedUrl, err := url.Parse(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to parse url", err)
	}

	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)
		edgexErr := loadProfilesFromUri(lc, path, dpc, secretProvider, addProfilesReq)
		if edgexErr != nil {
			return edgexErr
		}
	} else {
		edgexErr := loadProfilesFromFile(lc, path, dpc, addProfilesReq)
		if edgexErr != nil {
			return edgexErr
		}
	}

	if len(addProfilesReq) == 0 {
		return nil
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, edgexErr := dpc.Add(ctx, addProfilesReq)
	return edgexErr
}

func loadProfilesFromFile(lc logger.LoggingClient, path string, dpc interfaces.DeviceProfileClient, addProfilesReq []requests.DeviceProfileRequest) errors.EdgeX {
	var edgexErr errors.EdgeX
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

	lc.Infof("Loading pre-defined profiles from %s(%d files found)", absPath, len(files))

	for _, file := range files {
		fullPath := filepath.Join(absPath, file.Name())
		addProfilesReq, edgexErr = processProfiles(lc, fullPath, nil, dpc, addProfilesReq)
		if edgexErr != nil {
			return nil
		}
	}
	return nil
}

func loadProfilesFromUri(lc logger.LoggingClient, inputUri string, dpc interfaces.DeviceProfileClient, secretProvider bootstrapInterfaces.SecretProvider, addProfilesReq []requests.DeviceProfileRequest) errors.EdgeX {
	var edgexErr errors.EdgeX
	bytes, err := file.Load(inputUri, timeout, secretProvider)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory", err)
	}

	if len(bytes) == 0 {
		return nil
	}

	var files []string
	err = json.Unmarshal(bytes, &files)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "could not unmarshal Provision Watcher contents", err)
	}
	if len(files) == 0 {
		return nil
	}

	baseUrl, _ := path.Split(inputUri)
	parsedUrl, err := url.Parse(inputUri)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "could not parse uri for Provision Watcher", err)
	}
	lc.Infof("Loading pre-defined profiles from %s(%d files found)", parsedUrl.Redacted(), len(files))

	for _, file := range files {
		fullPath, err := url.JoinPath(baseUrl, file)
		if err != nil {
			lc.Error("could not join uri path for profile %s: %v", file, err)
			continue
		}
		addProfilesReq, edgexErr = processProfiles(lc, fullPath, secretProvider, dpc, addProfilesReq)
		if edgexErr != nil {
			return edgexErr
		}
	}
	return nil
}

func processProfiles(lc logger.LoggingClient, path string, secretProvider bootstrapInterfaces.SecretProvider, dpc interfaces.DeviceProfileClient, addProfilesReq []requests.DeviceProfileRequest) ([]requests.DeviceProfileRequest, errors.EdgeX) {
	var profile dtos.DeviceProfile
	bytes, err := file.Load(path, 10*time.Second, secretProvider)
	if err != nil {
		lc.Errorf("Failed to read %s: %v", path, err)
		return addProfilesReq, nil
	}
	if strings.HasSuffix(path, yamlExt) || strings.HasSuffix(path, ymlExt) {
		err = yaml.Unmarshal(bytes, &profile)
		if err != nil {
			lc.Errorf("Failed to YAML decode device profile: %v", err)
			return addProfilesReq, nil
		}
	} else if strings.HasSuffix(path, jsonExt) {
		err = json.Unmarshal(bytes, &profile)
		if err != nil {
			lc.Errorf("Failed to JSON decode device profile: %v", err)
			return addProfilesReq, nil
		}
	} else {
		return addProfilesReq, nil
	}

	res, err := dpc.DeviceProfileByName(context.Background(), profile.Name)
	if err == nil {
		lc.Infof("Profile %s exists, using the existing one", profile.Name)
		_, exist := cache.Profiles().ForName(profile.Name)
		if !exist {
			err = cache.Profiles().Add(dtos.ToDeviceProfileModel(res.Profile))
			if err != nil {
				return addProfilesReq, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to cache the profile %s", res.Profile.Name), err)
			}
		}
	} else {
		lc.Infof("Profile %s not found in Metadata, adding it ...", profile.Name)
		req := requests.NewDeviceProfileRequest(profile)
		addProfilesReq = append(addProfilesReq, req)
	}
	return addProfilesReq, nil
}
