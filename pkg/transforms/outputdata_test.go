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
	"encoding/json"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetOutputDataString(t *testing.T) {
	expected := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`
	target := NewOutputData()

	continuePipeline, result := target.SetOutputData(context, expected)

	assert.True(t, continuePipeline)
	assert.NotNil(t, result)

	actual := string(context.OutputData)
	assert.Equal(t, expected, actual)
}

func TestSetOutputDataBytes(t *testing.T) {
	var expected []byte
	expected = []byte(`<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`)
	target := NewOutputData()

	continuePipeline, result := target.SetOutputData(context, expected)
	assert.True(t, continuePipeline)
	assert.NotNil(t, result)

	actual := string(context.OutputData)
	assert.Equal(t, string(expected), actual)
}

func TestSetOutputDataEvent(t *testing.T) {
	target := NewOutputData()

	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	expected, _ := json.Marshal(eventIn)

	continuePipeline, result := target.SetOutputData(context, eventIn)
	assert.True(t, continuePipeline)
	assert.NotNil(t, result)

	actual := string(context.OutputData)
	assert.Equal(t, string(expected), actual)
}

func TestSetOutputDataNoData(t *testing.T) {
	target := NewOutputData()
	continuePipeline, result := target.SetOutputData(context)
	assert.Nil(t, result)
	assert.False(t, continuePipeline)
}

func TestSetOutputDataMultipleParametersValid(t *testing.T) {
	expected := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin></Event>`
	target := NewOutputData()

	continuePipeline, result := target.SetOutputData(context, expected, "", "", "")
	assert.True(t, continuePipeline)
	assert.NotNil(t, result)

	actual := string(context.OutputData)
	assert.Equal(t, expected, actual)
}

func TestSetOutputDataBadType(t *testing.T) {
	target := NewOutputData()

	// Channels are not marshalable to JSON and generate an error
	continuePipeline, result := target.SetOutputData(context, make(chan int))
	assert.False(t, continuePipeline)
	require.NotNil(t, result)
	assert.Contains(t, result.(error).Error(), "passed in data must be of type")
}
