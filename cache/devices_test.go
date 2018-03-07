// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
// TODO: one file in each package should contain a doc comment for the package.
// Alternatively, this doc can go in a file called doc.go.
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
