// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"bytes"
	"context"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/application"
)

type Executor struct {
	deviceName   string
	resource     string
	onChange     bool
	lastReadings dtos.BaseReading
	duration     time.Duration
	stop         bool
	rwMutex      *sync.RWMutex
}

// Run triggers this Executor executes the handler for the resource periodically
func (e *Executor) Run(ctx context.Context, wg *sync.WaitGroup, buffer chan bool, dic *di.Container) {
	wg.Add(1)
	defer wg.Done()

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(e.duration):
			if e.stop {
				return
			}
			ds := container.DeviceServiceFrom(dic.Get)
			if ds.AdminState == models.Locked {
				lc.Info("AutoEvent - stopped for locked device service")
				return
			}

			lc.Debugf("AutoEvent - reading %s", e.resource)
			evt, err := readResource(e, dic)
			if err != nil {
				lc.Errorf("AutoEvent - error occurs when reading resource %s: %v", e.resource, err)
				continue
			}

			if evt != nil {
				if e.onChange {
					if e.compareReadings(evt.Readings[0]) {
						lc.Debugf("AutoEvent - readings are the same as previous one %v", e.lastReadings)
						continue
					}
				}
				// After the auto event executes a read command, it will create a goroutine to send out events.
				// When the concurrent auto event amount becomes large, core-data might be hard to handle so many HTTP requests at the same time.
				// The device service will get some network errors like EOF or Connection reset by peer.
				// By adding a buffer here, the user can use the Service.AsyncBufferSize configuration to control the goroutine for sending events.
				go func() {
					buffer <- true
					// TODO: uncomment when v2 client refactoring is finished
					//common.SendEvent(event, lc, container.CoredataEventClientFrom(dic.Get))
					<-buffer
				}()
			} else {
				lc.Debugf("AutoEvent - no event generated when reading resource %s", e.resource)
			}
		}
	}
}

func readResource(e *Executor, dic *di.Container) (event *dtos.Event, err errors.EdgeX) {
	vars := make(map[string]string, 2)
	vars[common.NameVar] = e.deviceName
	vars[common.CommandVar] = e.resource

	res, err := application.CommandHandler(true, false, "", vars, "", dic)
	if err != nil {
		return event, err
	}
	return &res.Event, nil
}

func (e *Executor) compareReadings(reading dtos.BaseReading) bool {
	e.rwMutex.RLock()
	defer e.rwMutex.RUnlock()

	if reading.Value != "" {
		if reading.Value != e.lastReadings.Value {
			e.lastReadings = reading
			return false
		}
		return true
	} else {
		if bytes.Compare(e.lastReadings.BinaryValue, reading.BinaryValue) != 0 {
			e.lastReadings = reading
			return false
		}
		return true
	}
}

// Stop marks this Executor stopped
func (e *Executor) Stop() {
	e.stop = true
}

// NewExecutor creates an Executor for an AutoEvent
func NewExecutor(deviceName string, ae models.AutoEvent) (*Executor, error) {
	// check Frequency
	duration, err := time.ParseDuration(ae.Frequency)
	if err != nil {
		return nil, err
	}

	return &Executor{
		deviceName: deviceName,
		resource:   ae.Resource,
		onChange:   ae.OnChange,
		duration:   duration,
		stop:       false,
		rwMutex:    &sync.RWMutex{}}, nil
}
