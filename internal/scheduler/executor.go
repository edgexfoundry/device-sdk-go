// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	nameVar    string = "name"
	commandVar string = "command"
)

type schEvtExec struct {
	sch    models.Schedule
	schEvt models.ScheduleEvent
}

func (se *schEvtExec) Run() {
	isCmd, err := path.Match(common.SchedulerExecCMDPattern, se.schEvt.Addressable.Path)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Schedule Event Path parsing failed: %v, %v", se.schEvt, err))
	}
	if isCmd {
		execCmd(se)
		return
	}

	isDiscovery, err := path.Match(common.APIDiscoveryRoute, se.schEvt.Addressable.Path)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Schedule Event Path parsing failed: %v, %v", se.schEvt, err))
	}
	if isDiscovery {
		handler.DiscoveryHandler(nil)
		return
	}

	common.LoggingClient.Error(fmt.Sprintf("There is no correct execution for Schedule Event: %v", se.schEvt))
}

func execCmd(se *schEvtExec) {
	addr := se.schEvt.Addressable
	deviceName, cmdName, err := parseCmdPath(addr.Path)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Schedule Event execution failed: %v, %v", se.schEvt, err))
		return
	}
	vars := make(map[string]string, 2)
	vars[nameVar] = deviceName
	vars[commandVar] = cmdName
	evt, appErr := handler.CommandHandler(vars, se.schEvt.Parameters, addr.HTTPMethod)
	if appErr != nil {
		common.LoggingClient.Error(fmt.Sprintf("Schecule Event %s execution failed, AppErr: %v", se.schEvt.Name, appErr))
		return
	}
	common.LoggingClient.Debug(fmt.Sprintf("Schecule Event %s executed result- Event: %v, AppErr: %v", se.schEvt.Name, evt, appErr))
}

func parseCmdPath(path string) (deviceName string, cmdName string, err error) {
	sections := strings.Split(path, "/")
	if len(sections) != 7 {
		return "", "", fmt.Errorf("parsing Command path failed: %s", path)
	} else {
		return sections[5], sections[6], nil
	}
}

func (se *schEvtExec) cronSpec() (string, error) {
	duration, err := iso8601ToDuration(se.sch.Frequency)
	return fmt.Sprintf("@every %v", duration), err
}

func iso8601ToDuration(str string) (time.Duration, error) {
	durationRegex, err := regexp.Compile(`P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("compile Regular Expression for ISO 8601 format failed: %v", err))
		return time.Duration(0), err
	}
	matches := durationRegex.FindStringSubmatch(str)
	if len(matches) == 0 {
		return time.Duration(0), fmt.Errorf("parsing ISO 8601 format to Duration failed")
	}

	var years, months, days, hours, minutes, seconds int64
	if years, err = parseInt64(matches[1]); err != nil {
		return time.Duration(0), err
	}
	if months, err = parseInt64(matches[2]); err != nil {
		return time.Duration(0), err
	}
	if days, err = parseInt64(matches[3]); err != nil {
		return time.Duration(0), err
	}
	if hours, err = parseInt64(matches[4]); err != nil {
		return time.Duration(0), err
	}
	if minutes, err = parseInt64(matches[5]); err != nil {
		return time.Duration(0), err
	}
	if seconds, err = parseInt64(matches[6]); err != nil {
		return time.Duration(0), err
	}

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)
	return time.Duration(years*24*365*hour + months*30*24*hour + days*24*hour + hours*hour + minutes*minute + seconds*second), nil
}

func parseInt64(value string) (int64, error) {
	if len(value) == 0 {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0, err
	}
	return int64(parsed), nil
}
