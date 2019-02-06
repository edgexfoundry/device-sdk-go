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
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Filter houses various built in filter transforms
type Filter struct {
	DeviceIDs []string
}

// FilterByDeviceID ...
func (f Filter) FilterByDeviceID(params ...interface{}) interface{} {

	println("FILTER BY DEVICEID")

	if len(params) != 1 {
		return nil
	}
	deviceIDs := f.DeviceIDs
	event := params[0].(models.Event)

	for _, devID := range deviceIDs {
		if event.Device == devID {
			// LoggingClient.Debug(fmt.Sprintf("Event accepted: %s", event.Device))
			return event
		}
	}
	return nil
	// fmt.Println(event.Data)
	// edgexcontext.Complete("")
}
