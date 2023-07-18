//
// Copyright (C) 2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/file"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
)

func LoadProvisionWatchers(path string, dic *di.Container) errors.EdgeX {
	var addProvisionWatchersReq []requests.AddProvisionWatcherRequest
	if path == "" {
		return nil
	}

	lc := container.LoggingClientFrom(dic.Get)

	parsedUrl, err := url.Parse(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to parse url", err)
	}
	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		secretProvider := container.SecretProviderFrom(dic.Get)
		edgexErr := loadProvisionWatchersFromUri(path, secretProvider, lc, addProvisionWatchersReq)
		if edgexErr != nil {
			return edgexErr
		}
	} else {
		edgexErr := loadProvisionWatchersFromFile(path, lc, addProvisionWatchersReq)
		if edgexErr != nil {
			return edgexErr
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

func loadProvisionWatchersFromFile(path string, lc logger.LoggingClient, addProvisionWatchersReq []requests.AddProvisionWatcherRequest) errors.EdgeX {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path", err)
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory", err)
	}

	if len(files) == 0 {
		return nil
	}

	lc.Infof("Loading pre-defined provision watchers from %s(%d files found)", absPath, len(files))

	for _, file := range files {
		filename := filepath.Join(absPath, file.Name())
		addProvisionWatchersReq = processProvisonWatcherFile(filename, nil, lc, addProvisionWatchersReq)
	}
	return nil
}

func loadProvisionWatchersFromUri(inputUri string, secretProvider interfaces.SecretProvider, lc logger.LoggingClient, addProvisionWatchersReq []requests.AddProvisionWatcherRequest) errors.EdgeX {
	parsedUrl, err := url.Parse(inputUri)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "could not parse uri for Provision Watcher", err)
	}

	bytes, err := file.Load(inputUri, secretProvider, lc)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to read uri %s", parsedUrl.Redacted()), err)
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
	lc.Infof("Loading pre-defined provision watchers from %s(%d files found)", parsedUrl.Redacted(), len(files))

	for _, file := range files {
		provisionWatcherUrl, err := url.JoinPath(baseUrl, file)
		if err != nil {
			lc.Error("could not join uri path for provision watcher %s: %v", file, err)
			continue
		}
		addProvisionWatchersReq = processProvisonWatcherFile(provisionWatcherUrl, secretProvider, lc, addProvisionWatchersReq)
	}
	return nil
}

func processProvisonWatcherFile(path string, secretProvider interfaces.SecretProvider, lc logger.LoggingClient, addProvisionWatchersReq []requests.AddProvisionWatcherRequest) []requests.AddProvisionWatcherRequest {
	var watcher dtos.ProvisionWatcher
	data, err := file.Load(path, secretProvider, lc)
	if err != nil {
		lc.Errorf("Failed to read Provision Watcher: %v", err)
		return addProvisionWatchersReq
	}

	if strings.HasSuffix(path, yamlExt) || strings.HasSuffix(path, ymlExt) {
		err = yaml.Unmarshal(data, &watcher)
		if err != nil {
			lc.Errorf("Failed to YAML decode Provision Watcher: %v", err)
			return addProvisionWatchersReq
		}
	} else if strings.HasSuffix(path, jsonExt) {
		err := json.Unmarshal(data, &watcher)
		if err != nil {
			lc.Errorf("Failed to JSON decode Provision Watcher: %v", err)
			return addProvisionWatchersReq
		}
	} else {
		return addProvisionWatchersReq
	}

	err = common.Validate(watcher)
	if err != nil {
		lc.Errorf("ProvisionWatcher %s validation failed: %v", watcher.Name, err)
		return addProvisionWatchersReq
	}

	if _, ok := cache.ProvisionWatchers().ForName(watcher.Name); ok {
		lc.Infof("ProvisionWatcher %s exists, using the existing one", watcher.Name)
	} else {
		lc.Infof("ProvisionWatcher %s not found in Metadata, adding it...", watcher.Name)
		req := requests.NewAddProvisionWatcherRequest(watcher)
		addProvisionWatchersReq = append(addProvisionWatchersReq, req)
	}
	return addProvisionWatchersReq
}
