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

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
)

// Filter houses various the parameters for which filter transforms filter on
type Filter struct {
	FilterValues []string
	FilterOut    bool
}

// NewFilterFor creates, initializes and returns a new instance of Filter
// that defaults FilterOut to false so it is filtering for specified values
func NewFilterFor(filterValues []string) Filter {
	return Filter{FilterValues: filterValues, FilterOut: false}
}

// NewFilterOut creates, initializes and returns a new instance of Filter
// that defaults FilterOut to ture so it is filtering out specified values
func NewFilterOut(filterValues []string) Filter {
	return Filter{FilterValues: filterValues, FilterOut: true}
}

// FilterByProfileName filters based on the specified Device Profile, aka Class of Device.
// If FilterOut is false, it filters out those Events not associated with the specified Device Profile listed in FilterValues.
// If FilterOut is true, it out those Events that are associated with the specified Device Profile listed in FilterValues.
func (f Filter) FilterByProfileName(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {
	mode := "For"
	if f.FilterOut {
		mode = "Out"
	}
	edgexcontext.LoggingClient.Debugf("Filtering %s by Device Profile. FilterValues are: '[%v]'", mode, f.FilterValues)

	if len(params) < 1 {
		return false, errors.New("no Event Received")
	}

	event, ok := params[0].(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	// No names to filter for, so pass events thru rather than filtering them all out.
	if len(f.FilterValues) == 0 {
		return true, event
	}

	for _, name := range f.FilterValues {
		if event.ProfileName == name {
			if f.FilterOut {
				edgexcontext.LoggingClient.Debugf("Event not accepted for Profile=%s", event.ProfileName)
				return false, nil
			} else {
				edgexcontext.LoggingClient.Debugf("Event accepted for Profile=%s", event.ProfileName)
				return true, event
			}
		}
	}

	// Will only get here if Event's ProfileName didn't match any names in FilterValues
	if f.FilterOut {
		edgexcontext.LoggingClient.Debugf("Event accepted for Profile=", event.ProfileName)
		return true, event
	} else {
		edgexcontext.LoggingClient.Debugf("Event not accepted for Profile=%s", event.ProfileName)
		return false, nil
	}
}

// FilterByDeviceName filters based on the specified Device Names, aka Instance of a Device.
// If FilterOut is false, it filters out those Events not associated with the specified Device Names listed in FilterValues.
// If FilterOut is true, it out those Events that are associated with the specified Device Names listed in FilterValues.
func (f Filter) FilterByDeviceName(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {
	mode := "For"
	if f.FilterOut {
		mode = "Out"
	}
	edgexcontext.LoggingClient.Debugf("Filtering %s by Device Name. FilterValues are: '[%v]'", mode, f.FilterValues)

	if len(params) < 1 {
		return false, errors.New("no Event Received")
	}

	event, ok := params[0].(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	// No names to filter for, so pass events thru rather than filtering them all out.
	if len(f.FilterValues) == 0 {
		return true, event
	}

	for _, name := range f.FilterValues {
		if event.DeviceName == name {
			if f.FilterOut {
				edgexcontext.LoggingClient.Debugf("Event not accepted for Device=%s", event.DeviceName)
				return false, nil
			} else {
				edgexcontext.LoggingClient.Debugf("Event accepted for Device=%s", event.DeviceName)
				return true, event
			}
		}
	}

	// Will only get here if Event's DeviceName didn't match any names in FilterValues
	if f.FilterOut {
		edgexcontext.LoggingClient.Debugf("Event accepted for Device=", event.DeviceName)
		return true, event
	} else {
		edgexcontext.LoggingClient.Debugf("Event not accepted for Device=%s", event.DeviceName)
		return false, nil
	}
}

// FilterByResourceName filters based on the specified Reading resource names, aka Instance of a Device.
// If FilterOut is false, it filters out those Event Readings not associated with the specified Resource Names listed in FilterValues.
// If FilterOut is true, it out those Event Readings that are associated with the specified Resource Names listed in FilterValues.
// This function will return an error and stop the pipeline if a non-edgex event is received or if no data is received.
func (f Filter) FilterByResourceName(edgexcontext *appcontext.Context, params ...interface{}) (continuePipeline bool, result interface{}) {
	mode := "For"
	if f.FilterOut {
		mode = "Out"
	}
	edgexcontext.LoggingClient.Debugf("Filtering %s by Resource Name. FilterValues are: '[%v]'", mode, f.FilterValues)

	if len(params) < 1 {
		return false, errors.New("no Event Received")
	}

	existingEvent, ok := params[0].(dtos.Event)
	if !ok {
		return false, errors.New("type received is not an Event")
	}

	// No filter values, so pass all event and all readings thru, rather than filtering them all out.
	if len(f.FilterValues) == 0 {
		return true, existingEvent
	}

	auxEvent := dtos.Event{
		DeviceName: existingEvent.DeviceName,
		Created:    existingEvent.Created,
		Origin:     existingEvent.Origin,
		Readings:   []dtos.BaseReading{},
	}

	if f.FilterOut {
		for _, reading := range existingEvent.Readings {
			readingFilteredOut := false
			for _, name := range f.FilterValues {
				if reading.ResourceName == name {
					readingFilteredOut = true
					break
				}
			}

			if !readingFilteredOut {
				edgexcontext.LoggingClient.Debugf("Reading accepted: %s", reading.ResourceName)
				auxEvent.Readings = append(auxEvent.Readings, reading)
			} else {
				edgexcontext.LoggingClient.Debugf("Reading not accepted: %s", reading.ResourceName)
			}
		}
	} else {
		for _, reading := range existingEvent.Readings {
			readingFilteredFor := false
			for _, name := range f.FilterValues {
				if reading.ResourceName == name {
					readingFilteredFor = true
					break
				}
			}

			if readingFilteredFor {
				edgexcontext.LoggingClient.Debugf("Reading accepted: %s", reading.ResourceName)
				auxEvent.Readings = append(auxEvent.Readings, reading)
			} else {
				edgexcontext.LoggingClient.Debugf("Reading not accepted: %s", reading.ResourceName)
			}
		}
	}

	if len(auxEvent.Readings) > 0 {
		edgexcontext.LoggingClient.Debugf("Event accepted: %d remaining reading(s)", len(auxEvent.Readings))
		return true, auxEvent
	}

	edgexcontext.LoggingClient.Debug("Event not accepted: 0 remaining readings")
	return false, nil
}
