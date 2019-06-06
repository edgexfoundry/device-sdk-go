// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/ugorji/go/codec"
)

// Test NewBoolValue function.
func TestNewBoolValue(t *testing.T) {
	var value bool
	cv, _ := NewBoolValue("resource", 0, value)
	if cv.Type != Bool {
		t.Errorf("NewBoolValue: invalid Type: %v", cv.Type)
	}
	if value == true {
		t.Errorf("NewBoolValue: invalid value: true")
	}
	v, err := cv.BoolValue()
	if err != nil {
		t.Errorf("NewBoolValue: failed to get bool value")
	}
	if v != value {
		t.Errorf("NewBoolValue: bool value is incorrect")
	}
	if cv.ValueToString() != "false" {
		t.Errorf("NewBoolValue: invalid reading Value: %s", cv.ValueToString())
	}

	value = true
	cv, _ = NewBoolValue("resource", 0, value)
	if cv.Type != Bool {
		t.Errorf("NewBoolValue: invalid Type: %v #2", cv.Type)
	}
	if value == false {
		t.Errorf("NewBoolValue: invalid value: false")
	}
	v, err = cv.BoolValue()
	if err != nil {
		t.Errorf("NewBoolValue: failed to get bool value")
	}
	if v != value {
		t.Errorf("NewBoolValue: bool value is incorrect")
	}
	if cv.ValueToString() != "true" {
		t.Errorf("NewBoolValue: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewStringValue function.
func TestNewStringValue(t *testing.T) {
	var value string
	cv := NewStringValue("resource", 0, value)
	if cv.Type != String {
		t.Errorf("NewStringValue: invalid Type: %v", cv.Type)
	}
	v, err := cv.StringValue()
	if err != nil {
		t.Errorf("NewStringValue: failed to get string value")
	}
	if v != value {
		t.Errorf("NewStringValue: string value is incorrect")
	}

	value = "this is a real string"
	cv = NewStringValue("resource", 0, value)
	if cv.Type != String {
		t.Errorf("NewStringValue: invalid Type: %v #2", cv.Type)
	}
	if value != cv.stringValue {
		t.Errorf("NewStringValue: cv.stringValue: %s doesn't match value: %s", cv.stringValue, value)
	}
	v, err = cv.StringValue()
	if err != nil {
		t.Errorf("NewStringValue: failed to get string value")
	}
	if v != value {
		t.Errorf("NewStringValue: string value is incorrect")
	}
	if cv.ValueToString() != "this is a real string" {
		t.Errorf("NewStringValue: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewUint8Value function.
func TestNewUint8Value(t *testing.T) {
	var value uint8
	cv, _ := NewUint8Value("resource", 0, value)
	if cv.Type != Uint8 {
		t.Errorf("NewUint8Value: invalid Type: %v", cv.Type)
	}
	var res uint8
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint8Value: cv.Uint8Value: %d doesn't match value: %d", value, res)
	}
	v, err := cv.Uint8Value()
	if err != nil {
		t.Errorf("NewUint8Value: failed to get uint8 value")
	}
	if v != value {
		t.Errorf("NewUint8Value: uint8 value is incorrect")
	}

	value = 42
	cv, _ = NewUint8Value("resource", 0, value)
	if cv.Type != Uint8 {
		t.Errorf("NewUint8Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint8Value: cv.Uint8Value: %d doesn't match value: %d (#2)", value, res)
	}
	v, err = cv.Uint8Value()
	if err != nil {
		t.Errorf("NewUint8Value: failed to get uint8 value")
	}
	if v != value {
		t.Errorf("NewUint8Value: uint8 value is incorrect")
	}
	if cv.ValueToString() != "42" {
		t.Errorf("NewUint8Value: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewUint16Value function.
func TestNewUint16Value(t *testing.T) {
	var value uint16
	cv, _ := NewUint16Value("resource", 0, value)
	if cv.Type != Uint16 {
		t.Errorf("NewUint16Value: invalid Type: %v", cv.Type)
	}
	var res uint16
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint16Value: cv.Uint16Value: %d doesn't match value: %d", value, res)
	}
	v, err := cv.Uint16Value()
	if err != nil {
		t.Errorf("NewUint16Value: failed to get uint16 value")
	}
	if v != value {
		t.Errorf("NewUint16Value: uint16 value is incorrect")
	}

	value = 65535
	cv, _ = NewUint16Value("resource", 0, value)
	if cv.Type != Uint16 {
		t.Errorf("NewUint16Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint16Value: cv.Uint16Value: %d doesn't match value: %d (#2)", value, res)
	}
	v, err = cv.Uint16Value()
	if err != nil {
		t.Errorf("NewUint16Value: failed to get uint16 value")
	}
	if v != value {
		t.Errorf("NewUint16Value: uint16 value is incorrect")
	}
	if cv.ValueToString() != "65535" {
		t.Errorf("NewUint16Value: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewUint32Value function.
func TestNewUint32Value(t *testing.T) {
	var value uint32
	cv, _ := NewUint32Value("resource", 0, value)
	if cv.Type != Uint32 {
		t.Errorf("NewUint32Value: invalid Type: %v", cv.Type)
	}
	var res uint32
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint32Value: cv.Uint32Value: %d doesn't match value: %d", value, res)
	}
	v, err := cv.Uint32Value()
	if err != nil {
		t.Errorf("NewUint32Value: failed to get uint32 value")
	}
	if v != value {
		t.Errorf("NewUint32Value: uint32 value is incorrect")
	}

	value = 4294967295
	cv, _ = NewUint32Value("resource", 0, value)
	if cv.Type != Uint32 {
		t.Errorf("NewUint32Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)

	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint32Value: cv.Uint32Value: %d doesn't match value: %d (#2)", value, res)
	}
	v, err = cv.Uint32Value()
	if err != nil {
		t.Errorf("NewUint32Value: failed to get uint32 value")
	}
	if v != value {
		t.Errorf("NewUint32Value: uint32 value is incorrect")
	}
	if cv.ValueToString() != "4294967295" {
		t.Errorf("NewUint32Value: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewUint64Value function.
func TestNewUint64Value(t *testing.T) {
	var value uint64
	var origin int64 = 42
	cv, _ := NewUint64Value("resource", origin, value)
	if cv.Type != Uint64 {
		t.Errorf("NewUint64Value: invalid Type: %v", cv.Type)
	}
	if cv.Origin != origin {
		t.Errorf("NewUint64Value: invalid Origin: %d", cv.Origin)
	}
	var res uint64
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint64Value: cv.Uint64Value: %d doesn't match value: %d", value, res)
	}
	v, err := cv.Uint64Value()
	if err != nil {
		t.Errorf("NewUint64Value: failed to get uint64 value")
	}
	if v != value {
		t.Errorf("NewUint64Value: uint64 value is incorrect")
	}

	value = 18446744073709551615
	cv, _ = NewUint64Value("resource", 0, value)
	if cv.Type != Uint64 {
		t.Errorf("NewUint64Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewUint64Value: cv.Uint64Value: %d doesn't match value: %d (#2)", value, res)
	}
	v, err = cv.Uint64Value()
	if err != nil {
		t.Errorf("NewUint64Value: failed to get uint64 value")
	}
	if v != value {
		t.Errorf("NewUint64Value: uint64 value is incorrect")
	}
	if cv.ValueToString() != "18446744073709551615" {
		t.Errorf("NewUint64Value: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewInt8Value function.
func TestNewInt8Value(t *testing.T) {
	var value int8 = -128
	cv, _ := NewInt8Value("resource", 0, value)
	if cv.Type != Int8 {
		t.Errorf("NewInt8Value: invalid Type: %v", cv.Type)
	}
	var res int8
	buf := bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt8Value: cv.Int8Value: %d doesn't match value: %d", value, res)
	}
	v, err := cv.Int8Value()
	if err != nil {
		t.Errorf("NewInt8Value: failed to get int8 value")
	}
	if v != value {
		t.Errorf("NewInt8Value: int8 value is incorrect")
	}
	if cv.ValueToString() != "-128" {
		t.Errorf("NewInt8Value #1: invalid reading Value: %s", cv.ValueToString())
	}

	value = 127
	cv, _ = NewInt8Value("resource", 0, value)
	if cv.Type != Int8 {
		t.Errorf("NewInt8Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt8Value: cv.Int8Value: %d doesn't match value: %d (#2)", value, res)
	}
	v, err = cv.Int8Value()
	if err != nil {
		t.Errorf("NewInt8Value: failed to get int8 value")
	}
	if v != value {
		t.Errorf("NewInt8Value: int8 value is incorrect")
	}
	if cv.ValueToString() != "127" {
		t.Errorf("NewInt8Value #2: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewInt16Value function.
func TestNewInt16Value(t *testing.T) {
	var value int16 = -32768
	cv, _ := NewInt16Value("resource", 0, value)
	if cv.Type != Int16 {
		t.Errorf("NewInt16Value: invalid Type: %v", cv.Type)
	}
	var res int16
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt16Value: cv.Int16Value: %d doesn't match value: %d", value, res)
	}
	v, err := cv.Int16Value()
	if err != nil {
		t.Errorf("NewInt16Value: failed to get int16 value")
	}
	if v != value {
		t.Errorf("NewInt16Value: int16 value is incorrect")
	}
	if cv.ValueToString() != "-32768" {
		t.Errorf("NewInt16Value #1: invalid reading Value: %s", cv.ValueToString())
	}

	value = 32767
	cv, _ = NewInt16Value("resource", 0, value)
	if cv.Type != Int16 {
		t.Errorf("NewInt16Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt16Value: cv.Int16Value: %d doesn't match value: %d (#2)", value, res)
	}
	v, err = cv.Int16Value()
	if err != nil {
		t.Errorf("NewInt16Value: failed to get int16 value")
	}
	if v != value {
		t.Errorf("NewInt16Value: int16 value is incorrect")
	}
	if cv.ValueToString() != "32767" {
		t.Errorf("NewInt16Value #2: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewInt32Value function.
func TestNewInt32Value(t *testing.T) {
	var value int32 = -2147483648
	cv, _ := NewInt32Value("resource", 0, value)
	if cv.Type != Int32 {
		t.Errorf("NewInt32Value: invalid Type: %v", cv.Type)
	}
	var res int32
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt32Value: cv.Int32Value: %d doesn't match value: %d", value, res)
	}
	v, err := cv.Int32Value()
	if err != nil {
		t.Errorf("NewInt32Value: failed to get int32 value")
	}
	if v != value {
		t.Errorf("NewInt32Value: int32 value is incorrect")
	}
	if cv.ValueToString() != "-2147483648" {
		t.Errorf("NewInt32Value #1: invalid reading Value: %s", cv.ValueToString())
	}

	value = 2147483647
	cv, _ = NewInt32Value("resource", 0, value)
	if cv.Type != Int32 {
		t.Errorf("NewInt32Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt32Value: cv.Int32Value: %d doesn't match value: %d (#2)", value, res)
	}
	v, err = cv.Int32Value()
	if err != nil {
		t.Errorf("NewInt32Value: failed to get int32 value")
	}
	if v != value {
		t.Errorf("NewInt32Value: int32 value is incorrect")
	}
	if cv.ValueToString() != "2147483647" {
		t.Errorf("NewInt32Value #2: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewInt64Value function.
func TestNewInt64Value(t *testing.T) {
	var value int64 = -9223372036854775808
	var origin int64 = 42
	cv, _ := NewInt64Value("resource", origin, value)
	if cv.Type != Int64 {
		t.Errorf("NewInt64Value: invalid Type: %v", cv.Type)
	}
	if cv.Origin != origin {
		t.Errorf("NewInt64Value: invalid Origin: %d", cv.Origin)
	}
	var res int64
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt64Value: cv.Int64Value: %d doesn't match value: %d", value, res)
	}
	v, err := cv.Int64Value()
	if err != nil {
		t.Errorf("NewInt64Value: failed to get int64 value")
	}
	if v != value {
		t.Errorf("NewInt64Value: int64 value is incorrect")
	}
	if cv.ValueToString() != "-9223372036854775808" {
		t.Errorf("NewInt64Value #1: invalid reading Value: %s", cv.ValueToString())
	}

	value = 9223372036854775807
	cv, _ = NewInt64Value("resource", 0, value)
	if cv.Type != Int64 {
		t.Errorf("NewInt64Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewInt64Value: cv.Int64Value: %d doesn't match value: %d (#2)", value, res)
	}
	v, err = cv.Int64Value()
	if err != nil {
		t.Errorf("NewInt64Value: failed to get int64 value")
	}
	if v != value {
		t.Errorf("NewInt64Value: int64 value is incorrect")
	}
	if cv.ValueToString() != "9223372036854775807" {
		t.Errorf("NewInt64Value #2: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewFloat32Value function.
func TestNewFloat32Value(t *testing.T) {
	var value float32 = math.SmallestNonzeroFloat32
	var origin int64 = time.Now().UnixNano()
	cv, _ := NewFloat32Value("resource", origin, value)
	if cv.Type != Float32 {
		t.Errorf("NewFloat32Value: invalid Type: %v", cv.Type)
	}
	if cv.Origin != origin {
		t.Errorf("NewFloat32Value: invalid Origin: %d", cv.Origin)
	}
	var res float32
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewFloat32Value: cv.Int64Value: %v doesn't match value: %v", value, res)
	}
	v, err := cv.Float32Value()
	if err != nil {
		t.Errorf("NewFloat32Value: failed to get float32 value")
	}
	if v != value {
		t.Errorf("NewFloat32Value: float32 value is incorrect")
	}
	if cv.ValueToString(contract.Base64Encoding) != "AAAAAQ==" {
		t.Errorf("NewFloat32Value #1: invalid reading Value: %s", cv.ValueToString())
	}

	value = math.MaxFloat32
	cv, _ = NewFloat32Value("resource", 0, value)
	if cv.Type != Float32 {
		t.Errorf("NewFloat32Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewFloat32Value: cv.Float32Value: %v doesn't match value: %v (#2)", value, res)
	}
	v, err = cv.Float32Value()
	if err != nil {
		t.Errorf("NewFloat32Value: failed to get float32 value")
	}
	if v != value {
		t.Errorf("NewFloat32Value: float32 value is incorrect")
	}
	if cv.ValueToString(contract.Base64Encoding) != "f3///w==" {
		t.Errorf("NewFloat32Value #2: invalid reading Value: %s", cv.ValueToString())
	}
}

// Test NewFloat64Value function.
func TestNewFloat64Value(t *testing.T) {
	var value float64 = math.SmallestNonzeroFloat64
	var origin int64 = time.Now().UnixNano()
	cv, _ := NewFloat64Value("resource", origin, value)
	if cv.Type != Float64 {
		t.Errorf("NewFloat64Value: invalid Type: %v", cv.Type)
	}
	if cv.Origin != origin {
		t.Errorf("NewFloat64Value: invalid Origin: %d", cv.Origin)
	}
	var res float64
	buf := bytes.NewReader(cv.NumericValue)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewFloat64Value: cv.Int64Value: %v doesn't match value: %v", value, res)
	}
	v, err := cv.Float64Value()
	if err != nil {
		t.Errorf("NewFloat64Value: failed to get float64 value")
	}
	if v != value {
		t.Errorf("NewFloat64Value: float64 value is incorrect")
	}
	if cv.ValueToString(contract.Base64Encoding) != "AAAAAAAAAAE=" {
		t.Errorf("NewFloat64Value #1: invalid reading Value: %s", cv.ValueToString())
	}

	value = math.MaxFloat64
	cv, _ = NewFloat64Value("resource", 0, value)
	if cv.Type != Float64 {
		t.Errorf("NewFloat64Value: invalid Type: %v #3", cv.Type)
	}
	buf = bytes.NewReader(cv.NumericValue)
	fmt.Printf("cv: %v\n", cv)
	binary.Read(buf, binary.BigEndian, &res)
	if value != res {
		t.Errorf("NewFloat64Value: cv.Float64Value: %v doesn't match value: %v (#2)", value, res)
	}
	v, err = cv.Float64Value()
	if err != nil {
		t.Errorf("NewFloat64Value: failed to get float64 value")
	}
	if v != value {
		t.Errorf("NewFloat64Value: float64 value is incorrect")
	}
	if cv.ValueToString(contract.Base64Encoding) != "f+////////8=" {
		t.Errorf("NewFloat64Value #2: invalid reading Value: %s", cv.ValueToString())
	}
}

// encodeMockEvent accepts a contract.Event and returns a CBOR encoded byte array
func encodeMockEvent(e contract.Event) ([]byte, error) {
	var handle codec.CborHandle
	var byteBuffer = make([]byte, 0, 64)
	enc := codec.NewEncoderBytes(&byteBuffer, &handle)
	err := enc.Encode(e)
	if err != nil {
		return []byte{}, err
	}
	return byteBuffer, nil
}

func createMockPayload(size int) ([]byte, error) {
	token := make([]byte, size)
	_, err := rand.Read(token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Test NewBinaryValue function and associated methods for binary encode/decode.
func TestNewBinaryValue(t *testing.T) {
	var origin int64 = time.Now().UnixNano()
	// assign instance of mockStructB as a CBOR encoded CommandValue payload
	var mock1 contract.Event
	var mock2 contract.Event
	mock1.Device = "Device01234567890"
	mock1.Created = origin
	mock1.ID = "MyStringIdentifier"
	mock1.Modified = origin + 123
	mock1.Pushed = 12345
	// To extend coverage cborMock1 becomes encoded byte array.
	// We will then confirm CommandValue particulars of binary payload are valid
	cborMock1, err := encodeMockEvent(mock1)
	if err != nil {
		t.Errorf("NewBinaryValue: Error encoding struct as binary value")
	}
	cv, errAssign := NewBinaryValue("resource", origin, cborMock1)
	if errAssign != nil {
		t.Errorf("NewBinaryValue: Error invoking NewBinaryValue [%v]", errAssign)
	}
	// Confirm CommandValue particulars
	if cv.Type != Binary {
		t.Errorf("Expected Binary type! invalid Type: %v", cv.Type)
	}
	if cv.Origin != origin {
		t.Errorf("Expected matching value! invalid Origin: %d != %d", cv.Origin, origin)
	}
	// Populate cv.BinValue as CBOR encoded CommandValue and decode back
	err1 := encodeBinaryValue(cv, mock1)
	if err1 != nil {
		t.Errorf("NewBinaryValue: Error encoding binary value")
	}
	// decode the original encoded value, compare original raw struct to decoded mock2 and mock3
	mock2 = contract.Event{}
	err2 := decodeBinaryValue(bytes.NewReader(cborMock1), &mock2)
	if err2 != nil {
		t.Errorf("NewBinaryValue: Error decoding binary value")
	}
	if !reflect.DeepEqual(mock1, mock2) {
		t.Errorf("NewBinaryValue: cv.BinValue: %v doesn't match value: %v", mock2, mock1)
	}

	mock3 := contract.Event{}
	err3 := decodeBinaryValue(bytes.NewReader(cv.BinValue), &mock3)
	if err3 != nil {
		t.Errorf("NewBinaryValue: Error decoding binary value")
	}
	if !reflect.DeepEqual(mock2, mock3) {
		t.Errorf("NewBinaryValue: cv.BinValue: %v doesn't match value: %v", mock2, mock3)
	}
}

// Test NewBinaryValueConstraints function and associated methods for binary encode/decode.
func TestNewBinaryValueWithinConstraints(t *testing.T) {
	var origin int64 = time.Now().UnixNano()
	// Confirm we receive error if arbitrary binary value exceeds policy limit (currently 16MB)
	payloadLimit := MaxBinaryBytes
	payloadSize := MaxBinaryBytes
	mock4, errCreate4 := createMockPayload(payloadSize)
	if errCreate4 != nil {
		t.Errorf("NewBinaryValue: createMockPayload: Could not create payload size: %v", payloadSize)
	}
	if binary.Size(mock4) != payloadSize {
		t.Errorf("NewBinaryValue: createMockPayload: Could not create requested payload size: %v != %v", payloadSize, binary.Size(mock4))
	}
	cv2, errAssign2 := NewBinaryValue("resource", origin, mock4)
	if errAssign2 == nil {
		// Requested CommandValue payload within limit for binary readings (16777216 bytes)
		if cv2 == nil {
			t.Errorf("NewBinaryValue: Empty CommandValue returned by NewBinaryValue!")
		} else if !bytes.Equal(cv2.BinValue, mock4) {
			t.Errorf("NewBinaryValue: Value assigned differs from value read!")
		}
		// PASS
	} else {
		t.Errorf("NewBinaryValue: Error occurred when assigning valid payload size! bytes [%v] > limit [%v] - %v", payloadSize, payloadLimit, errAssign2)
	}
}

// Test NewBinaryValueConstraints function and associated methods for binary encode/decode.
func TestNewBinaryValueExceedsConstraints(t *testing.T) {
	var origin int64 = time.Now().UnixNano()
	// Confirm we receive error if arbitrary binary value exceeds policy limit (currently 16MB)
	payloadLimit := MaxBinaryBytes
	payloadSize := MaxBinaryBytes + 1
	mock4, errCreate4 := createMockPayload(payloadSize)
	if errCreate4 != nil {
		t.Errorf("NewBinaryValue: createMockPayload: Could not create payload size: %v", payloadSize)
	}
	if binary.Size(mock4) != payloadSize {
		t.Errorf("NewBinaryValue: createMockPayload: Could not create requested payload size: %v != %v", payloadSize, binary.Size(mock4))
	}
	cv2, errAssign2 := NewBinaryValue("resource", origin, mock4)
	if errAssign2 == nil {
		t.Errorf("NewBinaryValue: Accepted invalid payload size! bytes [%v] > limit [%v]", payloadSize, payloadLimit)
		// FAIL
	} else {
		// Requested CommandValue payload exceeds limit for binary readings (16777216 bytes)
		if cv2 != nil {
			t.Errorf("NewBinaryValue: Empty CommandValue returned by NewBinaryValue!")
		}
		// PASS
	}
}
