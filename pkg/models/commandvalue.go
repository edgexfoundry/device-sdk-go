// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

const (
	// Policy limits should be located in global config namespace
	// Currently assigning 16MB (binary), 16 * 2^20 bytes
	MaxBinaryBytes = 16777216
	// DefaultFoloatEncoding indicates the representation of floating value of reading.
	// It would be configurable in system level in the future
	DefaultFloatEncoding = models.Base64Encoding
)

// CommandValue is the struct to represent the reading value of a Get command coming
// from ProtocolDrivers or the parameter of a Put command sending to ProtocolDrivers.
type CommandValue struct {
	// DeviceResourceName is the name of Device Resource for this command
	DeviceResourceName string
	// Origin is an int64 value which indicates the time the reading
	// contained in the CommandValue was read by the ProtocolDriver
	// instance.
	Origin int64
	// Type is a ValueType value which indicates what type of
	// value was returned from the ProtocolDriver instance in
	// response to HandleCommand being called to handle a single
	// ResourceOperation.
	Type string
	// NumericValue is a byte slice with a maximum capacity of
	// 64 bytes, used to hold a numeric value returned by a
	// ProtocolDriver instance. The value can be converted to
	// its native type by referring to the the value of ResType.
	NumericValue []byte
	// stringValue is a string value returned as a value by a ProtocolDriver instance.
	stringValue string
	// BinValue is a binary value with a maximum capacity of 16 MB,
	// used to hold binary values returned by a ProtocolDriver instance.
	BinValue []byte
}

// NewBoolValue creates a CommandValue of Type Bool with the given value.
func NewBoolValue(DeviceResourceName string, origin int64, value bool) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeBool}
	err = encodeValue(cv, value)
	return
}

// NewBoolArrayValue creates a CommandValue of Type BoolArray with the given value.
func NewBoolArrayValue(DeviceResourceName string, origin int64, value []bool) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeBoolArray,
		stringValue:        string(jsonValue),
	}

	return
}

// NewStringValue creates a CommandValue of Type string with the given value.
func NewStringValue(DeviceResourceName string, origin int64, value string) (cv *CommandValue) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeString, stringValue: value}
	return
}

// NewUint8Value creates a CommandValue of Type Uint8 with the given value.
func NewUint8Value(DeviceResourceName string, origin int64, value uint8) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeUint8}
	err = encodeValue(cv, value)
	return
}

// NewUint8ArrayValue creates a CommandValue of Type Uint8Array with the given value.
func NewUint8ArrayValue(DeviceResourceName string, origin int64, value []uint8) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeUint8Array,
		stringValue:        strings.Join(strings.Fields(fmt.Sprintf("%d", value)), ","),
		BinValue:           jsonValue,
	}

	return
}

// NewUint16Value creates a CommandValue of Type Uint16 with the given value.
func NewUint16Value(DeviceResourceName string, origin int64, value uint16) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeUint16}
	err = encodeValue(cv, value)
	return
}

// NewUint16ArrayValue creates a CommandValue of Type Uint16Array with the given value.
func NewUint16ArrayValue(DeviceResourceName string, origin int64, value []uint16) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeUint16Array,
		stringValue:        strings.Join(strings.Fields(fmt.Sprintf("%d", value)), ","),
		BinValue:           jsonValue,
	}

	return
}

// NewUint32Value creates a CommandValue of Type Uint32 with the given value.
func NewUint32Value(DeviceResourceName string, origin int64, value uint32) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeUint32}
	err = encodeValue(cv, value)
	return
}

// NewUint32ArrayValue creates a CommandValue of Type Uint32Array with the given value.
func NewUint32ArrayValue(DeviceResourceName string, origin int64, value []uint32) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeUint32Array,
		stringValue:        strings.Join(strings.Fields(fmt.Sprintf("%d", value)), ","),
		BinValue:           jsonValue,
	}

	return
}

// NewUint64Value creates a CommandValue of Type Uint64 with the given value.
func NewUint64Value(DeviceResourceName string, origin int64, value uint64) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeUint64}
	err = encodeValue(cv, value)
	return
}

// NewUint64ArrayValue creates a CommandValue of Type Uint64Array with the given value.
func NewUint64ArrayValue(DeviceResourceName string, origin int64, value []uint64) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeUint64Array,
		stringValue:        strings.Join(strings.Fields(fmt.Sprintf("%d", value)), ","),
		BinValue:           jsonValue,
	}

	return
}

// NewInt8Value creates a CommandValue of Type Int8 with the given value.
func NewInt8Value(DeviceResourceName string, origin int64, value int8) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeInt8}
	err = encodeValue(cv, value)
	return
}

// NewInt8ArrayValue creates a CommandValue of Type Int8Array with the given value.
func NewInt8ArrayValue(DeviceResourceName string, origin int64, value []int8) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeInt8Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewInt16Value creates a CommandValue of Type Int16 with the given value.
func NewInt16Value(DeviceResourceName string, origin int64, value int16) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeInt16}
	err = encodeValue(cv, value)
	return
}

// NewInt16ArrayValue creates a CommandValue of Type Int16Array with the given value.
func NewInt16ArrayValue(DeviceResourceName string, origin int64, value []int16) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeInt16Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewInt32Value creates a CommandValue of Type Int32 with the given value.
func NewInt32Value(DeviceResourceName string, origin int64, value int32) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeInt32}
	err = encodeValue(cv, value)
	return
}

// NewInt32ArrayValue creates a CommandValue of Type Int32Array with the given value.
func NewInt32ArrayValue(DeviceResourceName string, origin int64, value []int32) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeInt32Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewInt64Value creates a CommandValue of Type Int64 with the given value.
func NewInt64Value(DeviceResourceName string, origin int64, value int64) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeInt64}
	err = encodeValue(cv, value)
	return
}

// NewInt64ArrayValue creates a CommandValue of Type Int64Array with the given value.
func NewInt64ArrayValue(DeviceResourceName string, origin int64, value []int64) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeInt64Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewFloat32Value creates a CommandValue of Type Float32 with the given value.
func NewFloat32Value(DeviceResourceName string, origin int64, value float32) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeFloat32}
	err = encodeValue(cv, value)
	return
}

// NewFloat32ArrayValue creates a CommandValue of Type Float32Array with the given value.
func NewFloat32ArrayValue(DeviceResourceName string, origin int64, value []float32) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeFloat32Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewFloat64Value creates a CommandValue of Type Float64 with the given value.
func NewFloat64Value(DeviceResourceName string, origin int64, value float64) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeFloat64}
	err = encodeValue(cv, value)
	return
}

// NewFloat64ArrayValue creates a CommandValue of Type Float64Array with the given value.
func NewFloat64ArrayValue(DeviceResourceName string, origin int64, value []float64) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               v2.ValueTypeFloat64Array,
		stringValue:        string(jsonValue),
	}

	return
}

//NewCommandValue create a CommandValue according to the Type supplied.
func NewCommandValue(DeviceResourceName string, origin int64, value interface{}, t string) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: t}
	switch t {
	case v2.ValueTypeBinary:
		// assign cv.BinValue
		cv.BinValue = value.([]byte)
	case v2.ValueTypeString:
		cv.stringValue = value.(string)
	default:
		err = encodeValue(cv, value)
	}
	return
}

// NewBinaryValue creates a CommandValue with binary payload and enforces the memory limit for event readings.
func NewBinaryValue(DeviceResourceName string, origin int64, value []byte) (cv *CommandValue, err error) {
	if binary.Size(value) > MaxBinaryBytes {
		return nil, fmt.Errorf("requested CommandValue payload exceeds limit for binary readings (%v bytes)", MaxBinaryBytes)
	}
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: v2.ValueTypeBinary, BinValue: value}
	return
}

func encodeValue(cv *CommandValue, value interface{}) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, value)
	if err == nil {
		cv.NumericValue = buf.Bytes()
	}
	return err
}

func decodeValue(reader io.Reader, value interface{}) error {
	err := binary.Read(reader, binary.BigEndian, value)
	return err
}

// ValueToString returns the string format of the value.
// In EdgeX, float value has two kinds of representation, Base64, and eNotation.
// Users can specify the floatEncoding in the properties value of the device profile, like floatEncoding: "Base64" or floatEncoding: "eNotation".
func (cv *CommandValue) ValueToString(encoding ...string) (str string) {
	if cv.Type == v2.ValueTypeString {
		str = cv.stringValue
		return
	}

	reader := bytes.NewReader(cv.NumericValue)

	switch cv.Type {
	case v2.ValueTypeBool:
		var res bool
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatBool(res)
	case v2.ValueTypeUint8:
		var res uint8
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case v2.ValueTypeUint16:
		var res uint16
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case v2.ValueTypeUint32:
		var res uint32
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case v2.ValueTypeUint64:
		var res uint64
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(res, 10)
	case v2.ValueTypeInt8:
		var res int8
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case v2.ValueTypeInt16:
		var res int16
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case v2.ValueTypeInt32:
		var res int32
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case v2.ValueTypeInt64:
		var res int64
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(res, 10)
	case v2.ValueTypeFloat32:
		floatEncoding := getFloatEncoding(encoding)

		if floatEncoding == models.ENotation {
			var res float32
			binary.Read(reader, binary.BigEndian, &res)
			str = fmt.Sprintf("%e", res)
		} else if floatEncoding == models.Base64Encoding {
			str = base64.StdEncoding.EncodeToString(cv.NumericValue)
		}
	case v2.ValueTypeFloat64:
		floatEncoding := getFloatEncoding(encoding)

		if floatEncoding == models.ENotation {
			var res float64
			binary.Read(reader, binary.BigEndian, &res)
			str = fmt.Sprintf("%e", res)
		} else if floatEncoding == models.Base64Encoding {
			str = base64.StdEncoding.EncodeToString(cv.NumericValue)
		}
	case v2.ValueTypeBinary:
		// produce string representation of first 20 bytes of binary value
		str = fmt.Sprintf(fmt.Sprintf("Binary: [%v...]", string(cv.BinValue[:20])))
	default:
		// ArrayType
		str = cv.stringValue
	}

	return
}

func getFloatEncoding(encoding []string) string {
	if len(encoding) > 0 {
		if encoding[0] == models.Base64Encoding {
			return models.Base64Encoding
		} else if encoding[0] == models.ENotation {
			return models.ENotation
		}
	}

	return DefaultFloatEncoding
}

// String returns a string representation of a CommandValue instance.
func (cv *CommandValue) String() (str string) {
	originStr := fmt.Sprintf("Origin: %d, ", cv.Origin)
	valueStr := cv.Type + ": " + cv.ValueToString()
	str = originStr + valueStr

	return
}

// BoolValue returns the value in bool data type, and returns error if the Type is not Bool.
func (cv *CommandValue) BoolValue() (bool, error) {
	var value bool
	if cv.Type != v2.ValueTypeBool {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// BoolArrayValue returns the value in an array of bool type, and returns error if the Type is not BoolArray.
func (cv *CommandValue) BoolArrayValue() ([]bool, error) {
	var value []bool
	if cv.Type != v2.ValueTypeBoolArray {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// StringValue returns the value in string data type, and returns error if the Type is not String.
func (cv *CommandValue) StringValue() (string, error) {
	value := cv.stringValue
	if cv.Type != v2.ValueTypeString {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	return value, nil
}

// Uint8Value returns the value in uint8 data type, and returns error if the Type is not Uint8.
func (cv *CommandValue) Uint8Value() (uint8, error) {
	var value uint8
	if cv.Type != v2.ValueTypeUint8 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Uint8ArrayValue returns the value in an array of uint8 type, and returns error if the Type is not Uint8Array.
func (cv *CommandValue) Uint8ArrayValue() ([]uint8, error) {
	var value []uint8
	if cv.Type != v2.ValueTypeUint8Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal(cv.BinValue, &value)
	return value, err
}

// Uint16Value returns the value in uint16 data type, and returns error if the Type is not Uint16.
func (cv *CommandValue) Uint16Value() (uint16, error) {
	var value uint16
	if cv.Type != v2.ValueTypeUint16 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Uint16ArrayValue returns the value in an array of uint16 type, and returns error if the Type is not Uint16Array.
func (cv *CommandValue) Uint16ArrayValue() ([]uint16, error) {
	var value []uint16
	if cv.Type != v2.ValueTypeUint16Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal(cv.BinValue, &value)
	return value, err
}

// Uint32Value returns the value in uint32 data type, and returns error if the Type is not Uint32.
func (cv *CommandValue) Uint32Value() (uint32, error) {
	var value uint32
	if cv.Type != v2.ValueTypeUint32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Uint32ArrayValue returns the value in an array of uint32 type, and returns error if the Type is not Uint32Array.
func (cv *CommandValue) Uint32ArrayValue() ([]uint32, error) {
	var value []uint32
	if cv.Type != v2.ValueTypeUint32Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal(cv.BinValue, &value)
	return value, err
}

// Uint64Value returns the value in uint64 data type, and returns error if the Type is not Uint64.
func (cv *CommandValue) Uint64Value() (uint64, error) {
	var value uint64
	if cv.Type != v2.ValueTypeUint64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Uint64ArrayValue returns the value in an array of uint64 type, and returns error if the Type is not Uint64Array.
func (cv *CommandValue) Uint64ArrayValue() ([]uint64, error) {
	var value []uint64
	if cv.Type != v2.ValueTypeUint64Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal(cv.BinValue, &value)
	return value, err
}

// Int8Value returns the value in int8 data type, and returns error if the Type is not Int8.
func (cv *CommandValue) Int8Value() (int8, error) {
	var value int8
	if cv.Type != v2.ValueTypeInt8 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Int8ArrayValue returns the value in an array of int8 type, and returns error if the Type is not Int8Array.
func (cv *CommandValue) Int8ArrayValue() ([]int8, error) {
	var value []int8
	if cv.Type != v2.ValueTypeInt8Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Int16Value returns the value in int16 data type, and returns error if the Type is not Int16.
func (cv *CommandValue) Int16Value() (int16, error) {
	var value int16
	if cv.Type != v2.ValueTypeInt16 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Int16ArrayValue returns the value in an array of int16 type, and returns error if the Type is not Int16Array.
func (cv *CommandValue) Int16ArrayValue() ([]int16, error) {
	var value []int16
	if cv.Type != v2.ValueTypeInt16Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Int32Value returns the value in int32 data type, and returns error if the Type is not Int32.
func (cv *CommandValue) Int32Value() (int32, error) {
	var value int32
	if cv.Type != v2.ValueTypeInt32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Int32ArrayValue returns the value in an array of int32 type, and returns error if the Type is not Int32Array.
func (cv *CommandValue) Int32ArrayValue() ([]int32, error) {
	var value []int32
	if cv.Type != v2.ValueTypeInt32Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Int64Value returns the value in int64 data type, and returns error if the Type is not Int64.
func (cv *CommandValue) Int64Value() (int64, error) {
	var value int64
	if cv.Type != v2.ValueTypeInt64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Int64ArrayValue returns the value in an array of int64 type, and returns error if the Type is not Int64Array.
func (cv *CommandValue) Int64ArrayValue() ([]int64, error) {
	var value []int64
	if cv.Type != v2.ValueTypeInt64Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Float32Value returns the value in float32 data type, and returns error if the Type is not Float32.
func (cv *CommandValue) Float32Value() (float32, error) {
	var value float32
	if cv.Type != v2.ValueTypeFloat32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Float32ArrayValue returns the value in an array of float32 type, and returns error if the Type is not Float32Array.
func (cv *CommandValue) Float32ArrayValue() ([]float32, error) {
	var value []float32
	if cv.Type != v2.ValueTypeFloat32Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Float64Value returns the value in float64 data type, and returns error if the Type is not Float64.
func (cv *CommandValue) Float64Value() (float64, error) {
	var value float64
	if cv.Type != v2.ValueTypeFloat64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Float64ArrayValue returns the value in an array of float64 type, and returns error if the Type is not Float64Array.
func (cv *CommandValue) Float64ArrayValue() ([]float64, error) {
	var value []float64
	if cv.Type != v2.ValueTypeFloat64Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// BinaryValue returns the value in []byte data type, and returns error if the Type is not Binary.
func (cv *CommandValue) BinaryValue() ([]byte, error) {
	var value []byte
	if cv.Type != v2.ValueTypeBinary {
		return value, fmt.Errorf("the CommandValue (%s) data type (%v) is not binary", cv.String(), cv.Type)
	}
	return cv.BinValue, nil
}
