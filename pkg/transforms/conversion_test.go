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

package transforms

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformToXML(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: deviceName1,
	}
	expectedResult := `<Event><ApiVersion></ApiVersion><Id></Id><DeviceName>device1</DeviceName><ProfileName></ProfileName><SourceName></SourceName><Created>0</Created><Origin>0</Origin></Event>`
	conv := NewConversion()

	continuePipeline, result := conv.TransformToXML(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, clients.ContentTypeXML, context.ResponseContentType())
	assert.Equal(t, expectedResult, result.(string))
}

func TestTransformToXMLNoData(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context, nil)

	assert.Equal(t, "No Event Received", result.(error).Error())
	assert.False(t, continuePipeline)
}

func TestTransformToXMLNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context, "")

	assert.Equal(t, "Unexpected type received", result.(error).Error())
	assert.False(t, continuePipeline)

}

func TestTransformToJSON(t *testing.T) {
	// Event from device 1
	eventIn := dtos.Event{
		DeviceName: deviceName1,
	}
	expectedResult := `{"apiVersion":"","id":"","deviceName":"device1","profileName":"","sourceName":"","origin":0,"readings":null}`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, clients.ContentTypeJSON, context.ResponseContentType())
	assert.Equal(t, expectedResult, result.(string))
}

func TestTransformToJSONNoEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, nil)

	assert.Equal(t, "No Event Received", result.(error).Error())
	assert.False(t, continuePipeline)

}

func TestTransformToJSONNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, "")
	require.EqualError(t, result.(error), "Unexpected type received")
	assert.False(t, continuePipeline)
}
