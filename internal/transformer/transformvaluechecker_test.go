// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"math"
	"testing"
)

func TestCheckTransformedValueInRange_numericDataType(t *testing.T) {
	var tests = []struct {
		origin      interface{}
		transformed float64
		inRange     bool
	}{
		{origin: uint8(10), transformed: float64(math.MaxUint8), inRange: true},
		{origin: uint8(10), transformed: float64(math.MaxUint8 + 1), inRange: false},
		{origin: uint16(10), transformed: float64(math.MaxUint16), inRange: true},
		{origin: uint16(10), transformed: float64(math.MaxUint16 + 1), inRange: false},
		{origin: uint32(10), transformed: float64(math.MaxUint32), inRange: true},
		{origin: uint32(10), transformed: float64(math.MaxUint32 + 1), inRange: false},
		{origin: uint64(10), transformed: float64(math.MaxUint64), inRange: true},
		{origin: uint64(10), transformed: float64(math.MaxUint64 * 2), inRange: false},
		{origin: int8(10), transformed: float64(math.MaxInt8), inRange: true},
		{origin: int8(10), transformed: float64(math.MaxInt8 + 1), inRange: false},
		{origin: int16(10), transformed: float64(math.MaxInt16), inRange: true},
		{origin: int16(10), transformed: float64(math.MaxInt16 + 1), inRange: false},
		{origin: int32(10), transformed: float64(math.MaxInt32), inRange: true},
		{origin: int32(10), transformed: float64(math.MaxInt32 + 1), inRange: false},
		{origin: int64(10), transformed: float64(math.MaxInt64), inRange: true},
		{origin: int64(10), transformed: float64(math.MaxUint64), inRange: false},
		{origin: float32(10), transformed: float64(math.MaxFloat32), inRange: true},
		{origin: float32(10), transformed: float64(math.MaxFloat32 * 2), inRange: false},
		{origin: float64(10), transformed: float64(math.MaxFloat64), inRange: true},
	}

	for _, tc := range tests {
		if checkTransformedValueInRange(tc.origin, tc.transformed, lc) != tc.inRange {
			t.Fatalf("Transformed value '%v' is not within the '%T' value type range", tc.transformed, tc.origin)
		}
	}
}

func TestCheckTransformedValueInRange_unsupportedDataType(t *testing.T) {
	origin := "123"
	transformed := float64(123)

	inRange := checkTransformedValueInRange(origin, transformed, lc)

	if inRange == true {
		t.Fatalf("Unexpected test result. Data type %T should not support range checking", origin)
	}
}
