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

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

// Tags contains the list of Tag key/values
type Tags struct {
	tags map[string]interface{}
}

// NewTags creates, initializes and returns a new instance of Tags using string values
// This factory method is Deprecated. Use NewGenericTags which allows generic interface values
func NewTags(tags map[string]string) Tags {
	newTags := Tags{
		tags: make(map[string]interface{}),
	}

	for tag, value := range tags {
		newTags.tags[tag] = value
	}

	return newTags
}

// NewGenericTags creates, initializes and returns a new instance of Tags using generic interface values
func NewGenericTags(tags map[string]interface{}) Tags {
	return Tags{
		tags: tags,
	}
}

// AddTags adds the pre-configured list of tags to the Event's tags collection.
func (t *Tags) AddTags(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	ctx.LoggingClient().Debugf("Adding tags to Event in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		return false, fmt.Errorf("function AddTags in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, fmt.Errorf("function AddTags in pipeline '%s', type received is not an Event", ctx.PipelineId())
	}

	if len(t.tags) > 0 {
		if event.Tags == nil {
			event.Tags = make(map[string]interface{})
		}

		for tag, value := range t.tags {
			event.Tags[tag] = value
		}
		ctx.LoggingClient().Debugf("Tags added to Event in pipeline '%s'. Event tags=%v", ctx.PipelineId(), event.Tags)
	} else {
		ctx.LoggingClient().Debugf("No tags added to Event in pipeline '%s'. Add tags list is empty.", ctx.PipelineId())
	}

	return true, event
}
