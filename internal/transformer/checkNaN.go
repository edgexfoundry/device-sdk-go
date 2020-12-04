// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"math"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
)

// NaNError is used to throw the NaN error for the floating-point value
type NaNError struct{}

func (e NaNError) Error() string {
	return "not a valid float value NaN"
}

func isNaN(cv *dsModels.CommandValue) (bool, error) {
	switch cv.Type {
	case dsModels.Float32:
		v, err := cv.Float32Value()
		if err != nil {
			return false, err
		}
		if math.IsNaN(float64(v)) {
			return true, nil
		}
	case dsModels.Float64:
		v, err := cv.Float64Value()
		if err != nil {
			return false, err
		}
		if math.IsNaN(v) {
			return true, nil
		}
	}
	return false, nil
}
