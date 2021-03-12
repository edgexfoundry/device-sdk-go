// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"math"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

func Test_isNaN(t *testing.T) {
	validFloat32, err := models.NewCommandValue("test-resource", v2.ValueTypeFloat32, float32(1.234))
	require.NoError(t, err)
	validFloat64, err := models.NewCommandValue("test-resource", v2.ValueTypeFloat64, 1.234)
	require.NoError(t, err)
	float32NaN, err := models.NewCommandValue("test-resource", v2.ValueTypeFloat32, float32(math.NaN()))
	require.NoError(t, err)
	float64NaN, err := models.NewCommandValue("test-resource", v2.ValueTypeFloat64, math.NaN())
	require.NoError(t, err)

	tests := []struct {
		name     string
		cv       *models.CommandValue
		expected bool
	}{
		{"valid float32 value", validFloat32, false},
		{"valid float64 value", validFloat64, false},
		{"float32 NaN error", float32NaN, true},
		{"float64 NaN error", float64NaN, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNaN, err := isNaN(tt.cv)
			assert.Equal(t, tt.expected, isNaN)
			assert.NoError(t, err)
		})
	}
}
