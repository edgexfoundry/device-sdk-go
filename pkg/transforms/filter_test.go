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

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	descriptor1 = "Descriptor1"
	descriptor2 = "Descriptor2"
)

func TestFilterByDeviceNameFound(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}

	filter := NewFilter([]string{"id1"})

	continuePipeline, result := filter.FilterByDeviceName(context, eventIn)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
	}
	if eventOut, ok := result.(*models.Event); ok {
		if eventOut.Device != "id1" {
			t.Fatal("device id does not match filter")
		}
	}
}

func TestFilterByDeviceNameNotFound(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}

	filter := NewFilter([]string{"id2"})

	continuePipeline, result := filter.FilterByDeviceName(context, eventIn)
	if result != nil {
		t.Fatal("result should be nil")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}

func TestFilterByDeviceNameNoParameters(t *testing.T) {
	filter := NewFilter([]string{"id2"})

	continuePipeline, result := filter.FilterByDeviceName(context)
	if result.(error).Error() != "no Event Received" {
		t.Fatal("Should have an error when no parameter was passed")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}

func TestFilterByDeviceNameNoFilterValues(t *testing.T) {
	expected := models.Event{
		Device: devID1,
	}

	filter := NewFilter(nil)

	continuePipeline, result := filter.FilterByDeviceName(context, expected)
	if !assert.NotNil(t, result, "Expected event to be passed thru") {
		t.Fatal()
	}

	actual, ok := result.(models.Event)
	if !assert.True(t, ok, "Expected result to be an Event") {
		t.Fatal()
	}

	assert.Equal(t, expected.Device, actual.Device, "Expected Event to be same as passed in")
	if !assert.True(t, continuePipeline, "Pipeline should'nt stop processing") {
		t.Fatal()
	}
}

func TestFilterByDeviceNameFoundExtraParameters(t *testing.T) {
	eventIn := models.Event{
		Device: devID1,
	}

	filter := NewFilter([]string{"id1"})

	continuePipeline, result := filter.FilterByDeviceName(context, eventIn, "application/event")
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if continuePipeline == false {
		t.Fatal("Pipeline should continue processing")
	}
	if eventOut, ok := result.(*models.Event); ok {
		if eventOut.Device != "id1" {
			t.Fatal("device id does not match filter")
		}
	}
}

func TestFilterByValueDescriptor(t *testing.T) {

	f1 := NewFilter([]string{descriptor1})
	f12 := NewFilter([]string{descriptor1, descriptor2})

	// event with a value descriptor 1
	event1 := models.Event{
		Device: devID1,
	}
	event1.Readings = append(event1.Readings, models.Reading{Name: descriptor1})

	// event with a value descriptor 2
	event2 := models.Event{}
	event2.Readings = append(event2.Readings, models.Reading{Name: descriptor2})

	// event with a value descriptor 1 and another 2
	event12 := models.Event{}
	event12.Readings = append(event12.Readings, models.Reading{Name: descriptor1})
	event12.Readings = append(event12.Readings, models.Reading{Name: descriptor2})

	continuePipeline, res := f1.FilterByValueDescriptor(context)
	if continuePipeline {
		t.Fatal("Pipeline should stop since no parameter was passed")
	}
	if res.(error).Error() != "no Event Received" {
		t.Fatal("Should have an error when no parameter was passed")
	}

	continuePipeline, res = f1.FilterByValueDescriptor(context, event1)
	if !continuePipeline {
		t.Fatal("Pipeline should continue")
	}
	if len(res.(models.Event).Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.(models.Event).Readings))
	}

	continuePipeline, res = f1.FilterByValueDescriptor(context, event12)
	if !continuePipeline {
		t.Fatal("Event should be continuePipeline")
	}
	if len(res.(models.Event).Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.(models.Event).Readings))
	}

	continuePipeline, res = f1.FilterByValueDescriptor(context, event2)
	if continuePipeline {
		t.Fatal("Event should be filtered out")
	}

	continuePipeline, res = f12.FilterByValueDescriptor(context, event1)
	if !continuePipeline {
		t.Fatal("Event should be continuePipeline")
	}
	if len(res.(models.Event).Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.(models.Event).Readings))
	}

	continuePipeline, res = f12.FilterByValueDescriptor(context, event12)
	if !continuePipeline {
		t.Fatal("Event should be continuePipeline")
	}
	if len(res.(models.Event).Readings) != 2 {
		t.Fatal("Event should be one reading, there are ", len(res.(models.Event).Readings))
	}

	continuePipeline, res = f12.FilterByValueDescriptor(context, event2)
	if !continuePipeline {
		t.Fatal("Event should be continuePipeline")
	}
	if len(res.(models.Event).Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.(models.Event).Readings))
	}
}

func TestFilterByValueDescriptorNoFilterValues(t *testing.T) {
	expected := models.Event{
		Device: devID1,
	}

	expected.Readings = append(expected.Readings, models.Reading{Name: descriptor1})

	filter := NewFilter(nil)

	continuePipeline, result := filter.FilterByValueDescriptor(context, expected)
	if !assert.NotNil(t, result, "Expected event to be passed thru") {
		t.Fatal()
	}

	actual, ok := result.(models.Event)
	if !assert.True(t, ok, "Expected result to be an Event") {
		t.Fatal()
	}

	if !assert.NotNil(t, actual.Readings, "Expected Reading passed thru") {
		t.Fatal()
	}

	assert.Equal(t, expected.Device, actual.Device, "Expected Event to be same as passed in")
	assert.Equal(t, expected.Readings[0].Name, actual.Readings[0].Name, "Expected Reading to be same as passed in")

	if !assert.True(t, continuePipeline, "Pipeline should'nt stop processing") {
		t.Fatal()
	}
}

func TestFilterByValueDescriptorExtraParameters(t *testing.T) {

	f1 := NewFilter([]string{descriptor1})

	// event with a value descriptor 1
	event1 := models.Event{
		Device: devID1,
	}
	event1.Readings = append(event1.Readings, models.Reading{Name: descriptor1})

	continuePipeline, res := f1.FilterByValueDescriptor(context, event1, "application/event")
	if !continuePipeline {
		t.Fatal("Pipeline should continue")
	}
	if len(res.(models.Event).Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.(models.Event).Readings))
	}
}
