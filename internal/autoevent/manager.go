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

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type Manager interface {
	StartAutoEvents(dic *di.Container) bool
	StopAutoEvents()
	RestartForDevice(deviceName string, dic *di.Container)
	StopForDevice(deviceName string)
}

type manager struct {
	execsMap map[string][]Executor
	ctx      context.Context
	wg       *sync.WaitGroup
}

var (
	createOnce sync.Once
	m          *manager
	mutex      sync.Mutex
)

// NewManager initiates the AutoEvent manager once
func NewManager(ctx context.Context, wg *sync.WaitGroup) {
	m = &manager{execsMap: make(map[string][]Executor), ctx: ctx, wg: wg}
}

func (m *manager) StartAutoEvents(dic *di.Container) bool {
	mutex.Lock()
	defer mutex.Unlock()
	createOnce.Do(func() {
		for _, d := range cache.Devices().All() {
			execs := triggerExecutors(d.Name, d.AutoEvents, m.ctx, m.wg, dic)
			m.execsMap[d.Name] = execs
		}
	})

	return true
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

func triggerExecutors(deviceName string, autoEvents []contract.AutoEvent, ctx context.Context, wg *sync.WaitGroup, dic *di.Container) []Executor {
	var execs []Executor
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	for _, autoEvent := range autoEvents {
		exec, err := NewExecutor(deviceName, autoEvent)
		if err != nil {
			lc.Error(fmt.Sprintf("AutoEvent for resource %s cannot be created, %v", autoEvent.Resource, err))
			// skip this AutoEvent if it causes error during creation
			continue
		}
		execs = append(execs, exec)
		go exec.Run(ctx, wg, dic)
	}
	return execs
}

// RestartForDevice stops all the AutoEvents of the specific Device
func (m *manager) RestartForDevice(deviceName string, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	m.StopForDevice(deviceName)
	d, ok := cache.Devices().ForName(deviceName)
	if !ok {
		lc.Error(fmt.Sprintf("there is no Device %s in cache to start AutoEvent", deviceName))
	}

	mutex.Lock()
	defer mutex.Unlock()
	execs := triggerExecutors(deviceName, d.AutoEvents, m.ctx, m.wg, dic)
	m.execsMap[deviceName] = execs
}

// StopForDevice stops all the AutoEvents of the specific Device
func (m *manager) StopForDevice(deviceName string) {
	mutex.Lock()
	defer mutex.Unlock()
	execs, ok := m.execsMap[deviceName]
	if ok {
		for _, e := range execs {
			e.Stop()
		}
		delete(m.execsMap, deviceName)
	}
}

// GetManager returns Manager instance
func GetManager() Manager {
	return m
}
