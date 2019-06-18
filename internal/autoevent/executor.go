// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type Executor interface {
	Run()
	Stop()
}

type executor struct {
	deviceName   string
	autoEvent    contract.AutoEvent
	lastReadings map[string]string
	duration     time.Duration
	stop         bool
	rwmutex      sync.RWMutex
}

// Run triggers this Executor executes the handler for the resource periodically
func (e *executor) Run() {
	for {
		if e.stop {
			break
		}
		time.Sleep(e.duration)

		common.LoggingClient.Debug(fmt.Sprintf("AutoEvent - executing %v", e.autoEvent))
		evt, appErr := readResource(e)
		if appErr != nil {
			common.LoggingClient.Error(fmt.Sprintf("AutoEvent - error occurs when reading resource %s",
				e.autoEvent.Resource))
			continue
		}

		if evt != nil {
			if e.autoEvent.OnChange {
				if compareReadings(e, evt.Readings) {
					common.LoggingClient.Debug(fmt.Sprintf("AutoEvent - readings are the same as previous one %v", e.lastReadings))
					continue
				}
				cacheReadings(e, evt.Readings)
			}
			common.LoggingClient.Debug(fmt.Sprintf("AutoEvent - pushing event %s", evt.String()))
			event := &dsModels.Event{Event: evt.Event}
			// Attach origin timestamp for readings if none yet specified
			if event.Origin == 0 {
				event.Origin = time.Now().UnixNano()
			}
			go common.SendEvent(event)
		}

		if evt == nil {
			common.LoggingClient.Info(fmt.Sprintf("AutoEvent - no event generated when reading resource %s",
				e.autoEvent.Resource))
			continue
		}
	}
}

func readResource(e *executor) (*dsModels.Event, common.AppError) {
	vars := make(map[string]string, 2)
	vars[common.NameVar] = e.deviceName
	vars[common.CommandVar] = e.autoEvent.Resource

	evt, appErr := handler.CommandHandler(vars, "", common.GetCmdMethod)
	return evt, appErr
}

func cacheReadings(e *executor, readings []contract.Reading) {
	e.rwmutex.Lock()
	defer e.rwmutex.Unlock()
	for _, r := range readings {
		e.lastReadings[r.Name] = r.Value
	}
}

func compareReadings(e *executor, readings []contract.Reading) bool {
	e.rwmutex.RLock()
	defer e.rwmutex.RUnlock()
	for _, r := range readings {
		v, ok := e.lastReadings[r.Name]
		if !ok || v != r.Value {
			return false
		}
	}
	return true
}

// Stop marks this Executor stopped
func (e *executor) Stop() {
	e.stop = true
}

// NewExecutor creates an Executor for an AutoEvent
func NewExecutor(deviceName string, ae contract.AutoEvent) (Executor, error) {
	// check Frequency
	duration, err := time.ParseDuration(ae.Frequency)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("AutoEvent Frequency %s cannot be parsed error, %v", ae.Frequency, err))
		return nil, err
	}

	return &executor{deviceName: deviceName, autoEvent: ae,
		lastReadings: make(map[string]string), duration: duration, stop: false}, nil
}
