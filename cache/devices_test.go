// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package cache

import "testing"

func TestNewDevices(t *testing.T) {

	d1 := NewDevices(nil)
	d2 := NewDevices(nil)
	if d1 != d2 {
		t.Error("Broken singleton")
	}
}

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
