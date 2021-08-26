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

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

// Filter houses various the parameters for which filter transforms filter on
type Filter struct {
	FilterValues []string
	FilterOut    bool
	ctx          interfaces.AppFunctionContext
}

// NewFilterFor creates, initializes and returns a new instance of Filter
// that defaults FilterOut to false, so it is filtering for specified values
func NewFilterFor(filterValues []string) Filter {
	return Filter{FilterValues: filterValues, FilterOut: false}
}

// NewFilterOut creates, initializes and returns a new instance of Filter
// that defaults FilterOut to true, so it is filtering out specified values
func NewFilterOut(filterValues []string) Filter {
	return Filter{FilterValues: filterValues, FilterOut: true}
}

// FilterByProfileName filters based on the specified Device Profile, aka Class of Device.
// If FilterOut is false, it filters out those Events not associated with the specified Device Profile listed in FilterValues.
// If FilterOut is true, it out those Events that are associated with the specified Device Profile listed in FilterValues.
func (f Filter) FilterByProfileName(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	f.ctx = ctx
	event, err := f.setupForFiltering("FilterByProfileName", "ProfileName", ctx.LoggingClient(), data)
	if err != nil {
		return false, err
	}

	ok := f.doEventFilter("ProfileName", event.ProfileName, ctx.LoggingClient())
	if ok {
		return true, *event
	}

	return false, nil

}

// FilterByDeviceName filters based on the specified Device Names, aka Instance of a Device.
// If FilterOut is false, it filters out those Events not associated with the specified Device Names listed in FilterValues.
// If FilterOut is true, it out those Events that are associated with the specified Device Names listed in FilterValues.
func (f Filter) FilterByDeviceName(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	f.ctx = ctx
	event, err := f.setupForFiltering("FilterByDeviceName", "DeviceName", ctx.LoggingClient(), data)
	if err != nil {
		return false, err
	}

	ok := f.doEventFilter("DeviceName", event.DeviceName, ctx.LoggingClient())
	if ok {
		return true, *event
	}

	return false, nil
}

// FilterBySourceName filters based on the specified Source for the Event, aka resource or command name.
// If FilterOut is false, it filters out those Events not associated with the specified Source listed in FilterValues.
// If FilterOut is true, it out those Events that are associated with the specified Source listed in FilterValues.
func (f Filter) FilterBySourceName(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	f.ctx = ctx
	event, err := f.setupForFiltering("FilterBySourceName", "SourceName", ctx.LoggingClient(), data)
	if err != nil {
		return false, err
	}

	ok := f.doEventFilter("SourceName", event.SourceName, ctx.LoggingClient())
	if ok {
		return true, *event
	}

	return false, nil
}

// FilterByResourceName filters based on the specified Reading resource names, aka Instance of a Device.
// If FilterOut is false, it filters out those Event Readings not associated with the specified Resource Names listed in FilterValues.
// If FilterOut is true, it out those Event Readings that are associated with the specified Resource Names listed in FilterValues.
// This function will return an error and stop the pipeline if a non-edgex event is received or if no data is received.
func (f Filter) FilterByResourceName(ctx interfaces.AppFunctionContext, data interface{}) (continuePipeline bool, result interface{}) {
	f.ctx = ctx
	existingEvent, err := f.setupForFiltering("FilterByResourceName", "ResourceName", ctx.LoggingClient(), data)
	if err != nil {
		return false, err
	}

	// No filter values, so pass all event and all readings through, rather than filtering them all out.
	if len(f.FilterValues) == 0 {
		return true, *existingEvent
	}

	// Create copy of Event which will contain any Reading that are not filtered out
	auxEvent := dtos.NewEvent(existingEvent.ProfileName, existingEvent.DeviceName, existingEvent.SourceName)
	auxEvent.Id = existingEvent.Id
	auxEvent.Origin = existingEvent.Origin
	auxEvent.Readings = []dtos.BaseReading{}

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
				ctx.LoggingClient().Debugf("Reading accepted in pipeline '%s' for resource %s", f.ctx.PipelineId(), reading.ResourceName)
				auxEvent.Readings = append(auxEvent.Readings, reading)
			} else {
				ctx.LoggingClient().Debugf("Reading not accepted in pipeline '%s' for resource %s", f.ctx.PipelineId(), reading.ResourceName)
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
				ctx.LoggingClient().Debugf("Reading accepted in pipeline '%s' for resource %s", f.ctx.PipelineId(), reading.ResourceName)
				auxEvent.Readings = append(auxEvent.Readings, reading)
			} else {
				ctx.LoggingClient().Debugf("Reading not accepted in pipeline '%s' for resource %s", f.ctx.PipelineId(), reading.ResourceName)
			}
		}
	}

	if len(auxEvent.Readings) > 0 {
		ctx.LoggingClient().Debugf("Event accepted: %d remaining reading(s) in pipeline '%s'", len(auxEvent.Readings), f.ctx.PipelineId())
		return true, auxEvent
	}

	ctx.LoggingClient().Debugf("Event not accepted: 0 remaining readings in pipeline '%s'", f.ctx.PipelineId())
	return false, nil
}

func (f Filter) setupForFiltering(funcName string, filterProperty string, lc logger.LoggingClient, data interface{}) (*dtos.Event, error) {
	mode := "For"
	if f.FilterOut {
		mode = "Out"
	}
	lc.Debugf("Filtering %s by %s in. FilterValues are: '[%v]'", mode, filterProperty, f.FilterValues)

	if data == nil {
		return nil, fmt.Errorf("%s: no Event Received in pipeline '%s'", funcName, f.ctx.PipelineId())
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return nil, fmt.Errorf("%s: type received is not an Event in pipeline '%s'", funcName, f.ctx.PipelineId())
	}

	return &event, nil
}

func (f Filter) doEventFilter(filterProperty string, value string, lc logger.LoggingClient) bool {
	// No names to filter for, so pass events through rather than filtering them all out.
	if len(f.FilterValues) == 0 {
		return true
	}

	for _, name := range f.FilterValues {
		if value == name {
			if f.FilterOut {
				lc.Debugf("Event not accepted for %s=%s in pipeline '%s'", filterProperty, value, f.ctx.PipelineId())
				return false
			} else {
				lc.Debugf("Event accepted for %s=%s in pipeline '%s'", filterProperty, value, f.ctx.PipelineId())
				return true
			}
		}
	}

	// Will only get here if Event's SourceName didn't match any names in FilterValues
	if f.FilterOut {
		lc.Debugf("Event accepted for %s=%s in pipeline '%s'", filterProperty, value, f.ctx.PipelineId())
		return true
	} else {
		lc.Debugf("Event not accepted for %s=%s in pipeline '%s'", filterProperty, value, f.ctx.PipelineId())
		return false
	}
}
