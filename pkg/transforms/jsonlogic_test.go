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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONLogicSimple(t *testing.T) {
	jsonLogic := NewJSONLogic(`{"==": [1, 1]}`)

	continuePipeline, result := jsonLogic.Evaluate(ctx, "{}")

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, "{}", result.(string))
}

func TestJSONLogicAdvanced(t *testing.T) {
	jsonLogic := NewJSONLogic(`{ "and" : [
		{"<" : [ { "var" : "temp" }, 110 ]},
		{"==" : [ { "var" : "sensor.type" }, "temperature" ] }
	  ] }`)

	data := `{ "temp" : 100, "sensor" : { "type" : "temperature" } }`
	continuePipeline, result := jsonLogic.Evaluate(ctx, data)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.JSONEq(t, data, result.(string))
}

func TestJSONLogicMalformedJSONRule(t *testing.T) {
	//missing quote
	jsonLogic := NewJSONLogic(`{"==: [1, 1]}`)

	continuePipeline, result := jsonLogic.Evaluate(ctx, `{}`)

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error))
}

func TestJSONLogicValidJSONBadRule(t *testing.T) {
	//missing quote
	jsonLogic := NewJSONLogic(`{"notAnOperator": [1, 1]}`)

	continuePipeline, result := jsonLogic.Evaluate(ctx, `{}`)

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
	require.IsType(t, fmt.Errorf(""), result)
	assert.Contains(t, result.(error).Error(), "unable to apply JSONLogic rule")
}

func TestJSONLogicNoData(t *testing.T) {
	//missing quote
	jsonLogic := NewJSONLogic(`{"notAnOperator": [1, 1]}`)

	continuePipeline, result := jsonLogic.Evaluate(ctx, nil)

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error))
}

func TestJSONLogicNonJSONData(t *testing.T) {
	//missing quote
	jsonLogic := NewJSONLogic(`{"==": [1, 1]}`)

	continuePipeline, result := jsonLogic.Evaluate(ctx, "iAmNotJson")

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error))
}
