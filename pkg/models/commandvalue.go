// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
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

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// ValueType indicates the type of value being passed back
// from a ProtocolDriver instance.
type ValueType int

const (
	// Bool indicates that the value is a bool,
	// stored in CommandValue's boolRes member.
	Bool ValueType = iota
	// BoolArray indicates that the value is array of bool,
	// store in CommandValue's stringValue member.
	BoolArray
	// String indicates that the value is a string,
	// stored in CommandValue's stringRes member.
	String
	// Uint8 indicates that the value is a uint8 that
	// is stored in CommandValue's NumericRes member.
	Uint8
	// Uint8Array indicates that the value is array of uint8,
	// store in CommandValue's stringValue member.
	Uint8Array
	// Uint16 indicates that the value is a uint16 that
	// is stored in CommandValue's NumericRes member.
	Uint16
	// Uint16Array indicates that the value is array of uint16,
	// store in CommandValue's stringValue member.
	Uint16Array
	// Uint32 indicates that the value is a uint32 that
	// is stored in CommandValue's NumericRes member.
	Uint32
	// Uint32Array indicates that the value is array of uint32,
	// store in CommandValue's stringValue member.
	Uint32Array
	// Uint64 indicates that the value is a uint64 that
	// is stored in CommandValue's NumericRes member.
	Uint64
	// Uint64Array indicates that the value is array of uint64,
	// store in CommandValue's stringValue member.
	Uint64Array
	// Int8 indicates that the value is a int8 that
	// is stored in CommandValue's NumericRes member.
	Int8
	// Int8Array indicates that the value is array of int8,
	// store in CommandValue's stringValue member.
	Int8Array
	// Int16 indicates that the value is a int16 that
	// is stored in CommandValue's NumericRes member.
	Int16
	// Int16Array indicates that the value is array of int16,
	// store in CommandValue's stringValue member.
	Int16Array
	// Int32 indicates that the value is a int32 that
	// is stored in CommandValue's NumericRes member.
	Int32
	// Int32Array indicates that the value is array of int32,
	// store in CommandValue's stringValue member.
	Int32Array
	// Int64 indicates that the value is a int64 that
	// is stored in CommandValue's NumericRes member.
	Int64
	// Int64Array indicates that the value is array of int64,
	// store in CommandValue's stringValue member.
	Int64Array
	// Float32 indicates that the value is a float32 that
	// is stored in CommandValue's NumericRes member.
	Float32
	// Float32Array indicates that the value is array of float32,
	// store in CommandValue's stringValue member.
	Float32Array
	// Float64 indicates that the value is a float64 that
	// is stored in CommandValue's NumericRes member.
	Float64
	// Float64Array indicates that the value is array of float64,
	// store in CommandValue's stringValue member.
	Float64Array
	// Binary indicates that the value is a binary payload that
	// is stored in CommandValue's ByteArrRes member.
	Binary
)

const (
	// Policy limits should be located in global config namespace
	// Currently assigning 16MB (binary), 16 * 2^20 bytes
	MaxBinaryBytes = 16777216
	// DefaultFoloatEncoding indicates the representation of floating value of reading.
	// It would be configurable in system level in the future
	DefaultFloatEncoding = contract.Base64Encoding
)

// ParseValueType could get ValueType from type name in string format
// if the type name cannot be parsed correctly, return String ValueType
func ParseValueType(typeName string) ValueType {
	switch strings.ToUpper(typeName) {
	case "BOOL":
		return Bool
	case "BOOLARRAY":
		return BoolArray
	case "STRING":
		return String
	case "UINT8":
		return Uint8
	case "UINT8ARRAY":
		return Uint8Array
	case "UINT16":
		return Uint16
	case "UINT16ARRAY":
		return Uint16Array
	case "UINT32":
		return Uint32
	case "UINT32ARRAY":
		return Uint32Array
	case "UINT64":
		return Uint64
	case "UINT64ARRAY":
		return Uint64Array
	case "INT8":
		return Int8
	case "INT8ARRAY":
		return Int8Array
	case "INT16":
		return Int16
	case "INT16ARRAY":
		return Int16Array
	case "INT32":
		return Int32
	case "INT32ARRAY":
		return Int32Array
	case "INT64":
		return Int64
	case "INT64ARRAY":
		return Int64Array
	case "FLOAT32":
		return Float32
	case "FLOAT32ARRAY":
		return Float32Array
	case "FLOAT64":
		return Float64
	case "FLOAT64ARRAY":
		return Float64Array
	case "BINARY":
		return Binary
	default:
		return String
	}
}

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
	Type ValueType
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
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Bool}
	err = encodeValue(cv, value)
	return
}

// NewBoolArrayValue creates a CommandValue of Type BoolArray with the given value.
func NewBoolArrayValue(DeviceResourceName string, origin int64, value []bool) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               BoolArray,
		stringValue:        string(jsonValue),
	}

	return
}

// NewStringValue creates a CommandValue of Type string with the given value.
func NewStringValue(DeviceResourceName string, origin int64, value string) (cv *CommandValue) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: String, stringValue: value}
	return
}

// NewUint8Value creates a CommandValue of Type Uint8 with the given value.
func NewUint8Value(DeviceResourceName string, origin int64, value uint8) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Uint8}
	err = encodeValue(cv, value)
	return
}

// NewUint8ArrayValue creates a CommandValue of Type Uint8Array with the given value.
func NewUint8ArrayValue(DeviceResourceName string, origin int64, value []uint8) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Uint8Array,
		stringValue:        strings.Join(strings.Fields(fmt.Sprintf("%d", value)), ","),
		BinValue:           jsonValue,
	}

	return
}

// NewUint16Value creates a CommandValue of Type Uint16 with the given value.
func NewUint16Value(DeviceResourceName string, origin int64, value uint16) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Uint16}
	err = encodeValue(cv, value)
	return
}

// NewUint16ArrayValue creates a CommandValue of Type Uint16Array with the given value.
func NewUint16ArrayValue(DeviceResourceName string, origin int64, value []uint16) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Uint16Array,
		stringValue:        strings.Join(strings.Fields(fmt.Sprintf("%d", value)), ","),
		BinValue:           jsonValue,
	}

	return
}

// NewUint32Value creates a CommandValue of Type Uint32 with the given value.
func NewUint32Value(DeviceResourceName string, origin int64, value uint32) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Uint32}
	err = encodeValue(cv, value)
	return
}

// NewUint32ArrayValue creates a CommandValue of Type Uint32Array with the given value.
func NewUint32ArrayValue(DeviceResourceName string, origin int64, value []uint32) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Uint32Array,
		stringValue:        strings.Join(strings.Fields(fmt.Sprintf("%d", value)), ","),
		BinValue:           jsonValue,
	}

	return
}

// NewUint64Value creates a CommandValue of Type Uint64 with the given value.
func NewUint64Value(DeviceResourceName string, origin int64, value uint64) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Uint64}
	err = encodeValue(cv, value)
	return
}

// NewUint64ArrayValue creates a CommandValue of Type Uint64Array with the given value.
func NewUint64ArrayValue(DeviceResourceName string, origin int64, value []uint64) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Uint64Array,
		stringValue:        strings.Join(strings.Fields(fmt.Sprintf("%d", value)), ","),
		BinValue:           jsonValue,
	}

	return
}

// NewInt8Value creates a CommandValue of Type Int8 with the given value.
func NewInt8Value(DeviceResourceName string, origin int64, value int8) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Int8}
	err = encodeValue(cv, value)
	return
}

// NewInt8ArrayValue creates a CommandValue of Type Int8Array with the given value.
func NewInt8ArrayValue(DeviceResourceName string, origin int64, value []int8) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Int8Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewInt16Value creates a CommandValue of Type Int16 with the given value.
func NewInt16Value(DeviceResourceName string, origin int64, value int16) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Int16}
	err = encodeValue(cv, value)
	return
}

// NewInt16ArrayValue creates a CommandValue of Type Int16Array with the given value.
func NewInt16ArrayValue(DeviceResourceName string, origin int64, value []int16) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Int16Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewInt32Value creates a CommandValue of Type Int32 with the given value.
func NewInt32Value(DeviceResourceName string, origin int64, value int32) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Int32}
	err = encodeValue(cv, value)
	return
}

// NewInt32ArrayValue creates a CommandValue of Type Int32Array with the given value.
func NewInt32ArrayValue(DeviceResourceName string, origin int64, value []int32) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Int32Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewInt64Value creates a CommandValue of Type Int64 with the given value.
func NewInt64Value(DeviceResourceName string, origin int64, value int64) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Int64}
	err = encodeValue(cv, value)
	return
}

// NewInt64ArrayValue creates a CommandValue of Type Int64Array with the given value.
func NewInt64ArrayValue(DeviceResourceName string, origin int64, value []int64) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Int64Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewFloat32Value creates a CommandValue of Type Float32 with the given value.
func NewFloat32Value(DeviceResourceName string, origin int64, value float32) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Float32}
	err = encodeValue(cv, value)
	return
}

// NewFloat32ArrayValue creates a CommandValue of Type Float32Array with the given value.
func NewFloat32ArrayValue(DeviceResourceName string, origin int64, value []float32) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Float32Array,
		stringValue:        string(jsonValue),
	}

	return
}

// NewFloat64Value creates a CommandValue of Type Float64 with the given value.
func NewFloat64Value(DeviceResourceName string, origin int64, value float64) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Float64}
	err = encodeValue(cv, value)
	return
}

// NewFloat64ArrayValue creates a CommandValue of Type Float64Array with the given value.
func NewFloat64ArrayValue(DeviceResourceName string, origin int64, value []float64) (cv *CommandValue, err error) {
	jsonValue, err := json.Marshal(value)
	cv = &CommandValue{
		DeviceResourceName: DeviceResourceName,
		Origin:             origin,
		Type:               Float64Array,
		stringValue:        string(jsonValue),
	}

	return
}

//NewCommandValue create a CommandValue according to the Type supplied.
func NewCommandValue(DeviceResourceName string, origin int64, value interface{}, t ValueType) (cv *CommandValue, err error) {
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: t}
	switch t {
	case Binary:
		// assign cv.BinValue
		cv.BinValue = value.([]byte)
	case String:
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
	cv = &CommandValue{DeviceResourceName: DeviceResourceName, Origin: origin, Type: Binary, BinValue: value}
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
	if cv.Type == String {
		str = cv.stringValue
		return
	}

	reader := bytes.NewReader(cv.NumericValue)

	switch cv.Type {
	case Bool:
		var res bool
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatBool(res)
	case Uint8:
		var res uint8
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case Uint16:
		var res uint16
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case Uint32:
		var res uint32
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(uint64(res), 10)
	case Uint64:
		var res uint64
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatUint(res, 10)
	case Int8:
		var res int8
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case Int16:
		var res int16
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case Int32:
		var res int32
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(int64(res), 10)
	case Int64:
		var res int64
		err := binary.Read(reader, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}
		str = strconv.FormatInt(res, 10)
	case Float32:
		floatEncoding := getFloatEncoding(encoding)

		if floatEncoding == contract.ENotation {
			var res float32
			binary.Read(reader, binary.BigEndian, &res)
			str = fmt.Sprintf("%e", res)
		} else if floatEncoding == contract.Base64Encoding {
			str = base64.StdEncoding.EncodeToString(cv.NumericValue)
		}
	case Float64:
		floatEncoding := getFloatEncoding(encoding)

		if floatEncoding == contract.ENotation {
			var res float64
			binary.Read(reader, binary.BigEndian, &res)
			str = fmt.Sprintf("%e", res)
		} else if floatEncoding == contract.Base64Encoding {
			str = base64.StdEncoding.EncodeToString(cv.NumericValue)
		}
	case Binary:
		// produce string representation of first 20 bytes of binary value
		str = fmt.Sprintf(fmt.Sprintf("Binary: [%v...]", string(cv.BinValue[:20])))
	default:
		// ArrayType
		str = cv.stringValue
	}

	return
}

// ValueTypeToString returns corresponding string representation of the ValueType.
func (cv *CommandValue) ValueTypeToString() string {
	switch cv.Type {
	case Bool:
		return contract.ValueTypeBool
	case BoolArray:
		return contract.ValueTypeBoolArray
	case String:
		return contract.ValueTypeString
	case Uint8:
		return contract.ValueTypeUint8
	case Uint8Array:
		return contract.ValueTypeUint8Array
	case Uint16:
		return contract.ValueTypeUint16
	case Uint16Array:
		return contract.ValueTypeUint16Array
	case Uint32:
		return contract.ValueTypeUint32
	case Uint32Array:
		return contract.ValueTypeUint32Array
	case Uint64:
		return contract.ValueTypeUint64
	case Uint64Array:
		return contract.ValueTypeUint64Array
	case Int8:
		return contract.ValueTypeInt8
	case Int8Array:
		return contract.ValueTypeInt8Array
	case Int16:
		return contract.ValueTypeInt16
	case Int16Array:
		return contract.ValueTypeInt16Array
	case Int32:
		return contract.ValueTypeInt32
	case Int32Array:
		return contract.ValueTypeInt32Array
	case Int64:
		return contract.ValueTypeInt64
	case Int64Array:
		return contract.ValueTypeInt64Array
	case Float32:
		return contract.ValueTypeFloat32
	case Float32Array:
		return contract.ValueTypeFloat32Array
	case Float64:
		return contract.ValueTypeFloat64
	case Float64Array:
		return contract.ValueTypeFloat64Array
	case Binary:
		return contract.ValueTypeBinary
	default:
		return ""
	}
}

func getFloatEncoding(encoding []string) string {
	if len(encoding) > 0 {
		if encoding[0] == contract.Base64Encoding {
			return contract.Base64Encoding
		} else if encoding[0] == contract.ENotation {
			return contract.ENotation
		}
	}

	return DefaultFloatEncoding
}

// String returns a string representation of a CommandValue instance.
func (cv *CommandValue) String() (str string) {

	originStr := fmt.Sprintf("Origin: %d, ", cv.Origin)

	var typeStr string

	switch cv.Type {
	case Bool:
		typeStr = "Bool: "
	case BoolArray:
		typeStr = "BoolArray: "
	case String:
		typeStr = "String: "
	case Uint8:
		typeStr = "Uint8: "
	case Uint8Array:
		typeStr = "Uint8Array: "
	case Uint16:
		typeStr = "Uint16: "
	case Uint16Array:
		typeStr = "Uint16Array: "
	case Uint32:
		typeStr = "Uint32: "
	case Uint32Array:
		typeStr = "Uint32Array: "
	case Uint64:
		typeStr = "Uint64: "
	case Uint64Array:
		typeStr = "Uint64Array: "
	case Int8:
		typeStr = "Int8: "
	case Int8Array:
		typeStr = "Int8Array: "
	case Int16:
		typeStr = "Int16: "
	case Int16Array:
		typeStr = "Int16Array: "
	case Int32:
		typeStr = "Int32: "
	case Int32Array:
		typeStr = "Int32Array: "
	case Int64:
		typeStr = "Int64: "
	case Int64Array:
		typeStr = "Int64Array: "
	case Float32:
		typeStr = "Float32: "
	case Float32Array:
		typeStr = "Float32Array: "
	case Float64:
		typeStr = "Float64: "
	case Float64Array:
		typeStr = "Float64Array: "
	case Binary:
		typeStr = "Binary: "
	}

	valueStr := typeStr + cv.ValueToString()

	str = originStr + valueStr

	return
}

// BoolValue returns the value in bool data type, and returns error if the Type is not Bool.
func (cv *CommandValue) BoolValue() (bool, error) {
	var value bool
	if cv.Type != Bool {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// BoolArrayValue returns the value in an array of bool type, and returns error if the Type is not BoolArray.
func (cv *CommandValue) BoolArrayValue() ([]bool, error) {
	var value []bool
	if cv.Type != BoolArray {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// StringValue returns the value in string data type, and returns error if the Type is not String.
func (cv *CommandValue) StringValue() (string, error) {
	value := cv.stringValue
	if cv.Type != String {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	return value, nil
}

// Uint8Value returns the value in uint8 data type, and returns error if the Type is not Uint8.
func (cv *CommandValue) Uint8Value() (uint8, error) {
	var value uint8
	if cv.Type != Uint8 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Uint8ArrayValue returns the value in an array of uint8 type, and returns error if the Type is not Uint8Array.
func (cv *CommandValue) Uint8ArrayValue() ([]uint8, error) {
	var value []uint8
	if cv.Type != Uint8Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal(cv.BinValue, &value)
	return value, err
}

// Uint16Value returns the value in uint16 data type, and returns error if the Type is not Uint16.
func (cv *CommandValue) Uint16Value() (uint16, error) {
	var value uint16
	if cv.Type != Uint16 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Uint16ArrayValue returns the value in an array of uint16 type, and returns error if the Type is not Uint16Array.
func (cv *CommandValue) Uint16ArrayValue() ([]uint16, error) {
	var value []uint16
	if cv.Type != Uint16Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal(cv.BinValue, &value)
	return value, err
}

// Uint32Value returns the value in uint32 data type, and returns error if the Type is not Uint32.
func (cv *CommandValue) Uint32Value() (uint32, error) {
	var value uint32
	if cv.Type != Uint32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Uint32ArrayValue returns the value in an array of uint32 type, and returns error if the Type is not Uint32Array.
func (cv *CommandValue) Uint32ArrayValue() ([]uint32, error) {
	var value []uint32
	if cv.Type != Uint32Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal(cv.BinValue, &value)
	return value, err
}

// Uint64Value returns the value in uint64 data type, and returns error if the Type is not Uint64.
func (cv *CommandValue) Uint64Value() (uint64, error) {
	var value uint64
	if cv.Type != Uint64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Uint64ArrayValue returns the value in an array of uint64 type, and returns error if the Type is not Uint64Array.
func (cv *CommandValue) Uint64ArrayValue() ([]uint64, error) {
	var value []uint64
	if cv.Type != Uint64Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal(cv.BinValue, &value)
	return value, err
}

// Int8Value returns the value in int8 data type, and returns error if the Type is not Int8.
func (cv *CommandValue) Int8Value() (int8, error) {
	var value int8
	if cv.Type != Int8 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Int8ArrayValue returns the value in an array of int8 type, and returns error if the Type is not Int8Array.
func (cv *CommandValue) Int8ArrayValue() ([]int8, error) {
	var value []int8
	if cv.Type != Int8Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Int16Value returns the value in int16 data type, and returns error if the Type is not Int16.
func (cv *CommandValue) Int16Value() (int16, error) {
	var value int16
	if cv.Type != Int16 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Int16ArrayValue returns the value in an array of int16 type, and returns error if the Type is not Int16Array.
func (cv *CommandValue) Int16ArrayValue() ([]int16, error) {
	var value []int16
	if cv.Type != Int16Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Int32Value returns the value in int32 data type, and returns error if the Type is not Int32.
func (cv *CommandValue) Int32Value() (int32, error) {
	var value int32
	if cv.Type != Int32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Int32ArrayValue returns the value in an array of int32 type, and returns error if the Type is not Int32Array.
func (cv *CommandValue) Int32ArrayValue() ([]int32, error) {
	var value []int32
	if cv.Type != Int32Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Int64Value returns the value in int64 data type, and returns error if the Type is not Int64.
func (cv *CommandValue) Int64Value() (int64, error) {
	var value int64
	if cv.Type != Int64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Int64ArrayValue returns the value in an array of int64 type, and returns error if the Type is not Int64Array.
func (cv *CommandValue) Int64ArrayValue() ([]int64, error) {
	var value []int64
	if cv.Type != Int64Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Float32Value returns the value in float32 data type, and returns error if the Type is not Float32.
func (cv *CommandValue) Float32Value() (float32, error) {
	var value float32
	if cv.Type != Float32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Float32ArrayValue returns the value in an array of float32 type, and returns error if the Type is not Float32Array.
func (cv *CommandValue) Float32ArrayValue() ([]float32, error) {
	var value []float32
	if cv.Type != Float32Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// Float64Value returns the value in float64 data type, and returns error if the Type is not Float64.
func (cv *CommandValue) Float64Value() (float64, error) {
	var value float64
	if cv.Type != Float64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

// Float64ArrayValue returns the value in an array of float64 type, and returns error if the Type is not Float64Array.
func (cv *CommandValue) Float64ArrayValue() ([]float64, error) {
	var value []float64
	if cv.Type != Float64Array {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := json.Unmarshal([]byte(cv.stringValue), &value)
	return value, err
}

// BinaryValue returns the value in []byte data type, and returns error if the Type is not Binary.
func (cv *CommandValue) BinaryValue() ([]byte, error) {
	var value []byte
	if cv.Type != Binary {
		return value, fmt.Errorf("the CommandValue (%s) data type (%v) is not binary", cv.String(), cv.Type)
	}
	return cv.BinValue, nil
}
