// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"errors"
	"math"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestTransformReadResult_NaN(t *testing.T) {
	ro := models.ResourceOperation{DeviceResource: "test-object"}
	float32Val, _ := dsModels.NewFloat32Value(ro.DeviceResource, 0, float32(math.NaN()))
	float64Val, _ := dsModels.NewFloat64Value(ro.DeviceResource, 0, math.NaN())

	tests := []struct {
		name string
		cv   *dsModels.CommandValue
	}{
		{"float32 NaN error", float32Val},
		{"float64 NaN error", float64Val},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pv := models.PropertyValue{}
			err := TransformReadResult(tt.cv, pv, lc)
			assert.True(t, errors.Is(err, NaNError{}), "transform result should be NaNError")
		})
	}
}
