// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-core-contracts/requests/states/operating"
	"github.com/google/uuid"
)

const (
	defaultBase   string = "0"
	defaultScale  string = "1.0"
	defaultOffset string = "0.0"
	defaultMask   string = "0"
	defaultShift  string = "0"

	Overflow = "overflow"
	NaN      = "NaN"
)

func TransformReadResult(cv *dsModels.CommandValue, pv contract.PropertyValue, lc logger.LoggingClient) error {
	if cv.Type == dsModels.String || cv.Type == dsModels.Bool || cv.Type == dsModels.Binary {
		return nil // do nothing for String, Bool and Binary
	}
	res, err := isNaN(cv)
	if err != nil {
		return err
	} else if res {
		return fmt.Errorf("NaN error for device resource '%s', error: %w", cv.DeviceResourceName, NaNError{})
	}

	value, err := commandValueForTransform(cv)
	newValue := value

	if pv.Mask != "" && pv.Mask != defaultMask &&
		(cv.Type == dsModels.Uint8 || cv.Type == dsModels.Uint16 || cv.Type == dsModels.Uint32 || cv.Type == dsModels.Uint64) {
		newValue, err = transformReadMask(newValue, pv.Mask, lc)
		if err != nil {
			return err
		}
	}

	if pv.Shift != "" && pv.Shift != defaultShift &&
		(cv.Type == dsModels.Uint8 || cv.Type == dsModels.Uint16 || cv.Type == dsModels.Uint32 || cv.Type == dsModels.Uint64) {
		newValue, err = transformReadShift(newValue, pv.Shift, lc)
		if err != nil {
			return fmt.Errorf("transform failed for device resource '%v', error: %w ", cv.DeviceResourceName, err)
		}
	}

	if pv.Base != "" && pv.Base != defaultBase {
		newValue, err = transformReadBase(newValue, pv.Base, lc)
		if err != nil {
			return fmt.Errorf("transform failed for device resource '%v', error: %w ", cv.DeviceResourceName, err)
		}
	}

	if pv.Scale != "" && pv.Scale != defaultScale {
		newValue, err = transformReadScale(newValue, pv.Scale, lc)
		if err != nil {
			return fmt.Errorf("transform failed for device resource '%v', error: %w ", cv.DeviceResourceName, err)
		}
	}

	if pv.Offset != "" && pv.Offset != defaultOffset {
		newValue, err = transformReadOffset(newValue, pv.Offset, lc)
		if err != nil {
			return fmt.Errorf("transform failed for device resource '%v', error: %w ", cv.DeviceResourceName, err)
		}
	}

	if value != newValue {
		err = replaceNewCommandValue(cv, newValue, lc)
	}
	return err
}

func transformReadBase(value interface{}, base string, lc logger.LoggingClient) (interface{}, error) {
	b, err := strconv.ParseFloat(base, 64)
	if err != nil {
		lc.Error(fmt.Sprintf("the base %s of PropertyValue cannot be parsed to float64: %v", base, err))
		return value, err
	} else if b == 0 {
		return value, nil // do nothing if Base = 0
	}

	var valueFloat64 float64
	switch v := value.(type) {
	case uint8:
		valueFloat64 = float64(v)
	case uint16:
		valueFloat64 = float64(v)
	case uint32:
		valueFloat64 = float64(v)
	case uint64:
		valueFloat64 = float64(v)
	case int8:
		valueFloat64 = float64(v)
	case int16:
		valueFloat64 = float64(v)
	case int32:
		valueFloat64 = float64(v)
	case int64:
		valueFloat64 = float64(v)
	case float32:
		valueFloat64 = float64(v)
	case float64:
		valueFloat64 = v
	}

	valueFloat64 = math.Pow(b, valueFloat64)
	inRange := checkTransformedValueInRange(value, valueFloat64, lc)
	if !inRange {
		return value, NewOverflowError(value, valueFloat64)
	}

	switch value.(type) {
	case uint8:
		value = uint8(valueFloat64)
	case uint16:
		value = uint16(valueFloat64)
	case uint32:
		value = uint32(valueFloat64)
	case uint64:
		value = uint64(valueFloat64)
	case int8:
		value = int8(valueFloat64)
	case int16:
		value = int16(valueFloat64)
	case int32:
		value = int32(valueFloat64)
	case int64:
		value = int64(valueFloat64)
	case float32:
		value = float32(valueFloat64)
	case float64:
		value = valueFloat64
	}
	return value, err
}

func transformReadScale(value interface{}, scale string, lc logger.LoggingClient) (interface{}, error) {
	switch v := value.(type) {
	case uint8:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint8(transformedValue)
	case uint16:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint16(transformedValue)
	case uint32:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint32(transformedValue)
	case uint64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint64(transformedValue)
	case int8:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int8(transformedValue)
	case int16:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int16(transformedValue)
	case int32:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int32(transformedValue)
	case int64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int64(transformedValue)
	case float32:
		s, err := strconv.ParseFloat(scale, 32)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := float32(s)
		transformedValue := float64(v * ns)

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = v * ns
	case float64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		value = v * s
	}

	return value, nil
}

func transformReadOffset(value interface{}, offset string, lc logger.LoggingClient) (interface{}, error) {
	switch v := value.(type) {
	case uint8:
		o, err := strconv.ParseUint(offset, 10, 8)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint8(transformedValue)
	case uint16:
		o, err := strconv.ParseUint(offset, 10, 16)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint16(transformedValue)
	case uint32:
		o, err := strconv.ParseUint(offset, 10, 32)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint32(transformedValue)
	case uint64:
		o, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := uint64(v + o)

		inRange := checkTransformedValueInRange(value, float64(transformedValue), lc)
		if !inRange {
			return value, NewOverflowError(value, float64(transformedValue))
		}

		value = uint64(v + o)
	case int8:
		o, err := strconv.ParseInt(offset, 10, 8)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int8(transformedValue)
	case int16:
		o, err := strconv.ParseInt(offset, 10, 16)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int16(transformedValue)
	case int32:
		o, err := strconv.ParseInt(offset, 10, 32)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int32(transformedValue)
	case int64:
		o, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := int64(v + o)

		inRange := checkTransformedValueInRange(value, float64(transformedValue), lc)
		if !inRange {
			return value, NewOverflowError(value, float64(transformedValue))
		}

		value = transformedValue
	case float32:
		o, err := strconv.ParseFloat(offset, 32)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue, lc)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = float32(transformedValue)
	case float64:
		o, err := strconv.ParseFloat(offset, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		value = v + o
	}

	return value, nil
}

func transformReadMask(value interface{}, mask string, lc logger.LoggingClient) (interface{}, error) {
	nv, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64)
	if err != nil {
		lc.Error(fmt.Sprintf("the value %s cannot be parsed to uint64: %v", value, err))
		return value, err
	}
	m, err := strconv.ParseUint(mask, 10, 64)
	if err != nil {
		return value, fmt.Errorf("invalid mask value, the mask %s should be unsigned and parsed to %T. %v", mask, m, err)
	}

	transformedValue := nv & m

	switch value.(type) {
	case uint8:
		value = uint8(transformedValue)
	case uint16:
		value = uint16(transformedValue)
	case uint32:
		value = uint32(transformedValue)
	case uint64:
		value = uint64(transformedValue)
	}

	return value, err
}

func transformReadShift(value interface{}, shift string, lc logger.LoggingClient) (interface{}, error) {
	nv, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64)
	if err != nil {
		lc.Error(fmt.Sprintf("the value %s cannot be parsed to uint64: %v", value, err))
		return value, err
	}
	signed, err := isSignedNumber(shift)
	if err != nil {
		return value, err
	}

	var transformedValue uint64
	if signed {
		signedShift, err := strconv.ParseInt(shift, 10, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the shift %s of PropertyValue cannot be parsed to %T: %v", shift, signedShift, err))
			return value, err
		}
		s := uint64(-signedShift)
		transformedValue = nv >> s
	} else {
		s, err := strconv.ParseUint(shift, 10, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the shift %s of PropertyValue cannot be parsed to %T: %v", shift, s, err))
			return value, err
		}
		transformedValue = nv << s
	}

	inRange := checkTransformedValueInRange(value, float64(transformedValue), lc)
	if !inRange {
		return value, NewOverflowError(value, float64(transformedValue))
	}

	switch value.(type) {
	case uint8:
		value = uint8(transformedValue)
	case uint16:
		value = uint16(transformedValue)
	case uint32:
		value = uint32(transformedValue)
	case uint64:
		value = uint64(transformedValue)
	}

	return value, err
}

func isSignedNumber(shift string) (bool, error) {
	s, err := strconv.ParseFloat(shift, 64)
	if err != nil {
		return false, fmt.Errorf("invalid shift value, the shift %v should be parsed to float64 for checking the sign of the number. %v", shift, err)
	}
	return math.Signbit(s), nil
}

func commandValueForTransform(cv *dsModels.CommandValue) (interface{}, error) {
	var v interface{}
	var err error = nil
	switch cv.Type {
	case dsModels.Uint8:
		v, err = cv.Uint8Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Uint16:
		v, err = cv.Uint16Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Uint32:
		v, err = cv.Uint32Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Uint64:
		v, err = cv.Uint64Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Int8:
		v, err = cv.Int8Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Int16:
		v, err = cv.Int16Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Int32:
		v, err = cv.Int32Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Int64:
		v, err = cv.Int64Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Float32:
		v, err = cv.Float32Value()
		if err != nil {
			return 0, err
		}
	case dsModels.Float64:
		v, err = cv.Float64Value()
		if err != nil {
			return 0, err
		}
	default:
		err = fmt.Errorf("wrong data type of CommandValue to transform: %s", cv.String())
	}
	return v, nil
}

func replaceNewCommandValue(cv *dsModels.CommandValue, newValue interface{}, lc logger.LoggingClient) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, newValue)
	if err != nil {
		lc.Error(fmt.Sprintf("binary.Write failed: %v", err))
	} else {
		cv.NumericValue = buf.Bytes()
	}
	return err
}

func CheckAssertion(
	cv *dsModels.CommandValue,
	assertion string,
	device *contract.Device,
	lc logger.LoggingClient,
	dc metadata.DeviceClient) error {
	if assertion != "" && cv.ValueToString() != assertion {
		device.OperatingState = contract.Disabled
		cache.Devices().Update(*device)
		ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
		go dc.UpdateOpStateByName(ctx, device.Name, operating.UpdateRequest{OperatingState: contract.Disabled})
		msg := fmt.Sprintf("assertion (%s) failed with value: %s", assertion, cv.ValueToString())
		lc.Error(msg)
		return fmt.Errorf(msg)
	}
	return nil
}

func MapCommandValue(value *dsModels.CommandValue, mappings map[string]string) (*dsModels.CommandValue, bool) {
	newValue, ok := mappings[value.ValueToString()]
	var result *dsModels.CommandValue
	if ok {
		result = dsModels.NewStringValue(value.DeviceResourceName, value.Origin, newValue)
	}
	return result, ok
}
