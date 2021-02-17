// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"math/rand"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareReadings(t *testing.T) {
	autoEvent := models.AutoEvent{Resource: "resource", OnChange: true, Frequency: "500ms"}
	e, err := NewExecutor("device-test", autoEvent)
	require.NoError(t, err)

	testReadings := []dtos.BaseReading{{ResourceName: "r1"}, {ResourceName: "r2"}}
	testReadings[0].ValueType = v2.ValueTypeInt8
	testReadings[0].Value = "1"
	testReadings[0].ValueType = v2.ValueTypeInt8
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
	readingsLengthChanged[2].ValueType = v2.ValueTypeBinary
	readingsLengthChanged[2].ResourceName = "b1"
	readingsLengthChanged[2].BinaryValue = make([]byte, 1000)
	rand.Read(readingsLengthChanged[2].BinaryValue)

	readingsBinaryValueChanged := make([]dtos.BaseReading, len(readingsLengthChanged))
	copy(readingsBinaryValueChanged, readingsLengthChanged)
	readingsBinaryValueChanged[2].BinaryValue = make([]byte, 1000)
	rand.Read(readingsBinaryValueChanged[2].BinaryValue)

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
