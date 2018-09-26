// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// ResultType indicates the type of result being passed back
// from a ProtocolDriver instance.
type ResultType int

const (
	// Bool indicates that the result is a bool,
	// stored in CommandResult's boolRes member.
	Bool ResultType = iota
	// String indicates that the result is a string,
	// stored in CommandResult's stringRes member.
	String
	// Uint8 indicates that the result is a uint8 that
	// is stored in CommandResult's NumericRes member.
	Uint8
	// Uint16 indicates that the result is a uint16 that
	// is stored in CommandResult's NumericRes member.
	Uint16
	// Uint32 indicates that the result is a uint32 that
	// is stored in CommandResult's NumericRes member.
	Uint32
	// Uint64 indicates that the result is a uint64 that
	// is stored in CommandResult's NumericRes member.
	Uint64
	// Int8 indicates that the result is a int8 that
	// is stored in CommandResult's NumericRes member.
	Int8
	// Int16 indicates that the result is a int16 that
	// is stored in CommandResult's NumericRes member.
	Int16
	// Int32 indicates that the result is a int32 that
	// is stored in CommandResult's NumericRes member.
	Int32
	// Int64 indicates that the result is a int64 that
	// is stored in CommandResult's NumericRes member.
	Int64
	// Float32 indicates that the result is a float32 that
	// is stored in CommandResult's NumericRes member.
	Float32
	// Float64 indicates that the result is a float64 that
	// is stored in CommandResult's NumericRes member.
	Float64
)

type CommandResult struct {
	// DeviceId identifies the device that produced this result.
	DeviceId string
	// DeviceName identifies the device that produced this result.
	DeviceName string
	// RO is a pointer to the ResourceOperation that triggered the
	// CommandResult to be returned from the ProtocolDriver instance.
	RO *models.ResourceOperation
	// VDR is a pointer to the associated ValueDescriptor.
	VDR *models.ValueDescriptor
	// Origin is an int64 value which indicates the time the reading
	// contained in the CommandResult was read by the ProtocolDriver
	// instance.
	Origin int64
	// Type is a ResultType value which indicates what type of
	// result was returned from the ProtocolDriver instance in
	// response to HandleCommand being called to handle a single
	// ResourceOperation.
	Type ResultType
	// NumericResult is a byte slice with a maximum capacity of
	// 64 bytes, used to hold a numeric result returned by a
	// ProtocolDriver instance. The value can be converted to
	// its native type by referring to the the value of ResType.
	NumericResult []byte
	// StringResult is a string value returned as a result by a ProtocolDriver instance.
	StringResult string
	// BoolResult is a bool value returned as a result by a ProtocolDriver instance.
	BoolResult bool
}

func NewBoolResult(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value bool) (cr *CommandResult) {

	// TODO: modify CommandResult to just store bool in NumericValue too...

	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Bool, BoolResult: value}

	fmt.Printf("result: %v\n", cr)
	return
}

func NewStringResult(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value string) (cr *CommandResult) {
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: String, StringResult: value}

	fmt.Printf("result: %v\n", cr)
	return
}

// TODO: re-factor to get rid of duplicate code

// NewUint8Result creates a CommandResult of Type Uint8 with the given value.
func NewUint8Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value uint8) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Uint8}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewUint16Result creates a CommandResult of Type Uint16 with the given value.
func NewUint16Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value uint16) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Uint16}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewUint32Result creates a CommandResult of Type Uint32 with the given value.
func NewUint32Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value uint32) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Uint32}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewUint64Result creates a CommandResult of Type Uint64 with the given value.
func NewUint64Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value uint64) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Uint64}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewInt8Result creates a CommandResult of Type Int8 with the given value.
func NewInt8Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value int8) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Int8}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewInt16Result creates a CommandResult of Type Int16 with the given value.
func NewInt16Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value int16) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Int16}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewInt32Result creates a CommandResult of Type Int32 with the given value.
func NewInt32Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value int32) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Int32}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewInt64Result creates a CommandResult of Type Int64 with the given value.
func NewInt64Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value int64) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Int64}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewFloat32Result creates a CommandResult of Type Float32 with the given value.
func NewFloat32Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value float32) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Float32}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// NewFloat64Result creates a CommandResult of Type Float64 with the given value.
func NewFloat64Result(ro *models.ResourceOperation, vdr *models.ValueDescriptor, origin int64, value float64) (cr *CommandResult) {
	buf := new(bytes.Buffer)
	cr = &CommandResult{RO: ro, VDR: vdr, Origin: origin, Type: Float64}
	err := binary.Write(buf, binary.BigEndian, value)
	if err != nil {
		fmt.Printf("binary.Write failed: %v\n", err)
	}

	cr.NumericResult = buf.Bytes()
	return
}

// TODO: add Get funcs for all basic types

// Transform applies any transform attributes contained in the
// CommandResult's ValueDescriptor to the result. If the transform
// operations result in an overflow for the specified type, an
// error is returned.
func (cr *CommandResult) Transform() error {
	return nil
}

// Reading returns a new Reading instance created from the the given CommandResult.
func (cr *CommandResult) Reading(devName string, doName string) *models.Reading {

	reading := &models.Reading{Name: doName, Device: devName}
	reading.Value = cr.toString()

	// if result has a non-zero Origin, use it
	if cr.Origin > 0 {
		reading.Origin = cr.Origin
	} else {
		t := time.Now()
		reading.Origin = t.Unix()
	}

	return reading
}

// String returns a string representation of a CommandResult instance.
func (cr *CommandResult) toString() (str string) {
	if cr.Type == Bool {

		if cr.BoolResult == true {
			str = "true"
		} else {
			str = "false"
		}

		return
	}

	if cr.Type == String {
		str = cr.StringResult
		return
	}

	buf := bytes.NewReader(cr.NumericResult)

	switch cr.Type {
	case Uint8:
		var res uint8
		err := binary.Read(buf, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}

		str = fmt.Sprintf("%d", res)
	case Uint16:
		var res uint16
		err := binary.Read(buf, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}

		str = fmt.Sprintf("%d", res)
	case Uint32:
		var res uint32
		err := binary.Read(buf, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}

		str = fmt.Sprintf("%d", res)
	case Uint64:
		var res uint64
		err := binary.Read(buf, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}

		str = fmt.Sprintf("%d", res)
	case Int8:
		var res int8
		err := binary.Read(buf, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}

		str = fmt.Sprintf("%d", res)
	case Int16:
		var res int16
		err := binary.Read(buf, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}

		str = fmt.Sprintf("%d", res)
	case Int32:
		var res int32
		err := binary.Read(buf, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}

		str = fmt.Sprintf("%d", res)
	case Int64:
		var res int64
		err := binary.Read(buf, binary.BigEndian, &res)
		if err != nil {
			str = err.Error()
		}

		str = fmt.Sprintf("%d", res)

	// TODO: implement base64 encoding of float results
	case Float32:
		var res float32
		str = fmt.Sprintf("%f", binary.Read(buf, binary.BigEndian, &res))
	case Float64:
		var res float64
		str = fmt.Sprintf("%f", binary.Read(buf, binary.BigEndian, &res))
	}

	return
}

// String returns a string representation of a CommandResult instance.
func (cr *CommandResult) String() (str string) {

	roStr := fmt.Sprintf("%v\n", cr.RO)
	vdrStr := fmt.Sprintf("%v\n", cr.VDR)
	originStr := fmt.Sprintf("%d\n", cr.Origin)

	var typeStr string

	switch cr.Type {
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

	resultStr := typeStr + cr.toString()

	str = roStr + vdrStr + originStr + resultStr

	return
}

// TransformResult applies transforms specified in the given
// PropertyValue instance.
func (cr *CommandResult) TransformResult(models.PropertyValue) bool {

	// TODO: implement base, scale, offset & assertions
	return true
}
