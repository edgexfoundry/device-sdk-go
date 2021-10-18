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

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/stretchr/testify/assert"
)

func TestPushToCore_ShouldFailPipelineOnError(t *testing.T) {
	coreData := NewCoreDataSimpleReading("MyProfile", "MyDevice", "MyResource", common.ValueTypeInt32)
	continuePipeline, result := coreData.PushToCoreData(ctx, "something")

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
}

func TestPushToCore_NoData(t *testing.T) {
	coreData := NewCoreDataSimpleReading("MyProfile", "MyDevice", "MyResource", common.ValueTypeInt32)
	continuePipeline, result := coreData.PushToCoreData(ctx, nil)

	assert.NotNil(t, result)
	assert.Contains(t, result.(error).Error(), "No Data Received")
	assert.False(t, continuePipeline)
}

func TestPushToCore_Object(t *testing.T) {
	myObject := struct {
		Name  string
		Value int32
	}{
		Name:  "my-object",
		Value: 1234,
	}

	coreData := NewCoreDataObjectReading("MyProfile", "MyDevice", "MyResource")
	continuePipeline, result := coreData.PushToCoreData(ctx, myObject)

	assert.NotNil(t, result)
	assert.True(t, continuePipeline)
	_, isError := result.(error)
	assert.False(t, isError)
}
