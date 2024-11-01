// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"crypto/rand"
	"runtime"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/panjf2000/ants/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareReadings(t *testing.T) {
	autoEvent := models.AutoEvent{SourceName: "sourceName", OnChange: true, Interval: "500ms"}
	pool, err := ants.NewPool(runtime.GOMAXPROCS(0), ants.WithNonblocking(true))
	require.NoError(t, err)
	e, err := NewExecutor("device-test", autoEvent, pool)
	require.NoError(t, err)

	testReadings := []dtos.BaseReading{{ResourceName: "r1"}, {ResourceName: "r2"}}
	testReadings[0].ValueType = common.ValueTypeInt8
	testReadings[0].Value = "1"
	testReadings[0].ValueType = common.ValueTypeInt8
	testReadings[1].Value = "2"

	firstReadings := testReadings

	readingsValueChanged := make([]dtos.BaseReading, len(firstReadings))
	copy(readingsValueChanged, firstReadings)
	readingsValueChanged[1].Value = "3"

	readingsResourceChanged := make([]dtos.BaseReading, len(readingsValueChanged))
	copy(readingsResourceChanged, readingsValueChanged)
	readingsResourceChanged[0].ResourceName = "c1"

	readingsValueUnchanged := readingsResourceChanged

	readingsLengthChanged := append(readingsValueUnchanged, dtos.BaseReading{})
	readingsLengthChanged[2].ValueType = common.ValueTypeBinary
	readingsLengthChanged[2].ResourceName = "b1"
	readingsLengthChanged[2].BinaryValue = make([]byte, 1000)
	_, randErr := rand.Read(readingsLengthChanged[2].BinaryValue) // nolint: gosec
	require.NoError(t, randErr)

	readingsBinaryValueChanged := make([]dtos.BaseReading, len(readingsLengthChanged))
	copy(readingsBinaryValueChanged, readingsLengthChanged)
	readingsBinaryValueChanged[2].BinaryValue = make([]byte, 1000)
	_, randErr = rand.Read(readingsBinaryValueChanged[2].BinaryValue)
	require.NoError(t, randErr)
	readingBinaryValueUnchanged := readingsBinaryValueChanged

	tests := []struct {
		name     string
		reading  []dtos.BaseReading
		expected bool
	}{
		{"false - lastReadings are nil", firstReadings, false},
		{"false - reading's value changed", readingsValueChanged, false},
		{"false - reading's resource name changed", readingsResourceChanged, false},
		{"true - readings unchanged", readingsValueUnchanged, true},
		{"false - readings length changed", readingsLengthChanged, false},
		{"false - reading's binary value changed", readingsBinaryValueChanged, false},
		{"true - readings unchanged", readingBinaryValueUnchanged, true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			res := e.compareReadings(testCase.reading)
			assert.Equal(t, testCase.expected, res, "compareReading result not as expected")
		})
	}
}

func TestOnChangeThreshold(t *testing.T) {
	deviceName := "testDevice"
	resourceName := "testResource"
	profileName := "testProfile"
	autoEvent := models.AutoEvent{SourceName: resourceName, OnChange: true, Interval: "500ms"}
	pool, err := ants.NewPool(runtime.GOMAXPROCS(0), ants.WithNonblocking(true))
	require.NoError(t, err)
	e, err := NewExecutor(deviceName, autoEvent, pool)
	require.NoError(t, err)

	tests := []struct {
		name                string
		valueType           string
		lastReadingValue    any
		currentReadingValue any
		onChangeThreshold   float64
		expectUnchanged     bool
	}{
		{"float32 unchanged is true", common.ValueTypeFloat32, float32(0), float32(0.01), 0.01, true},
		{"float32 unchanged is false", common.ValueTypeFloat32, float32(0), float32(0.02), 0.01, false},
		{"float64 unchanged is true", common.ValueTypeFloat64, float64(0), float64(0.01), 0.01, true},
		{"float64 unchanged is false", common.ValueTypeFloat64, float64(0), float64(0.02), 0.01, false},
		{"uint8 unchanged is true", common.ValueTypeUint8, uint8(0), uint8(1), 1, true},
		{"uint8 unchanged is false", common.ValueTypeUint8, uint8(0), uint8(2), 1, false},
		{"uint16 unchanged is true", common.ValueTypeUint16, uint16(0), uint16(1), 1, true},
		{"uint16 unchanged is false", common.ValueTypeUint16, uint16(0), uint16(2), 1, false},
		{"uint32 unchanged is true", common.ValueTypeUint32, uint32(0), uint32(1), 1, true},
		{"uint32 unchanged is false", common.ValueTypeUint32, uint32(0), uint32(2), 1, false},
		{"uint64 unchanged is true", common.ValueTypeUint64, uint64(0), uint64(1), 1, true},
		{"uint64 unchanged is false", common.ValueTypeUint64, uint64(0), uint64(2), 1, false},
		{"int8 unchanged is true", common.ValueTypeInt8, int8(0), int8(1), 1, true},
		{"int8 unchanged is false", common.ValueTypeInt8, int8(0), int8(2), 1, false},
		{"int16 unchanged is true", common.ValueTypeInt16, int16(0), int16(1), 1, true},
		{"int16 unchanged is false", common.ValueTypeInt16, int16(0), int16(2), 1, false},
		{"int32 unchanged is true", common.ValueTypeInt32, int32(0), int32(1), 1, true},
		{"int32 unchanged is false", common.ValueTypeInt32, int32(0), int32(2), 1, false},
		{"int64 unchanged is true", common.ValueTypeInt64, int64(0), int64(1), 1, true},
		{"int64 unchanged is false", common.ValueTypeInt64, int64(0), int64(2), 1, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			lastReading, err := dtos.NewSimpleReading(profileName, deviceName, resourceName, testCase.valueType, testCase.lastReadingValue)
			require.NoError(t, err)
			currentReading, err := dtos.NewSimpleReading(profileName, deviceName, resourceName, testCase.valueType, testCase.currentReadingValue)
			require.NoError(t, err)
			e.lastReadings = map[string]any{lastReading.ResourceName: lastReading.Value}
			e.onChangeThreshold = testCase.onChangeThreshold

			res := e.compareReadings([]dtos.BaseReading{currentReading})

			assert.Equal(t, testCase.expectUnchanged, res, "compareReading result not as expected")
		})
	}
}
