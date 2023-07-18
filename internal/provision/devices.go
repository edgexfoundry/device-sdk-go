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
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/file"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
)

func LoadDevices(path string, dic *di.Container) errors.EdgeX {
	var addDevicesReq []requests.AddDeviceRequest
	var edgexErr errors.EdgeX
	if path == "" {
		return nil
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	parsedUrl, err := url.Parse(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to parse devices path as a url", err)
	}
	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)
		addDevicesReq, edgexErr = loadDevicesFromUri(path, serviceName, secretProvider, lc)
		if edgexErr != nil {
			return edgexErr
		}
	} else {
		addDevicesReq, edgexErr = loadDevicesFromFile(path, serviceName, lc)
		if edgexErr != nil {
			return edgexErr
		}
	}

	if len(addDevicesReq) == 0 {
		return nil
	}
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) //nolint: staticcheck
	responses, edgexErr := dc.Add(ctx, addDevicesReq)
	if edgexErr != nil {
		return edgexErr
	}

	err = nil
	for _, response := range responses {
		if response.StatusCode != http.StatusCreated {
			if response.StatusCode == http.StatusConflict {
				lc.Warnf("%s. Device may be owned by other device service instance.", response.Message)
				continue
			}

			err = multierror.Append(err, fmt.Errorf("add device failed: %s", response.Message))
		}
	}

	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

func loadDevicesFromFile(path string, serviceName string, lc logger.LoggingClient) ([]requests.AddDeviceRequest, errors.EdgeX) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return []requests.AddDeviceRequest{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path for devices", err)
	}

	files, err := os.ReadDir(absPath)
	if err != nil {
		return []requests.AddDeviceRequest{}, errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory for devices", err)
	}

	if len(files) == 0 {
		return []requests.AddDeviceRequest{}, nil
	}

	lc.Infof("Loading pre-defined devices from %s(%d files found)", absPath, len(files))
	var addDevicesReq, processedDevicesReq []requests.AddDeviceRequest
	for _, file := range files {
		fullPath := filepath.Join(absPath, file.Name())
		processedDevicesReq = processDevices(fullPath, fullPath, serviceName, nil, lc)
		if len(processedDevicesReq) > 0 {
			addDevicesReq = append(addDevicesReq, processedDevicesReq...)
		}
	}
	return addDevicesReq, nil
}

func loadDevicesFromUri(inputUri string, serviceName string, secretProvider interfaces.SecretProvider, lc logger.LoggingClient) ([]requests.AddDeviceRequest, errors.EdgeX) {
	parsedUrl, err := url.Parse(inputUri)
	if err != nil {
		return []requests.AddDeviceRequest{}, errors.NewCommonEdgeX(errors.KindServerError, "could not parse URI for Devices", err)
	}

	bytes, err := file.Load(inputUri, secretProvider, lc)
	if err != nil {
		return []requests.AddDeviceRequest{}, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to read devices list from URI %s", parsedUrl.Redacted()), err)
	}

	if len(bytes) == 0 {
		return []requests.AddDeviceRequest{}, nil
	}

	var files []string

	err = json.Unmarshal(bytes, &files)
	if err != nil {
		return []requests.AddDeviceRequest{}, errors.NewCommonEdgeX(errors.KindServerError, "could not unmarshal Devices list contents", err)
	}

	if len(files) == 0 {
		return []requests.AddDeviceRequest{}, nil
	}

	baseUrl, _ := path.Split(inputUri)
	lc.Infof("Loading pre-defined devices from %s(%d files found)", parsedUrl.Redacted(), len(files))
	var addDevicesReq, processedDevicesReq []requests.AddDeviceRequest
	for _, file := range files {
		fullPath, err := url.JoinPath(baseUrl, file)
		if err != nil {
			lc.Error("could not join uri path for device %s: %v", file, err)
			continue
		}
		parsedFullPath, err := url.Parse(fullPath)
		if err != nil {
			lc.Error("could not join uri path for device %s: %v", file, err)
			continue
		}
		processedDevicesReq = processDevices(fullPath, parsedFullPath.Redacted(), serviceName, secretProvider, lc)
		if len(processedDevicesReq) > 0 {
			addDevicesReq = append(addDevicesReq, processedDevicesReq...)
		}
	}
	return addDevicesReq, nil
}

func processDevices(fullPath string, displayPath string, serviceName string, secretProvider interfaces.SecretProvider, lc logger.LoggingClient) []requests.AddDeviceRequest {
	var devices []dtos.Device
	var addDevicesReq []requests.AddDeviceRequest

	fileType := GetFileType(fullPath)

	// if the file type is not yaml or json, it cannot be parsed - just return to not break the loop for other devices
	if fileType == OTHER {
		return []requests.AddDeviceRequest{}
	}

	content, err := file.Load(fullPath, secretProvider, lc)
	if err != nil {
		lc.Errorf("Failed to read Devices from %s: %v", displayPath, err)
		return []requests.AddDeviceRequest{}
	}

	switch fileType {
	case YAML:
		d := struct {
			DeviceList []dtos.Device `yaml:"deviceList"`
		}{}
		err = yaml.Unmarshal(content, &d)
		if err != nil {
			lc.Errorf("Failed to YAML decode Devices from %s: %v", displayPath, err)
			return []requests.AddDeviceRequest{}
		}
		devices = d.DeviceList
	case JSON:
		err = json.Unmarshal(content, &devices)
		if err != nil {
			lc.Errorf("Failed to JSON decode Devices from %s: %v", displayPath, err)
			return []requests.AddDeviceRequest{}
		}
	}

	for _, device := range devices {
		if _, ok := cache.Devices().ForName(device.Name); ok {
			lc.Infof("Device %s exists, using the existing one", device.Name)
		} else {
			lc.Infof("Device %s not found in Metadata, adding it ...", device.Name)
			device.ServiceName = serviceName
			device.AdminState = models.Unlocked
			device.OperatingState = models.Up
			req := requests.NewAddDeviceRequest(device)
			addDevicesReq = append(addDevicesReq, req)
		}
	}
	return addDevicesReq
}
