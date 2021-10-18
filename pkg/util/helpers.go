//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package util

import (
	"encoding/json"
	"errors"
	"strings"
)

//SplitComma - use custom split func, on commas, instead of .Split to eliminate empty values (i.e Test,,,)
func SplitComma(c rune) bool {
	return c == ','
}

//SplitColon - use custom split func, on colons, instead of .Split to eliminate empty values (i.e Test,,,)
func SplitColon(c rune) bool {
	return c == ':'
}

//DeleteEmptyAndTrim removes empty strings from a slice
func DeleteEmptyAndTrim(s []string) []string {
	var r []string
	for _, str := range s {
		str = strings.TrimSpace(str)
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

//CoerceType will accept a string, []byte, or json.Marshaller type and convert it to a []byte for use and consistency in the SDK
func CoerceType(param interface{}) ([]byte, error) {
	var data []byte
	var err error

	switch p := param.(type) {
	case string:
		input := p
		data = []byte(input)

	case []byte:
		data = p

	default:
		data, err = json.Marshal(param)
		if err != nil {
			return nil, errors.New(
				"marshaling input data to JSON failed, " +
					"passed in data must be of type []byte, string, or support marshaling to JSON",
			)
		}
	}

	return data, nil
}
