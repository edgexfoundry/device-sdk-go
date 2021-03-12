// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkTransformedValueInRange(t *testing.T) {
	var tests = []struct {
		name        string
		origin      interface{}
		transformed float64
		expected    bool
	}{
		{"valid - uint8 in range", uint8(10), float64(math.MaxUint8), true},
		{"invalid - uint8 out of range", uint8(10), float64(math.MaxUint8 + 1), false},
		{"invalid - uint8 transformed to float", uint8(10), float64(123.456), false},
		{"invalid - uint8 transformed to signed number", uint8(10), float64(-123), false},
		{"valid - uint16 in range", uint16(10), float64(math.MaxUint16), true},
		{"invalid - uint16 out of range", uint16(10), float64(math.MaxUint16 + 1), false},
		{"invalid - uint16 transformed to float", uint16(10), float64(123.456), false},
		{"invalid - uint16 transformed to signed number", uint16(10), float64(-123), false},
		{"valid - uint32 in range", uint32(10), float64(math.MaxUint32), true},
		{"invalid - uint32 out of range", uint32(10), float64(math.MaxUint32 + 1), false},
		{"invalid - uint32 transformed to float", uint32(10), float64(123.456), false},
		{"invalid - uint32 transformed to signed number", uint32(10), float64(-123), false},
		{"valid - uint64 in range", uint64(10), float64(math.MaxUint64), true},
		{"invalid - uint64 out of range", uint64(10), float64(math.MaxUint64 * 2), false},
		{"invalid - uint64 transformed to float", uint64(10), float64(123.456), false},
		{"invalid - uint64 transformed to signed number", uint64(10), float64(-123), false},
		{"valid - int8 in range", int8(10), float64(math.MaxInt8), true},
		{"invalid - int8 out of range", int8(10), float64(math.MaxInt8 + 1), false},
		{"invalid - int8 transformed to float", int8(10), float64(-123.456), false},
		{"valid - int16 in range", int16(10), float64(math.MaxInt16), true},
		{"invalid - int16 out of range", int16(10), float64(math.MaxInt16 + 1), false},
		{"invalid - int16 transformed to float", int16(10), float64(-123.456), false},
		{"valid - int32 in range", int32(10), float64(math.MaxInt32), true},
		{"invalid - int32 out of range", int32(10), float64(math.MaxInt32 + 1), false},
		{"invalid - int32 transformed to float", int32(10), float64(-123.456), false},
		{"valid - int64 in range", int64(10), float64(math.MaxInt64), true},
		{"invalid - int64 out of range", int64(10), float64(math.MaxUint64), false},
		{"invalid - int64 transformed to float", int64(10), float64(-123.456), false},
		{"valid - float32 in range", float32(10), float64(math.MaxFloat32), true},
		{"invalid - float32 out of range", float32(10), float64(math.MaxFloat32 * 2), false},
		{"valid - float64 in range", float64(10), float64(math.MaxFloat64), true},
		{"invalid - unsupported origin value type", "invalid", float64(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := checkTransformedValueInRange(tt.origin, tt.transformed)
			assert.Equal(t, tt.expected, res)
		})
	}
}
