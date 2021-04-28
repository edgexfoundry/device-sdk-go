// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validate(t *testing.T) {
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
			err := validate(tt.valueType, tt.value)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandValue_ValueToString(t *testing.T) {
	uintCommandValue, err := NewCommandValue("test-resource", v2.ValueTypeUint8, uint8(1))
	require.NoError(t, err)
	intCommandValue, err := NewCommandValue("test-resource", v2.ValueTypeInt8, int8(-1))
	require.NoError(t, err)
	floatCommandValue, err := NewCommandValue("test-resource", v2.ValueTypeFloat64, float64(123.456))
	require.NoError(t, err)
	stringCommandValue, err := NewCommandValue("test-resource", v2.ValueTypeString, "string")
	require.NoError(t, err)
	boolCommandValue, err := NewCommandValue("test-resource", v2.ValueTypeBool, true)
	require.NoError(t, err)
	binaryValue := make([]byte, 100)
	rand.Read(binaryValue)
	binaryCommandValue, err := NewCommandValue("test-resource", v2.ValueTypeBinary, binaryValue)
	require.NoError(t, err)

	tests := []struct {
		name     string
		cv       *CommandValue
		expected string
	}{
		{"valid - CommandValue with uint Value", uintCommandValue, "1"},
		{"valid - CommandValue with int Value", intCommandValue, "-1"},
		{"valid - CommandValue with float Value", floatCommandValue, "123.456"},
		{"valid - CommandValue with string Value", stringCommandValue, "string"},
		{"valid - CommandValue with boolean Value", boolCommandValue, "true"},
		{"valid - CommandValue with binary Value", binaryCommandValue, fmt.Sprintf("Binary: [%v...]", string(binaryValue[:20]))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.cv.ValueToString()
			assert.Equal(t, res, tt.expected)
		})
	}
}

func TestCommandValue_BoolValue(t *testing.T) {
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBool, Value: true}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: true}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBool, Value: "true"}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    bool
		expectedErr bool
	}{
		{"valid - CommandValue with bool Value", valid, true, false},
		{"invalid - ValueType is not Bool", invalidType, true, true},
		{"invalid - Value is not boolean", invalidValue, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.BoolValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_BoolArrayValue(t *testing.T) {
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBoolArray, Value: []bool{true, false, true}}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: true}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBoolArray, Value: "true"}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []bool
		expectedErr bool
	}{
		{"valid - CommandValue with []bool Value", valid, []bool{true, false, true}, false},
		{"invalid - ValueType is not BoolArray", invalidType, nil, true},
		{"invalid - Value is not []bool", invalidValue, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.BoolArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_StringValue(t *testing.T) {
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeString, Value: "test"}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: "test"}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeString, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    string
		expectedErr bool
	}{
		{"valid - CommandValue with string Value", valid, "test", false},
		{"invalid - ValueType is not String", invalidType, "test", true},
		{"invalid - Value is not string", invalidValue, "test", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.StringValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Uint8Value(t *testing.T) {
	value := uint8(1)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint8, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint8, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    uint8
		expectedErr bool
	}{
		{"valid - CommandValue with uint8 Value", valid, value, false},
		{"invalid - ValueType is not Uint8", invalidType, value, true},
		{"invalid - Value is not uint8", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Uint8Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Uint8ArrayValue(t *testing.T) {
	value := []uint8{1}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint8Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint8Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []uint8
		expectedErr bool
	}{
		{"valid - CommandValue with []uint8 Value", valid, value, false},
		{"invalid - ValueType is not Uint8Array", invalidType, value, true},
		{"invalid - Value is not []uint8", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Uint8ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Uint16Value(t *testing.T) {
	value := uint16(1)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint16, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint16, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    uint16
		expectedErr bool
	}{
		{"valid - CommandValue with uint16 Value", valid, value, false},
		{"invalid - ValueType is not Uint16", invalidType, value, true},
		{"invalid - Value is not uint16", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Uint16Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Uint16ArrayValue(t *testing.T) {
	value := []uint16{1}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint16Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint16Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []uint16
		expectedErr bool
	}{
		{"valid - CommandValue with []uint16 Value", valid, value, false},
		{"invalid - ValueType is not Uint16Array", invalidType, value, true},
		{"invalid - Value is not []uint16", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Uint16ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Uint32Value(t *testing.T) {
	value := uint32(1)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint32, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint32, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    uint32
		expectedErr bool
	}{
		{"valid - CommandValue with uint32 Value", valid, value, false},
		{"invalid - ValueType is not Uint32", invalidType, value, true},
		{"invalid - Value is not uint32", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Uint32Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Uint32ArrayValue(t *testing.T) {
	value := []uint32{1}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint32Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint32Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []uint32
		expectedErr bool
	}{
		{"valid - CommandValue with []uint32 Value", valid, value, false},
		{"invalid - ValueType is not Uint32Array", invalidType, value, true},
		{"invalid - Value is not []uint32", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Uint32ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Uint64Value(t *testing.T) {
	value := uint64(1)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint64, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint64, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    uint64
		expectedErr bool
	}{
		{"valid - CommandValue with uint64 Value", valid, value, false},
		{"invalid - ValueType is not Uint64", invalidType, value, true},
		{"invalid - Value is not uint64", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Uint64Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Uint64ArrayValue(t *testing.T) {
	value := []uint64{1}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint64Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeUint64Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []uint64
		expectedErr bool
	}{
		{"valid - CommandValue with []uint64 Value", valid, value, false},
		{"invalid - ValueType is not Uint64Array", invalidType, value, true},
		{"invalid - Value is not []uint64", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Uint64ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Int8Value(t *testing.T) {
	value := int8(-1)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt8, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt8, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    int8
		expectedErr bool
	}{
		{"valid - CommandValue with int8 Value", valid, value, false},
		{"invalid - ValueType is not Int8", invalidType, value, true},
		{"invalid - Value is not int8", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Int8Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Int8ArrayValue(t *testing.T) {
	value := []int8{-1}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt8Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt8Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []int8
		expectedErr bool
	}{
		{"valid - CommandValue with []int8 Value", valid, value, false},
		{"invalid - ValueType is not Int8Array", invalidType, value, true},
		{"invalid - Value is not []int8", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Int8ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Int16Value(t *testing.T) {
	value := int16(-1)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt16, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt16, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    int16
		expectedErr bool
	}{
		{"valid - CommandValue with int16 Value", valid, value, false},
		{"invalid - ValueType is not Int16", invalidType, value, true},
		{"invalid - Value is not int16", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Int16Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Int16ArrayValue(t *testing.T) {
	value := []int16{-1}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt16Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt16Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []int16
		expectedErr bool
	}{
		{"valid - CommandValue with []int16 Value", valid, value, false},
		{"invalid - ValueType is not Int16Array", invalidType, value, true},
		{"invalid - Value is not []int16", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Int16ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Int32Value(t *testing.T) {
	value := int32(-1)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt32, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt32, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    int32
		expectedErr bool
	}{
		{"valid - CommandValue with int32 Value", valid, value, false},
		{"invalid - ValueType is not Int32", invalidType, value, true},
		{"invalid - Value is not int32", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Int32Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Int32ArrayValue(t *testing.T) {
	value := []int32{-1}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt32Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt32Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []int32
		expectedErr bool
	}{
		{"valid - CommandValue with []int32 Value", valid, value, false},
		{"invalid - ValueType is not Int32Array", invalidType, value, true},
		{"invalid - Value is not []int32", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Int32ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Int64Value(t *testing.T) {
	value := int64(-1)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt64, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt64, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    int64
		expectedErr bool
	}{
		{"valid - CommandValue with int64 Value", valid, value, false},
		{"invalid - ValueType is not Int64", invalidType, value, true},
		{"invalid - Value is not int64", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Int64Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Int64ArrayValue(t *testing.T) {
	value := []int64{-1}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt64Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeInt64Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []int64
		expectedErr bool
	}{
		{"valid - CommandValue with []int64 Value", valid, value, false},
		{"invalid - ValueType is not Int64Array", invalidType, value, true},
		{"invalid - Value is not []int64", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Int64ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Float32Value(t *testing.T) {
	value := float32(13.456)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeFloat32, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeFloat32, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    float32
		expectedErr bool
	}{
		{"valid - CommandValue with float32 Value", valid, value, false},
		{"invalid - ValueType is not Float32", invalidType, value, true},
		{"invalid - Value is not float32", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Float32Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Float32ArrayValue(t *testing.T) {
	value := []float32{12.345}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeFloat32Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeFloat32Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []float32
		expectedErr bool
	}{
		{"valid - CommandValue with []float32 Value", valid, value, false},
		{"invalid - ValueType is not Float32Array", invalidType, value, true},
		{"invalid - Value is not []float32", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Float32ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Float64Value(t *testing.T) {
	value := float64(13.456)
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeFloat64, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeFloat64, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    float64
		expectedErr bool
	}{
		{"valid - CommandValue with float64 Value", valid, value, false},
		{"invalid - ValueType is not Float64", invalidType, value, true},
		{"invalid - Value is not float64", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Float64Value()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}

func TestCommandValue_Float64ArrayValue(t *testing.T) {
	value := []float64{12.345}
	valid := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeFloat64Array, Value: value}
	invalidType := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeBinary, Value: value}
	invalidValue := &CommandValue{DeviceResourceName: "test-resource", Type: v2.ValueTypeFloat64Array, Value: true}

	tests := []struct {
		name        string
		cv          *CommandValue
		expected    []float64
		expectedErr bool
	}{
		{"valid - CommandValue with []float64 Value", valid, value, false},
		{"invalid - ValueType is not Float64Array", invalidType, value, true},
		{"invalid - Value is not []float64", invalidValue, value, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.cv.Float64ArrayValue()
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Equal(t, res, tt.expected)
			}
		})
	}
}
