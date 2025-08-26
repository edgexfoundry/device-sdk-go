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
	"net/url"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/file"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const (
	jsonExt = ".json"
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

func LoadProfiles(path string, overwrite bool, dic *di.Container) errors.EdgeX {
	var addProfilesReq []requests.DeviceProfileRequest
	var updateProfilesReq []requests.DeviceProfileRequest
	var edgexErr errors.EdgeX
	if path == "" {
		return nil
	}
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	parsedUrl, err := url.Parse(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to parse Device Profile path as a URI", err)
	}

	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)
		addProfilesReq, updateProfilesReq, edgexErr = loadProfilesFromURI(path, parsedUrl, overwrite, dpc, secretProvider, lc)
		if edgexErr != nil {
			return edgexErr
		}
	} else {
		addProfilesReq, updateProfilesReq, edgexErr = loadProfilesFromFile(path, overwrite, dpc, lc)
		if edgexErr != nil {
			return edgexErr
		}
	}

	if len(addProfilesReq) == 0 && len(updateProfilesReq) == 0 {
		return nil
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, edgexErr = dpc.Add(ctx, addProfilesReq)
	if edgexErr != nil {
		return edgexErr
	}

	_, edgexErr = dpc.Update(ctx, updateProfilesReq)

	return edgexErr
}

func loadProfilesFromFile(path string, overwrite bool, dpc interfaces.DeviceProfileClient, lc logger.LoggingClient) ([]requests.DeviceProfileRequest, []requests.DeviceProfileRequest, errors.EdgeX) {
	var edgexErr errors.EdgeX
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path for profiles", err)
	}

	files, err := os.ReadDir(absPath)
	if err != nil {
		return nil, nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory for profiles", err)
	}

	if len(files) == 0 {
		return nil, nil, nil
	}

	lc.Infof("Loading pre-defined Device Profiles from %s(%d files found)", absPath, len(files))
	var addProfilesReq, processedProfilesReq []requests.DeviceProfileRequest
	var updateProfilesReq, processedUpdateProfilesReq []requests.DeviceProfileRequest
	for _, file := range files {
		fullPath := filepath.Join(absPath, file.Name())
		processedProfilesReq, processedUpdateProfilesReq, edgexErr = processProfiles(fullPath, fullPath, overwrite, nil, lc, dpc)
		if edgexErr != nil {
			return nil, nil, edgexErr
		}
		if len(processedProfilesReq) > 0 {
			addProfilesReq = append(addProfilesReq, processedProfilesReq...)
		}

		if len(processedUpdateProfilesReq) > 0 {
			updateProfilesReq = append(updateProfilesReq, processedUpdateProfilesReq...)
		}
	}
	return addProfilesReq, updateProfilesReq, nil
}

func loadProfilesFromURI(inputURI string, parsedURI *url.URL, overwrite bool, dpc interfaces.DeviceProfileClient, secretProvider bootstrapInterfaces.SecretProvider, lc logger.LoggingClient) ([]requests.DeviceProfileRequest, []requests.DeviceProfileRequest, errors.EdgeX) {
	// the input URI contains the index file containing the Profile list to be loaded
	bytes, err := file.Load(inputURI, secretProvider, lc)
	if err != nil {
		return nil, nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to load Device Profile list from URI %s", parsedURI.Redacted()), err)
	}

	var files map[string]string
	err = json.Unmarshal(bytes, &files)
	if err != nil {
		return nil, nil, errors.NewCommonEdgeX(errors.KindServerError, "could not unmarshal Profile list contents", err)
	}
	if len(files) == 0 {
		lc.Infof("Index file %s for Device Profiles list is empty", parsedURI.Redacted())
		return nil, nil, nil
	}

	lc.Infof("Loading pre-defined Device Profiles from %s(%d files found)", parsedURI.Redacted(), len(files))
	var addProfilesReq, processedProfilesReq []requests.DeviceProfileRequest
	var updateProfilesReq, processedUpdateProfilesReq []requests.DeviceProfileRequest
	for name, file := range files {
		add, update, edgexErr := checkDeviceProfile(name, overwrite, dpc, lc)
		if !add && !update {
			if edgexErr != nil {
				return addProfilesReq, updateProfilesReq, edgexErr
			}
		} else if add || update {
			fullPath, redactedPath := GetFullAndRedactedURI(parsedURI, file, "Device Profile", lc)
			processedProfilesReq, updateProfilesReq, edgexErr = processProfiles(fullPath, redactedPath, overwrite, secretProvider, lc, dpc)
			if edgexErr != nil {
				return nil, nil, edgexErr
			}
			if len(processedProfilesReq) > 0 {
				addProfilesReq = append(addProfilesReq, processedProfilesReq...)
			}

			if len(processedUpdateProfilesReq) > 0 {
				updateProfilesReq = append(updateProfilesReq, processedUpdateProfilesReq...)
			}
		}
	}
	return addProfilesReq, updateProfilesReq, nil
}

func processProfiles(fullPath, displayPath string, overwrite bool, secretProvider bootstrapInterfaces.SecretProvider, lc logger.LoggingClient, dpc interfaces.DeviceProfileClient) ([]requests.DeviceProfileRequest, []requests.DeviceProfileRequest, errors.EdgeX) {
	var profile dtos.DeviceProfile
	var addProfilesReq []requests.DeviceProfileRequest
	var updateProfilesReq []requests.DeviceProfileRequest

	fileType := GetFileType(fullPath)

	// if the file type is not yaml or json, it cannot be parsed - just return to not break the loop for other devices
	if fileType == OTHER {
		return nil, nil, nil
	}

	content, err := file.Load(fullPath, secretProvider, lc)
	if err != nil {
		lc.Errorf("Failed to read Device Profile from %s: %v", displayPath, err)
		return nil, nil, nil
	}

	switch fileType {
	case YAML:
		err = yaml.Unmarshal(content, &profile)
		if err != nil {
			lc.Errorf("Failed to YAML decode Device Profile from %s: %v", displayPath, err)
			return nil, nil, nil
		}
	case JSON:
		err = json.Unmarshal(content, &profile)
		if err != nil {
			lc.Errorf("Failed to JSON decode Device Profile from %s: %v", displayPath, err)
			return nil, nil, nil
		}
	}

	add, update, edgexErr := checkDeviceProfile(profile.Name, overwrite, dpc, lc)
	if add {
		lc.Infof("Device Profile %s not found in Metadata, adding it ...", profile.Name)
		req := requests.NewDeviceProfileRequest(profile)
		addProfilesReq = append(addProfilesReq, req)
	} else if update {
		req := requests.NewDeviceProfileRequest(profile)
		updateProfilesReq = append(updateProfilesReq, req)
	} else {
		if edgexErr != nil {
			return addProfilesReq, updateProfilesReq, edgexErr
		}
	}
	return addProfilesReq, updateProfilesReq, nil
}

func checkDeviceProfile(name string, overwrite bool, dpc interfaces.DeviceProfileClient, lc logger.LoggingClient) (bool, bool, errors.EdgeX) {
	res, err := dpc.DeviceProfileByName(context.Background(), name)
	if err == nil {
		if overwrite {
			lc.Infof("Device Profile %s exists, overwriting with profiles from local files", name)
			return false, true, nil
		} else {
			lc.Infof("Device Profile %s exists, using the existing one", name)
			err = cache.Profiles().CheckAndAdd(dtos.ToDeviceProfileModel(res.Profile))
			if err != nil {
				return false, false, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to cache the profile %s", res.Profile.Name), err)
			}
			return false, false, nil
		}
	}
	return true, false, nil
}
