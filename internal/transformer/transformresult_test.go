// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"errors"
	"fmt"
	"math"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var lc logger.LoggingClient

func init() {
	lc = logger.NewClientStdOut("device-sdk-test", false, "DEBUG")
}

func TestTransformReadResult_base_unt8(t *testing.T) {
	val := uint8(2)
	base := "10"
	expected := uint8(100)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint8)
	}
}

func TestTransformReadResult_base_unt8_overflow(t *testing.T) {
	val := uint8(10)
	base := "3"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_unt8(t *testing.T) {
	val := uint8(math.MaxUint8 / 5)
	scale := "5"
	expected := uint8(math.MaxUint8)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v", result, expected)
	}
	if cv.Type != dsModels.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint8)
	}
}

func TestTransformReadResult_scale_unt8_overflow(t *testing.T) {
	val := uint8(math.MaxUint8 / 5)
	scale := "6"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_unt8(t *testing.T) {
	val := uint8(math.MaxUint8 - 1)
	offset := "1"
	expected := uint8(math.MaxUint8)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint8)
	}
}

func TestTransformReadResult_offset_unt8_overflow(t *testing.T) {
	val := uint8(math.MaxUint8)
	offset := "1"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_unt16(t *testing.T) {
	val := uint16(2)
	base := "200"
	expected := uint16(40000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint16)
	}
}

func TestTransformReadResult_base_uint16_overflow(t *testing.T) {
	val := uint16(200)
	base := "3"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_uint16(t *testing.T) {
	val := uint16(math.MaxUint16 / 5)
	scale := "5"
	expected := uint16(math.MaxUint16)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint16)
	}
}

func TestTransformReadResult_scale_uint16_overflow(t *testing.T) {
	val := uint16(math.MaxUint16 / 5)
	scale := "6"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_uint16(t *testing.T) {
	val := uint16(math.MaxUint16 - 1)
	offset := "1"
	expected := uint16(math.MaxUint16)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint16)
	}
}

func TestTransformReadResult_offset_uint16_overflow(t *testing.T) {
	val := uint16(math.MaxUint16)
	offset := "1"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_uint32(t *testing.T) {
	val := uint32(2)
	base := "20000"
	expected := uint32(400000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint32)
	}
}

func TestTransformReadResult_base_uint32_overflow(t *testing.T) {
	val := uint32(4000000)
	base := "1000"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_uint32(t *testing.T) {
	val := uint32(math.MaxUint32 / 5)
	scale := "5"
	expected := uint32(math.MaxUint32)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint32)
	}
}

func TestTransformReadResult_scale_uint32_overflow(t *testing.T) {
	val := uint32(math.MaxUint32 / 5)
	scale := "6"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_uint32(t *testing.T) {
	val := uint32(math.MaxUint32 - 1)
	offset := "1"
	expected := uint32(math.MaxUint32)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint32)
	}
}

func TestTransformReadResult_offset_uint32_overflow(t *testing.T) {
	val := uint32(math.MaxUint32)
	offset := "1"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_uint64(t *testing.T) {
	val := uint64(2)
	base := "20000"
	expected := uint64(400000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint64)
	}
}

func TestTransformReadResult_scale_uint64(t *testing.T) {
	val := uint64(20000)
	scale := "20000"
	expected := uint64(400000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint64)
	}
}

func TestTransformReadResult_offset_uint64(t *testing.T) {
	val := uint64(math.MaxUint64) - uint64(1)
	offset := "1"
	expected := uint64(math.MaxUint64)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint64)
	}
}

func TestTransformReadResult_base_int8(t *testing.T) {
	val := int8(2)
	base := "10"
	expected := int8(100)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int8)
	}
}

func TestTransformReadResult_base_int8_overflow(t *testing.T) {
	val := int8(10)
	base := "3"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_int8(t *testing.T) {
	val := int8(10)
	scale := "10"
	expected := int8(100)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v", result, expected)
	}
	if cv.Type != dsModels.Int8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int8)
	}
}

func TestTransformReadResult_scale_int8_overflow(t *testing.T) {
	val := uint8(10)
	scale := "30"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_int8(t *testing.T) {
	val := int8(math.MaxInt8 - 1)
	offset := "1"
	expected := int8(math.MaxInt8)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int8)
	}
}

func TestTransformReadResult_offset_int8_overflow(t *testing.T) {
	val := int8(math.MaxInt8)
	offset := "1"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_int16(t *testing.T) {
	val := int16(2)
	base := "100"
	expected := int16(10000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int16)
	}
}

func TestTransformReadResult_base_int16_overflow(t *testing.T) {
	val := int16(100)
	base := "3"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_int16(t *testing.T) {
	val := int16(10000)
	scale := "3"
	expected := int16(30000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int16)
	}
}

func TestTransformReadResult_scale_int16_overflow(t *testing.T) {
	val := int16(10000)
	scale := "4"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_int16(t *testing.T) {
	val := int16(math.MaxInt16 - 1)
	offset := "1"
	expected := int16(math.MaxInt16)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int16)
	}
}

func TestTransformReadResult_offset_int16_overflow(t *testing.T) {
	val := int16(math.MaxInt16)
	offset := "1"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_int32(t *testing.T) {
	val := int32(2)
	base := "20000"
	expected := int32(400000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int32)
	}
}

func TestTransformReadResult_base_int32_overflow(t *testing.T) {
	val := int32(20000)
	base := "3"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_int32(t *testing.T) {
	val := int32(200000000)
	scale := "10"
	expected := int32(2000000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int32)
	}
}

func TestTransformReadResult_scale_int32_overflow(t *testing.T) {
	val := int32(200000000)
	scale := "15"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_int32(t *testing.T) {
	val := int32(math.MaxInt32 - 1)
	offset := "1"
	expected := int32(math.MaxInt32)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int32)
	}
}

func TestTransformReadResult_offset_int32_overflow(t *testing.T) {
	val := int32(math.MaxInt32)
	offset := "1"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_int64(t *testing.T) {
	val := int64(2)
	base := "20000"
	expected := int64(400000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int64)
	}
}

func TestTransformReadResult_scale_int64(t *testing.T) {
	val := int64(20000)
	scale := "20000"
	expected := int64(400000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int64)
	}
}

func TestTransformReadResult_offset_int64(t *testing.T) {
	val := int64(math.MaxInt64) - int64(1)
	offset := "1"
	expected := int64(math.MaxInt64)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewInt64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Int64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Int64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Int64)
	}
}

func TestTransformReadResult_base_float32(t *testing.T) {
	val := float32(1.1)
	base := "2"
	expected := float32(2.143547)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Float32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Float32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Float32)
	}
}

func TestTransformReadResult_base_float32_overflow(t *testing.T) {
	val := float32(math.MaxFloat32)
	base := "2"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_float32(t *testing.T) {
	val := float32(12.1)
	scale := "10"
	expected := float32(121)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Float32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Float32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Float32)
	}
}

func TestTransformReadResult_scale_float32_overflow(t *testing.T) {
	val := float32(math.MaxFloat32 / 2)
	scale := "3"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_float32(t *testing.T) {
	val := float32(1.1)
	offset := "1"
	expected := float32(2.1)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Float32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Float32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Float32)
	}
}

func TestTransformReadResult_offset_float32_overflow(t *testing.T) {
	val := float32(math.MaxFloat32)
	offset := fmt.Sprintf("%v", math.MaxFloat32)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if !errors.As(err, &OverflowError{}) {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_float64(t *testing.T) {
	val := 1.1
	base := "2"
	expected := 2.1435469250725863
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Float64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Float64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Float64)
	}
}

func TestTransformReadResult_scale_float64(t *testing.T) {
	val := float32(200000000)
	scale := "10"
	expected := float32(2000000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Float32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Float32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Float32)
	}
}

func TestTransformReadResult_offset_float64(t *testing.T) {
	val := float64(1.1)
	offset := "1"
	expected := float64(2.1)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewFloat64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Float64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Float64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Float64)
	}
}

func TestTransformReadResult_mask_uint8(t *testing.T) {
	val := uint8(math.MaxUint8)
	mask := "15"
	expected := uint8(15)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Mask: mask,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint8)
	}
}

func TestTransformReadResult_mask_uint16(t *testing.T) {
	val := uint16(math.MaxUint16)
	mask := "256"
	expected := uint16(256)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Mask: mask,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint16)
	}
}

func TestTransformReadResult_mask_uint32(t *testing.T) {
	val := uint32(math.MaxUint32)
	mask := "256"
	expected := uint32(256)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Mask: mask,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint32)
	}
}

func TestTransformReadResult_mask_uint64(t *testing.T) {
	val := uint64(math.MaxUint64)
	mask := "256"
	expected := uint64(256)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Mask: mask,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint64)
	}
}

func TestTransformReadResult_mask_should_be_unsigned(t *testing.T) {
	val := uint64(math.MaxUint64)
	mask := "-256"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Mask: mask,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil || err.Error() != "invalid mask value, the mask -256 should be unsigned and parsed to uint64. strconv.ParseUint: parsing \"-256\": invalid syntax" {
		t.Fatalf("Unexpected test result, transform function should throw the correct error. %v", err)
	}
}

func TestTransformReadResult_shift_uint8(t *testing.T) {
	val := uint8(6)
	shift := "5"
	expected := uint8(192)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint8)
	}
}

func TestTransformReadResult_signedShift_uint8(t *testing.T) {
	val := uint8(96)
	shift := "-4"
	expected := uint8(6)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint8Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint8Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint8)
	}
}

func TestTransformReadResult_shift_uint16(t *testing.T) {
	val := uint16(128)
	shift := "8"
	expected := uint16(32768)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint16)
	}
}

func TestTransformReadResult_signedShift_uint16(t *testing.T) {
	val := uint16(32768)
	shift := "-8"
	expected := uint16(128)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint16Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint16Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint16)
	}
}

func TestTransformReadResult_shift_uint32(t *testing.T) {
	val := uint32(120000)
	shift := "3"
	expected := uint32(960000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint32)
	}
}

func TestTransformReadResult_signedShift_uint32(t *testing.T) {
	val := uint32(120000)
	shift := "-2"
	expected := uint32(30000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint32Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint32Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint32)
	}
}

func TestTransformReadResult_shift_uint64(t *testing.T) {
	val := uint64(1000000000)
	shift := "4"
	expected := uint64(16000000000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint64)
	}
}

func TestTransformReadResult_signedShift_uint64(t *testing.T) {
	val := uint64(1000000000)
	shift := "-4"
	expected := uint64(62500000)
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	result, err := cv.Uint64Value()
	if err != nil {
		t.Fatalf("Fail to transform read result, error: %v", err)
	}
	if result != expected {
		t.Fatalf("Unexpect test result, result '%v' should be '%v'", result, expected)
	}
	if cv.Type != dsModels.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, dsModels.Uint64)
	}
}

func TestTransformReadResult_shift_should_be_valid(t *testing.T) {
	val := uint64(1000000000)
	shift := "-+4"
	ro := contract.ResourceOperation{DeviceResource: "test-object"}
	cv, err := dsModels.NewUint64Value(ro.DeviceResource, 0, val)
	pv := contract.PropertyValue{
		Shift: shift,
	}

	err = TransformReadResult(cv, pv, lc)

	if err == nil {
		t.Fatalf("Unexpected test result, transform function should throw the error.")
	}
}
