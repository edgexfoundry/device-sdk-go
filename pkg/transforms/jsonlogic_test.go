package transforms

import (
	"testing"

	jlogic "github.com/diegoholiveira/jsonlogic"
	"github.com/stretchr/testify/assert"
)

func TestJSONLogicSimple(t *testing.T) {
	jsonlogic := NewJSONLogic(`{"==": [1, 1]}`)

	continuePipeline, result := jsonlogic.Evaluate(context, "{}")

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.Equal(t, "{}", result.(string))
}

func TestJSONLogicAdvanced(t *testing.T) {
	jsonlogic := NewJSONLogic(`{ "and" : [
		{"<" : [ { "var" : "temp" }, 110 ]},
		{"==" : [ { "var" : "sensor.type" }, "temperature" ] }
	  ] }`)

	data := `{ "temp" : 100, "sensor" : { "type" : "temperature" } }`
	continuePipeline, result := jsonlogic.Evaluate(context, data)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	assert.JSONEq(t, data, result.(string))
}

func TestJSONLogicMalformedJSONRule(t *testing.T) {
	//missing quote
	jsonlogic := NewJSONLogic(`{"==: [1, 1]}`)

	continuePipeline, result := jsonlogic.Evaluate(context, `{}`)

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error))
}

func TestJSONLogicValidJSONBadRule(t *testing.T) {
	//missing quote
	jsonlogic := NewJSONLogic(`{"notanoperator": [1, 1]}`)

	continuePipeline, result := jsonlogic.Evaluate(context, `{}`)

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
	assert.Equal(t, "The operator \"notanoperator\" is not supported", result.(jlogic.ErrInvalidOperator).Error())
}

func TestJSONLogicNoData(t *testing.T) {
	//missing quote
	jsonlogic := NewJSONLogic(`{"notanoperator": [1, 1]}`)

	continuePipeline, result := jsonlogic.Evaluate(context)

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error))
}

func TestJSONLogicNonJSONData(t *testing.T) {
	//missing quote
	jsonlogic := NewJSONLogic(`{"==": [1, 1]}`)

	continuePipeline, result := jsonlogic.Evaluate(context, "iamnotjson")

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
	assert.Error(t, result.(error))
}
