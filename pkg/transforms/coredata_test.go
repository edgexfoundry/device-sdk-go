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

	"github.com/stretchr/testify/assert"
)

func TestPushToCore_ShouldFailPipelineOnError(t *testing.T) {
	coreData := NewCoreData()
	coreData.DeviceName = "my-device"
	coreData.ReadingName = "my-device-resource"
	continuePipeline, result := coreData.PushToCoreData(context, "something")

	assert.NotNil(t, result)
	assert.False(t, continuePipeline)
}

func TestPushToCore_NoData(t *testing.T) {
	coreData := NewCoreData()
	coreData.DeviceName = "my-device"
	coreData.ReadingName = "my-device-resource"
	continuePipeline, result := coreData.PushToCoreData(context, nil)

	assert.NotNil(t, result)
	assert.Equal(t, "No Data Received", result.(error).Error())
	assert.False(t, continuePipeline)
}
