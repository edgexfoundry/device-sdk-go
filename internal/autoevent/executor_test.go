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
	readings := make([]contract.Reading, 4)
	readings[0] = contract.Reading{Name: "Temperature", Value: "10"}
	readings[1] = contract.Reading{Name: "Humidity", Value: "50"}
	readings[2] = contract.Reading{Name: "Pressure", Value: "3"}
	readings[3] = contract.Reading{Name: "Image", BinaryValue: []byte("This is a image")}

	autoEvent := contract.AutoEvent{Frequency: "500ms"}
	e, err := NewExecutor("hasBinaryTrue", autoEvent)
	if err != nil {
		t.Errorf("Autoevent executor creation failed: %v", err)
	}
	resultFalse := compareReadings(e.(*executor), readings, true)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false in the first place")
	}

	readings[1] = contract.Reading{Name: "Humidity", Value: "51"}
	resultFalse = compareReadings(e.(*executor), readings, true)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false")
	}

	readings[3] = contract.Reading{Name: "Image", BinaryValue: []byte("This is not a image")}
	resultFalse = compareReadings(e.(*executor), readings, true)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false")
	}

	resultTrue := compareReadings(e.(*executor), readings, true)
	if !resultTrue {
		t.Error("compare readings with cache failed, the result should be true with unchanged readings")
	}

	e, err = NewExecutor("hasBinaryFalse", autoEvent)
	if err != nil {
		t.Errorf("Autoevent executor creation failed: %v", err)
	}
	// This scenario should not happen in real case
	resultFalse = compareReadings(e.(*executor), readings, false)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false in the first place")
	}

	readings[0] = contract.Reading{Name: "Temperature", Value: "20"}
	resultFalse = compareReadings(e.(*executor), readings, false)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false")
	}

	readings[3] = contract.Reading{Name: "Image", BinaryValue: []byte("This is a image")}
	resultTrue = compareReadings(e.(*executor), readings, false)
	if !resultTrue {
		t.Error("compare readings with cache failed, the result should always be true in such scenario")
	}

	resultTrue = compareReadings(e.(*executor), readings, false)
	if !resultTrue {
		t.Error("compare readings with cache failed, the result should be true with unchanged readings")
	}
}
