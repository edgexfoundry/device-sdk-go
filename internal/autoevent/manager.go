// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/internal/common"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type Manager interface {
	StartAutoEvents()
	StopAutoEvents()
	RestartForDevice(deviceName string)
	StopForDevice(deviceName string)
}

var (
	createOnce sync.Once
	m          *manager
	mutex      sync.Mutex
)

type manager struct {
	execsMap  map[string][]Executor
	startOnce sync.Once
	ctx       context.Context
	wg        *sync.WaitGroup
}

func (m *manager) StartAutoEvents() {
	mutex.Lock()
	m.startOnce.Do(func() {
		for _, d := range cache.Devices().All() {
			execs := triggerExecutors(d.Name, d.AutoEvents, m.ctx, m.wg)
			m.execsMap[d.Name] = execs
		}
	})
	mutex.Unlock()
}

func (m *manager) StopAutoEvents() {
	mutex.Lock()
	for k, v := range m.execsMap {
		for _, e := range v {
			e.Stop()
		}
		delete(m.execsMap, k)
	}
	mutex.Unlock()
}

func triggerExecutors(deviceName string, autoEvents []contract.AutoEvent, ctx context.Context, wg *sync.WaitGroup) []Executor {
	var execs []Executor
	for _, autoEvent := range autoEvents {
		exec, err := NewExecutor(deviceName, autoEvent)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("AutoEvent for resource %s cannot be created, %v", autoEvent.Resource, err))
			// skip this AutoEvent if it causes error during creation
			continue
		}
		execs = append(execs, exec)
		go exec.Run(ctx, wg)
	}
	return execs
}

// RestartForDevice stops all the AutoEvents of the specific Device
func (m *manager) RestartForDevice(deviceName string) {
	m.StopForDevice(deviceName)
	d, ok := cache.Devices().ForName(deviceName)
	if !ok {
		common.LoggingClient.Error(fmt.Sprintf("there is no Device %s in cache to start AutoEvent", deviceName))
	}

	mutex.Lock()
	execs := triggerExecutors(deviceName, d.AutoEvents, m.ctx, m.wg)
	m.execsMap[deviceName] = execs
	mutex.Unlock()
}

// StopForDevice stops all the AutoEvents of the specific Device
func (m *manager) StopForDevice(deviceName string) {
	mutex.Lock()
	execs, ok := m.execsMap[deviceName]
	if ok {
		for _, e := range execs {
			e.Stop()
		}
		delete(m.execsMap, deviceName)
	}
	mutex.Unlock()
}

// NewManager initiates the AutoEvent manager once
func NewManager(ctx context.Context, wg *sync.WaitGroup) {
	createOnce.Do(func() {
		m = &manager{execsMap: make(map[string][]Executor), ctx: ctx, wg: wg}
	})
}

// GetManager returns Manager instance
func GetManager() Manager {
	return m
}
