// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
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
