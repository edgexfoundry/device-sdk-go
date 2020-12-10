// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestCompareReadings(t *testing.T) {
	readings := make([]contract.Reading, 4)
	readings[0] = contract.Reading{Name: "Temperature", Value: "10"}
	readings[1] = contract.Reading{Name: "Humidity", Value: "50"}
	readings[2] = contract.Reading{Name: "Pressure", Value: "3"}
	readings[3] = contract.Reading{Name: "Image", BinaryValue: []byte("This is a image")}

	lc := logger.NewClientStdOut("device-sdk-test", false, "DEBUG")
	autoEvent := contract.AutoEvent{Frequency: "500ms"}
	e, err := NewExecutor("hasBinaryTrue", autoEvent)
	if err != nil {
		t.Errorf("Autoevent executor creation failed: %v", err)
	}
	resultFalse := compareReadings(e, readings, true, lc)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false in the first place")
	}

	readings[1] = contract.Reading{Name: "Humidity", Value: "51"}
	resultFalse = compareReadings(e, readings, true, lc)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false")
	}

	readings[3] = contract.Reading{Name: "Image", BinaryValue: []byte("This is not a image")}
	resultFalse = compareReadings(e, readings, true, lc)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false")
	}

	resultTrue := compareReadings(e, readings, true, lc)
	if !resultTrue {
		t.Error("compare readings with cache failed, the result should be true with unchanged readings")
	}

	e, err = NewExecutor("hasBinaryFalse", autoEvent)
	if err != nil {
		t.Errorf("Autoevent executor creation failed: %v", err)
	}
	// This scenario should not happen in real case
	resultFalse = compareReadings(e, readings, false, lc)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false in the first place")
	}

	readings[0] = contract.Reading{Name: "Temperature", Value: "20"}
	resultFalse = compareReadings(e, readings, false, lc)
	if resultFalse {
		t.Error("compare readings with cache failed, the result should be false")
	}

	readings[3] = contract.Reading{Name: "Image", BinaryValue: []byte("This is a image")}
	resultTrue = compareReadings(e, readings, false, lc)
	if !resultTrue {
		t.Error("compare readings with cache failed, the result should always be true in such scenario")
	}

	resultTrue = compareReadings(e, readings, false, lc)
	if !resultTrue {
		t.Error("compare readings with cache failed, the result should be true with unchanged readings")
	}
}
