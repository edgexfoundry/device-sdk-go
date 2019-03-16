// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func init() {
	lc := logger.NewClient("sdk", false, "./test.log", "DEBUG")
	common.LoggingClient = lc
}

func TestTransformReadResult_base_unt8(t *testing.T) {
	val := uint8(10)
	base := "2"
	expected := uint8(100)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint8)
	}
}

func TestTransformReadResult_base_unt8_overflow(t *testing.T) {
	val := uint8(10)
	base := "3"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_unt8(t *testing.T) {
	val := uint8(math.MaxUint8 / 5)
	scale := "5"
	expected := uint8(math.MaxUint8)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint8)
	}
}

func TestTransformReadResult_scale_unt8_overflow(t *testing.T) {
	val := uint8(math.MaxUint8 / 5)
	scale := "6"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_unt8(t *testing.T) {
	val := uint8(math.MaxUint8 - 1)
	offset := "1"
	expected := uint8(math.MaxUint8)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint8)
	}
}

func TestTransformReadResult_offset_unt8_overflow(t *testing.T) {
	val := uint8(math.MaxUint8)
	offset := "1"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_unt16(t *testing.T) {
	val := uint16(200)
	base := "2"
	expected := uint16(40000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint16)
	}
}

func TestTransformReadResult_base_uint16_overflow(t *testing.T) {
	val := uint16(200)
	base := "3"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_uint16(t *testing.T) {
	val := uint16(math.MaxUint16 / 5)
	scale := "5"
	expected := uint16(math.MaxUint16)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint16)
	}
}

func TestTransformReadResult_scale_uint16_overflow(t *testing.T) {
	val := uint16(math.MaxUint16 / 5)
	scale := "6"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_uint16(t *testing.T) {
	val := uint16(math.MaxUint16 - 1)
	offset := "1"
	expected := uint16(math.MaxUint16)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint16)
	}
}

func TestTransformReadResult_offset_uint16_overflow(t *testing.T) {
	val := uint16(math.MaxUint16)
	offset := "1"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_uint32(t *testing.T) {
	val := uint32(20000)
	base := "2"
	expected := uint32(400000000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint32)
	}
}

func TestTransformReadResult_base_uint32_overflow(t *testing.T) {
	val := uint32(4000000)
	base := "1000"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_uint32(t *testing.T) {
	val := uint32(math.MaxUint32 / 5)
	scale := "5"
	expected := uint32(math.MaxUint32)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint32)
	}
}

func TestTransformReadResult_scale_uint32_overflow(t *testing.T) {
	val := uint32(math.MaxUint32 / 5)
	scale := "6"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_uint32(t *testing.T) {
	val := uint32(math.MaxUint32 - 1)
	offset := "1"
	expected := uint32(math.MaxUint32)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint32)
	}
}

func TestTransformReadResult_offset_uint32_overflow(t *testing.T) {
	val := uint32(math.MaxUint32)
	offset := "1"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_uint64(t *testing.T) {
	val := uint64(20000)
	base := "2"
	expected := uint64(400000000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint64Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint64)
	}
}

func TestTransformReadResult_scale_uint64(t *testing.T) {
	val := uint64(20000)
	scale := "20000"
	expected := uint64(400000000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint64Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint64)
	}
}

func TestTransformReadResult_offset_uint64(t *testing.T) {
	val := uint64(math.MaxUint64) - uint64(1)
	offset := "1"
	expected := uint64(math.MaxUint64)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint64Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Uint64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Uint64)
	}
}

func TestTransformReadResult_base_int8(t *testing.T) {
	val := int8(10)
	base := "2"
	expected := int8(100)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int8)
	}
}

func TestTransformReadResult_base_int8_overflow(t *testing.T) {
	val := int8(10)
	base := "3"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_int8(t *testing.T) {
	val := int8(10)
	scale := "10"
	expected := int8(100)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int8)
	}
}

func TestTransformReadResult_scale_int8_overflow(t *testing.T) {
	val := uint8(10)
	scale := "30"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewUint8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_int8(t *testing.T) {
	val := int8(math.MaxInt8 - 1)
	offset := "1"
	expected := int8(math.MaxInt8)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int8 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int8)
	}
}

func TestTransformReadResult_offset_int8_overflow(t *testing.T) {
	val := int8(math.MaxInt8)
	offset := "1"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt8Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_int16(t *testing.T) {
	val := int16(100)
	base := "2"
	expected := int16(10000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int16)
	}
}

func TestTransformReadResult_base_int16_overflow(t *testing.T) {
	val := int16(100)
	base := "3"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_int16(t *testing.T) {
	val := int16(10000)
	scale := "3"
	expected := int16(30000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int16)
	}
}

func TestTransformReadResult_scale_int16_overflow(t *testing.T) {
	val := int16(10000)
	scale := "4"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_int16(t *testing.T) {
	val := int16(math.MaxInt16 - 1)
	offset := "1"
	expected := int16(math.MaxInt16)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int16 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int16)
	}
}

func TestTransformReadResult_offset_int16_overflow(t *testing.T) {
	val := int16(math.MaxInt16)
	offset := "1"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt16Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_int32(t *testing.T) {
	val := int32(20000)
	base := "2"
	expected := int32(400000000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int32)
	}
}

func TestTransformReadResult_base_int32_overflow(t *testing.T) {
	val := int32(20000)
	base := "3"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_int32(t *testing.T) {
	val := int32(200000000)
	scale := "10"
	expected := int32(2000000000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int32)
	}
}

func TestTransformReadResult_scale_int32_overflow(t *testing.T) {
	val := int32(200000000)
	scale := "15"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_int32(t *testing.T) {
	val := int32(math.MaxInt32 - 1)
	offset := "1"
	expected := int32(math.MaxInt32)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int32)
	}
}

func TestTransformReadResult_offset_int32_overflow(t *testing.T) {
	val := int32(math.MaxInt32)
	offset := "1"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_int64(t *testing.T) {
	val := int64(20000)
	base := "2"
	expected := int64(400000000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt64Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int64)
	}
}

func TestTransformReadResult_scale_int64(t *testing.T) {
	val := int64(20000)
	scale := "20000"
	expected := int64(400000000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt64Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int64)
	}
}

func TestTransformReadResult_offset_int64(t *testing.T) {
	val := int64(math.MaxInt64) - int64(1)
	offset := "1"
	expected := int64(math.MaxInt64)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewInt64Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Int64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Int64)
	}
}

func TestTransformReadResult_base_float32(t *testing.T) {
	val := float32(1.1)
	base := "2"
	expected := float32(1.21)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Float32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Float32)
	}
}

func TestTransformReadResult_base_float32_overflow(t *testing.T) {
	val := float32(math.MaxFloat32)
	base := "2"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with base '%v' should be overflow", val, base)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_scale_float32(t *testing.T) {
	val := float32(12.1)
	scale := "10"
	expected := float32(121)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Float32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Float32)
	}
}

func TestTransformReadResult_scale_float32_overflow(t *testing.T) {
	val := float32(math.MaxFloat32 / 2)
	scale := "3"
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with scale '%v' should be overflow", val, scale)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_offset_float32(t *testing.T) {
	val := float32(1.1)
	offset := "1"
	expected := float32(2.1)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Float32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Float32)
	}
}

func TestTransformReadResult_offset_float32_overflow(t *testing.T) {
	val := float32(math.MaxFloat32)
	offset := fmt.Sprintf("%v", math.MaxFloat32)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

	if err == nil {
		t.Fatalf("Unexpect test result, transform reading '%v' with offset '%v' should be overflow", val, offset)
	}
	if err, ok := errors.Cause(err).(OverflowError); !ok {
		t.Fatalf("Unexpect test result, error should be OverflowError, %v", err)
	}
}

func TestTransformReadResult_base_float64(t *testing.T) {
	val := float64(11)
	base := "2"
	expected := float64(121)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat64Value(&ro, 0, val)
	pv := models.PropertyValue{
		Base: base,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Float64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Float64)
	}
}

func TestTransformReadResult_scale_float64(t *testing.T) {
	val := float32(200000000)
	scale := "10"
	expected := float32(2000000000)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat32Value(&ro, 0, val)
	pv := models.PropertyValue{
		Scale: scale,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Float32 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Float32)
	}
}

func TestTransformReadResult_offset_float64(t *testing.T) {
	val := float64(1.1)
	offset := "1"
	expected := float64(2.1)
	ro := models.ResourceOperation{Object: "test-object"}
	cv, err := ds_models.NewFloat64Value(&ro, 0, val)
	pv := models.PropertyValue{
		Offset: offset,
	}

	err = TransformReadResult(cv, pv)

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
	if cv.Type != ds_models.Float64 {
		t.Fatalf("Unexpect test result, value type '%v' should be '%v'", cv.Type, ds_models.Float64)
	}
}
