// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"testing"

	"github.com/edgexfoundry/device-sdk-go/mock"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// TODO:
//   TestCompareCommands
//   TestCompareDevices
//   TestCompareDeviceProfiles
//   TestCompareDeviceResources
//   TestCompareResources
//   TestCompareResourceOperations
//   TestCompareServices

func TestCompareStrings(t *testing.T) {
	strings1 := []string{"one", "two", "three"}
	strings2 := []string{"one", "two"}
	strings3 := []string{"one", "two", "THREE"}

	if !compareStrings(strings1, strings1) {
		t.Error("Equal slices fail check!")
	}

	if compareStrings(strings1, strings2) {
		t.Error("Different size slices are OK!")
	}

	if compareStrings(strings1, strings3) {
		t.Error("Slice with different strings are OK!")
	}
}

func TestCompareStrStrMap(t *testing.T) {
	map1 := map[string]string{
		"album":  "electric ladyland",
		"artist": "jimi hendrix",
		"guitar": "white strat",
	}

	map2 := map[string]string{
		"album":  "electric ladyland",
		"artist": "jimi hendrix",
	}

	map3 := map[string]string{
		"album":  "iv",
		"artist": "led zeppelin",
		"guitar": "les paul",
	}

	if !compareStrStrMap(map1, map1) {
		t.Error("Equal maps fail check")
	}

	if compareStrStrMap(map1, map2) {
		t.Error("Different size maps are OK!")
	}

	if compareStrStrMap(map1, map3) {
		t.Error("Maps with different content are OK!")
	}
}

func TestUpdateDevice(t *testing.T) {
	svc = &Service{}
	svc.ac = mock.AddressableClientMock{}
	svc.dc = &mock.DeviceClientMock{}
	devicesMap := map[string]*models.Device{
		"meter": {Name: "meter", AdminState: models.Locked, Addressable: models.Addressable{Name: "addressable-meter"}},
	}

	dc = &deviceCache{devices: devicesMap, names: map[string]string{}}

	device := models.Device{Name: "meter", AdminState: models.Unlocked, Addressable: models.Addressable{Name: "addressable-meter"}}
	err := dc.Update(&device)

	if err != nil {
		t.Fatal(err.Error())
	}

	if dc.Devices()["meter"].AdminState != models.Unlocked {
		t.Fatal("AdminState should be updated")
	}

}
