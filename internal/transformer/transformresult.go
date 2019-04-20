// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	defaultBase   string = "0"
	defaultScale  string = "1.0"
	defaultOffset string = "0.0"
)

func TransformReadResult(cv *dsModels.CommandValue, pv contract.PropertyValue) error {
	if cv.Type == dsModels.String || cv.Type == dsModels.Bool || cv.Type == dsModels.Binary {
		return nil // do nothing for String, Bool and Binary
	}

	value, err := commandValueForTransform(cv)
	newValue := value

	if pv.Base != "" && pv.Base != defaultBase {
		newValue, err = transformReadBase(newValue, pv.Base)
		if overflowError, ok := err.(OverflowError); ok {
			return errors.Wrap(overflowError, fmt.Sprintf("Overflow failed for device resource '%v' ", cv.RO.Object))
		} else if err != nil {
			return err
		}
	}

	if pv.Scale != "" && pv.Scale != defaultScale {
		newValue, err = transformReadScale(newValue, pv.Scale)
		if overflowError, ok := err.(OverflowError); ok {
			return errors.Wrap(overflowError, fmt.Sprintf("Overflow failed for device resource '%v' ", cv.RO.Object))
		} else if err != nil {
			return err
		}
	}

	if pv.Offset != "" && pv.Offset != defaultOffset {
		newValue, err = transformReadOffset(newValue, pv.Offset)
		if overflowError, ok := err.(OverflowError); ok {
			return errors.Wrap(overflowError, fmt.Sprintf("Overflow failed for device resource: %v", cv.RO.Object))
		} else if err != nil {
			return err
		}
	}

	if value != newValue {
		err = replaceNewCommandValue(cv, newValue)
	}
	return err
}

func transformReadBase(value interface{}, base string) (interface{}, error) {
	b, err := strconv.ParseFloat(base, 64)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("the base %s of PropertyValue cannot be parsed to float64: %v", base, err))
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

	valueFloat64 = math.Pow(valueFloat64, b)
	inRange := checkTransformedValueInRange(value, valueFloat64)
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

func transformReadScale(value interface{}, scale string) (interface{}, error) {
	switch v := value.(type) {
	case uint8:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint8(transformedValue)
	case uint16:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint16(transformedValue)
	case uint32:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint32(transformedValue)
	case uint64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint64(transformedValue)
	case int8:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int8(transformedValue)
	case int16:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int16(transformedValue)
	case int32:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int32(transformedValue)
	case int64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		transformedValue := nv * s

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int64(transformedValue)
	case float32:
		s, err := strconv.ParseFloat(scale, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := float32(s)
		transformedValue := float64(v * ns)

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = v * ns
	case float64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		value = v * s
	}

	return value, nil
}

func transformReadOffset(value interface{}, offset string) (interface{}, error) {
	switch v := value.(type) {
	case uint8:
		o, err := strconv.ParseUint(offset, 10, 8)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint8(transformedValue)
	case uint16:
		o, err := strconv.ParseUint(offset, 10, 16)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint16(transformedValue)
	case uint32:
		o, err := strconv.ParseUint(offset, 10, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = uint32(transformedValue)
	case uint64:
		o, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := uint64(v + o)

		inRange := checkTransformedValueInRange(value, float64(transformedValue))
		if !inRange {
			return value, NewOverflowError(value, float64(transformedValue))
		}

		value = uint64(v + o)
	case int8:
		o, err := strconv.ParseInt(offset, 10, 8)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int8(transformedValue)
	case int16:
		o, err := strconv.ParseInt(offset, 10, 16)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int16(transformedValue)
	case int32:
		o, err := strconv.ParseInt(offset, 10, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = int32(transformedValue)
	case int64:
		o, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := int64(v + o)

		inRange := checkTransformedValueInRange(value, float64(transformedValue))
		if !inRange {
			return value, NewOverflowError(value, float64(transformedValue))
		}

		value = transformedValue
	case float32:
		o, err := strconv.ParseFloat(offset, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		transformedValue := float64(v) + float64(o)

		inRange := checkTransformedValueInRange(value, transformedValue)
		if !inRange {
			return value, NewOverflowError(value, transformedValue)
		}

		value = float32(transformedValue)
	case float64:
		o, err := strconv.ParseFloat(offset, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		value = v + o
	}

	return value, nil
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

func replaceNewCommandValue(cv *dsModels.CommandValue, newValue interface{}) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, newValue)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("binary.Write failed: %v", err))
	} else {
		cv.NumericValue = buf.Bytes()
	}
	return err
}

func CheckAssertion(cv *dsModels.CommandValue, assertion string, device *contract.Device) error {
	if assertion != "" && cv.ValueToString() != assertion {
		device.OperatingState = contract.Disabled
		cache.Devices().Update(*device)
		ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
		go common.DeviceClient.UpdateOpStateByName(device.Name, contract.Disabled, ctx)
		msg := fmt.Sprintf("assertion (%s) failed with value: %s", assertion, cv.ValueToString())
		common.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}
	return nil
}

func MapCommandValue(value *dsModels.CommandValue) (*dsModels.CommandValue, bool) {
	mappings := value.RO.Mappings
	newValue, ok := mappings[value.ValueToString()]
	var result *dsModels.CommandValue
	if ok {
		result = dsModels.NewStringValue(value.RO, value.Origin, newValue)
	}
	return result, ok
}
