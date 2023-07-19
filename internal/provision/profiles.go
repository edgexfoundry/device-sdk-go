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
	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/file"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"path/filepath"
)

const (
	jsonExt = ".json"
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

func LoadProfiles(path string, dic *di.Container) errors.EdgeX {
	var addProfilesReq []requests.DeviceProfileRequest
	var edgexErr errors.EdgeX
	if path == "" {
		return nil
	}
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	parsedUrl, err := url.Parse(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to parse profile path as a URI", err)
	}

	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)
		addProfilesReq, edgexErr = loadProfilesFromURI(path, parsedUrl, dpc, secretProvider, lc)
		if edgexErr != nil {
			return edgexErr
		}
	} else {
		addProfilesReq, edgexErr = loadProfilesFromFile(path, dpc, lc)
		if edgexErr != nil {
			return edgexErr
		}
	}

	if len(addProfilesReq) == 0 {
		return nil
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, edgexErr = dpc.Add(ctx, addProfilesReq)
	return edgexErr
}

func loadProfilesFromFile(path string, dpc interfaces.DeviceProfileClient, lc logger.LoggingClient) ([]requests.DeviceProfileRequest, errors.EdgeX) {
	var edgexErr errors.EdgeX
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path for profiles", err)
	}

	files, err := os.ReadDir(absPath)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory for profiles", err)
	}

	if len(files) == 0 {
		return nil, nil
	}

	lc.Infof("Loading pre-defined profiles from %s(%d files found)", absPath, len(files))
	var addProfilesReq, processedProfilesReq []requests.DeviceProfileRequest
	for _, file := range files {
		fullPath := filepath.Join(absPath, file.Name())
		processedProfilesReq, edgexErr = processProfiles(fullPath, fullPath, nil, lc, dpc)
		if edgexErr != nil {
			lc.Error(edgexErr.Error())
			return addProfilesReq, nil
		}
		if len(processedProfilesReq) > 0 {
			addProfilesReq = append(addProfilesReq, processedProfilesReq...)
		}
	}
	return addProfilesReq, nil
}

func loadProfilesFromURI(inputURI string, parsedURI *url.URL, dpc interfaces.DeviceProfileClient, secretProvider bootstrapInterfaces.SecretProvider, lc logger.LoggingClient) ([]requests.DeviceProfileRequest, errors.EdgeX) {
	var edgexErr errors.EdgeX
	redactedURI, err := url.JoinPath(parsedURI.Host, parsedURI.Path)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to create a query-free URI for the profile list", err)
	}
	bytes, err := file.Load(inputURI, secretProvider, lc)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to load Profile list from URI %s", redactedURI), err)
	}

	if len(bytes) == 0 {
		return nil, nil
	}

	var files []string
	err = json.Unmarshal(bytes, &files)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "could not unmarshal Profile list contents", err)
	}
	if len(files) == 0 {
		return nil, nil
	}

	lc.Infof("Loading pre-defined profiles from %s(%d files found)", redactedURI, len(files))
	var addProfilesReq, processedProfilesReq []requests.DeviceProfileRequest
	for _, file := range files {
		fullPath, redactedPath := GetFullAndRedactedURI(parsedURI, file, "profile", lc)
		if fullPath == "" || redactedPath == "" {
			continue
		}

		processedProfilesReq, edgexErr = processProfiles(fullPath, redactedPath, secretProvider, lc, dpc)
		if edgexErr != nil {
			return nil, edgexErr
		}
		if len(processedProfilesReq) > 0 {
			addProfilesReq = append(addProfilesReq, processedProfilesReq...)
		}
	}
	return addProfilesReq, nil
}

func processProfiles(fullPath, displayPath string, secretProvider bootstrapInterfaces.SecretProvider, lc logger.LoggingClient, dpc interfaces.DeviceProfileClient) ([]requests.DeviceProfileRequest, errors.EdgeX) {
	var profile dtos.DeviceProfile
	var addProfilesReq []requests.DeviceProfileRequest

	fileType := GetFileType(fullPath)

	// if the file type is not yaml or json, it cannot be parsed - just return to not break the loop for other devices
	if fileType == OTHER {
		return nil, nil
	}

	content, err := file.Load(fullPath, secretProvider, lc)
	if err != nil {
		lc.Errorf("Failed to read Profiles from %s: %v", displayPath, err)
		return nil, nil
	}

	switch fileType {
	case YAML:
		err = yaml.Unmarshal(content, &profile)
		if err != nil {
			lc.Errorf("Failed to YAML decode device profile from %s: %v", displayPath, err)
			return nil, nil
		}
	case JSON:
		err = json.Unmarshal(content, &profile)
		if err != nil {
			lc.Errorf("Failed to JSON decode device profile from %s: %v", displayPath, err)
			return nil, nil
		}
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
