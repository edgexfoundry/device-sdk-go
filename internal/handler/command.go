// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.
func CommandHandler(vars map[string]string, body string, method string) (*models.Event, common.AppError) {
	dKey := vars["id"]
	cmd := vars["command"]

	var ok bool
	var d models.Device
	if dKey != "" {
		d, ok = cache.Devices().ForId(dKey)
	} else {
		dKey = vars["name"]
		d, ok = cache.Devices().ForName(dKey)
	}
	if !ok {
		msg := fmt.Sprintf("Device: %s not found; %s", dKey, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewNotFoundError(msg, nil)
	}

	if d.AdminState == "LOCKED" {
		msg := fmt.Sprintf("%s is locked; %s", d.Name, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewLockedError(msg, nil)
	}

	// TODO: need to mark device when operation in progress, so it can't be removed till completed

	// NOTE: as currently implemented, CommandExists checks the existence of a deviceprofile
	// *resource* name, not a *command* name! A deviceprofile's command section is only used
	// to trigger valuedescriptor creation.
	exists, err := cache.Profiles().CommandExists(d.Profile.Name, cmd)

	// TODO: once cache locking has been implemented, this should never happen
	if err != nil {
		msg := fmt.Sprintf("internal error; Device: %s searching %s in cache failed; %s", d.Name, cmd, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewServerError(msg, err)
	}

	if !exists {
		msg := fmt.Sprintf("%s for Device: %s not found; %s", cmd, d.Name, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewNotFoundError(msg, nil)
	}

	if strings.ToLower(method) == "get" {
		return execReadCmd(&d, cmd)
	} else {
		appErr := execWriteCmd(&d, cmd, body)
		return nil, appErr
	}
}

func execReadCmd(device *models.Device, cmd string) (*models.Event, common.AppError) {
	readings := make([]models.Reading, 0, common.CurrentConfig.Device.MaxCmdOps)

	// make ResourceOperations
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, "get")
	if err != nil {
		common.LoggingClient.Error(err.Error())
		return nil, common.NewNotFoundError(err.Error(), err)
	}

	if len(ros) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("Handler - execReadCmd: MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: GET",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LoggingClient.Error(msg)
		return nil, common.NewServerError(msg, nil)
	}

	reqs := make([]ds_models.CommandRequest, len(ros))

	for i, op := range ros {
		drName := op.Object
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: deviceResource: %s", drName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, drName)
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: deviceResource: %v", dr))
		if !ok {
			msg := fmt.Sprintf("Handler - execReadCmd: no deviceResource: %s for dev: %s cmd: %s method: GET", drName, device.Name, cmd)
			common.LoggingClient.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		reqs[i].RO = op
		reqs[i].DeviceResource = dr
	}

	results, err := common.Driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		msg := fmt.Sprintf("Handler - execReadCmd: error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return nil, common.NewServerError(msg, err)
	}

	var transformsOK bool = true

	for _, cv := range results {
		// get the device resource associated with the rsp.RO
		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, cv.RO.Object)
		if !ok {
			msg := fmt.Sprintf("Handler - execReadCmd: no deviceResource: %s for dev: %s in Command Result %v", cv.RO.Object, device.Name, cv)
			common.LoggingClient.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		if common.CurrentConfig.Device.DataTransform {
			err = transformer.TransformReadResult(cv, dr.Properties.Value)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf("Handler - execReadCmd: CommandValue (%s) transformed failed: %v", cv.String(), err))
				transformsOK = false
			}
		}

		err = transformer.CheckAssertion(cv, dr.Properties.Value.Assertion, device)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("Handler - execReadCmd: Assertion failed for device resource: %s, with value: %v", cv.String(), err))
			cv = ds_models.NewStringValue(cv.RO, cv.Origin, fmt.Sprintf("Assertion failed for device resource, with value: %s and assertion: %s", cv.String(), dr.Properties.Value.Assertion))
		}

		if len(cv.RO.Mappings) > 0 {
			newCV, ok := transformer.MapCommandValue(cv)
			if ok {
				cv = newCV
			} else {
				common.LoggingClient.Warn(fmt.Sprintf("Handler - execReadCmd: Resource Operation (%v) mapping value (%s) failed with the mapping table: %v", cv.RO, cv.String(), cv.RO.Mappings))
				//transformsOK = false  // issue #89 will discuss how to handle there is no mapping matched
			}
		}

		// TODO: the Java SDK supports a RO secondary device resource(object).
		// If defined, then a RO result will generate a reading for the
		// secondary object. As this use case isn't defined and/or used in
		// any of the existing Java device services, this concept hasn't
		// been implemened in gxds. TBD at the devices f2f whether this
		// be killed completely.

		reading := common.CommandValueToReading(cv, device.Name)
		readings = append(readings, *reading)

		common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: device: %s RO: %v reading: %v", device.Name, cv.RO, reading))
	}

	if !transformsOK {
		msg := fmt.Sprintf("Transform failed for dev: %s cmd: %s method: GET", device.Name, cmd)
		common.LoggingClient.Error(msg)
		common.LoggingClient.Debug(fmt.Sprintf("Readings: %v", readings))
		return nil, common.NewServerError(msg, nil)
	}

	// push to Core Data
	event := &models.Event{Device: device.Name, Readings: readings}
	event.Origin = time.Now().UnixNano() / int64(time.Millisecond)
	go common.SendEvent(event)

	// TODO: enforce config.MaxCmdValueLen; need to include overhead for
	// the rest of the reading JSON + Event JSON length?  Should there be
	// a separate JSON body max limit for retvals & command parameters?

	return event, nil
}

func execWriteCmd(device *models.Device, cmd string, params string) common.AppError {
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, "set")
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: can't find ResrouceOperations in Profile(%s) and Command(%s), %v", device.Profile.Name, cmd, err)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	if len(ros) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("Handler - execWriteCmd: MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: PUT",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LoggingClient.Error(msg)
		return common.NewServerError(msg, nil)
	}

	roMap := roSliceToMap(ros)

	cvs, err := parseWriteParams(device.Profile.Name, roMap, params)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: Put parameters parsing failed: %s", params)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	reqs := make([]ds_models.CommandRequest, len(cvs))
	for i, cv := range cvs {
		drName := cv.RO.Object
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteCmd: putting deviceResource: %s", drName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, drName)
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteCmd: putting deviceResource: %v", dr))
		if !ok {
			msg := fmt.Sprintf("Handler - execWriteCmd: no deviceResource: %s for dev: %s cmd: %s method: GET", drName, device.Name, cmd)
			common.LoggingClient.Error(msg)
			return common.NewServerError(msg, nil)
		}

		reqs[i].RO = *cv.RO
		reqs[i].DeviceResource = dr

		if common.CurrentConfig.Device.DataTransform {
			err = transformer.TransformWriteParameter(cv, dr.Properties.Value)
			if err != nil {
				msg := fmt.Sprintf("Handler - execWriteCmd: CommandValue (%s) transformed failed: %v", cv.String(), err)
				common.LoggingClient.Error(msg)
				return common.NewServerError(msg, err)
			}
		}
	}

	err = common.Driver.HandleWriteCommands(device.Name, device.Protocols, reqs, cvs)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteCmd: error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return common.NewServerError(msg, err)
	}

	return nil
}

func parseWriteParams(profileName string, roMap map[string]*models.ResourceOperation, params string) ([]*ds_models.CommandValue, error) {
	var paramMap map[string]string
	err := json.Unmarshal([]byte(params), &paramMap)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Handler - parseWriteParams: parsing Write parameters failed %s, %v", params, err))
		return []*ds_models.CommandValue{}, err
	}

	if len(paramMap) == 0 {
		err := fmt.Errorf("no parameters specified")
		return []*ds_models.CommandValue{}, err
	}

	result := make([]*ds_models.CommandValue, 0, len(paramMap))
	for k, v := range paramMap {
		ro, ok := roMap[k]
		if ok {
			if len(ro.Mappings) > 0 {
				newV, ok := ro.Mappings[v]
				if ok {
					v = newV
				} else {
					msg := fmt.Sprintf("Handler - parseWriteParams: Resource Operation (%v) mapping value (%s) failed with the mapping table: %v", ro, v, ro.Mappings)
					common.LoggingClient.Warn(msg)
					//return result, fmt.Errorf(msg) // issue #89 will discuss how to handle there is no mapping matched
				}
			}
			cv, err := createCommandValueForParam(profileName, ro, v)
			if err == nil {
				result = append(result, cv)
			} else {
				return result, err
			}
		} else {
			err := fmt.Errorf("the parameter %s cannot find the matched ResourceOperation", k)
			return []*ds_models.CommandValue{}, err
		}
	}

	return result, nil
}

func roSliceToMap(ros []models.ResourceOperation) map[string]*models.ResourceOperation {
	roMap := make(map[string]*models.ResourceOperation, len(ros))
	for i, ro := range ros {
		roMap[ro.Parameter] = &ros[i]
	}
	return roMap
}

func createCommandValueForParam(profileName string, ro *models.ResourceOperation, v string) (*ds_models.CommandValue, error) {
	var result *ds_models.CommandValue
	var err error
	var value interface{}
	var t ds_models.ValueType

	dr, ok := cache.Profiles().DeviceResource(profileName, ro.Object)
	if !ok {
		msg := fmt.Sprintf("createCommandValueForParam: no deviceResource: %s", ro.Object)
		common.LoggingClient.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	origin := time.Now().UnixNano() / int64(time.Millisecond)

	switch strings.ToLower(dr.Properties.Value.Type) {
	case "bool":
		value, err = strconv.ParseBool(v)
		t = ds_models.Bool
	case "string":
		value = v
		t = ds_models.String
	case "uint8":
		n, e := strconv.ParseUint(v, 10, 8)
		value = uint8(n)
		err = e
		t = ds_models.Uint8
	case "uint16":
		n, e := strconv.ParseUint(v, 10, 16)
		value = uint16(n)
		err = e
		t = ds_models.Uint16
	case "uint32":
		n, e := strconv.ParseUint(v, 10, 32)
		value = uint32(n)
		err = e
		t = ds_models.Uint32
	case "uint64":
		value, err = strconv.ParseUint(v, 10, 64)
		t = ds_models.Uint64
	case "int8":
		n, e := strconv.ParseInt(v, 10, 8)
		value = int8(n)
		err = e
		t = ds_models.Int8
	case "int16":
		n, e := strconv.ParseInt(v, 10, 16)
		value = int16(n)
		err = e
		t = ds_models.Int16
	case "int32":
		n, e := strconv.ParseInt(v, 10, 32)
		value = int32(n)
		err = e
		t = ds_models.Int32
	case "int64":
		value, err = strconv.ParseInt(v, 10, 64)
		t = ds_models.Int64
	case "float32":
		n, e := strconv.ParseFloat(v, 32)
		value = float32(n)
		err = e
		t = ds_models.Float32
	case "float64":
		value, err = strconv.ParseFloat(v, 64)
		t = ds_models.Float64
	}

	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Handler - Command: Parsing parameter value (%s) to %s failed: %v", v, dr.Properties.Value.Type, err))
		return result, err
	}

	result, err = ds_models.NewCommandValue(ro, origin, value, t)

	return result, err
}

func CommandAllHandler(cmd string, body string, method string) ([]*models.Event, common.AppError) {
	common.LoggingClient.Debug(fmt.Sprintf("Handler - CommandAll: execute the %s command %s from all operational devices", method, cmd))
	devices := filterOperationalDevices(cache.Devices().All())

	devCount := len(devices)
	var waitGroup sync.WaitGroup
	waitGroup.Add(devCount)
	cmdResults := make(chan struct {
		event  *models.Event
		appErr common.AppError
	}, devCount)

	for i, _ := range devices {
		go func(device *models.Device) {
			defer waitGroup.Done()
			var event *models.Event = nil
			var appErr common.AppError = nil
			if strings.ToLower(method) == "get" {
				event, appErr = execReadCmd(device, cmd)
			} else {
				appErr = execWriteCmd(device, cmd, body)
			}
			cmdResults <- struct {
				event  *models.Event
				appErr common.AppError
			}{event, appErr}
		}(devices[i])
	}
	waitGroup.Wait()
	close(cmdResults)

	errCount := 0
	getResults := make([]*models.Event, 0, devCount)
	var appErr common.AppError
	for r := range cmdResults {
		if r.appErr != nil {
			errCount++
			common.LoggingClient.Error("Handler - CommandAll: " + r.appErr.Message())
			appErr = r.appErr // only the last error will be returned
		} else if r.event != nil {
			getResults = append(getResults, r.event)
		}
	}

	if errCount < devCount {
		common.LoggingClient.Info("Handler - CommandAll: part of commands executed successfully, returning 200 OK")
		appErr = nil
	}

	return getResults, appErr

}

func filterOperationalDevices(devices []models.Device) []*models.Device {
	result := make([]*models.Device, 0, len(devices))
	for i, d := range devices {
		if (d.AdminState == models.Locked) || (d.OperatingState == models.Disabled) {
			continue
		}
		result = append(result, &devices[i])
	}
	return result
}
