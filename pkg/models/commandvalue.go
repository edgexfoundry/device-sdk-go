// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// ValueType indicates the type of value being passed back
// from a ProtocolDriver instance.
type ValueType int

const (
	// Bool indicates that the value is a bool,
	// stored in CommandValue's boolRes member.
	Bool ValueType = iota
	// String indicates that the value is a string,
	// stored in CommandValue's stringRes member.
	String
	// Uint8 indicates that the value is a uint8 that
	// is stored in CommandValue's NumericRes member.
	Uint8
	// Uint16 indicates that the value is a uint16 that
	// is stored in CommandValue's NumericRes member.
	Uint16
	// Uint32 indicates that the value is a uint32 that
	// is stored in CommandValue's NumericRes member.
	Uint32
	// Uint64 indicates that the value is a uint64 that
	// is stored in CommandValue's NumericRes member.
	Uint64
	// Int8 indicates that the value is a int8 that
	// is stored in CommandValue's NumericRes member.
	Int8
	// Int16 indicates that the value is a int16 that
	// is stored in CommandValue's NumericRes member.
	Int16
	// Int32 indicates that the value is a int32 that
	// is stored in CommandValue's NumericRes member.
	Int32
	// Int64 indicates that the value is a int64 that
	// is stored in CommandValue's NumericRes member.
	Int64
	// Float32 indicates that the value is a float32 that
	// is stored in CommandValue's NumericRes member.
	Float32
	// Float64 indicates that the value is a float64 that
	// is stored in CommandValue's NumericRes member.
	Float64
)

type CommandValue struct {
	// RO is a pointer to the ResourceOperation that triggered the
	// CommandResult to be returned from the ProtocolDriver instance.
	RO *models.ResourceOperation
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
}

func NewBoolValue(ro *models.ResourceOperation, origin int64, value bool) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Bool}
	err = encodeValue(cv, value)
	return
}

func NewStringValue(ro *models.ResourceOperation, origin int64, value string) (cv *CommandValue) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: String, stringValue: value}
	return
}

// NewUint8Value creates a CommandValue of Type Uint8 with the given value.
func NewUint8Value(ro *models.ResourceOperation, origin int64, value uint8) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Uint8}
	err = encodeValue(cv, value)
	return
}

// NewUint16Value creates a CommandValue of Type Uint16 with the given value.
func NewUint16Value(ro *models.ResourceOperation, origin int64, value uint16) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Uint16}
	err = encodeValue(cv, value)
	return
}

// NewUint32Value creates a CommandValue of Type Uint32 with the given value.
func NewUint32Value(ro *models.ResourceOperation, origin int64, value uint32) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Uint32}
	err = encodeValue(cv, value)
	return
}

// NewUint64Value creates a CommandValue of Type Uint64 with the given value.
func NewUint64Value(ro *models.ResourceOperation, origin int64, value uint64) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Uint64}
	err = encodeValue(cv, value)
	return
}

// NewInt8Value creates a CommandValue of Type Int8 with the given value.
func NewInt8Value(ro *models.ResourceOperation, origin int64, value int8) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Int8}
	err = encodeValue(cv, value)
	return
}

// NewInt16Value creates a CommandValue of Type Int16 with the given value.
func NewInt16Value(ro *models.ResourceOperation, origin int64, value int16) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Int16}
	err = encodeValue(cv, value)
	return
}

// NewInt32Value creates a CommandValue of Type Int32 with the given value.
func NewInt32Value(ro *models.ResourceOperation, origin int64, value int32) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Int32}
	err = encodeValue(cv, value)
	return
}

// NewInt64Value creates a CommandValue of Type Int64 with the given value.
func NewInt64Value(ro *models.ResourceOperation, origin int64, value int64) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Int64}
	err = encodeValue(cv, value)
	return
}

// NewFloat32Value creates a CommandValue of Type Float32 with the given value.
func NewFloat32Value(ro *models.ResourceOperation, origin int64, value float32) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Float32}
	err = encodeValue(cv, value)
	return
}

// NewFloat64Value creates a CommandValue of Type Float64 with the given value.
func NewFloat64Value(ro *models.ResourceOperation, origin int64, value float64) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: Float64}
	err = encodeValue(cv, value)
	return
}

//NewCommandValue create a CommandValue according to the Type supplied
func NewCommandValue(ro *models.ResourceOperation, origin int64, value interface{}, t ValueType) (cv *CommandValue, err error) {
	cv = &CommandValue{RO: ro, Origin: origin, Type: t}
	if t != String {
		err = encodeValue(cv, value)
	} else {
		cv.stringValue = value.(string)
	}
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

//ValueToString returns the string format of the value
func (cv *CommandValue) ValueToString() (str string) {
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
		//var res float32
		//binary.Read(reader, binary.BigEndian, &res)
		//str = strconv.FormatFloat(float64(res), 'f', -1, 32)
		str = base64.StdEncoding.EncodeToString(cv.NumericValue)
	case Float64:
		//var res float64
		//binary.Read(reader, binary.BigEndian, &res)
		//str = strconv.FormatFloat(res, 'f', -1, 64)
		str = base64.StdEncoding.EncodeToString(cv.NumericValue)
	}

	return
}

// String returns a string representation of a CommandValue instance.
func (cv *CommandValue) String() (str string) {

	originStr := fmt.Sprintf("Origin: %d, ", cv.Origin)

	var typeStr string

	switch cv.Type {
	case Bool:
		typeStr = "Bool: "
	case String:
		typeStr = "String: "
	case Uint8:
		typeStr = "Uint8: "
	case Uint16:
		typeStr = "Uint16: "
	case Uint32:
		typeStr = "Uint32: "
	case Uint64:
		typeStr = "Uint64: "
	case Int8:
		typeStr = "Int8: "
	case Int16:
		typeStr = "Int16: "
	case Int32:
		typeStr = "Int32: "
	case Int64:
		typeStr = "Int64: "
	case Float32:
		typeStr = "Float32: "
	case Float64:
		typeStr = "Float64: "
	}

	valueStr := typeStr + cv.ValueToString()

	str = originStr + valueStr

	return
}

func (cv *CommandValue) BoolValue() (bool, error) {
	var value bool
	if cv.Type != Bool {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) StringValue() (string, error) {
	value := cv.stringValue
	if cv.Type != String {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	return value, nil
}

func (cv *CommandValue) Uint8Value() (uint8, error) {
	var value uint8
	if cv.Type != Uint8 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Uint16Value() (uint16, error) {
	var value uint16
	if cv.Type != Uint16 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Uint32Value() (uint32, error) {
	var value uint32
	if cv.Type != Uint32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Uint64Value() (uint64, error) {
	var value uint64
	if cv.Type != Uint64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Int8Value() (int8, error) {
	var value int8
	if cv.Type != Int8 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Int16Value() (int16, error) {
	var value int16
	if cv.Type != Int16 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Int32Value() (int32, error) {
	var value int32
	if cv.Type != Int32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Int64Value() (int64, error) {
	var value int64
	if cv.Type != Int64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Float32Value() (float32, error) {
	var value float32
	if cv.Type != Float32 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}

func (cv *CommandValue) Float64Value() (float64, error) {
	var value float64
	if cv.Type != Float64 {
		return value, fmt.Errorf("the data type is not %T", value)
	}
	err := decodeValue(bytes.NewReader(cv.NumericValue), &value)
	return value, err
}
