// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"testing"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestCompareReadings(t *testing.T) {
	readings := make([]contract.Reading, 3)
	readings[0] = contract.Reading{Name: "Temperature", Value: "10"}
	readings[1] = contract.Reading{Name: "Humidity", Value: "50"}
	readings[2] = contract.Reading{Name: "Pressure", Value: "3"}

	autoEvent := contract.AutoEvent{Frequency: "500ms"}
	e, _ := NewExecutor("meter", autoEvent)
	cacheReadings(e.(*executor), readings)
	resultTrue := compareReadings(e.(*executor), readings)
	if !resultTrue {
		t.Error("compare reading with cache failed, the result should be true")
	}

	readings[1] = contract.Reading{Name: "Humidity", Value: "51"}
	resultFalse := compareReadings(e.(*executor), readings)
	if resultFalse {
		t.Error("compare reading with cache failed, the result should be false")
	}
}
