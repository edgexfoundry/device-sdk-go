// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"math"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

func TransformWriteParameter(cv *dsModels.CommandValue, pv models.PropertyValue, lc logger.LoggingClient) error {
	var err error
	if cv.Type == v2.ValueTypeString || cv.Type == v2.ValueTypeBool || cv.Type == v2.ValueTypeBinary {
		return nil // do nothing for String, Bool and Binary
	}

	value, err := commandValueForTransform(cv)
	newValue := value

	if pv.Offset != "" && pv.Offset != defaultOffset {
		newValue, err = transformWriteOffset(newValue, pv.Offset, lc)
		if err != nil {
			return err
		}
	}

	if pv.Scale != "" && pv.Scale != defaultScale {
		newValue, err = transformWriteScale(newValue, pv.Scale, lc)
		if err != nil {
			return err
		}
	}

	if pv.Base != "" && pv.Base != defaultBase {
		newValue, err = transformWriteBase(newValue, pv.Base, lc)
	}

	if value != newValue {
		err = replaceNewCommandValue(cv, newValue, lc)
	}
	return err
}

func transformWriteBase(value interface{}, base string, lc logger.LoggingClient) (interface{}, error) {
	b, err := strconv.ParseFloat(base, 64)
	if err != nil {
		lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to float64: %v", base, err))
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

	// inverse of a base transform for a value
	valueFloat64 = math.Log(valueFloat64) / math.Log(b)

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

func transformWriteScale(value interface{}, scale string, lc logger.LoggingClient) (interface{}, error) {
	switch v := value.(type) {
	case uint8:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		value = uint8(nv / s)
	case uint16:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		value = uint16(nv / s)
	case uint32:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		value = uint32(nv / s)
	case uint64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		value = uint64(nv / s)
	case int8:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		value = int8(nv / s)
	case int16:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		value = int16(nv / s)
	case int32:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		value = int32(nv / s)
	case int64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		nv := float64(v)
		value = int64(nv / s)
	case float32:
		s, err := strconv.ParseFloat(scale, 32)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		ns := float32(s)
		value = v / ns
	case float64:
		s, err := strconv.ParseFloat(scale, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the scale %s of PropertyValue cannot be parsed to %T: %v", scale, v, err))
			return value, err
		}
		value = v / s
	}

	return value, nil
}

func transformWriteOffset(value interface{}, offset string, lc logger.LoggingClient) (interface{}, error) {
	switch v := value.(type) {
	case uint8:
		o, err := strconv.ParseUint(offset, 10, 8)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := uint8(o)
		value = v - no
	case uint16:
		o, err := strconv.ParseUint(offset, 10, 16)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := uint16(o)
		value = v - no
	case uint32:
		o, err := strconv.ParseUint(offset, 10, 32)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := uint32(o)
		value = v - no
	case uint64:
		o, err := strconv.ParseUint(offset, 10, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		value = v - o
	case int8:
		o, err := strconv.ParseInt(offset, 10, 8)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := int8(o)
		value = v - no
	case int16:
		o, err := strconv.ParseInt(offset, 10, 16)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := int16(o)
		value = v - no
	case int32:
		o, err := strconv.ParseInt(offset, 10, 32)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := int32(o)
		value = v - no
	case int64:
		o, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		value = v - o
	case float32:
		o, err := strconv.ParseFloat(offset, 32)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		no := float32(o)
		value = v - no
	case float64:
		o, err := strconv.ParseFloat(offset, 64)
		if err != nil {
			lc.Error(fmt.Sprintf("the offset %s of PropertyValue cannot be parsed to %T: %v", offset, v, err))
			return value, err
		}
		value = v - o
	}

	return value, nil
}
