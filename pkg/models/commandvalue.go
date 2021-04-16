// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/binary"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
)

const (
	// Policy limits should be located in global config namespace
	// Currently assigning 16MB (binary), 16 * 2^20 bytes
	MaxBinaryBytes = 16777216
)

// CommandValue is the struct to represent the reading value of a Get command coming
// from ProtocolDrivers or the parameter of a Put command sending to ProtocolDrivers.
type CommandValue struct {
	// DeviceResourceName is the name of Device Resource for this command
	DeviceResourceName string
	// Type indicates what type of value was returned from the ProtocolDriver instance in
	// response to HandleCommand being called to handle a single ResourceOperation.
	Type string
	// Value holds value returned by a ProtocolDriver instance.
	// The value can be converted to its native type by referring to ValueType.
	Value interface{}
	// Origin is an int64 value which indicates the time the reading
	// contained in the CommandValue was read by the ProtocolDriver
	// instance.
	Origin int64
}

// NewCommandValue create a CommandValue according to the valueType supplied.
func NewCommandValue(deviceResourceName string, valueType string, value interface{}) (*CommandValue, error) {
	err := validate(valueType, value)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to create CommandValue", err)
	}

	return &CommandValue{
		DeviceResourceName: deviceResourceName,
		Type:               valueType,
		Value:              value}, nil
}

// ValueToString returns the string format of the value.
func (cv *CommandValue) ValueToString() string {
	if cv.Type == v2.ValueTypeBinary {
		binaryValue := cv.Value.([]byte)
		return fmt.Sprintf("Binary: [%v...]", string(binaryValue[:20]))
	}
	return fmt.Sprintf("%v", cv.Value)
}

// String returns a string representation of a CommandValue instance.
func (cv *CommandValue) String() string {
	return fmt.Sprintf("DeviceResource: %s, %s: %s", cv.DeviceResourceName, cv.Type, cv.ValueToString())
}

// BoolValue returns the value in bool data type, and returns error if the Type is not Bool.
func (cv *CommandValue) BoolValue() (bool, error) {
	var value bool
	if cv.Type != v2.ValueTypeBool {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeBool)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(bool)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// BoolArrayValue returns the value in an array of bool type, and returns error if the Type is not BoolArray.
func (cv *CommandValue) BoolArrayValue() ([]bool, error) {
	var value []bool
	if cv.Type != v2.ValueTypeBoolArray {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeBoolArray)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]bool)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// StringValue returns the value in string data type, and returns error if the Type is not String.
func (cv *CommandValue) StringValue() (string, error) {
	var value string
	if cv.Type != v2.ValueTypeString {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeString)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(string)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Uint8Value returns the value in uint8 data type, and returns error if the Type is not Uint8.
func (cv *CommandValue) Uint8Value() (uint8, error) {
	var value uint8
	if cv.Type != v2.ValueTypeUint8 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeUint8)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(uint8)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Uint8ArrayValue returns the value in an array of uint8 type, and returns error if the Type is not Uint8Array.
func (cv *CommandValue) Uint8ArrayValue() ([]uint8, error) {
	var value []uint8
	if cv.Type != v2.ValueTypeUint8Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeUint8Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]uint8)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Uint16Value returns the value in uint16 data type, and returns error if the Type is not Uint16.
func (cv *CommandValue) Uint16Value() (uint16, error) {
	var value uint16
	if cv.Type != v2.ValueTypeUint16 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeUint16)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(uint16)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Uint16ArrayValue returns the value in an array of uint16 type, and returns error if the Type is not Uint16Array.
func (cv *CommandValue) Uint16ArrayValue() ([]uint16, error) {
	var value []uint16
	if cv.Type != v2.ValueTypeUint16Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeUint16Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]uint16)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Uint32Value returns the value in uint32 data type, and returns error if the Type is not Uint32.
func (cv *CommandValue) Uint32Value() (uint32, error) {
	var value uint32
	if cv.Type != v2.ValueTypeUint32 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeUint32)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(uint32)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Uint32ArrayValue returns the value in an array of uint32 type, and returns error if the Type is not Uint32Array.
func (cv *CommandValue) Uint32ArrayValue() ([]uint32, error) {
	var value []uint32
	if cv.Type != v2.ValueTypeUint32Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeUint32Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]uint32)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Uint64Value returns the value in uint64 data type, and returns error if the Type is not Uint64.
func (cv *CommandValue) Uint64Value() (uint64, error) {
	var value uint64
	if cv.Type != v2.ValueTypeUint64 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeUint64)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(uint64)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Uint64ArrayValue returns the value in an array of uint64 type, and returns error if the Type is not Uint64Array.
func (cv *CommandValue) Uint64ArrayValue() ([]uint64, error) {
	var value []uint64
	if cv.Type != v2.ValueTypeUint64Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeUint64Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]uint64)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Int8Value returns the value in int8 data type, and returns error if the Type is not Int8.
func (cv *CommandValue) Int8Value() (int8, error) {
	var value int8
	if cv.Type != v2.ValueTypeInt8 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeInt8)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(int8)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Int8ArrayValue returns the value in an array of int8 type, and returns error if the Type is not Int8Array.
func (cv *CommandValue) Int8ArrayValue() ([]int8, error) {
	var value []int8
	if cv.Type != v2.ValueTypeInt8Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeInt8Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]int8)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Int16Value returns the value in int16 data type, and returns error if the Type is not Int16.
func (cv *CommandValue) Int16Value() (int16, error) {
	var value int16
	if cv.Type != v2.ValueTypeInt16 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeInt16)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(int16)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Int16ArrayValue returns the value in an array of int16 type, and returns error if the Type is not Int16Array.
func (cv *CommandValue) Int16ArrayValue() ([]int16, error) {
	var value []int16
	if cv.Type != v2.ValueTypeInt16Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeInt16Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]int16)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Int32Value returns the value in int32 data type, and returns error if the Type is not Int32.
func (cv *CommandValue) Int32Value() (int32, error) {
	var value int32
	if cv.Type != v2.ValueTypeInt32 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeInt32)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(int32)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Int32ArrayValue returns the value in an array of int32 type, and returns error if the Type is not Int32Array.
func (cv *CommandValue) Int32ArrayValue() ([]int32, error) {
	var value []int32
	if cv.Type != v2.ValueTypeInt32Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeInt32Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]int32)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Int64Value returns the value in int64 data type, and returns error if the Type is not Int64.
func (cv *CommandValue) Int64Value() (int64, error) {
	var value int64
	if cv.Type != v2.ValueTypeInt64 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeInt64)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(int64)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Int64ArrayValue returns the value in an array of int64 type, and returns error if the Type is not Int64Array.
func (cv *CommandValue) Int64ArrayValue() ([]int64, error) {
	var value []int64
	if cv.Type != v2.ValueTypeInt64Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeInt64Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]int64)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Float32Value returns the value in float32 data type, and returns error if the Type is not Float32.
func (cv *CommandValue) Float32Value() (float32, error) {
	var value float32
	if cv.Type != v2.ValueTypeFloat32 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeFloat32)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(float32)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Float32ArrayValue returns the value in an array of float32 type, and returns error if the Type is not Float32Array.
func (cv *CommandValue) Float32ArrayValue() ([]float32, error) {
	var value []float32
	if cv.Type != v2.ValueTypeFloat32Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeFloat32Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]float32)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Float64Value returns the value in float64 data type, and returns error if the Type is not Float64.
func (cv *CommandValue) Float64Value() (float64, error) {
	var value float64
	if cv.Type != v2.ValueTypeFloat64 {
		errMsg := fmt.Sprintf("cannot convert CommandValue of %s to %s", cv.Type, v2.ValueTypeFloat64)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.(float64)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// Float64ArrayValue returns the value in an array of float64 type, and returns error if the Type is not Float64Array.
func (cv *CommandValue) Float64ArrayValue() ([]float64, error) {
	var value []float64
	if cv.Type != v2.ValueTypeFloat64Array {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeFloat64Array)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]float64)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// BinaryValue returns the value in []byte data type, and returns error if the Type is not Binary.
func (cv *CommandValue) BinaryValue() ([]byte, error) {
	var value []byte
	if cv.Type != v2.ValueTypeBinary {
		errMsg := fmt.Sprintf("cannot convert %s to %s", cv.Type, v2.ValueTypeBinary)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	value, ok := cv.Value.([]byte)
	if !ok {
		errMsg := fmt.Sprintf("failed to transfrom %v to %T", cv.Value, value)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return value, nil
}

// validate checks if the given value can be converted to specified valueType by
// performing type assertion
func validate(valueType string, value interface{}) error {
	var ok bool
	switch valueType {
	case v2.ValueTypeString:
		_, ok = value.(string)
	case v2.ValueTypeBool:
		_, ok = value.(bool)
	case v2.ValueTypeBoolArray:
		_, ok = value.([]bool)
	case v2.ValueTypeUint8:
		_, ok = value.(uint8)
	case v2.ValueTypeUint8Array:
		_, ok = value.([]uint8)
	case v2.ValueTypeUint16:
		_, ok = value.(uint16)
	case v2.ValueTypeUint16Array:
		_, ok = value.([]uint16)
	case v2.ValueTypeUint32:
		_, ok = value.(uint32)
	case v2.ValueTypeUint32Array:
		_, ok = value.([]uint32)
	case v2.ValueTypeUint64:
		_, ok = value.(uint64)
	case v2.ValueTypeUint64Array:
		_, ok = value.([]uint64)
	case v2.ValueTypeInt8:
		_, ok = value.(int8)
	case v2.ValueTypeInt8Array:
		_, ok = value.([]int8)
	case v2.ValueTypeInt16:
		_, ok = value.(int16)
	case v2.ValueTypeInt16Array:
		_, ok = value.([]int16)
	case v2.ValueTypeInt32:
		_, ok = value.(int32)
	case v2.ValueTypeInt32Array:
		_, ok = value.([]int32)
	case v2.ValueTypeInt64:
		_, ok = value.(int64)
	case v2.ValueTypeInt64Array:
		_, ok = value.([]int64)
	case v2.ValueTypeFloat32:
		_, ok = value.(float32)
	case v2.ValueTypeFloat32Array:
		_, ok = value.([]float32)
	case v2.ValueTypeFloat64:
		_, ok = value.(float64)
	case v2.ValueTypeFloat64Array:
		_, ok = value.([]float64)
	case v2.ValueTypeBinary:
		_, ok = value.([]byte)
		if binary.Size(value) > MaxBinaryBytes {
			errMsg := fmt.Sprintf("value payload exceeds limit for binary readings (%v bytes)", MaxBinaryBytes)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
		}
	default:
		return errors.NewCommonEdgeX(errors.KindServerError, "unrecognized value type", nil)
	}

	if !ok {
		errMsg := fmt.Sprintf("failed to convert interface value %v to Type %s", value, valueType)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return nil
}
