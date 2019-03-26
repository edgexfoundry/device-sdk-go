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

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	descriptor1 = "Descriptor1"
	descriptor2 = "Descriptor2"
)

func init() {
	lc := logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
	context = &appcontext.Context{
		LoggingClient: lc,
	}
}
func TestFilterByDeviceIDFound(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	filter := Filter{
		FilterValues: []string{"id1"},
	}
	continuePipeline, result := filter.FilterByDeviceID(context, eventIn)
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
func TestFilterByDeviceIDNotFound(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	filter := Filter{
		FilterValues: []string{"id2"},
	}
	continuePipeline, result := filter.FilterByDeviceID(context, eventIn)
	if result != nil {
		t.Fatal("result should be nil")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
}
func TestFilterByDeviceIDNoParameters(t *testing.T) {
	filter := Filter{
		FilterValues: []string{"id2"},
	}
	continuePipeline, result := filter.FilterByDeviceID(context)
	if result.(error).Error() != "No Event Received" {
		t.Fatal("Should have an error when no parameter was passed")
	}
	if continuePipeline == true {
		t.Fatal("Pipeline should stop processing")
	}
	// if result != errors.new()("") {
	// 	t.Fatal("Pipeline should return no paramater error")
	// }
}

func TestFilterValue(t *testing.T) {

	f1 := Filter{
		FilterValues: []string{descriptor1},
	}
	f12 := Filter{
		FilterValues: []string{descriptor1, descriptor2},
	}

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
	if res.(error).Error() != "No Event Received" {
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
