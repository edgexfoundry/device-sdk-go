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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
)

// BatchMode Enum for choosing behavior of Batch. Default is CountAndTime.
type BatchMode int

const (
	BatchByCountOnly = iota
	BatchByTimeOnly
	BatchByTimeAndCount
)

type atomicBatchData struct {
	mutex sync.Mutex
	data  [][]byte
}

func (d *atomicBatchData) append(toBeAdded []byte) [][]byte {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data = append(d.data, toBeAdded)
	result := d.data
	return result
}

func (d *atomicBatchData) all() [][]byte {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	result := d.data
	return result
}

func (d *atomicBatchData) removeAll() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data = nil
}

func (d *atomicBatchData) length() int {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	result := len(d.data)
	return result
}

// BatchConfig ...
type BatchConfig struct {
	IsEventData    bool
	timeInterval   string
	parsedDuration time.Duration
	batchThreshold int
	batchMode      BatchMode
	batchData      atomicBatchData
	timerActive    common.AtomicBool
	done           chan bool
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
func (batch *BatchConfig) Batch(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		// We didn't receive a result
		return false, fmt.Errorf("function Batch in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	ctx.LoggingClient().Debugf("Batching Data in pipeline '%s'", ctx.PipelineId())
	byteData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}
	// always append data
	batch.batchData.append(byteData)

	// If its time only or time and count
	if batch.batchMode != BatchByCountOnly {
		if !batch.timerActive.Value() {
			batch.timerActive.Set(true)
			select {
			case <-batch.done:
				ctx.LoggingClient().Debugf("Batch count has been reached in pipeline '%s'", ctx.PipelineId())
			case <-time.After(batch.parsedDuration):
				ctx.LoggingClient().Debugf("Timer has elapsed in pipeline '%s'", ctx.PipelineId())
			}
			batch.timerActive.Set(false)
		} else {
			if batch.batchMode == BatchByTimeOnly {
				return false, nil
			}
		}
	}

	if batch.batchMode != BatchByTimeOnly {
		//Only want to check the threshold if the timer is running and in TimeAndCount mode OR if we are
		// in CountOnly mode
		if batch.batchMode == BatchByCountOnly || (batch.timerActive.Value() && batch.batchMode == BatchByTimeAndCount) {
			// if we have not reached the threshold, then stop pipeline and continue batching
			if batch.batchData.length() < batch.batchThreshold {
				return false, nil
			}
			// if in BatchByCountOnly mode, there are no listeners so this would hang indefinitely
			if batch.done != nil {
				batch.done <- true
			}
		}
	}

	ctx.LoggingClient().Debugf("Forwarding Batched Data in pipeline '%s'", ctx.PipelineId())
	// we've met the threshold, lets clear out the buffer and send it forward in the pipeline
	if batch.batchData.length() > 0 {
		var resultData interface{}
		copyOfData := batch.batchData.all()
		resultData = copyOfData
		if batch.IsEventData {
			ctx.LoggingClient().Debug("Marshaling batched data to []Event")
			var events []dtos.Event
			for _, data := range copyOfData {
				event := dtos.Event{}
				if err := json.Unmarshal(data, &event); err != nil {
					return false, fmt.Errorf("unable to marshal batched data to slice of Events in pipeline '%s': %s", ctx.PipelineId(), err.Error())
				}
				events = append(events, event)
			}

			resultData = events
		}

		batch.batchData.removeAll()
		return true, resultData
	}
	return false, nil
}
