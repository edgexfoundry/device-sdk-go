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

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	profileName1 = "profile1"
	profileName2 = "profile2"

	deviceName1 = "device1"
	deviceName2 = "device2"

	sourceName1 = "source1"
	sourceName2 = "source2"

	resource1 = "resource1"
	resource2 = "resource2"
	resource3 = "resource3"
)

func TestFilter_FilterByProfileName(t *testing.T) {
	profile1Event := dtos.NewEvent(profileName1, deviceName1, sourceName1)

	tests := []struct {
		Name              string
		Filters           []string
		FilterOut         bool
		EventIn           *dtos.Event
		ExpectedNilResult bool
	}{
		{"filter for - no event", []string{profileName1}, true, nil, true},
		{"filter for - no filter values", []string{}, false, &profile1Event, false},
		{"filter for with extra data - found", []string{profileName1}, false, &profile1Event, false},
		{"filter for - found", []string{profileName1}, false, &profile1Event, false},
		{"filter for - not found", []string{profileName2}, false, &profile1Event, true},

		{"filter out - no event", []string{profileName1}, true, nil, true},
		{"filter out - no filter values", []string{}, true, &profile1Event, false},
		{"filter out extra param - found", []string{profileName1}, true, &profile1Event, true},
		{"filter out - found", []string{profileName1}, true, &profile1Event, true},
		{"filter out - not found", []string{profileName2}, true, &profile1Event, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var filter Filter
			if test.FilterOut {
				filter = NewFilterOut(test.Filters)
			} else {
				filter = NewFilterFor(test.Filters)
			}

			expectedContinue := !test.ExpectedNilResult

			if test.EventIn == nil {
				continuePipeline, result := filter.FilterByProfileName(ctx, nil)
				assert.Contains(t, result.(error).Error(), "FilterByProfileName: no Event Received")
				assert.False(t, continuePipeline)
			} else {
				continuePipeline, result := filter.FilterByProfileName(ctx, *test.EventIn)
				assert.Equal(t, expectedContinue, continuePipeline)
				assert.Equal(t, test.ExpectedNilResult, result == nil)
				if result != nil && test.EventIn != nil {
					assert.Equal(t, *test.EventIn, result)
				}
			}
		})
	}
}

func TestFilter_FilterByDeviceName(t *testing.T) {
	device1Event := dtos.NewEvent(profileName1, deviceName1, sourceName1)

	tests := []struct {
		Name              string
		Filters           []string
		FilterOut         bool
		EventIn           *dtos.Event
		ExpectedNilResult bool
	}{
		{"filter for - no event", []string{deviceName1}, false, nil, true},
		{"filter for - no filter values", []string{}, false, &device1Event, false},
		{"filter for with extra data - found", []string{deviceName1}, false, &device1Event, false},
		{"filter for - found", []string{deviceName1}, false, &device1Event, false},
		{"filter for - not found", []string{deviceName2}, false, &device1Event, true},

		{"filter out - no event", []string{deviceName1}, true, nil, true},
		{"filter out - no filter values", []string{}, true, &device1Event, false},
		{"filter out extra param - found", []string{deviceName1}, true, &device1Event, true},
		{"filter out - found", []string{deviceName1}, true, &device1Event, true},
		{"filter out - not found", []string{deviceName2}, true, &device1Event, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var filter Filter
			if test.FilterOut {
				filter = NewFilterOut(test.Filters)
			} else {
				filter = NewFilterFor(test.Filters)
			}

			expectedContinue := !test.ExpectedNilResult

			if test.EventIn == nil {
				continuePipeline, result := filter.FilterByDeviceName(ctx, nil)
				assert.Contains(t, result.(error).Error(), "FilterByDeviceName: no Event Received")
				assert.False(t, continuePipeline)
			} else {
				continuePipeline, result := filter.FilterByDeviceName(ctx, *test.EventIn)
				assert.Equal(t, expectedContinue, continuePipeline)
				assert.Equal(t, test.ExpectedNilResult, result == nil)
				if result != nil && test.EventIn != nil {
					assert.Equal(t, *test.EventIn, result)
				}
			}
		})
	}
}

func TestFilter_FilterBySourceName(t *testing.T) {
	source1Event := dtos.NewEvent(profileName1, deviceName1, sourceName1)

	tests := []struct {
		Name              string
		Filters           []string
		FilterOut         bool
		EventIn           *dtos.Event
		ExpectedNilResult bool
	}{
		{"filter for - no event", []string{sourceName1}, true, nil, true},
		{"filter for - no filter values", []string{}, false, &source1Event, false},
		{"filter for with extra data - found", []string{sourceName1}, false, &source1Event, false},
		{"filter for - found", []string{sourceName1}, false, &source1Event, false},
		{"filter for - not found", []string{sourceName2}, false, &source1Event, true},

		{"filter out - no event", []string{sourceName1}, true, nil, true},
		{"filter out - no filter values", []string{}, true, &source1Event, false},
		{"filter out extra param - found", []string{sourceName1}, true, &source1Event, true},
		{"filter out - found", []string{sourceName1}, true, &source1Event, true},
		{"filter out - not found", []string{sourceName2}, true, &source1Event, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var filter Filter
			if test.FilterOut {
				filter = NewFilterOut(test.Filters)
			} else {
				filter = NewFilterFor(test.Filters)
			}

			expectedContinue := !test.ExpectedNilResult

			if test.EventIn == nil {
				continuePipeline, result := filter.FilterBySourceName(ctx, nil)
				assert.Contains(t, result.(error).Error(), "FilterBySourceName: no Event Received")
				assert.False(t, continuePipeline)
			} else {
				continuePipeline, result := filter.FilterBySourceName(ctx, *test.EventIn)
				assert.Equal(t, expectedContinue, continuePipeline)
				assert.Equal(t, test.ExpectedNilResult, result == nil)
				if result != nil && test.EventIn != nil {
					assert.Equal(t, *test.EventIn, result)
				}
			}
		})
	}
}

func TestFilter_FilterByResourceName(t *testing.T) {
	// event with a reading for resource 1
	resource1Event := dtos.NewEvent(profileName1, deviceName1, sourceName1)
	err := resource1Event.AddSimpleReading(resource1, common.ValueTypeInt32, int32(123))
	require.NoError(t, err)

	// event with a reading for resource 2
	resource2Event := dtos.NewEvent(profileName1, deviceName1, sourceName1)
	err = resource2Event.AddSimpleReading(resource2, common.ValueTypeInt32, int32(123))
	require.NoError(t, err)

	// event with a reading for resource 3
	resource3Event := dtos.NewEvent(profileName1, deviceName1, sourceName1)
	err = resource3Event.AddSimpleReading(resource3, common.ValueTypeInt32, int32(123))
	require.NoError(t, err)

	// event with readings for resource 1 & 2
	twoResourceEvent := dtos.NewEvent(profileName1, deviceName1, sourceName1)
	err = twoResourceEvent.AddSimpleReading(resource1, common.ValueTypeInt32, int32(123))
	require.NoError(t, err)
	err = twoResourceEvent.AddSimpleReading(resource2, common.ValueTypeInt32, int32(123))
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		Filters              []string
		FilterOut            bool
		EventIn              *dtos.Event
		ExpectedNilResult    bool
		ExpectedReadingCount int
	}{
		{"filter for - no event", []string{resource1}, false, nil, true, 0},
		{"filter for extra param - found", []string{resource1}, false, &resource1Event, false, 1},
		{"filter for 0 in R1 - no change", []string{}, false, &resource1Event, false, 1},
		{"filter for 1 in R1 - 1 of 1 found", []string{resource1}, false, &resource1Event, false, 1},
		{"filter for 1 in 2R - 1 of 2 found", []string{resource1}, false, &twoResourceEvent, false, 1},
		{"filter for 2 in R1 - 1 of 1 found", []string{resource1, resource2}, false, &resource1Event, false, 1},
		{"filter for 2 in 2R - 2 of 2 found", []string{resource1, resource2}, false, &twoResourceEvent, false, 2},
		{"filter for 2 in R2 - 1 of 2 found", []string{resource1, resource2}, false, &resource2Event, false, 1},
		{"filter for 1 in R2 - not found", []string{resource1}, false, &resource2Event, true, 0},

		{"filter out - no event", []string{resource1}, true, nil, true, 0},
		{"filter out extra param - found", []string{resource1}, true, &resource1Event, true, 0},
		{"filter out 0 in R1 - no change", []string{}, true, &resource1Event, false, 1},
		{"filter out 1 in R1 - 1 of 1 found", []string{resource1}, true, &resource1Event, true, 0},
		{"filter out 1 in R2 - not found", []string{resource1}, true, &resource2Event, false, 1},
		{"filter out 1 in 2R - 1 of 2 found", []string{resource1}, true, &twoResourceEvent, false, 1},
		{"filter out 2 in R1 - 1 of 1 found", []string{resource1, resource2}, true, &resource1Event, true, 0},
		{"filter out 2 in R2 - 1 of 1 found", []string{resource1, resource2}, true, &resource2Event, true, 0},
		{"filter out 2 in 2R - 2 of 2 found", []string{resource1, resource2}, true, &twoResourceEvent, true, 0},
		{"filter out 2 in R3 - not found", []string{resource1, resource2}, true, &resource3Event, false, 1},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var filter Filter
			if test.FilterOut {
				filter = NewFilterOut(test.Filters)
			} else {
				filter = NewFilterFor(test.Filters)
			}

			expectedContinue := !test.ExpectedNilResult

			if test.EventIn == nil {
				continuePipeline, result := filter.FilterByResourceName(ctx, nil)
				assert.Contains(t, result.(error).Error(), "FilterByResourceName: no Event Received")
				assert.False(t, continuePipeline)
			} else {
				continuePipeline, result := filter.FilterByResourceName(ctx, *test.EventIn)
				assert.Equal(t, expectedContinue, continuePipeline)
				assert.Equal(t, test.ExpectedNilResult, result == nil)
				if result != nil {
					actualEvent, ok := result.(dtos.Event)
					require.True(t, ok)
					assert.Equal(t, test.ExpectedReadingCount, len(actualEvent.Readings))

					// Make sure the event is still valid
					request := requests.NewAddEventRequest(actualEvent)
					err = request.Validate()
					require.NoError(t, err)
				}
			}
		})
	}
}
