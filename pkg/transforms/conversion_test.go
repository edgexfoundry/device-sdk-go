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

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var context *appcontext.Context

const (
	devID1        = "id1"
	devID2        = "id2"
	readingName1  = "sensor1"
	readingValue1 = "123.45"
)

func init() {
	lc := logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
	context = &appcontext.Context{
		LoggingClient: lc,
	}
}
func TestTransformToXML(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`
	conv := Conversion{}

	continuePipeline, result := conv.TransformToXML(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))
}
func TestTransformToXMLNoParameters(t *testing.T) {
	conv := Conversion{}
	continuePipeline, result := conv.TransformToXML(context)

	assert.Equal(t, "No Event Received", result.(error).Error())
	assert.False(t, continuePipeline)
}
func TestTransformToXMLNotAnEvent(t *testing.T) {
	conv := Conversion{}
	continuePipeline, result := conv.TransformToXML(context, "")

	assert.Equal(t, "Unexpected type received", result.(error).Error())
	assert.False(t, continuePipeline)

}
func TestTransformToXMLMultipleParametersValid(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`
	conv := Conversion{}
	continuePipeline, result := conv.TransformToXML(context, eventIn, "", "", "")
	if result == nil {
		t.Fatal("result should not be nil")
	}

	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))
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
	expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id2</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`
	conv := Conversion{}
	continuePipeline, result := conv.TransformToXML(context, eventIn2, eventIn1, "", "")

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

}

func TestTransformToJSON(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `{"device":"id1"}`
	conv := Conversion{}
	continuePipeline, result := conv.TransformToJSON(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))
}
func TestTransformToJSONNoEvent(t *testing.T) {
	conv := Conversion{}
	continuePipeline, result := conv.TransformToJSON(context)

	assert.Equal(t, "No Event Received", result.(error).Error())
	assert.False(t, continuePipeline)

}
func TestTransformToJSONNotAnEvent(t *testing.T) {
	conv := Conversion{}
	continuePipeline, result := conv.TransformToJSON(context, "")
	if result.(error).Error() != "Unexpected type received" {
		t.Fatal("Should have an error when wrong type was passed")
	}
	assert.False(t, continuePipeline)

}
func TestTransformToJSONMultipleParametersValid(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	expectedResult := `{"device":"id1"}`
	conv := Conversion{}
	continuePipeline, result := conv.TransformToJSON(context, eventIn, "", "", "")
	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

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
	continuePipeline, result := conv.TransformToJSON(context, eventIn2, eventIn1, "", "")

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

}
