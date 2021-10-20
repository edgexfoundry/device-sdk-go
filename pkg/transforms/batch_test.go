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
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dataToBatch = [3]string{"Test1", "Test2", "Test3"}

func TestBatchNoData(t *testing.T) {

	bs, _ := NewBatchByCount(1)
	continuePipeline, err := bs.Batch(ctx, nil)
	assert.False(t, continuePipeline)
	assert.Contains(t, err.(error).Error(), "No Data Received")

}
func TestBatchInCountMode(t *testing.T) {

	bs, _ := NewBatchByCount(3)

	continuePipeline1, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
	assert.False(t, continuePipeline1)
	assert.Len(t, bs.batchData.all(), 1, "Should have 1 record")

	continuePipeline2, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
	assert.False(t, continuePipeline2)
	assert.Len(t, bs.batchData.all(), 2, "Should have 2 records")

	continuePipeline3, result3 := bs.Batch(ctx, []byte(dataToBatch[0]))
	assert.True(t, continuePipeline3)
	assert.Len(t, result3, 3, "Should have 3 records")
	assert.Len(t, bs.batchData.all(), 0, "Records should have been cleared")

	continuePipeline4, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
	assert.False(t, continuePipeline4)
	assert.Len(t, bs.batchData.all(), 1, "Should have 1 record")

	continuePipeline5, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
	assert.False(t, continuePipeline5)
	assert.Len(t, bs.batchData.all(), 2, "Should have 2 records")

	continuePipeline6, result4 := bs.Batch(ctx, []byte(dataToBatch[0]))
	assert.True(t, continuePipeline6)
	assert.Len(t, result4, 3, "Should have 3 records")
	assert.Len(t, bs.batchData.all(), 0, "Records should have been cleared")
}
func TestBatchInTimeAndCountMode_TimeElapsed(t *testing.T) {

	bs, _ := NewBatchByTimeAndCount("2s", 10)
	var wgAll sync.WaitGroup
	var wgFirst sync.WaitGroup
	wgAll.Add(3)
	wgFirst.Add(1)

	go func() {
		go func() {
			time.Sleep(time.Second * 1)
			wgFirst.Done()
		}()

		// Key to this test is this call occurs first and will be blocked until
		// batch time interval has elapsed. In the mean time the other go func have to execute
		// before the batch time interval has elapsed, so the sleep above has to be less than the
		// batch time interval.
		continuePipeline1, result := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.True(t, continuePipeline1)
		if continuePipeline1 {
			assert.Equal(t, 3, len(result.([][]byte)))
			assert.Len(t, bs.batchData.all(), 0, "Should have 0 records")
		}
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		continuePipeline2, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline2)
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		continuePipeline3, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline3)
		wgAll.Done()
	}()
	wgAll.Wait()
}

func TestBatchInTimeAndCountMode_CountMet(t *testing.T) {

	bs, _ := NewBatchByTimeAndCount("90s", 3)
	var wgAll sync.WaitGroup
	var wgFirst sync.WaitGroup
	var wgSecond sync.WaitGroup
	wgAll.Add(3)
	wgFirst.Add(1)
	wgSecond.Add(1)

	go func() {
		go func() {
			time.Sleep(time.Second * 10)
			wgFirst.Done()
		}()
		continuePipeline1, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline1)
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		go func() {
			time.Sleep(time.Second * 10)
			wgSecond.Done()
		}()
		continuePipeline2, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline2)
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		wgSecond.Wait()
		continuePipeline3, result := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.True(t, continuePipeline3)
		assert.Equal(t, 3, len(result.([][]byte)))
		assert.Nil(t, bs.batchData.all(), "Should have 0 records")
		wgAll.Done()
	}()
	wgAll.Wait()
}
func TestBatchInTimeMode(t *testing.T) {

	bs, _ := NewBatchByTime("3s")
	var wgAll sync.WaitGroup
	var wgFirst sync.WaitGroup
	wgAll.Add(3)
	wgFirst.Add(1)

	go func() {
		go func() {
			time.Sleep(1000)
			wgFirst.Done()
		}()

		continuePipeline1, result := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.True(t, continuePipeline1)
		assert.Equal(t, 3, len(result.([][]byte)))
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		continuePipeline2, result := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline2)
		assert.Nil(t, result)
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		continuePipeline3, _ := bs.Batch(ctx, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline3)
		wgAll.Done()
	}()
	wgAll.Wait()
}

func TestBatchIsEventData(t *testing.T) {
	events := []dtos.Event{
		dtos.NewEvent("p1", "d1", "s1"),
		dtos.NewEvent("p1", "d1", "s1"),
		dtos.NewEvent("p1", "d1", "s1"),
	}
	err := events[0].AddSimpleReading("r1", common.ValueTypeString, "Hello")
	require.NoError(t, err)
	err = events[1].AddSimpleReading("r2", common.ValueTypeInt8, int8(34))
	require.NoError(t, err)
	err = events[2].AddSimpleReading("r3", common.ValueTypeFloat64, 89.90)
	require.NoError(t, err)

	tests := []struct {
		name        string
		isEventData bool
	}{
		{"Is Events", true},
		{"Is Not Events", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bbc, err := NewBatchByCount(3)
			require.NoError(t, err)
			bbc.IsEventData = test.isEventData

			if test.isEventData {
				continuePipeline, result := bbc.Batch(ctx, events[0])
				require.False(t, continuePipeline)
				require.Nil(t, result)

				continuePipeline, result = bbc.Batch(ctx, events[1])
				require.False(t, continuePipeline)
				require.Nil(t, result)

				continuePipeline, result = bbc.Batch(ctx, events[2])
				require.True(t, continuePipeline)
				require.NotNil(t, result)

				batchedEvents, ok := result.([]dtos.Event)
				require.True(t, ok)
				assert.Equal(t, events, batchedEvents)
			} else {
				continuePipeline, result := bbc.Batch(ctx, dataToBatch[0])
				require.False(t, continuePipeline)
				require.Nil(t, result)

				continuePipeline, result = bbc.Batch(ctx, dataToBatch[1])
				require.False(t, continuePipeline)
				require.Nil(t, result)

				continuePipeline, result = bbc.Batch(ctx, dataToBatch[2])
				require.True(t, continuePipeline)
				require.NotNil(t, result)

				_, ok := result.([]dtos.Event)
				require.False(t, ok)
				_, ok = result.([][]byte)
				require.True(t, ok)
			}
		})
	}
}
