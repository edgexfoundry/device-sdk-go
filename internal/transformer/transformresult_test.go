// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"math"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

func Test_transformReadBase(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		base        float64
		expected    interface{}
		expectedErr bool
	}{
		{"valid - uint8 base transformation", uint8(2), 10, uint8(100), false},
		{"invalid - uint8 base transformation overflow", uint8(10), 10, nil, true},
		{"valid - uint16 base transformation", uint16(2), 100, uint16(10000), false},
		{"invalid - uint16 base transformation overflow", uint16(100), 10, nil, true},
		{"valid - uint32 base transformation", uint32(2), 10000, uint32(100000000), false},
		{"invalid - uint32 base transformation overflow", uint32(10000), 10, nil, true},
		{"valid - uint64 base transformation", uint64(2), 100000, uint64(10000000000), false},
		{"invalid - uint64 base transformation overflow", uint64(100000000), 10, nil, true},
		{"valid - int8 base transformation", int8(2), 10, int8(100), false},
		{"invalid - int8 base transformation overflow", int8(10), 10, nil, true},
		{"valid - int16 base transformation", int16(2), 100, int16(10000), false},
		{"invalid - int16 base transformation overflow", int16(100), 10, nil, true},
		{"valid - int32 base transformation", int32(2), 10000, int32(100000000), false},
		{"invalid - int32 base transformation overflow", int32(10000), 10, nil, true},
		{"valid - int64 base transformation", int64(2), 100000, int64(10000000000), false},
		{"invalid - int64 base transformation overflow", int64(100000000), 10, nil, true},
		{"valid - float32 base transformation", float32(1.1), 2, float32(2.143547), false},
		{"invalid - float32 base transformation overflow", math.MaxFloat32, 2, nil, true},
		{"valid - float64 base transformation", float64(1.1), 2, float64(2.1435469250725863), false},
		{"invalid - float64 base transformation overflow", math.MaxFloat64, 2, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := transformBase(tt.value, tt.base, true)
			if !tt.expectedErr {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			} else {
				require.Error(t, err)
				assert.Equal(t, errors.Kind(err), errors.KindOverflowError, "expect Overflow ErrKind")
			}
		})
	}
}

func Test_transformReadScale(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		scale       float64
		expected    interface{}
		expectedErr bool
	}{
		{"valid - uint8 scale transformation", uint8(math.MaxUint8 / 5), 5, uint8(math.MaxUint8), false},
		{"invalid - uint8 scale transformation overflow", uint8(math.MaxUint8 / 5), 6, nil, true},
		{"valid - uint16 scale transformation", uint16(math.MaxUint16 / 5), 5, uint16(math.MaxUint16), false},
		{"invalid - uint16 scale transformation overflow", uint16(math.MaxUint16 / 5), 6, nil, true},
		{"valid - uint32 scale transformation", uint32(math.MaxUint32 / 5), 5, uint32(math.MaxUint32), false},
		{"invalid - uint32 scale transformation overflow", uint32(math.MaxUint32 / 5), 6, nil, true},
		{"valid - uint64 scale transformation", uint64(math.MaxUint64 / 5), 5, uint64(math.MaxUint64), false},
		{"valid - int8 scale transformation", int8(10), 10, int8(100), false},
		{"invalid - int8 scale transformation overflow", int8(10), 30, nil, true},
		{"valid - int16 scale transformation", int16(10000), 3, int16(30000), false},
		{"invalid - int16 scale transformation overflow", int16(10000), 4, nil, true},
		{"valid - int32 scale transformation", int32(1000000000), 2, int32(2000000000), false},
		{"invalid - int32 scale transformation overflow", int32(1000000000), 3, nil, true},
		{"valid - int64 scale transformation", int64(1000000000), 1000000000, int64(1000000000000000000), false},
		{"valid - float32 scale transformation", float32(12.1), 10, float32(121), false},
		{"invalid - float32 scale transformation overflow", float32(math.MaxFloat32 / 2), 3, nil, true},
		{"valid - float64 scale transformation", float64(111111111.1), 2, float64(222222222.2), false},
		{"invalid - float64 scale transformation overflow", float64(math.MaxFloat64), 2, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := transformScale(tt.value, tt.scale, true)
			if !tt.expectedErr {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			} else {
				require.Error(t, err)
				assert.Equal(t, errors.Kind(err), errors.KindOverflowError, "expect Overflow ErrKind")
			}
		})
	}
}

func Test_transformReadOffset(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		offset      float64
		expected    interface{}
		expectedErr bool
	}{
		{"valid - uint8 offset transformation", uint8(math.MaxUint8 - 1), 1, uint8(math.MaxUint8), false},
		{"invalid - uint8 offset transformation overflow", uint8(math.MaxUint8), 1, nil, true},
		{"valid - uint16 offset transformation", uint16(math.MaxUint16 - 1), 1, uint16(math.MaxUint16), false},
		{"invalid - uint16 offset transformation overflow", uint16(math.MaxUint16), 1, nil, true},
		{"valid - uint32 offset transformation", uint32(math.MaxUint32 - 1), 1, uint32(math.MaxUint32), false},
		{"invalid - uint32 offset transformation overflow", uint32(math.MaxUint32), 1, nil, true},
		{"valid - uint64 offset transformation", uint64(math.MaxUint64) - uint64(1), 1, uint64(math.MaxUint64), false},
		{"valid - int8 offset transformation", int8(math.MaxInt8 - 1), 1, int8(math.MaxInt8), false},
		{"invalid - int8 offset transformation overflow", int8(math.MaxInt8), 1, nil, true},
		{"valid - int16 offset transformation", int16(math.MaxInt16 - 1), 1, int16(math.MaxInt16), false},
		{"invalid - int16 offset transformation overflow", int16(math.MaxInt16), 1, nil, true},
		{"valid - int32 offset transformation", int32(math.MaxInt32 - 1), 1, int32(math.MaxInt32), false},
		{"invalid - int32 offset transformation overflow", int32(math.MaxInt32), 1, nil, true},
		{"valid - int64 offset transformation", int64(math.MaxInt64 - 1), 1, int64(math.MaxInt64), false},
		{"valid - float32 offset transformation", float32(1.1), 1, float32(2.1), false},
		{"invalid - float32 offset transformation overflow", float32(math.MaxFloat32), math.MaxFloat32, nil, true},
		{"valid - float64 offset transformation", float64(1.1), 1, float64(2.1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := transformOffset(tt.value, tt.offset, true)
			if !tt.expectedErr {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			} else {
				require.Error(t, err)
				assert.Equal(t, errors.Kind(err), errors.KindOverflowError, "expect Overflow ErrKind")
			}
		})
	}
}

func Test_transformMask(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		mask     uint64
		expected interface{}
	}{
		{"valid - uint8 mask transformation", uint8(math.MaxUint8), 15, uint8(15)},
		{"valid - uint16 mask transformation", uint16(math.MaxUint16), 256, uint16(256)},
		{"valid - uint32 mask transformation", uint32(math.MaxUint32), 256, uint32(256)},
		{"valid - uint64 mask transformation", uint64(math.MaxUint64), 256, uint64(256)},
		{"valid - int8 mask transformation", int8(math.MaxInt8), 15, int8(15)},
		{"valid - int16 mask transformation", int16(math.MaxInt16), 127, int16(127)},
		{"valid - int32 mask transformation", int32(math.MaxInt32), 32767, int32(32767)},
		{"valid - int64 mask transformation", int64(math.MaxInt64), 2147483647, int64(2147483647)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := transformMask(tt.value, tt.mask)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func Test_transformShift(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		shift    int64
		expected interface{}
	}{
		{"valid - uint8 shift transformation with negative shift value", uint8(0b1), -4, uint8(0b10000)},
		{"valid - uint8 shift transformation with positive shift value", uint8(0b11111111), 4, uint8(0b00001111)},
		{"valid - uint16 shift transformation with negative shift value", uint16(0b1), -8, uint16(0b100000000)},
		{"valid - uint16 shift transformation with positive shift value", uint16(0b1111111100000000), 8, uint16(0b0000000011111111)},
		{"valid - uint32 shift transformation with negative shift value", uint32(0b1), -16, uint32(0b10000000000000000)},
		{"valid - uint32 shift transformation with positive shift value", uint32(0b11111111111111110000000000000000), 16, uint32(0b00000000000000001111111111111111)},
		{"valid - uint64 shift transformation with negative shift value", uint64(0b1), -32, uint64(0b100000000000000000000000000000000)},
		{"valid - uint64 shift transformation with positive shift value", uint64(0b1111111111111111111111111111111100000000000000000000000000000000), 32, uint64(0b0000000000000000000000000000000011111111111111111111111111111111)},
		{"valid - int8 shift transformation with negative shift value", int8(0b1), -4, int8(0b10000)},
		{"valid - int8 shift transformation with positive shift value", int8(0b1111111), 4, int8(0b0000111)},
		{"valid - int16 shift transformation with negative shift value", int16(0b1), -8, int16(0b100000000)},
		{"valid - int16 shift transformation with positive shift value", int16(0b111111100000000), 8, int16(0b000000001111111)},
		{"valid - int32 shift transformation with negative shift value", int32(0b1), -16, int32(0b10000000000000000)},
		{"valid - int32 shift transformation with positive shift value", int32(0b1111111111111110000000000000000), 16, int32(0b0000000000000000111111111111111)},
		{"valid - int64 shift transformation with negative shift value", int64(0b1), -32, int64(0b100000000000000000000000000000000)},
		{"valid - int64 shift transformation with positive shift value", int64(0b111111111111111111111111111111100000000000000000000000000000000), 32, int64(0b000000000000000000000000000000001111111111111111111111111111111)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := transformShift(tt.value, tt.shift)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func Test_commandValueForTransform(t *testing.T) {
	u8, err := models.NewCommandValue("test-resource", common.ValueTypeUint8, uint8(0))
	require.NoError(t, err)
	u16, err := models.NewCommandValue("test-resource", common.ValueTypeUint16, uint16(0))
	require.NoError(t, err)
	u32, err := models.NewCommandValue("test-resource", common.ValueTypeUint32, uint32(0))
	require.NoError(t, err)
	u64, err := models.NewCommandValue("test-resource", common.ValueTypeUint64, uint64(0))
	require.NoError(t, err)
	i8, err := models.NewCommandValue("test-resource", common.ValueTypeInt8, int8(0))
	require.NoError(t, err)
	i16, err := models.NewCommandValue("test-resource", common.ValueTypeInt16, int16(0))
	require.NoError(t, err)
	i32, err := models.NewCommandValue("test-resource", common.ValueTypeInt32, int32(0))
	require.NoError(t, err)
	i64, err := models.NewCommandValue("test-resource", common.ValueTypeInt64, int64(0))
	require.NoError(t, err)
	f32, err := models.NewCommandValue("test-resource", common.ValueTypeFloat32, float32(0))
	require.NoError(t, err)
	f64, err := models.NewCommandValue("test-resource", common.ValueTypeFloat64, float64(0))
	require.NoError(t, err)
	s, err := models.NewCommandValue("test-resource", common.ValueTypeString, "invalid")
	require.NoError(t, err)

	tests := []struct {
		name          string
		cv            *models.CommandValue
		expectedValue interface{}
		expectedErr   bool
	}{
		{"valid - uint8 CommandValue", u8, uint8(0), false},
		{"valid - uint16 CommandValue", u16, uint16(0), false},
		{"valid - uint32 CommandValue", u32, uint32(0), false},
		{"valid - uint64 CommandValue", u64, uint64(0), false},
		{"valid - int8 CommandValue", i8, int8(0), false},
		{"valid - int16 CommandValue", i16, int16(0), false},
		{"valid - int32 CommandValue", i32, int32(0), false},
		{"valid - int64 CommandValue", i64, int64(0), false},
		{"valid - float32 CommandValue", f32, float32(0), false},
		{"valid - float64 CommandValue", f64, float64(0), false},
		{"invalid - unsupported type for transformation", s, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := commandValueForTransform(tt.cv)
			if !tt.expectedErr {
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func Test_mapCommandValue(t *testing.T) {
	numericValue, err := models.NewCommandValue("test-resource", common.ValueTypeFloat32, float32(123.456))
	require.NoError(t, err)
	stringValue, err := models.NewCommandValue("test-resource", common.ValueTypeString, "key")
	require.NoError(t, err)
	arrayValue, err := models.NewCommandValue("test-resource", common.ValueTypeInt8Array, []int8{1, 2, 3})
	require.NoError(t, err)
	invalid, err := models.NewCommandValue("test-resource", common.ValueTypeString, "invalid")
	require.NoError(t, err)

	mappings := map[string]string{
		"123.456": "value",
		"key":     "value",
		"[1 2 3]": "value",
	}

	tests := []struct {
		name    string
		cv      *models.CommandValue
		success bool
	}{
		{"valid - CommandValue with numeric mapping", numericValue, true},
		{"valid - CommandValue with string mapping", stringValue, true},
		{"valid - CommandValue with array mapping", arrayValue, true},
		{"invalid - mapping not found", invalid, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := mapCommandValue(tt.cv, mappings)
			require.Equal(t, ok, tt.success)
			if ok {
				assert.Equal(t, res.Value, "value")
			}
		})
	}
}
