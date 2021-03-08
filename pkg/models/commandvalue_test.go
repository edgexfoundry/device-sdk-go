// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"math/rand"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewCommandValue(t *testing.T) {
	exceedBinary := make([]byte, MaxBinaryBytes+1)
	rand.Read(exceedBinary)
	tests := []struct {
		name        string
		valueType   string
		value       interface{}
		expectedErr bool
	}{
		{"invalid - value doesn't match with valueType", v2.ValueTypeInt64, "invalid", true},
		{"invalid - array type with incorrect elements", v2.ValueTypeUint64Array, []int8{-1, -2, -3}, true},
		{"invalid - binary payload exceeds MaxBinaryBytes size", v2.ValueTypeBinary, exceedBinary, true},
		{"valid - binary payload doesn't exceed MaxBinaryBytes size", v2.ValueTypeBinary, []byte{1, 2, 3}, false},
		{"valid - normal type", v2.ValueTypeInt8, int8(8), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCommandValue("test-resource", tt.valueType, tt.value)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
