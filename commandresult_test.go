// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

// Test NewBoolResult function.
func TestNewBoolResult(t *testing.T) {
	var result bool

	cr := NewBoolResult(nil, nil, 0, result)

	if cr.Type != Bool {
		t.Errorf("NewBoolResult: invalid Type: %v", cr.Type)
	}

	if result == true {
		t.Errorf("NewBoolResult: invalid result: true")
	}

	result = true

	cr = NewBoolResult(nil, nil, 0, result)

	if cr.Type != Bool {
		t.Errorf("NewBoolResult: invalid Type: %v #2", cr.Type)
	}

	if result == false {
		t.Errorf("NewBoolResult: invalid result: false")
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("bool reading: %v\n", reading)
}

// Test NewStringResult function.
func TestNewStringResult(t *testing.T) {
	var result string

	cr := NewStringResult(nil, nil, 0, result)
	if cr.Type != String {
		t.Errorf("NewStringResult: invalid Type: %v", cr.Type)
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("string reading: %v\n", reading)

	result = "this is a real string"
	cr = NewStringResult(nil, nil, 0, result)
	if cr.Type != String {
		t.Errorf("NewStringResult: invalid Type: %v #2", cr.Type)
	}

	if result != cr.StringResult {
		t.Errorf("NewStringResult: cr.StringResult: %s doesn't match result: %s", cr.StringResult, result)
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("string reading #2: %v\n", reading)
}

// Test NewUint8Result function.
func TestNewUint8Result(t *testing.T) {
	var result uint8

	cr := NewUint8Result(nil, nil, 0, result)
	if cr.Type != Uint8 {
		t.Errorf("NewUint8Result: invalid Type: %v", cr.Type)
	}

	var res uint8
	buf := bytes.NewReader(cr.NumericResult)
	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewUint8Result: cr.Uint8Result: %d doesn't match result: %d", result, res)
	}

	reading := cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("uint8 reading: %v\n", reading)

	result = 42
	cr = NewUint8Result(nil, nil, 0, result)
	if cr.Type != Uint8 {
		t.Errorf("NewUint8Result: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewUint8Result: cr.Uint8Result: %d doesn't match result: %d (#2)", result, res)
	}

	reading = cr.Reading("FakeDevice", "FakeDeviceObject")
	fmt.Printf("uint8 reading #2: %v\n", reading)
}

// Test NewUint16Result function.
func TestNewUint16Result(t *testing.T) {
	var result uint16

	cr := NewUint16Result(nil, nil, 0, result)
	if cr.Type != Uint16 {
		t.Errorf("NewUint16Result: invalid Type: %v", cr.Type)
	}

	var res uint16
	buf := bytes.NewReader(cr.NumericResult)
	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewUint16Result: cr.Uint16Result: %d doesn't match result: %d", result, res)
	}

	result = 65535
	cr = NewUint16Result(nil, nil, 0, result)
	if cr.Type != Uint16 {
		t.Errorf("NewUint16Result: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewUint16Result: cr.Uint16Result: %d doesn't match result: %d (#2)", result, res)
	}
}

// Test NewUint32Result function.
func TestNewUint32Result(t *testing.T) {
	var result uint32

	cr := NewUint32Result(nil, nil, 0, result)
	if cr.Type != Uint32 {
		t.Errorf("NewUint32Result: invalid Type: %v", cr.Type)
	}

	var res uint32
	buf := bytes.NewReader(cr.NumericResult)
	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewUint32Result: cr.Uint32Result: %d doesn't match result: %d", result, res)
	}

	result = 4294967295
	cr = NewUint32Result(nil, nil, 0, result)
	if cr.Type != Uint32 {
		t.Errorf("NewUint32Result: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewUint32Result: cr.Uint32Result: %d doesn't match result: %d (#2)", result, res)
	}
}

// Test NewUint64Result function.
func TestNewUint64Result(t *testing.T) {
	var result uint64
	var origin int64 = 42

	cr := NewUint64Result(nil, nil, origin, result)
	if cr.Type != Uint64 {
		t.Errorf("NewUint64Result: invalid Type: %v", cr.Type)
	}

	if cr.Origin != origin {
		t.Errorf("NewUint64Result: invalid Origin: %d", cr.Origin)
	}

	var res uint64
	buf := bytes.NewReader(cr.NumericResult)
	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewUint64Result: cr.Uint64Result: %d doesn't match result: %d", result, res)
	}

	result = 18446744073709551615

	cr = NewUint64Result(nil, nil, 0, result)
	if cr.Type != Uint64 {
		t.Errorf("NewUint64Result: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewUint64Result: cr.Uint64Result: %d doesn't match result: %d (#2)", result, res)
	}
}

// Test NewInt8Result function.
func TestNewInt8Result(t *testing.T) {
	var result int8 = -128

	cr := NewInt8Result(nil, nil, 0, result)
	if cr.Type != Int8 {
		t.Errorf("NewInt8Result: invalid Type: %v", cr.Type)
	}

	var res int8
	buf := bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewInt8Result: cr.Int8Result: %d doesn't match result: %d", result, res)
	}

	result = 127
	cr = NewInt8Result(nil, nil, 0, result)
	if cr.Type != Int8 {
		t.Errorf("NewInt8Result: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewInt8Result: cr.Int8Result: %d doesn't match result: %d (#2)", result, res)
	}
}

// Test NewInt16Result function.
func TestNewInt16Result(t *testing.T) {
	var result int16 = -32768

	cr := NewInt16Result(nil, nil, 0, result)
	if cr.Type != Int16 {
		t.Errorf("NewInt16Result: invalid Type: %v", cr.Type)
	}

	var res int16
	buf := bytes.NewReader(cr.NumericResult)
	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewInt16Result: cr.Int16Result: %d doesn't match result: %d", result, res)
	}

	result = 32767
	cr = NewInt16Result(nil, nil, 0, result)
	if cr.Type != Int16 {
		t.Errorf("NewInt16Result: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewInt16Result: cr.Int16Result: %d doesn't match result: %d (#2)", result, res)
	}
}

// Test NewInt32Result function.
func TestNewInt32Result(t *testing.T) {
	var result int32 = -2147483648

	cr := NewInt32Result(nil, nil, 0, result)
	if cr.Type != Int32 {
		t.Errorf("NewInt32Result: invalid Type: %v", cr.Type)
	}

	var res int32
	buf := bytes.NewReader(cr.NumericResult)
	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewInt32Result: cr.Int32Result: %d doesn't match result: %d", result, res)
	}

	result = 2147483647
	cr = NewInt32Result(nil, nil, 0, result)
	if cr.Type != Int32 {
		t.Errorf("NewInt32Result: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewInt32Result: cr.Int32Result: %d doesn't match result: %d (#2)", result, res)
	}
}

// Test NewInt64Result function.
func TestNewInt64Result(t *testing.T) {
	var result int64 = -9223372036854775808
	var origin int64 = 42

	cr := NewInt64Result(nil, nil, origin, result)
	if cr.Type != Int64 {
		t.Errorf("NewInt64Result: invalid Type: %v", cr.Type)
	}

	if cr.Origin != origin {
		t.Errorf("NewInt64Result: invalid Origin: %d", cr.Origin)
	}

	var res int64
	buf := bytes.NewReader(cr.NumericResult)
	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewInt64Result: cr.Int64Result: %d doesn't match result: %d", result, res)
	}

	result = 9223372036854775807

	cr = NewInt64Result(nil, nil, 0, result)
	if cr.Type != Int64 {
		t.Errorf("NewInt64Result: invalid Type: %v #3", cr.Type)
	}

	buf = bytes.NewReader(cr.NumericResult)
	fmt.Printf("cr: %v\n", cr)

	binary.Read(buf, binary.BigEndian, &res)
	if result != res {
		t.Errorf("NewInt64Result: cr.Int64Result: %d doesn't match result: %d (#2)", result, res)
	}
}
