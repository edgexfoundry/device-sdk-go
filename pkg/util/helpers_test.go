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
	"reflect"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/stretchr/testify/assert"
)

func TestSplitComma(t *testing.T) {
	commaDelimited := "Hel lo,, ,Test,Hi, "
	results := strings.FieldsFunc(commaDelimited, SplitComma)
	// Should have 4 elements (space counts as an element)
	assert.Equal(t, 5, len(results))
	assert.Equal(t, "Hel lo", results[0])
	assert.Equal(t, " ", results[1])
	assert.Equal(t, "Test", results[2])
	assert.Equal(t, "Hi", results[3])
	assert.Equal(t, " ", results[4])
}
func TestSplitCommaEmpty(t *testing.T) {
	commaDelimited := ""
	results := strings.FieldsFunc(commaDelimited, SplitComma)
	// Should have 4 elements (space counts as an element)
	assert.Equal(t, 0, len(results))
}

func TestDeleteEmptyAndTrim(t *testing.T) {
	target := []string{" Hel lo", "test ", " "}
	results := DeleteEmptyAndTrim(target)
	// Should have 4 elements (space counts as an element)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "Hel lo", results[0])
	assert.Equal(t, "test", results[1])
}

func TestCoerceTypeStringToByteArray(t *testing.T) {
	myData := "" //string
	var expectedType []byte
	result, err := CoerceType(myData)
	assert.NoError(t, err)
	assert.IsType(t, reflect.TypeOf(expectedType), reflect.TypeOf(result))
}

func TestCoerceTypeByteArrayToByteArray(t *testing.T) {
	var myData []byte //string
	var expectedType []byte
	result, err := CoerceType(myData)
	assert.NoError(t, err)
	assert.IsType(t, reflect.TypeOf(expectedType), reflect.TypeOf(result))
}
func TestCoerceTypeJSONMarshallerToByteArray(t *testing.T) {
	myData := dtos.Event{
		DeviceName: "deviceId",
	}
	var expectedType []byte
	result, err := CoerceType(myData)
	assert.NoError(t, err)
	assert.IsType(t, reflect.TypeOf(expectedType), reflect.TypeOf(result))
}
func TestCoerceTypeNotSupportedToByteArray(t *testing.T) {
	// Channels are not marshalable to JSON and generate an error
	myData := make(chan int)
	var expectedType []byte
	result, err := CoerceType(myData)
	assert.Error(t, err)
	assert.IsType(t, reflect.TypeOf(expectedType), reflect.TypeOf(result))
}
