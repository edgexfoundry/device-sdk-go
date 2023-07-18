// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"net/url"
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

func GetFullAndParsedURI(baseUrl, file, description string, lc logger.LoggingClient) (string, *url.URL) {
	fullPath, err := url.JoinPath(baseUrl, file)
	if err != nil {
		lc.Error("could not join URI path for %s %s: %v", description, file, err)
		return "", nil
	}
	parsedFullPath, err := url.Parse(fullPath)
	if err != nil {
		lc.Error("could not parse URI path for %s %s: %v", description, file, err)
		return "", nil
	}
	return fullPath, parsedFullPath
}
