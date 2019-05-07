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
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.
func CommandHandler(vars map[string]string, body string, method string) (*dsModels.Event, common.AppError) {
	dKey := vars[common.IdVar]
	cmd := vars[common.CommandVar]

	var ok bool
	var d contract.Device
	if dKey != "" {
		d, ok = cache.Devices().ForId(dKey)
	} else {
		dKey = vars[common.NameVar]
		d, ok = cache.Devices().ForName(dKey)
	}
	if !ok {
		msg := fmt.Sprintf("Device: %s not found; %s", dKey, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewNotFoundError(msg, nil)
	}

	if d.AdminState == contract.Locked {
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

	if strings.ToLower(method) == common.GetCmdMethod {
		return execReadCmd(&d, cmd)
	} else {
		appErr := execWriteCmd(&d, cmd, body)
		return nil, appErr
	}
}

func execReadCmd(device *contract.Device, cmd string) (*dsModels.Event, common.AppError) {
	readings := make([]contract.Reading, 0, common.CurrentConfig.Device.MaxCmdOps)

	// make ResourceOperations
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, common.GetCmdMethod)
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

	reqs := make([]dsModels.CommandRequest, len(ros))

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

		reqs[i].DeviceResourceName = dr.Name
		reqs[i].Attributes = dr.Attributes
		reqs[i].Type = dsModels.ParseValueType(dr.Properties.Value.Type)
	}

	results, err := common.Driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		msg := fmt.Sprintf("Handler - execReadCmd: error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return nil, common.NewServerError(msg, err)
	}

	var transformsOK bool = true

	for _, cv := range results {
		// get the device resource associated with the rsp.RO
		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, cv.DeviceResourceName)
		if !ok {
			msg := fmt.Sprintf("Handler - execReadCmd: no deviceResource: %s for dev: %s in Command Result %v", cv.DeviceResourceName, device.Name, cv)
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
			cv = dsModels.NewStringValue(cv.DeviceResourceName, cv.Origin, fmt.Sprintf("Assertion failed for device resource, with value: %s and assertion: %s", cv.String(), dr.Properties.Value.Assertion))
		}

		ro, err := cache.Profiles().ResourceOperation(device.Profile.Name, cv.DeviceResourceName, common.GetCmdMethod)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf("Handler - execReadCmd: getting resource operation failed: %s", err.Error()))
			transformsOK = false
		}

		if len(ro.Mappings) > 0 {
			newCV, ok := transformer.MapCommandValue(cv, ro.Mappings)
			if ok {
				cv = newCV
			} else {
				common.LoggingClient.Warn(fmt.Sprintf("Handler - execReadCmd: Resource Operation (%s) mapping value (%s) failed with the mapping table: %v", ro.Resource, cv.String(), ro.Mappings))
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

		common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: device: %s DeviceResource: %v reading: %v", device.Name, cv.DeviceResourceName, reading))
	}

	if !transformsOK {
		msg := fmt.Sprintf("Transform failed for dev: %s cmd: %s method: GET", device.Name, cmd)
		common.LoggingClient.Error(msg)
		common.LoggingClient.Debug(fmt.Sprintf("Readings: %v", readings))
		return nil, common.NewServerError(msg, nil)
	}

	// push to Core Data
	cevent := contract.Event{Device: device.Name, Readings: readings}
	event := &dsModels.Event{Event: cevent}
	event.Origin = time.Now().UnixNano() / int64(time.Millisecond)

	// TODO: enforce config.MaxCmdValueLen; need to include overhead for
	// the rest of the reading JSON + Event JSON length?  Should there be
	// a separate JSON body max limit for retvals & command parameters?

	return event, nil
}

func execWriteCmd(device *contract.Device, cmd string, params string) common.AppError {
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, common.SetCmdMethod)
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

	reqs := make([]dsModels.CommandRequest, len(cvs))
	for i, cv := range cvs {
		drName := cv.DeviceResourceName
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

		reqs[i].DeviceResourceName = cv.DeviceResourceName
		reqs[i].Attributes = dr.Attributes
		reqs[i].Type = cv.Type

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

func parseWriteParams(profileName string, roMap map[string]*contract.ResourceOperation, params string) ([]*dsModels.CommandValue, error) {
	var paramMap map[string]string
	err := json.Unmarshal([]byte(params), &paramMap)
	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Handler - parseWriteParams: parsing Write parameters failed %s, %v", params, err))
		return []*dsModels.CommandValue{}, err
	}

	if len(paramMap) == 0 {
		err := fmt.Errorf("no parameters specified")
		return []*dsModels.CommandValue{}, err
	}

	result := make([]*dsModels.CommandValue, 0, len(paramMap))
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
			return []*dsModels.CommandValue{}, err
		}
	}

	return result, nil
}

func roSliceToMap(ros []contract.ResourceOperation) map[string]*contract.ResourceOperation {
	roMap := make(map[string]*contract.ResourceOperation, len(ros))
	for i, ro := range ros {
		roMap[ro.Object] = &ros[i]
	}
	return roMap
}

func createCommandValueForParam(profileName string, ro *contract.ResourceOperation, v string) (*dsModels.CommandValue, error) {
	var result *dsModels.CommandValue
	var err error
	var value interface{}
	var t dsModels.ValueType

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
		t = dsModels.Bool
	case "string":
		value = v
		t = dsModels.String
	case "uint8":
		n, e := strconv.ParseUint(v, 10, 8)
		value = uint8(n)
		err = e
		t = dsModels.Uint8
	case "uint16":
		n, e := strconv.ParseUint(v, 10, 16)
		value = uint16(n)
		err = e
		t = dsModels.Uint16
	case "uint32":
		n, e := strconv.ParseUint(v, 10, 32)
		value = uint32(n)
		err = e
		t = dsModels.Uint32
	case "uint64":
		value, err = strconv.ParseUint(v, 10, 64)
		t = dsModels.Uint64
	case "int8":
		n, e := strconv.ParseInt(v, 10, 8)
		value = int8(n)
		err = e
		t = dsModels.Int8
	case "int16":
		n, e := strconv.ParseInt(v, 10, 16)
		value = int16(n)
		err = e
		t = dsModels.Int16
	case "int32":
		n, e := strconv.ParseInt(v, 10, 32)
		value = int32(n)
		err = e
		t = dsModels.Int32
	case "int64":
		value, err = strconv.ParseInt(v, 10, 64)
		t = dsModels.Int64
	case "float32":
		n, e := strconv.ParseFloat(v, 32)
		value = float32(n)
		err = e
		t = dsModels.Float32
	case "float64":
		value, err = strconv.ParseFloat(v, 64)
		t = dsModels.Float64
	}

	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("Handler - Command: Parsing parameter value (%s) to %s failed: %v", v, dr.Properties.Value.Type, err))
		return result, err
	}

	result, err = dsModels.NewCommandValue(dr.Name, origin, value, t)

	return result, err
}

func CommandAllHandler(cmd string, body string, method string) ([]*dsModels.Event, common.AppError) {
	common.LoggingClient.Debug(fmt.Sprintf("Handler - CommandAll: execute the %s command %s from all operational devices", method, cmd))
	devices := filterOperationalDevices(cache.Devices().All())

	devCount := len(devices)
	var waitGroup sync.WaitGroup
	waitGroup.Add(devCount)
	cmdResults := make(chan struct {
		event  *dsModels.Event
		appErr common.AppError
	}, devCount)

	for i, _ := range devices {
		go func(device *contract.Device) {
			defer waitGroup.Done()
			var event *dsModels.Event = nil
			var appErr common.AppError = nil
			if strings.ToLower(method) == common.GetCmdMethod {
				event, appErr = execReadCmd(device, cmd)
			} else {
				appErr = execWriteCmd(device, cmd, body)
			}
			cmdResults <- struct {
				event  *dsModels.Event
				appErr common.AppError
			}{event, appErr}
		}(devices[i])
	}
	waitGroup.Wait()
	close(cmdResults)

	errCount := 0
	getResults := make([]*dsModels.Event, 0, devCount)
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

func filterOperationalDevices(devices []contract.Device) []*contract.Device {
	result := make([]*contract.Device, 0, len(devices))
	for i, d := range devices {
		if (d.AdminState == contract.Locked) || (d.OperatingState == contract.Disabled) {
			continue
		}
		result = append(result, &devices[i])
	}
	return result
}
