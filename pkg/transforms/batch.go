package transforms

import (
	"errors"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"
)

// BatchMode Enum for choosing behavior of Batch. Default is CountAndTime.
type BatchMode int

const (
	BatchByCountOnly = iota
	BatchByTimeOnly
	BatchByTimeAndCount
)

// BatchConfig ...
type BatchConfig struct {
	timeInterval                string
	parsedDuration              time.Duration
	batchThreshold              int
	batchMode                   BatchMode
	batchData                   [][]byte
	continuedPipelineTransforms []appcontext.AppFunction
	timerActive                 bool
	done                        chan bool
}

// NewBatchByTime create, initializes  and returns a new instance for BatchConfig
func NewBatchByTime(timeInterval string) (*BatchConfig, error) {
	config := BatchConfig{
		timeInterval: timeInterval,
		batchMode:    BatchByTimeOnly, //Default to CountAndTime
	}
	var err error
	config.parsedDuration, err = time.ParseDuration(config.timeInterval)
	if err != nil {
		return nil, err
	}
	config.done = make(chan bool)

	return &config, nil
}

// NewBatchByCount create, initializes  and returns a new instance for BatchConfig
func NewBatchByCount(batchThreshold int) (*BatchConfig, error) {
	config := BatchConfig{
		batchThreshold: batchThreshold,
		batchMode:      BatchByCountOnly, //Default to CountAndTime
	}

	return &config, nil
}

// NewBatchByTimeAndCount create, initializes  and returns a new instance for BatchConfig
func NewBatchByTimeAndCount(timeInterval string, batchThreshold int) (*BatchConfig, error) {
	config := BatchConfig{
		timeInterval:   timeInterval,
		batchThreshold: batchThreshold,
		batchMode:      BatchByTimeAndCount, //Default to CountAndTime
	}
	var err error
	config.parsedDuration, err = time.ParseDuration(config.timeInterval)
	if err != nil {
		return nil, err
	}
	config.done = make(chan bool)

	return &config, nil
}

// Batch ...
func (batch *BatchConfig) Batch(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}

	edgexcontext.LoggingClient.Debug("Batching Data")
	data, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}
	// always append data
	batch.batchData = append(batch.batchData, data)

	// If its time only or time and count
	if batch.batchMode != BatchByCountOnly {
		if batch.timerActive == false {
			batch.timerActive = true
			for {
				select {
				case <-batch.done:
					edgexcontext.LoggingClient.Debug("Batch count has been reached")
				case <-time.After(batch.parsedDuration):
					edgexcontext.LoggingClient.Debug("Timer has elapsed")
				}
				break
			}
			batch.timerActive = false
		} else {
			if batch.batchMode == BatchByTimeOnly {
				return false, nil
			}
		}
	}

	if batch.batchMode != BatchByTimeOnly {
		//Only want to check the threshold if the timer is running and in TimeAndCount mode OR if we are
		// in CountOnly mode
		if batch.batchMode == BatchByCountOnly || (batch.timerActive == true && batch.batchMode == BatchByTimeAndCount) {
			// if we have not reached the threshold, then stop pipeline and continue batching
			if len(batch.batchData) < batch.batchThreshold {
				return false, nil
			}
			// if in BatchByCountOnly mode, there are no listeners so this would hang indefinitely
			if batch.done != nil {
				batch.done <- true
			}
		}
	}

	edgexcontext.LoggingClient.Debug("Forwarding Batched Data...")
	// we've met the threshold, lets clear out the buffer and send it forward in the pipeline
	if len(batch.batchData) > 0 {
		copy := batch.batchData
		batch.batchData = nil
		return true, copy
	}
	return false, nil
}
