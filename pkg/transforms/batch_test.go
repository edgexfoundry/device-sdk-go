package transforms

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var dataToBatch = [3]string{"Test1", "Test2", "Test3"}

func TestBatchNoData(t *testing.T) {

	bs, _ := NewBatchByCount(1)
	continuePipeline, err := bs.Batch(context)
	assert.False(t, continuePipeline)
	assert.Equal(t, "No Data Received", err.(error).Error())

}
func TestBatchInCountMode(t *testing.T) {

	bs, _ := NewBatchByCount(3)

	continuePipeline1, _ := bs.Batch(context, []byte(dataToBatch[0]))
	assert.False(t, continuePipeline1)
	assert.Len(t, bs.batchData, 1, "Should have 1 record")

	continuePipeline2, _ := bs.Batch(context, []byte(dataToBatch[0]))
	assert.False(t, continuePipeline2)
	assert.Len(t, bs.batchData, 2, "Should have 2 records")

	continuePipeline3, result3 := bs.Batch(context, []byte(dataToBatch[0]))
	assert.True(t, continuePipeline3)
	assert.Len(t, result3, 3, "Should have 3 records")
	assert.Len(t, bs.batchData, 0, "Records should have been cleared")

	continuePipeline4, _ := bs.Batch(context, []byte(dataToBatch[0]))
	assert.False(t, continuePipeline4)
	assert.Len(t, bs.batchData, 1, "Should have 1 record")

	continuePipeline5, _ := bs.Batch(context, []byte(dataToBatch[0]))
	assert.False(t, continuePipeline5)
	assert.Len(t, bs.batchData, 2, "Should have 2 records")

	continuePipeline6, result4 := bs.Batch(context, []byte(dataToBatch[0]))
	assert.True(t, continuePipeline6)
	assert.Len(t, result4, 3, "Should have 3 records")
	assert.Len(t, bs.batchData, 0, "Records should have been cleared")
}
func TestBatchInTimeAndCountMode_TimeElapsed(t *testing.T) {

	bs, _ := NewBatchByTimeAndCount("2s", 10)
	var wgAll sync.WaitGroup
	var wgFirst sync.WaitGroup
	wgAll.Add(3)
	wgFirst.Add(1)

	go func() {
		go func() {
			time.Sleep(time.Second * 2)
			wgFirst.Done()
		}()
		continuePipeline1, result := bs.Batch(context, []byte(dataToBatch[0]))
		assert.True(t, continuePipeline1)
		assert.Equal(t, 3, len(result.([][]byte)))
		assert.Len(t, bs.batchData, 0, "Should have 0 records")
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		continuePipeline2, _ := bs.Batch(context, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline2)
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		continuePipeline3, _ := bs.Batch(context, []byte(dataToBatch[0]))
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
			time.Sleep(200)
			wgFirst.Done()
		}()
		continuePipeline1, _ := bs.Batch(context, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline1)
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		go func() {
			time.Sleep(200)
			wgSecond.Done()
		}()
		continuePipeline2, _ := bs.Batch(context, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline2)
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		wgSecond.Wait()
		continuePipeline3, result := bs.Batch(context, []byte(dataToBatch[0]))
		assert.True(t, continuePipeline3)
		assert.Equal(t, 3, len(result.([][]byte)))
		assert.Nil(t, bs.batchData, "Should have 0 records")
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
			time.Sleep(100)
			wgFirst.Done()
		}()

		continuePipeline1, result := bs.Batch(context, []byte(dataToBatch[0]))
		assert.True(t, continuePipeline1)
		assert.Equal(t, 3, len(result.([][]byte)))
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		continuePipeline2, result := bs.Batch(context, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline2)
		assert.Nil(t, result)
		wgAll.Done()
	}()
	go func() {
		wgFirst.Wait()
		continuePipeline3, _ := bs.Batch(context, []byte(dataToBatch[0]))
		assert.False(t, continuePipeline3)
		wgAll.Done()
	}()
	wgAll.Wait()
}
