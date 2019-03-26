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
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Filter houses various the parameters for which filter transforms filter on
type Filter struct {
	FilterValues []string
}

// FilterByDeviceID filters events received from CoreData based off the specified DeviceIDs.
// This function returns an Event
func (f Filter) FilterByDeviceID(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {

	edgexcontext.LoggingClient.Debug("Filter by DeviceID")

	if len(params) != 1 {
		return false, errors.New("No Event Received")
	}
	deviceIDs := f.FilterValues
	event := params[0].(models.Event)
	for _, devID := range deviceIDs {
		if event.Device == devID {
			// LoggingClient.Debug(fmt.Sprintf("Event accepted: %s", event.Device))
			return true, event
		}
	}
	return false, nil
	// fmt.Println(event.Data)

}

// FilterByValueDescriptor - filters events received from CoreData based on specified value descriptors
// This function returns an Event
func (f Filter) FilterByValueDescriptor(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {

	edgexcontext.LoggingClient.Debug("Filter by ValueDescriptorID")

	if len(params) != 1 {
		return false, errors.New("No Event Received")
	}

	existingEvent := params[0].(models.Event)
	auxEvent := models.Event{
		Pushed:   existingEvent.Pushed,
		Device:   existingEvent.Device,
		Created:  existingEvent.Created,
		Modified: existingEvent.Modified,
		Origin:   existingEvent.Origin,
		Readings: []models.Reading{},
	}

	for _, filterID := range f.FilterValues {
		for _, reading := range existingEvent.Readings {
			if reading.Name == filterID {
				// LoggingClient.Debug(fmt.Sprintf("Reading filtered: %s", reading.Name))
				auxEvent.Readings = append(auxEvent.Readings, reading)
			}
		}
	}
	thereExistReadings := len(auxEvent.Readings) > 0
	var returnResult models.Event
	if thereExistReadings {
		returnResult = auxEvent
	}
	return thereExistReadings, returnResult
}
