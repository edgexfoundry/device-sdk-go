//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateWriteMaximum(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		max         float64
		expectedErr bool
	}{
		{"valid - uint8 compare succeed", uint8(math.MaxUint8 - 1), float64(math.MaxUint8), false},
		{"invalid - uint8 compare failed", uint8(math.MaxUint8), float64(math.MaxUint8 - 1), true},
		{"valid - uint16 compare succeed", uint16(math.MaxUint16 - 1), float64(math.MaxUint16), false},
		{"invalid - uint16 compare failed", uint16(math.MaxUint16), float64(math.MaxUint16 - 1), true},
		{"valid - uint32 compare succeed", uint32(math.MaxUint32 - 1), float64(math.MaxUint32), false},
		{"invalid - uint32 compare failed", uint32(math.MaxUint32), float64(math.MaxUint32 - 1), true},
		{"valid - uint64 compare succeed", uint64(1000000000), float64(1000000001), false},
		{"invalid - uint64 compare failed", uint64(1000000001), float64(1000000000), true},
		{"valid - int8 compare succeed", int8(math.MaxInt8 - 1), float64(math.MaxInt8), false},
		{"invalid - int8 compare failed", int8(math.MaxInt8), float64(math.MaxInt8 - 1), true},
		{"valid - int16 compare succeed", int16(math.MaxInt16 - 1), float64(math.MaxInt16), false},
		{"invalid - int16 compare failed", int16(math.MaxInt16), float64(math.MaxInt16 - 1), true},
		{"valid - int32 compare succeed", int32(math.MaxInt32 - 1), float64(math.MaxInt32), false},
		{"invalid - int32 compare failed", int32(math.MaxInt32), float64(math.MaxInt32 - 1), true},
		{"valid - int64 compare succeed", int64(1000000000), float64(1000000001), false},
		{"invalid - int64 compare failed", int64(1000000001), float64(1000000000), true},
		{"valid - float32 compare succeed", float32(12345.6789), float64(123456.789), false},
		{"invalid - float32 compare failed", float32(123456.789), float64(12345.6789), true},
		{"valid - float64 compare succeed", float64(12345.6789), float64(123456.789), false},
		{"invalid - float64 compare failed", float64(123456.789), float64(12345.6789), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWriteMaximum(tt.value, tt.max)
			if !tt.expectedErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func Test_validateWriteMinimum(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		min         float64
		expectedErr bool
	}{
		{"valid - uint8 compare succeed", uint8(1), float64(0), false},
		{"invalid - uint8 compare failed", uint8(0), float64(1), true},
		{"valid - uint16 compare succeed", uint16(1), float64(0), false},
		{"invalid - uint16 compare failed", uint16(0), float64(1), true},
		{"valid - uint32 compare succeed", uint32(1), float64(0), false},
		{"invalid - uint32 compare failed", uint32(0), float64(1), true},
		{"valid - uint64 compare succeed", uint64(1), float64(0), false},
		{"invalid - uint64 compare failed", uint64(0), float64(1), true},
		{"valid - int8 compare succeed", int8(math.MinInt8 + 1), float64(math.MinInt8), false},
		{"invalid - int8 compare failed", int8(math.MinInt8), float64(math.MinInt8 + 1), true},
		{"valid - int16 compare succeed", int16(math.MinInt16 + 1), float64(math.MinInt16), false},
		{"invalid - int16 compare failed", int16(math.MinInt16), float64(math.MinInt16 + 1), true},
		{"valid - int32 compare succeed", int32(math.MinInt32 + 1), float64(math.MinInt32), false},
		{"invalid - int32 compare failed", int32(math.MinInt32), float64(math.MinInt32 + 1), true},
		{"valid - int64 compare succeed", int64(math.MinInt64 + 1), float64(math.MinInt64), false},
		{"invalid - int64 compare failed", int64(-1000000001), float64(-1000000000), true},
		{"valid - float32 compare succeed", float32(123456.789), float64(12345.6789), false},
		{"invalid - float32 compare failed", float32(12345.6789), float64(123456.789), true},
		{"valid - float64 compare succeed", float64(123456.789), float64(12345.6789), false},
		{"invalid - float64 compare failed", float64(12345.6789), float64(123456.789), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWriteMinimum(tt.value, tt.min)
			if !tt.expectedErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
