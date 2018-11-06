// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"testing"
)

func TestBuildAddr(t *testing.T) {
	addr := BuildAddr("test.xyz", "8000")

	if addr != "http://test.xyz:8000" {
		t.Errorf("Expected 'http://test.xyz:8000' but got: %s", addr)
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

	if !CompareStrings(strings1, strings1) {
		t.Error("Equal slices fail check!")
	}

	if CompareStrings(strings1, strings2) {
		t.Error("Different size slices are OK!")
	}

	if CompareStrings(strings1, strings3) {
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

	if !CompareStrStrMap(map1, map1) {
		t.Error("Equal maps fail check")
	}

	if CompareStrStrMap(map1, map2) {
		t.Error("Different size maps are OK!")
	}

	if CompareStrStrMap(map1, map3) {
		t.Error("Maps with different content are OK!")
	}
}
