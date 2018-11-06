// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"math"
	"strconv"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	defaultBase   string = "0"
	defaultScale  string = "1.0"
	defaultOffset string = "0.0"
)

func TransformReadResult(cv *ds_models.CommandValue, pv models.PropertyValue) error {
	if cv.Type == ds_models.String || cv.Type == ds_models.Bool {
		return nil // do nothing for String and Bool
	}

	value, err := commandValueForTransform(cv)
	newValue := value

	if pv.Base != "" && pv.Base != defaultBase {
		newValue, err = transformReadBase(newValue, pv.Base)
		if err != nil {
			return err
		}
	}

	if pv.Scale != "" && pv.Scale != defaultScale {
		newValue, err = transformReadScale(newValue, pv.Scale)
		if err != nil {
			return err
		}
	}

	if pv.Offset != "" && pv.Offset != defaultOffset {
		newValue, err = transformReadOffset(newValue, pv.Offset)
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

	valueFloat64 = math.Pow(b, valueFloat64)

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
		s, err := strconv.ParseUint(scale, 10, 8)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := uint8(s)
		value = v * ns
	case uint16:
		s, err := strconv.ParseUint(scale, 10, 16)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := uint16(s)
		value = v * ns
	case uint32:
		s, err := strconv.ParseUint(scale, 10, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := uint32(s)
		value = v * ns
	case uint64:
		s, err := strconv.ParseUint(scale, 10, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		value = v * s
	case int8:
		s, err := strconv.ParseInt(scale, 10, 8)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := int8(s)
		value = v * ns
	case int16:
		s, err := strconv.ParseInt(scale, 10, 16)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := int16(s)
		value = v * ns
	case int32:
		s, err := strconv.ParseInt(scale, 10, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := int32(s)
		value = v * ns
	case int64:
		s, err := strconv.ParseInt(scale, 10, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		value = v * s
	case float32:
		s, err := strconv.ParseFloat(scale, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := float32(s)
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
		no := uint8(o)
		value = v + no
	case uint16:
		o, err := strconv.ParseUint(offset, 10, 16)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := uint16(o)
		value = v + no
	case uint32:
		o, err := strconv.ParseUint(offset, 10, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := uint32(o)
		value = v + no
	case uint64:
		o, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		value = v + o
	case int8:
		o, err := strconv.ParseInt(offset, 10, 8)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := int8(o)
		value = v + no
	case int16:
		o, err := strconv.ParseInt(offset, 10, 16)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := int16(o)
		value = v + no
	case int32:
		o, err := strconv.ParseInt(offset, 10, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := int32(o)
		value = v + no
	case int64:
		o, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		value = v + o
	case float32:
		o, err := strconv.ParseFloat(offset, 32)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := float32(o)
		value = v + no
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

func commandValueForTransform(cv *ds_models.CommandValue) (interface{}, error) {
	var v interface{}
	var err error = nil
	switch cv.Type {
	case ds_models.Uint8:
		v, err = cv.Uint8Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Uint16:
		v, err = cv.Uint16Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Uint32:
		v, err = cv.Uint32Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Uint64:
		v, err = cv.Uint64Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Int8:
		v, err = cv.Int8Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Int16:
		v, err = cv.Int16Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Int32:
		v, err = cv.Int32Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Int64:
		v, err = cv.Int64Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Float32:
		v, err = cv.Float32Value()
		if err != nil {
			return 0, err
		}
	case ds_models.Float64:
		v, err = cv.Float64Value()
		if err != nil {
			return 0, err
		}
	default:
		err = fmt.Errorf("wrong data type of CommandValue to transform: %s", cv.String())
	}
	return v, nil
}

func replaceNewCommandValue(cv *ds_models.CommandValue, newValue interface{}) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, newValue)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("binary.Write failed: %v", err))
	} else {
		cv.NumericValue = buf.Bytes()
	}
	return err
}

func CheckAssertion(cv *ds_models.CommandValue, assertion string, device *models.Device) error {
	if assertion != "" && cv.ValueToString() != assertion {
		device.OperatingState = models.Disabled
		cache.Devices().Update(*device)
		go common.DeviceClient.UpdateOpStateByName(device.Name, models.Disabled)
		msg := fmt.Sprintf("assertion (%s) failed with value: %s", assertion, cv.ValueToString())
		common.LoggingClient.Error(msg)
		return fmt.Errorf(msg)
	}
	return nil
}

func MapCommandValue(value *ds_models.CommandValue) (*ds_models.CommandValue, bool) {
	mappings := value.RO.Mappings
	newValue, ok := mappings[value.ValueToString()]
	var result *ds_models.CommandValue
	if ok {
		result = ds_models.NewStringValue(value.RO, value.Origin, newValue)
	}
	return result, ok
}
