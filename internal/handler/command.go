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
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/ugorji/go/codec"
)

// Note, every HTTP request to ServeHTTP is made in a separate goroutine, which
// means care needs to be taken with respect to shared data accessed through *Server.
func CommandHandler(
	vars map[string]string,
	body string,
	method string,
	queryParams string,
	r *http.Request) (*dsModels.Event, common.AppError) {

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

	if d.OperatingState == contract.Disabled {
		msg := fmt.Sprintf("%s is disabled; %s", d.Name, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewLockedError(msg, nil)
	}

	// TODO: need to mark device when operation in progress, so it can't be removed till completed

	cmdExists, err := cache.Profiles().CommandExists(d.Profile.Name, cmd, method)

	// TODO: once cache locking has been implemented, this should never happen
	if err != nil {
		msg := fmt.Sprintf("internal error; Device: %s searching %s in cache failed; %s", d.Name, cmd, method)
		common.LoggingClient.Error(msg)
		return nil, common.NewServerError(msg, err)
	}

	if !cmdExists {
		dr, drExists := cache.Profiles().DeviceResource(d.Profile.Name, cmd)
		if !drExists {
			msg := fmt.Sprintf("%s for Device: %s not found; %s", cmd, d.Name, method)
			common.LoggingClient.Error(msg)
			return nil, common.NewNotFoundError(msg, nil)
		}

		if strings.ToLower(method) == common.GetCmdMethod {
			return execReadDeviceResource(&d, &dr, queryParams)
		} else {
			appErr := execWriteDeviceResource(&d, &dr, body)
			return nil, appErr
		}
	}

	if strings.ToLower(method) == common.GetCmdMethod {
		return execReadCmd(&d, cmd, queryParams)
	} else {
		appErr := execWriteCmd(&d, cmd, body, r)
		return nil, appErr
	}
}

func execReadDeviceResource(
	device *contract.Device,
	dr *contract.DeviceResource,
	queryParams string) (*dsModels.Event, common.AppError) {

	var reqs []dsModels.CommandRequest
	var req dsModels.CommandRequest
	common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: deviceResource: %s", dr.Name))

	req.DeviceResourceName = dr.Name
	req.Attributes = dr.Attributes
	if queryParams != "" {
		if len(req.Attributes) <= 0 {
			req.Attributes = make(map[string]string)
		}
		m := common.FilterQueryParams(queryParams)
		req.Attributes[common.URLRawQuery] = m.Encode()
	}
	req.Type = dsModels.ParseValueType(dr.Properties.Value.Type)
	reqs = append(reqs, req)

	results, err := common.Driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		msg := fmt.Sprintf(
			"Handler - execReadCmd: error for Device: %s DeviceResource: %s, %v",
			device.Name,
			dr.Name,
			err)
		return nil, common.NewServerError(msg, err)
	}

	return cvsToEvent(device, results, dr.Name)
}

func cvsToEvent(device *contract.Device, cvs []*dsModels.CommandValue, cmd string) (*dsModels.Event, common.AppError) {
	readings := make([]contract.Reading, 0, common.CurrentConfig.Device.MaxCmdOps)
	var transformsOK = true
	var err error

	for _, cv := range cvs {
		// get the device resource associated with the rsp.RO
		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, cv.DeviceResourceName)
		if !ok {
			msg := fmt.Sprintf(
				"Handler - execReadCmd: no deviceResource: %s for dev: %s in Command Result %v",
				cv.DeviceResourceName,
				device.Name,
				cv)
			common.LoggingClient.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		if common.CurrentConfig.Device.DataTransform {
			err = transformer.TransformReadResult(cv, dr.Properties.Value)
			if err != nil {
				common.LoggingClient.Error(fmt.Sprintf(
					"Handler - execReadCmd: CommandValue (%s) transformed failed: %v",
					cv.String(),
					err))
				transformsOK = false
			}
		}

		err = transformer.CheckAssertion(cv, dr.Properties.Value.Assertion, device)
		if err != nil {
			common.LoggingClient.Error(fmt.Sprintf(
				"Handler - execReadCmd: Assertion failed for device resource: %s, with value: %v",
				cv.String(),
				err))
			cv = dsModels.NewStringValue(
				cv.DeviceResourceName,
				cv.Origin,
				fmt.Sprintf(
					"Assertion failed for device resource, with value: %s and assertion: %s",
					cv.String(),
					dr.Properties.Value.Assertion))
		}

		ro, err := cache.Profiles().ResourceOperation(device.Profile.Name, cv.DeviceResourceName, common.GetCmdMethod)
		if err != nil {
			common.LoggingClient.Debug(fmt.Sprintf("getting resource operation failed: %s", err.Error()))
		} else if len(ro.Mappings) > 0 {
			newCV, ok := transformer.MapCommandValue(cv, ro.Mappings)
			if ok {
				cv = newCV
			} else {
				common.LoggingClient.Warn(fmt.Sprintf(
					"Handler - execReadCmd: Resource Operation (%s) mapping value (%s) failed with the mapping table: %v",
					ro.DeviceCommand,
					cv.String(),
					ro.Mappings))
				//transformsOK = false  // issue #89 will discuss how to handle there is no mapping matched
			}
		}

		// TODO: the Java SDK supports a RO secondary device resource.
		// If defined, then a RO result will generate a reading for the
		// secondary device resource. As this use case isn't defined and/or used in
		// any of the existing Java device services, this concept hasn't
		// been implemened in gxds. TBD at the devices f2f whether this
		// be killed completely.

		reading := common.CommandValueToReading(cv, device.Name, dr.Properties.Value.FloatEncoding)
		readings = append(readings, *reading)

		common.LoggingClient.Debug(fmt.Sprintf(
			"Handler - execReadCmd: device: %s DeviceResource: %v reading: %v",
			device.Name,
			cv.DeviceResourceName,
			reading))
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
	event.Origin = common.GetUniqueOrigin()

	// TODO: enforce config.MaxCmdValueLen; need to include overhead for
	// the rest of the reading JSON + Event JSON length?  Should there be
	// a separate JSON body max limit for retvals & command parameters?

	return event, nil
}

func execReadCmd(device *contract.Device, cmd string, queryParams string) (*dsModels.Event, common.AppError) {
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
		drName := op.DeviceResource
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: deviceResource: %s", drName))

		// TODO: add recursive support for resource command chaining. This occurs when a
		// deviceprofile resource command operation references another resource command
		// instead of a device resource (see BoschXDK for reference).

		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, drName)
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execReadCmd: deviceResource: %v", dr))
		if !ok {
			msg := fmt.Sprintf(
				"Handler - execReadCmd: no deviceResource: %s for dev: %s cmd: %s method: GET",
				drName,
				device.Name,
				cmd)
			common.LoggingClient.Error(msg)
			return nil, common.NewServerError(msg, nil)
		}

		reqs[i].DeviceResourceName = dr.Name
		reqs[i].Attributes = dr.Attributes
		if queryParams != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]string)
			}
			m := common.FilterQueryParams(queryParams)
			reqs[i].Attributes[common.URLRawQuery] = m.Encode()
		}
		reqs[i].Type = dsModels.ParseValueType(dr.Properties.Value.Type)
	}

	results, err := common.Driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		msg := fmt.Sprintf("Handler - execReadCmd: error for Device: %s cmd: %s, %v", device.Name, cmd, err)
		return nil, common.NewServerError(msg, err)
	}

	return cvsToEvent(device, results, cmd)
}

func execWriteDeviceResource(device *contract.Device, dr *contract.DeviceResource, params string) common.AppError {
	paramMap, err := parseParams(params, nil)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteDeviceResource: Put parameters parsing failed: %s", params)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	v, ok := paramMap[dr.Name]
	if !ok && dr.Properties.Value.DefaultValue != "" {
		v = dr.Properties.Value.DefaultValue
	} else if !ok {
		msg := fmt.Sprintf("there is no %s in parameters and no default value in DeviceResource", dr.Name)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, fmt.Errorf(msg))
	}

	cv, err := createCommandValueFromDR(dr, v)
	if err != nil {
		msg := fmt.Sprintf("Handler - execWriteDeviceResource: Put parameters parsing failed: %s", params)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	reqs := make([]dsModels.CommandRequest, 1)
	common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteDeviceResource: putting deviceResource: %s", dr.Name))
	reqs[0].DeviceResourceName = cv.DeviceResourceName
	reqs[0].Attributes = dr.Attributes
	reqs[0].Type = cv.Type

	if common.CurrentConfig.Device.DataTransform {
		err = transformer.TransformWriteParameter(cv, dr.Properties.Value)
		if err != nil {
			msg := fmt.Sprintf(
				"Handler - execWriteDeviceResource: CommandValue (%s) transformed failed: %v",
				cv.String(),
				err)
			common.LoggingClient.Error(msg)
			return common.NewServerError(msg, err)
		}
	}

	err = common.Driver.HandleWriteCommands(device.Name, device.Protocols, reqs, []*dsModels.CommandValue{cv})
	if err != nil {
		msg := fmt.Sprintf(
			"Handler - execWriteDeviceResource: error for Device: %s Device Resource: %s, %v",
			device.Name,
			dr.Name,
			err)
		return common.NewServerError(msg, err)
	}

	return nil
}

func execWriteCmd(device *contract.Device, cmd string, params string, r *http.Request) common.AppError {
	ros, err := cache.Profiles().ResourceOperations(device.Profile.Name, cmd, common.SetCmdMethod)
	if err != nil {
		msg := fmt.Sprintf(
			"Handler - execWriteCmd: can't find ResrouceOperations in Profile(%s) and Command(%s), %v",
			device.Profile.Name,
			cmd,
			err)
		common.LoggingClient.Error(msg)
		return common.NewBadRequestError(msg, err)
	}

	if len(ros) > common.CurrentConfig.Device.MaxCmdOps {
		msg := fmt.Sprintf("Handler - execWriteCmd: MaxCmdOps (%d) execeeded for dev: %s cmd: %s method: PUT",
			common.CurrentConfig.Device.MaxCmdOps, device.Name, cmd)
		common.LoggingClient.Error(msg)
		return common.NewServerError(msg, nil)
	}

	cvs, err := parseWriteParams(device.Profile.Name, ros, params, r)
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
		common.LoggingClient.Debug(fmt.Sprintf("Handler - execWriteCmd: putting deviceResource: %s", drName))
		if !ok {
			msg := fmt.Sprintf(
				"Handler - execWriteCmd: no deviceResource: %s for dev: %s cmd: %s method: GET",
				drName,
				device.Name,
				cmd)
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

func parseWriteParams(profileName string,
	ros []contract.ResourceOperation,
	params string,
	r *http.Request) ([]*dsModels.CommandValue, error) {

	paramMap, err := parseParams(params, r)
	if err != nil {
		return []*dsModels.CommandValue{}, err
	}

	result := make([]*dsModels.CommandValue, 0, len(paramMap))
	for _, ro := range ros {
		common.LoggingClient.Debug(fmt.Sprintf("looking for %s in the request parameters", ro.DeviceResource))
		p, ok := paramMap[ro.DeviceResource]
		if !ok {
			dr, ok := cache.Profiles().DeviceResource(profileName, ro.DeviceResource)
			if !ok {
				err := fmt.Errorf(
					"the parameter %s does not match any DeviceResource in DeviceProfile",
					ro.DeviceResource)
				return []*dsModels.CommandValue{}, err
			}

			if ro.Parameter != "" {
				common.LoggingClient.Debug(fmt.Sprintf("there is no %s in the request parameters, retrieving value from the Parameter field from the ResourceOperation", ro.DeviceResource))
				p = ro.Parameter
			} else if dr.Properties.Value.DefaultValue != "" {
				common.LoggingClient.Debug(fmt.Sprintf("there is no %s in the request parameters, retrieving value from the DefaultValue field from the ValueProperty", ro.DeviceResource))
				p = dr.Properties.Value.DefaultValue
			} else {
				err := fmt.Errorf(
					"the parameter %s is not defined in the request body and there is no default value",
					ro.DeviceResource)
				return []*dsModels.CommandValue{}, err
			}
		}

		if len(ro.Mappings) > 0 {
			newP, ok := ro.Mappings[p]
			if ok {
				p = newP
			} else {
				msg := fmt.Sprintf(
					"parseWriteParams: Resource (%s) mapping value (%s) failed with the mapping table: %v",
					ro.DeviceResource,
					p,
					ro.Mappings)
				common.LoggingClient.Warn(msg)
				//return result, fmt.Errorf(msg) // issue #89 will discuss how to handle there is no mapping matched
			}
		}

		cv, err := createCommandValueFromRO(profileName, &ro, p)
		if err == nil {
			result = append(result, cv)
		} else {
			return result, err
		}
	}

	return result, nil
}

func parseParams(params string, r *http.Request) (paramMap map[string]string, err error) {

	// Check the header value
	if r.Header.Get(clients.ContentType) == clients.ContentTypeCBOR {
		// Handle Content-type CBOR and populate paramMap
		err = codec.NewDecoderBytes([]byte(params), &codec.CborHandle{}).Decode(&paramMap)
	} else if r.Header.Get(clients.ContentType) == clients.ContentTypeJSON {
		// Handle Content-type JSON and populate paramMap
		err = json.Unmarshal([]byte(params), &paramMap)
	}

	if err != nil {
		common.LoggingClient.Error(fmt.Sprintf("parsing Write parameters failed %s, %v", params, err))
		return
	}

	if len(paramMap) == 0 {
		err = fmt.Errorf("no parameters specified")
		return
	}
	return
}

func createCommandValueFromRO(
	profileName string,
	ro *contract.ResourceOperation,
	v string) (*dsModels.CommandValue, error) {

	dr, ok := cache.Profiles().DeviceResource(profileName, ro.DeviceResource)
	if !ok {
		msg := fmt.Sprintf("createCommandValueForParam: no deviceResource: %s", ro.DeviceResource)
		common.LoggingClient.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	return createCommandValueFromDR(&dr, v)
}

func createCommandValueFromDR(dr *contract.DeviceResource, v string) (*dsModels.CommandValue, error) {
	var result *dsModels.CommandValue
	var err error
	var value interface{}
	var t dsModels.ValueType
	origin := time.Now().UnixNano()

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
		common.LoggingClient.Error(fmt.Sprintf(
			"Handler - Command: Parsing parameter value (%s) to %s failed: %v",
			v,
			dr.Properties.Value.Type,
			err))
		return result, err
	}

	result, err = dsModels.NewCommandValue(dr.Name, origin, value, t)

	return result, err
}

func CommandAllHandler(
	cmd string,
	body string,
	method string,
	queryParams string) ([]*dsModels.Event, common.AppError) {

	common.LoggingClient.Debug(fmt.Sprintf(
		"Handler - CommandAll: execute the %s command %s from all operational devices",
		method,
		cmd))
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
				event, appErr = execReadCmd(device, cmd, queryParams)
			} else {
				appErr = execWriteCmd(device, cmd, body, nil)
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
