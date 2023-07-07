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
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v3"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
)

func LoadDevices(path string, dic *di.Container) errors.EdgeX {
	var addDevicesReq []requests.AddDeviceRequest

	if path == "" {
		return nil
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	parsedUrl, err := url.Parse(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to parse url", err)
	}
	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)
		edgexErr := loadDevicesFromUri(lc, path, serviceName, secretProvider, addDevicesReq)
		if edgexErr != nil {
			return edgexErr
		}
	} else {
		edgexErr := loadDevicesFromFile(lc, path, serviceName, addDevicesReq)
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

func loadDevicesFromFile(lc logger.LoggingClient, path string, serviceName string, addDevicesReq []requests.AddDeviceRequest) errors.EdgeX {
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

	lc.Infof("Loading pre-defined devices from %s(%d files found)", absPath, len(files))

	for _, file := range files {
		fullPath := filepath.Join(absPath, file.Name())
		addDevicesReq = processDevices(lc, fullPath, serviceName, nil, addDevicesReq)
	}
	return nil
}

func loadDevicesFromUri(lc logger.LoggingClient, inputUri string, serviceName string, secretProvider interfaces.SecretProvider, addDevicesReq []requests.AddDeviceRequest) errors.EdgeX {
	parsedUrl, err := url.Parse(inputUri)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "could not parse uri for Provision Watcher", err)
	}

	bytes, err := file.Load(inputUri, timeout, secretProvider)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to read uri %s", parsedUrl.Redacted()), err)
	}

	if len(bytes) == 0 {
		return nil
	}

	var files []string

	err = json.Unmarshal(bytes, &files)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "could not unmarshal Device contents", err)
	}

	if len(files) == 0 {
		return nil
	}

	baseUrl, _ := path.Split(inputUri)
	lc.Infof("Loading pre-defined devices from %s(%d files found)", parsedUrl.Redacted(), len(files))

	for _, file := range files {
		fullPath, err := url.JoinPath(baseUrl, file)
		if err != nil {
			lc.Error("could not join uri path for device %s: %v", file, err)
			continue
		}
		addDevicesReq = processDevices(lc, fullPath, serviceName, nil, addDevicesReq)
	}
	return nil
}

func processDevices(lc logger.LoggingClient, fullPath string, serviceName string, secretProvider interfaces.SecretProvider, addDevicesReq []requests.AddDeviceRequest) []requests.AddDeviceRequest {
	var devices []dtos.Device
	content, err := file.Load(fullPath, timeout, secretProvider)
	if err != nil {
		lc.Errorf("Failed to read %s: %v", fullPath, err)
		return addDevicesReq
	}
	if strings.HasSuffix(fullPath, yamlExt) || strings.HasSuffix(fullPath, ymlExt) {
		d := struct {
			DeviceList []dtos.Device `yaml:"deviceList"`
		}{}
		err = yaml.Unmarshal(content, &d)
		if err != nil {
			lc.Errorf("Failed to YAML decode %s: %v", fullPath, err)
			return addDevicesReq
		}
		devices = d.DeviceList
	} else if strings.HasSuffix(fullPath, ".json") {
		err = json.Unmarshal(content, &devices)
		if err != nil {
			lc.Errorf("Failed to JSON decode %s: %v", fullPath, err)
			return addDevicesReq
		}
	} else {
		return addDevicesReq
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
