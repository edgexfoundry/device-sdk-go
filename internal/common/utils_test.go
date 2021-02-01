// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"testing"
)

func TestGetUniqueOrigin(t *testing.T) {
	origin1 := GetUniqueOrigin()
	origin2 := GetUniqueOrigin()

	if origin1 >= origin2 {
		t.Errorf("origin1: %d should <= origin2: %d", origin1, origin2)
	}
}
