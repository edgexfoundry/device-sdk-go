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

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	descriptor1 = "Descriptor1"
	descriptor2 = "Descriptor2"
	descriptor3 = "Descriptor3"
)

func TestFilterByDeviceNameFound(t *testing.T) {
	// Event from DeviceName 1
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	filter := NewFilter([]string{"id1"})

	continuePipeline, result := filter.FilterByDeviceName(context, eventIn)
	require.NotNil(t, result)
	require.True(t, continuePipeline)

	if eventOut, ok := result.(*dtos.Event); ok {
		assert.Equal(t, "id1", eventOut.DeviceName)
	}
}

func TestFilterOutByDeviceNameFound(t *testing.T) {
	// Event from DeviceName 1
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	filter := NewFilter([]string{devID1})
	filter.FilterOut = true

	continuePipeline, result := filter.FilterByDeviceName(context, eventIn)
	assert.Nil(t, result)
	assert.False(t, continuePipeline)
}

func TestFilterByDeviceNameNotFound(t *testing.T) {
	// Event from DeviceName 1
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	filter := NewFilter([]string{"id2"})

	continuePipeline, result := filter.FilterByDeviceName(context, eventIn)
	assert.Nil(t, result)
	assert.False(t, continuePipeline)
}

func TestFilterOutByDeviceNameNotFound(t *testing.T) {
	// Event from DeviceName 1
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	filter := NewFilter([]string{devID2})
	filter.FilterOut = true

	continuePipeline, result := filter.FilterByDeviceName(context, eventIn)
	assert.NotNil(t, result)
	assert.True(t, continuePipeline)

	if eventOut, ok := result.(*dtos.Event); ok {
		assert.Equal(t, devID1, eventOut.DeviceName)
	}
}

func TestFilterByDeviceNameNoParameters(t *testing.T) {
	filter := NewFilter([]string{"id2"})

	continuePipeline, result := filter.FilterByDeviceName(context)
	assert.EqualError(t, result.(error), "no Event Received")
	assert.False(t, continuePipeline)
}

func TestFilterByDeviceNameNoFilterValues(t *testing.T) {
	expected := dtos.Event{
		DeviceName: devID1,
	}

	filter := NewFilter(nil)

	continuePipeline, result := filter.FilterByDeviceName(context, expected)
	require.NotNil(t, result, "Expected event to be passed thru")

	actual, ok := result.(dtos.Event)
	require.True(t, ok, "Expected result to be an Event")

	assert.Equal(t, expected.DeviceName, actual.DeviceName, "Expected Event to be same as passed in")
	assert.True(t, continuePipeline, "Pipeline shouldn't stop processing")
}

func TestFilterByDeviceNameFoundExtraParameters(t *testing.T) {
	eventIn := dtos.Event{
		DeviceName: devID1,
	}

	filter := NewFilter([]string{"id1"})

	continuePipeline, result := filter.FilterByDeviceName(context, eventIn, "application/event")
	require.NotNil(t, result)
	assert.True(t, continuePipeline, "Pipeline should continue processing")

	if eventOut, ok := result.(*dtos.Event); ok {
		assert.Equal(t, "id1", eventOut.DeviceName, "DeviceName does not match filter")
	}
}

func TestFilterByValueDescriptor(t *testing.T) {
	f1 := NewFilter([]string{descriptor1})
	f12 := NewFilter([]string{descriptor1, descriptor2})

	// event with a value descriptor 1
	event1 := dtos.Event{
		DeviceName: devID1,
	}
	event1.Readings = append(event1.Readings, dtos.BaseReading{ResourceName: descriptor1})

	// event with a value descriptor 2
	event2 := dtos.Event{}
	event2.Readings = append(event2.Readings, dtos.BaseReading{ResourceName: descriptor2})

	// event with a value descriptor 1 and another 2
	event12 := dtos.Event{}
	event12.Readings = append(event12.Readings, dtos.BaseReading{ResourceName: descriptor1})
	event12.Readings = append(event12.Readings, dtos.BaseReading{ResourceName: descriptor2})

	continuePipeline, res := f1.FilterByValueDescriptor(context)
	assert.False(t, continuePipeline, "Pipeline should stop since no parameter was passed")
	assert.EqualError(t, res.(error), "no Event Received")

	continuePipeline, res = f1.FilterByValueDescriptor(context, event1)
	assert.True(t, continuePipeline, "Pipeline should continue")
	assert.Len(t, res.(dtos.Event).Readings, 1, "Event should have one reading")

	continuePipeline, res = f1.FilterByValueDescriptor(context, event12)
	assert.True(t, continuePipeline, "Pipeline should continue")
	assert.Len(t, res.(dtos.Event).Readings, 1, "Event should have one reading")

	continuePipeline, res = f1.FilterByValueDescriptor(context, event2)
	assert.False(t, continuePipeline, "Event should be filtered out")

	continuePipeline, res = f12.FilterByValueDescriptor(context, event1)
	assert.True(t, continuePipeline, "Pipeline should continue")
	assert.Len(t, res.(dtos.Event).Readings, 1, "Event should have one reading")

	continuePipeline, res = f12.FilterByValueDescriptor(context, event12)
	assert.True(t, continuePipeline, "Pipeline should continue")
	assert.Len(t, res.(dtos.Event).Readings, 2, "Event should have one reading")

	continuePipeline, res = f12.FilterByValueDescriptor(context, event2)
	assert.True(t, continuePipeline, "Pipeline should continue")
	assert.Len(t, res.(dtos.Event).Readings, 1, "Event should have one reading")
}

func TestFilterOutByValueDescriptor(t *testing.T) {
	f1 := NewFilter([]string{descriptor1})
	f1.FilterOut = true
	f12 := NewFilter([]string{descriptor1, descriptor2})
	f12.FilterOut = true

	// event with a value descriptor 1
	event1 := dtos.Event{
		DeviceName: devID1,
	}
	event1.Readings = append(event1.Readings, dtos.BaseReading{ResourceName: descriptor1})

	// event with a value descriptor 2
	event2 := dtos.Event{}
	event2.Readings = append(event2.Readings, dtos.BaseReading{ResourceName: descriptor2})

	// event with a value descriptor 1 and another 2
	event12 := dtos.Event{}
	event12.Readings = append(event12.Readings, dtos.BaseReading{ResourceName: descriptor1})
	event12.Readings = append(event12.Readings, dtos.BaseReading{ResourceName: descriptor2})

	// event with a value descriptor 3
	event3 := dtos.Event{}
	event3.Readings = append(event3.Readings, dtos.BaseReading{ResourceName: descriptor3})

	continuePipeline, res := f1.FilterByValueDescriptor(context)
	assert.False(t, continuePipeline, "Pipeline should stop since no parameter was passed")
	assert.EqualError(t, res.(error), "no Event Received")

	continuePipeline, res = f1.FilterByValueDescriptor(context, event1)
	assert.False(t, continuePipeline, "Pipeline should NOT continue")
	assert.Len(t, res.(dtos.Event).Readings, 0, "Event should have no readings")

	continuePipeline, res = f1.FilterByValueDescriptor(context, event2)
	assert.True(t, continuePipeline, "Pipeline should continue")
	assert.Len(t, res.(dtos.Event).Readings, 1, "Event should have one reading")

	continuePipeline, res = f1.FilterByValueDescriptor(context, event12)
	assert.True(t, continuePipeline, "Event should be filtered out")
	assert.Len(t, res.(dtos.Event).Readings, 1, "Event should have one reading")

	continuePipeline, res = f12.FilterByValueDescriptor(context, event1)
	assert.False(t, continuePipeline, "Pipeline should NOT continue")
	assert.Len(t, res.(dtos.Event).Readings, 0, "Event should have one reading")

	continuePipeline, res = f12.FilterByValueDescriptor(context, event2)
	assert.False(t, continuePipeline, "Pipeline should NOT continue")
	assert.Len(t, res.(dtos.Event).Readings, 0, "Event should have no reading")

	continuePipeline, res = f12.FilterByValueDescriptor(context, event12)
	assert.False(t, continuePipeline, "Pipeline should NOT continue")
	assert.Len(t, res.(dtos.Event).Readings, 0, "Event should have no reading")

	continuePipeline, res = f12.FilterByValueDescriptor(context, event3)
	assert.True(t, continuePipeline, "Event should be filtered out")
	assert.Len(t, res.(dtos.Event).Readings, 1, "Event should have one reading")
}

func TestFilterByValueDescriptorNoFilterValues(t *testing.T) {
	expected := dtos.Event{
		DeviceName: devID1,
	}

	expected.Readings = append(expected.Readings, dtos.BaseReading{ResourceName: descriptor1})

	filter := NewFilter(nil)

	continuePipeline, result := filter.FilterByValueDescriptor(context, expected)
	require.NotNil(t, result, "Expected event to be passed thru")

	actual, ok := result.(dtos.Event)
	require.True(t, ok, "Expected result to be an Event")
	require.NotNil(t, actual.Readings, "Expected Reading passed thru")
	assert.Equal(t, expected.DeviceName, actual.DeviceName, "Expected Event to be same as passed in")
	require.True(t, len(expected.Readings) == 1)
	assert.Equal(t, expected.Readings[0].ResourceName, actual.Readings[0].ResourceName, "Expected Reading to be same as passed in")
	assert.True(t, continuePipeline, "Pipeline shouldn't stop processing")
}

func TestFilterByValueDescriptorExtraParameters(t *testing.T) {
	f1 := NewFilter([]string{descriptor1})

	// event with a value descriptor 1
	event1 := dtos.Event{
		DeviceName: devID1,
	}
	event1.Readings = append(event1.Readings, dtos.BaseReading{ResourceName: descriptor1})

	continuePipeline, res := f1.FilterByValueDescriptor(context, event1, "application/event")
	assert.True(t, continuePipeline, "Pipeline should continue")
	assert.Len(t, res.(dtos.Event).Readings, 1, "Event should have one reading")
}
