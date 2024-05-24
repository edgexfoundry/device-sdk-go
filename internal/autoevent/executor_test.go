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

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/panjf2000/ants"
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
