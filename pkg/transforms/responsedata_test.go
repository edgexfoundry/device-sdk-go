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
	"encoding/json"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetResponseDataString(t *testing.T) {
	expected := getExpectedEventXml(t)
	target := NewResponseData()

	continuePipeline, result := target.SetResponseData(ctx, expected)

	assert.True(t, continuePipeline)
	assert.NotNil(t, result)

	actual := string(ctx.ResponseData())
	assert.Equal(t, expected, actual)
}

func TestSetResponseDataBytes(t *testing.T) {
	expected := []byte(getExpectedEventXml(t))
	target := NewResponseData()

	continuePipeline, result := target.SetResponseData(ctx, expected)
	assert.True(t, continuePipeline)
	assert.NotNil(t, result)

	actual := string(ctx.ResponseData())
	assert.Equal(t, string(expected), actual)
}

func TestSetResponseDataEvent(t *testing.T) {
	target := NewResponseData()

	eventIn := dtos.Event{
		DeviceName: deviceName1,
	}

	expected, _ := json.Marshal(eventIn)

	continuePipeline, result := target.SetResponseData(ctx, eventIn)
	assert.True(t, continuePipeline)
	assert.NotNil(t, result)

	actual := string(ctx.ResponseData())
	assert.Equal(t, string(expected), actual)
}

func TestSetResponseDataNoData(t *testing.T) {
	target := NewResponseData()
	continuePipeline, result := target.SetResponseData(ctx, nil)
	assert.Contains(t, result.(error).Error(), "No Data Received")
	assert.False(t, continuePipeline)
}

func TestSetResponseDataBadType(t *testing.T) {
	target := NewResponseData()

	// Channels are not marshalable to JSON and generate an error
	continuePipeline, result := target.SetResponseData(ctx, make(chan int))
	assert.False(t, continuePipeline)
	require.NotNil(t, result)
	assert.Contains(t, result.(error).Error(), "passed in data must be of type")
}

func getExpectedEventXml(t *testing.T) string {
	event := dtos.NewEvent("profile1", "dev1", "source1")
	err := event.AddSimpleReading("resource1", common.ValueTypeInt32, int32(32))
	require.NoError(t, err)

	xml, err := event.ToXML()
	require.NoError(t, err)
	return xml
}
