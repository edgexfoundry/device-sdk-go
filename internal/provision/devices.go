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
	"net/http"
	"os"
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
	lc.Infof("Loading pre-defined devices from %s(%d files found)", absPath, len(files))

	var addDevicesReq []requests.AddDeviceRequest
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	for _, file := range files {
		var devices []dtos.Device
		fullPath := filepath.Join(absPath, file.Name())
		if strings.HasSuffix(fullPath, yamlExt) || strings.HasSuffix(fullPath, ymlExt) {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}
			d := struct {
				DeviceList []dtos.Device `yaml:"deviceList"`
			}{}
			err = yaml.Unmarshal(content, &d)
			if err != nil {
				lc.Errorf("Failed to YAML decode %s: %v", fullPath, err)
				continue
			}
			devices = d.DeviceList
		} else if strings.HasSuffix(fullPath, ".json") {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}
			err = json.Unmarshal(content, &devices)
			if err != nil {
				lc.Errorf("Failed to JSON decode %s: %v", fullPath, err)
				continue
			}
		} else {
			continue
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
