// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestCompareReadings(t *testing.T) {
	readings := make([]models.Reading, 3)
	readings[0] = models.Reading{Name: "Temperature", Value: "10"}
	readings[1] = models.Reading{Name: "Humidity", Value: "50"}
	readings[2] = models.Reading{Name: "Pressure", Value: "3"}

	autoEvent := models.AutoEvent{Frequency: "500ms"}
	e, _ := NewExecutor("meter", autoEvent)
	cacheReadings(e.(*executor), readings)
	resultTrue := compareReadings(e.(*executor), readings)
	if !resultTrue {
		t.Error("compare reading with cache failed, the result should be true")
	}

	readings[1] = models.Reading{Name: "Humidity", Value: "51"}
	resultFalse := compareReadings(e.(*executor), readings)
	if resultFalse {
		t.Error("compare reading with cache failed, the result should be false")
	}
}
