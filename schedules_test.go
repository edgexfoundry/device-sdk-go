//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"testing"

	"github.com/edgexfoundry/device-sdk-go/mock"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func setup() {
	var loggingConfig = LoggingInfo{File: "./device-simple.log", RemoteURL: ""}
	var config = Config{Logging: loggingConfig}
	svc = &Service{c: &config}
	svc.scc = &mock.ScheduleClientMock{}
	svc.scec = &mock.ScheduleEventClientMock{}
	svc.ac = &mock.AddressableClientMock{}
	svc.lc = logger.NewClient("test_service", false, svc.c.Logging.File)
}

func TestNewSchedules(t *testing.T) {
	setup()

	var defaultSchedules = []models.Schedule{
		{Name: "hourly"},
		{Name: "daily"},
	}
	var defaultScheduleEvents = []models.ScheduleEvent{
		{Name: "hourly Read", Schedule: "hourly"},
		{Name: "daily clean", Schedule: "daily"},
	}

	var config = Config{
		Schedules:      defaultSchedules,
		ScheduleEvents: defaultScheduleEvents,
	}

	scheduleCache := getScheduleCache(&config)

	if len(scheduleCache.schedules) != len(defaultSchedules) && len(scheduleCache.scheduleEvents) != len(defaultScheduleEvents) {
		t.Error("Not expect getScheduleCache result!")
	}

}

func TestDefaultScheduleIsExisted(t *testing.T) {
	setup()

	var scheduleName = "test-schedule-name"
	var expect = true

	var result = isScheduleExist(scheduleName)

	if result != expect {
		t.Error("Schedule not exist!")
	}
}

func TestScheduleCache_GetScheduleByName(t *testing.T) {
	var scheduleName = "5sec-schedule"

	expect := models.Schedule{Name: scheduleName}

	var defaultSchedules = []models.Schedule{expect}

	var scheduleCache = &ScheduleCache{schedules: defaultSchedules}

	_, err := scheduleCache.GetScheduleByName(scheduleName)

	if err != nil {
		t.Fatal(err)
	}
}

func TestScheduleCache_GetScheduleEventByName(t *testing.T) {
	var eventName = "readTemperature"

	expect := models.ScheduleEvent{Name: eventName}

	var defaultScheduleEvents = []models.ScheduleEvent{expect}

	var scheduleCache = &ScheduleCache{scheduleEvents: defaultScheduleEvents}

	_, err := scheduleCache.GetScheduleEventByName(eventName)

	if err != nil {
		t.Fatal(err)
	}
}

func TestScheduleCache_UpdateSchedule(t *testing.T) {
	var defaultSchedules = []models.Schedule{
		{Name: "5sec-schedule", RunOnce: false},
	}
	var scheduleCache = &ScheduleCache{schedules: defaultSchedules}
	var schedule = &models.Schedule{Name: "5sec-schedule", RunOnce: true}

	err := scheduleCache.UpdateSchedule(schedule)

	updateScheduleEvent, err := scheduleCache.GetScheduleByName("5sec-schedule")
	if updateScheduleEvent.RunOnce != true || err != nil {
		t.Fatal(err)
	}
}

func TestScheduleCache_UpdateScheduleEvent(t *testing.T) {
	var defaultScheduleEvents = []models.ScheduleEvent{
		{Name: "readTemperature", Schedule: "5sec-schedule"},
	}
	var scheduleCache = &ScheduleCache{scheduleEvents: defaultScheduleEvents}
	var scheduleEvent = &models.ScheduleEvent{Name: "readTemperature", Schedule: "10sec-schedule"}

	err := scheduleCache.UpdateScheduleEvent(scheduleEvent)

	updateScheduleEvent, err := scheduleCache.GetScheduleEventByName("readTemperature")
	if updateScheduleEvent.Schedule != "10sec-schedule" || err != nil {
		t.Fatal(err)
	}
}
