// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"math"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
)

func checkTransformedValueInRange(origin interface{}, transformed float64) bool {
	inRange := false
	switch origin.(type) {
	case uint8:
		if transformed >= 0 && transformed <= math.MaxUint8 {
			inRange = true
		}
	case uint16:
		if transformed >= 0 && transformed <= math.MaxUint16 {
			inRange = true
		}
	case uint32:
		if transformed >= 0 && transformed <= math.MaxUint32 {
			inRange = true
		}
	case uint64:
		// if the variable isn't casted to float64, this statement will cause error on 32bit system
		maxiMum := float64(math.MaxUint64)
		if transformed >= 0 && transformed <= maxiMum {
			inRange = true
		}
	case int8:
		if transformed >= math.MinInt8 && transformed <= math.MaxInt8 {
			inRange = true
		}
	case int16:
		if transformed >= math.MinInt16 && transformed <= math.MaxInt16 {
			inRange = true
		}
	case int32:
		if transformed >= math.MinInt32 && transformed <= math.MaxInt32 {
			inRange = true
		}
	case int64:
		if transformed >= math.MinInt64 && transformed <= math.MaxInt64 {
			inRange = true
		}
	case float32:
		if math.Abs(transformed) >= math.SmallestNonzeroFloat32 && math.Abs(transformed) <= math.MaxFloat32 {
			inRange = true
		}
	case float64:
		if math.Abs(transformed) >= math.SmallestNonzeroFloat64 && math.Abs(transformed) <= math.MaxFloat64 {
			inRange = true
		}
	default:
		common.LoggingClient.Error(fmt.Sprintf("data type %T doesn't support range checking", origin))
	}

	return inRange
}
