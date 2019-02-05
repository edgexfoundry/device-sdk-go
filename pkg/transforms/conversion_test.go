//
// Copyright (c) 2019 Intel Corporation
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

package transforms

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	devID1        = "id1"
	devID2        = "id2"
	readingName1  = "sensor1"
	readingValue1 = "123.45"
)

func TestTransformToXML(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin><Event></Event></Event>`
	conv := Conversion{}
	result := conv.TransformToXML(&eventIn)
	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}
func TestTransformToXMLNoParameters(t *testing.T) {
	conv := Conversion{}
	result := conv.TransformToXML()
	if result != nil {
		t.Fatal("result should be nil")
	}
}
func TestTransformToXMLNotAnEvent(t *testing.T) {
	conv := Conversion{}
	result := conv.TransformToXML("")
	if result != nil {
		t.Fatal("result should be nil")
	}
}
func TestTransformToXMLMultipleParametersNotValid(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin><Event></Event></Event>`
	conv := Conversion{}
	result := conv.TransformToXML(&eventIn, "", "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}
func TestTransformToXMLMultipleParametersTwoEvents(t *testing.T) {
	// Event from device 1
	eventIn1 := models.Event{
		Device: devID1,
	}
	// Event from device 1
	eventIn2 := models.Event{
		Device: devID2,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id2</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin><Event></Event></Event>`
	conv := Conversion{}
	result := conv.TransformToXML(&eventIn2, &eventIn1, "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}

func TestTransformToJSON(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `{"device":"id1"}`
	conv := Conversion{}
	result := conv.TransformToJSON(&eventIn)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}
func TestTransformToJSONNoEvent(t *testing.T) {
	conv := Conversion{}
	result := conv.TransformToJSON()
	if result != nil {
		t.Fatal("result should be nil")
	}
}
func TestTransformToJSONNotAnEvent(t *testing.T) {
	conv := Conversion{}
	result := conv.TransformToJSON("")
	if result != nil {
		t.Fatal("result should be nil")
	}
}
func TestTransformToJSONMultipleParametersNotValid(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `{"device":"id1"}`
	conv := Conversion{}
	result := conv.TransformToJSON(&eventIn, "", "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}
func TestTransformToJSONMultipleParametersTwoEvents(t *testing.T) {
	// Event from device 1
	eventIn1 := models.Event{
		Device: devID1,
	}
	// Event from device 2
	eventIn2 := models.Event{
		Device: devID2,
	}
	expectedResult := `{"device":"id2"}`
	conv := Conversion{}
	result := conv.TransformToJSON(&eventIn2, &eventIn1, "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}
