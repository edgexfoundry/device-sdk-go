// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"net/url"
	"path"
	"strings"
)

type FileType int

const (
	YAML FileType = iota
	JSON
	OTHER
)

func GetFileType(fullPath string) FileType {
	if strings.HasSuffix(fullPath, yamlExt) || strings.HasSuffix(fullPath, ymlExt) {
		return YAML
	} else if strings.HasSuffix(fullPath, ".json") {
		return JSON
	} else {
		return OTHER
	}
}

func GetFullAndRedactedURI(baseURI *url.URL, file, description string, lc logger.LoggingClient) (string, string) {
	basePath, _ := path.Split(baseURI.Path)
	newPath, err := url.JoinPath(basePath, file)
	if err != nil {
		lc.Error("could not join URI path for %s %s/%s: %v", description, basePath, file, err)
		return "", ""
	}
	var fullURI url.URL
	err = utils.DeepCopy(baseURI, &fullURI)
	if err != nil {
		lc.Error("could not copy URI for %s %s: %v", description, newPath, err)
		return "", ""
	}
	fullURI.User = baseURI.User
	fullURI.Path = newPath
	return fullURI.String(), fullURI.Redacted()
}
