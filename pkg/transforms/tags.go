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
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

// Tags contains the list of Tag key/values
type Tags struct {
	tags map[string]string
}

// NewTags creates, initializes and returns a new instance of Tags
func NewTags(tags map[string]string) Tags {
	return Tags{
		tags: tags,
	}
}

// AddTags adds the pre-configured list of tags to the Event's tags collection.
func (t *Tags) AddTags(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	ctx.LoggingClient().Debug("Adding tags to Event")

	if data == nil {
		return false, errors.New("no Event Received")
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	if len(t.tags) > 0 {
		if event.Tags == nil {
			event.Tags = make(map[string]string)
		}

		for tag, value := range t.tags {
			event.Tags[tag] = value
		}
		ctx.LoggingClient().Debugf("Tags added to Event. Event tags=%v", event.Tags)
	} else {
		ctx.LoggingClient().Debug("No tags added to Event. Add tags list is empty.")
	}

	return true, event
}
