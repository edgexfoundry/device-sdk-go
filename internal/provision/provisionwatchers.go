//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
)

func LoadProvisionWatchers(path string, dic *di.Container) errors.EdgeX {
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

	lc := container.LoggingClientFrom(dic.Get)
	lc.Infof("Loading pre-defined provision watchers from %s(%d files found)", absPath, len(files))

	var addProvisionWatchersReq []requests.AddProvisionWatcherRequest
	for _, file := range files {
		var watcher dtos.ProvisionWatcher
		filename := filepath.Join(absPath, file.Name())
		data, err := os.ReadFile(filename)
		if err != nil {
			lc.Errorf("Failed to read %s: %v", filename, err)
			continue
		}

		if strings.HasSuffix(filename, yamlExt) || strings.HasSuffix(filename, ymlExt) {
			err = yaml.Unmarshal(data, &watcher)
			if err != nil {
				lc.Errorf("Failed to YAML decode %s: %v", filename, err)
				continue
			}
		} else if strings.HasSuffix(filename, jsonExt) {
			err := json.Unmarshal(data, &watcher)
			if err != nil {
				lc.Errorf("Failed to JSON decode %s: %v", filename, err)
				continue
			}
		} else {
			continue
		}

		err = common.Validate(watcher)
		if err != nil {
			lc.Errorf("ProvisionWatcher %s validation failed: %v", watcher.Name, err)
			continue
		}

		if _, ok := cache.ProvisionWatchers().ForName(watcher.Name); ok {
			lc.Infof("ProvisionWatcher %s exists, using the existing one", watcher.Name)
		} else {
			lc.Infof("ProvisionWatcher %s not found in Metadata, adding it...", watcher.Name)
			req := requests.NewAddProvisionWatcherRequest(watcher)
			addProvisionWatchersReq = append(addProvisionWatchersReq, req)
		}
	}

	if len(addProvisionWatchersReq) == 0 {
		return nil
	}

	pwc := container.ProvisionWatcherClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) //nolint: staticcheck
	_, edgexErr := pwc.Add(ctx, addProvisionWatchersReq)
	return edgexErr
}
