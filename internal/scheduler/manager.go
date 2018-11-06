// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/robfig/cron"
)

var (
	schMgrOnce sync.Once
	cr         *cron.Cron
)

func StartScheduler() {
	schMgrOnce.Do(func() {
		cr = cron.New()
		schEvtExecs := loadSchEvts()
		for i, _ := range schEvtExecs {
			common.LoggingClient.Info(fmt.Sprintf("Initializing Schedule Event Executor: %v", *schEvtExecs[i]))
			spec, err := schEvtExecs[i].cronSpec()
			if err != nil {
				common.LoggingClient.Error(err.Error())
				continue
			}
			cr.AddJob(spec, schEvtExecs[i])
		}
		common.LoggingClient.Info("Starting internal Scheduler")
		cr.Start()
		common.LoggingClient.Info("Started internal Scheduler")
	})
}

func StopScheduler() {
	common.LoggingClient.Info("Stopping internal Scheduler")
	cr.Stop()
	common.LoggingClient.Info("Stopped internal Scheduler")
}

func loadSchEvts() []*schEvtExec {
	schEvts := cache.ScheduleEvents().All()
	result := make([]*schEvtExec, len(schEvts))
	for i, schEvt := range schEvts {
		common.LoggingClient.Debug(fmt.Sprintf("Loading Schedule Event %s", schEvt.Name))
		sch, ok := cache.Schedules().ForName(schEvt.Schedule)
		if !ok {
			common.LoggingClient.Error(fmt.Sprintf("Schedule %s for Schedule Event %s cannot be found in cache", schEvt.Schedule, schEvt.Name))
			continue
		}
		exec := schEvtExec{schEvt: schEvt, sch: sch}
		result[i] = &exec
	}
	return result
}
