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
	"errors"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/urlclient/local"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var context *appcontext.Context

const (
	devID1 = "id1"
	devID2 = "id2"
)

func init() {
	lc := logger.NewMockClient()
	eventClient := coredata.NewEventClient(local.New("http://test" + clients.ApiEventRoute))
	mockSP := &mocks.SecretProvider{}
	mockSP.On("GetSecrets", "/path", "Secret-Header-Name").Return(map[string]string{"Secret-Header-Name": "value"}, nil)
	mockSP.On("GetSecrets", "/path", "Secret-Header-Name-2").Return(nil, errors.New("FAKE NOT FOUND ERROR"))

	context = &appcontext.Context{
		LoggingClient:  lc,
		EventClient:    eventClient,
		SecretProvider: mockSP,
	}
}

func TestTransformToXML(t *testing.T) {
	// Event from device 1
	eventIn := dtos.Event{
		DeviceName: devID1,
	}
	expectedResult := `<Event><ApiVersion></ApiVersion><Id></Id><DeviceName>id1</DeviceName><ProfileName></ProfileName><Created>0</Created><Origin>0</Origin></Event>`
	conv := NewConversion()

	continuePipeline, result := conv.TransformToXML(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, clients.ContentTypeXML, context.ResponseContentType)
	assert.Equal(t, expectedResult, result.(string))
}
func TestTransformToXMLNoParameters(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context)

	assert.Equal(t, "No Event Received", result.(error).Error())
	assert.False(t, continuePipeline)
}
func TestTransformToXMLNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context, "")

	assert.Equal(t, "Unexpected type received", result.(error).Error())
	assert.False(t, continuePipeline)

}
func TestTransformToXMLMultipleParametersValid(t *testing.T) {
	// Event from device 1
	eventIn := dtos.Event{
		DeviceName: devID1,
	}
	expectedResult := `<Event><ApiVersion></ApiVersion><Id></Id><DeviceName>id1</DeviceName><ProfileName></ProfileName><Created>0</Created><Origin>0</Origin></Event>`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context, eventIn, "", "", "")
	require.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))
}
func TestTransformToXMLMultipleParametersTwoEvents(t *testing.T) {
	// Event from device 1
	eventIn1 := dtos.Event{
		DeviceName: devID1,
	}
	// Event from device 1
	eventIn2 := dtos.Event{
		DeviceName: devID2,
	}
	expectedResult := `<Event><ApiVersion></ApiVersion><Id></Id><DeviceName>id2</DeviceName><ProfileName></ProfileName><Created>0</Created><Origin>0</Origin></Event>`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToXML(context, eventIn2, eventIn1, "", "")

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

}

func TestTransformToJSON(t *testing.T) {
	// Event from device 1
	eventIn := dtos.Event{
		DeviceName: devID1,
	}
	expectedResult := `{"id":"","deviceName":"id1","profileName":"","created":0,"origin":0,"readings":null}`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, eventIn)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, clients.ContentTypeJSON, context.ResponseContentType)
	assert.Equal(t, expectedResult, result.(string))
}
func TestTransformToJSONNoEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context)

	assert.Equal(t, "No Event Received", result.(error).Error())
	assert.False(t, continuePipeline)

}
func TestTransformToJSONNotAnEvent(t *testing.T) {
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, "")
	require.EqualError(t, result.(error), "Unexpected type received")
	assert.False(t, continuePipeline)

}
func TestTransformToJSONMultipleParametersValid(t *testing.T) {
	// Event from device 1
	eventIn := dtos.Event{
		DeviceName: devID1,
	}
	expectedResult := `{"id":"","deviceName":"id1","profileName":"","created":0,"origin":0,"readings":null}`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, eventIn, "", "", "")
	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

}
func TestTransformToJSONMultipleParametersTwoEvents(t *testing.T) {
	// Event from device 1
	eventIn1 := dtos.Event{
		DeviceName: devID1,
	}
	// Event from device 2
	eventIn2 := dtos.Event{
		DeviceName: devID2,
	}
	expectedResult := `{"id":"","deviceName":"id2","profileName":"","created":0,"origin":0,"readings":null}`
	conv := NewConversion()
	continuePipeline, result := conv.TransformToJSON(context, eventIn2, eventIn1, "", "")

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, expectedResult, result.(string))

}
