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

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

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
	continuePipeline, result := conv.TransformToXML(excontext.Context{}, eventIn)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
	}
	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}
func TestTransformToXMLNoParameters(t *testing.T) {
	conv := Conversion{}
	continuePipeline, result := conv.TransformToXML(excontext.Context{})
	if result.(error).Error() != "No Event Received" {
		t.Fatal("result should be an error that says \"No Event Received\"")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}
func TestTransformToXMLNotAnEvent(t *testing.T) {
	conv := Conversion{}
	continuePipeline, result := conv.TransformToXML(excontext.Context{}, "")
	if result.(error).Error() != "Unexpected type received" {
		t.Fatal("result should be an error that says \"Unexpected type received\"")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}
func TestTransformToXMLMultipleParametersValid(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin><Event></Event></Event>`
	conv := Conversion{}
	continuePipeline, result := conv.TransformToXML(excontext.Context{}, eventIn, "", "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
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
	continuePipeline, result := conv.TransformToXML(excontext.Context{}, eventIn2, eventIn1, "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
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
	continuePipeline, result := conv.TransformToJSON(excontext.Context{}, eventIn)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
	}
	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}
func TestTransformToJSONNoEvent(t *testing.T) {
	conv := Conversion{}
	continuePipeline, result := conv.TransformToJSON(excontext.Context{})
	if result.(error).Error() != "No Event Received" {
		t.Fatal("result should be an error that says \"No Event Received\"")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}
func TestTransformToJSONNotAnEvent(t *testing.T) {
	conv := Conversion{}
	continuePipeline, result := conv.TransformToJSON(excontext.Context{}, "")
	if result.(error).Error() != "Unexpected type received" {
		t.Fatal("Should have an error when wrong type was passed")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}
func TestTransformToJSONMultipleParametersValid(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `{"device":"id1"}`
	conv := Conversion{}
	continuePipeline, result := conv.TransformToJSON(excontext.Context{}, eventIn, "", "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
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
	continuePipeline, result := conv.TransformToJSON(excontext.Context{}, eventIn2, eventIn1, "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
	}
	if result.(string) != expectedResult {
		t.Fatal("result does not match expectedResult")
	}
}
