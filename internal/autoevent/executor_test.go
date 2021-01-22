// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareReadings(t *testing.T) {
	autoEvent := models.AutoEvent{Resource: "resource", OnChange: true, Frequency: "500ms"}
	e, err := NewExecutor("device-test", autoEvent)
	require.NoError(t, err)

	firstNonBinaryReading := dtos.BaseReading{}
	firstNonBinaryReading.Value = "1"
	nonBinaryReadingValueChanged := firstNonBinaryReading
	nonBinaryReadingValueChanged.Value = "2"
	nonBinaryReadingValueUnchanged := nonBinaryReadingValueChanged
	firstBinaryReading := dtos.BaseReading{}
	firstBinaryReading.BinaryValue = []byte{1, 2, 3}
	binaryReadingValueChanged := firstBinaryReading
	binaryReadingValueChanged.BinaryValue = []byte{4, 5, 6}
	binaryReadingValueUnchanged := binaryReadingValueChanged

	tests := []struct {
		name     string
		reading  dtos.BaseReading
		expected bool
	}{
		{"false - lastReading is nil", firstNonBinaryReading, false},
		{"false - reading value changed", nonBinaryReadingValueChanged, false},
		{"true - reading value unchanged", nonBinaryReadingValueUnchanged, true},
		{"false - lastReading is not binary", firstBinaryReading, false},
		{"false - binary value changed", binaryReadingValueChanged, false},
		{"true - binary value unchanged", binaryReadingValueUnchanged, true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			res := e.compareReadings(testCase.reading)
			assert.Equal(t, testCase.expected, res, "compareReading result not as expected")
		})
	}
}
