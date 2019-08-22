package transforms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkAsPushed(t *testing.T) {
	coreData := NewCoreData()

	continuePipeline, result := coreData.MarkAsPushed(context, "something")

	assert.NotNil(t, result)
	assert.Equal(t, "something", result)
	assert.True(t, continuePipeline)
}
