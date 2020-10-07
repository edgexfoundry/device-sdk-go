// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type CommandProcessor struct {
	device         *contract.Device
	deviceResource *contract.DeviceResource
	cmd            string
	params         string
	dic            *di.Container
}

func NewCommandProcessor(device *contract.Device, dr *contract.DeviceResource, id string, cmd string, params string, dic *di.Container) *CommandProcessor {
	return &CommandProcessor{
		device:         device,
		deviceResource: dr,
		cmd:            cmd,
		params:         params,
		dic:            dic,
	}
}

func CommandHandler(isRead bool, vars map[string]string, body string, id string, dic *di.Container) (event *dsModels.Event, err edgexErr.EdgeX) {
	var device contract.Device
	deviceKey := vars[sdkCommon.IdVar]

	defer func() {
		if err == nil {
			go sdkCommon.UpdateLastConnected(
				device.Name,
				container.ConfigurationFrom(dic.Get),
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.MetadataDeviceClientFrom(dic.Get))
		}
	}()

	// check provided device exists
	device, exist := cache.Devices().ForId(deviceKey)
	if !exist {
		deviceKey = vars[sdkCommon.NameVar]
		device, exist = cache.Devices().ForName(deviceKey)
		if !exist {
			return nil, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "device not found", nil)
		}
	}

	// check device's AdminState
	if device.AdminState == contract.Locked {
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServiceLocked, fmt.Sprintf("device %s locked", device.Name), nil)
	}

	var method string
	if isRead {
		method = sdkCommon.GetCmdMethod
	} else {
		method = sdkCommon.SetCmdMethod
	}
	cmd := vars[sdkCommon.CommandVar]
	helper := NewCommandProcessor(&device, nil, id, cmd, body, dic)
	cmdExists, _ := cache.Profiles().CommandExists(device.Profile.Name, cmd, method)
	if !cmdExists {
		dr, drExists := cache.Profiles().DeviceResource(device.Profile.Name, cmd)
		if !drExists {
			return nil, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "command not found", nil)
		}

		helper = NewCommandProcessor(&device, &dr, id, cmd, body, dic)
		if isRead {
			return helper.ReadDeviceResource()
		} else {
			return nil, helper.WriteDeviceResource()
		}
	} else {
		if isRead {
			return helper.ReadCommand()
		} else {
			return nil, helper.WriteCommand()
		}
	}
}

func (c *CommandProcessor) ReadDeviceResource() (*dsModels.Event, edgexErr.EdgeX) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	lc.Debug(fmt.Sprintf("Application - readDeviceResource: reading deviceResource: %s", c.deviceResource.Name))

	// check provided deviceResource is not write-only
	if c.deviceResource.Properties.Value.ReadWrite == "W" {
		errMsg := fmt.Sprintf("deviceResource %s is marked as write-only", c.deviceResource.Name)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, nil)
	}

	var req dsModels.CommandRequest
	var reqs []dsModels.CommandRequest

	// prepare CommandRequest
	req.DeviceResourceName = c.deviceResource.Name
	req.Attributes = c.deviceResource.Attributes
	if c.params != "" {
		if len(req.Attributes) <= 0 {
			req.Attributes = make(map[string]string)
		}
		req.Attributes[sdkCommon.URLRawQuery] = c.params
	}
	req.Type = dsModels.ParseValueType(c.deviceResource.Properties.Value.Type)
	reqs = append(reqs, req)

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(c.dic.Get)
	results, err := driver.HandleReadCommands(c.device.Name, c.device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading DeviceResourece %s for %s: %v", c.deviceResource.Name, c.device.Name, err)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	event, err := commandValuesToEvent(c.device, results, c.deviceResource.Name, lc, container.MetadataDeviceClientFrom(c.dic.Get), container.ConfigurationFrom(c.dic.Get))
	if err != nil {
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to convert CommandValue to Event", err)
	}

	return event, nil
}

func (c *CommandProcessor) ReadCommand() (*dsModels.Event, edgexErr.EdgeX) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	lc.Debug(fmt.Sprintf("Application - readCmd: reading cmd: %s", c.cmd))

	// check GET ResourceOperation(s) exist for provided command
	ros, err := cache.Profiles().ResourceOperations(c.device.Profile.Name, c.cmd, sdkCommon.GetCmdMethod)
	if err != nil {
		errMsg := fmt.Sprintf("GET ResourceOperation(s) for %s command not found", c.cmd)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, err)
	}

	// check ResourceOperation count does not exceed MaxCmdOps defined in configuration
	configuration := container.ConfigurationFrom(c.dic.Get)
	if len(ros) > configuration.Device.MaxCmdOps {
		errMsg := fmt.Sprintf("GET command %s exceed device %s MaxCmdOps (%d)", c.cmd, c.device.Name, configuration.Device.MaxCmdOps)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
	}

	// prepare CommandRequests
	reqs := make([]dsModels.CommandRequest, len(ros))
	for i, op := range ros {
		drName := op.DeviceResource
		// check the deviceResource in ResourceOperation actually exist
		dr, ok := cache.Profiles().DeviceResource(c.device.Profile.Name, drName)
		if !ok {
			errMsg := fmt.Sprintf("deviceResource %s in GET commnd %s for %s not defined", drName, c.cmd, c.device.Name)
			return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
		}

		// check the deviceResource isn't write-only
		if dr.Properties.Value.ReadWrite == "W" {
			errMsg := fmt.Sprintf("deviceResource %s in GET command %s is marked as write-only", drName, c.cmd)
			return nil, edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, nil)
		}

		reqs[i].DeviceResourceName = dr.Name
		reqs[i].Attributes = dr.Attributes
		if c.params != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]string)
			}
			reqs[i].Attributes[sdkCommon.URLRawQuery] = c.params
		}
		reqs[i].Type = dsModels.ParseValueType(dr.Properties.Value.Type)
	}

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(c.dic.Get)
	results, err := driver.HandleReadCommands(c.device.Name, c.device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading DeviceCommand %s for %s: %v", c.cmd, c.device.Name, err)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	event, err := commandValuesToEvent(c.device, results, c.cmd, lc, container.MetadataDeviceClientFrom(c.dic.Get), configuration)
	if err != nil {
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to transform CommandValue to Event", err)
	}

	return event, nil
}

func (c *CommandProcessor) WriteDeviceResource() edgexErr.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	lc.Debug(fmt.Sprintf("Application - writeDeviceResource: writting deviceResource: %s", c.deviceResource.Name))

	// check provided deviceResource is not read-only
	if c.deviceResource.Properties.Value.ReadWrite == "R" {
		errMsg := fmt.Sprintf("deviceResource %s is marked as read-only", c.deviceResource.Name)
		return edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, nil)
	}

	// parse request body string
	paramMap, err := parseParams(c.params)
	if err != nil {
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to parse PUT parameters", err)
	}

	// check request body contains provided deviceResource
	v, ok := paramMap[c.deviceResource.Name]
	if !ok {
		if c.deviceResource.Properties.Value.DefaultValue != "" {
			v = c.deviceResource.Properties.Value.DefaultValue
		} else {
			errMsg := fmt.Sprintf("deviceResource %s not found in request body and no default value defined", c.deviceResource.Name)
			return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
		}
	}

	// create CommandValue
	cv, err := createCommandValueFromDeviceResource(c.deviceResource, v)
	if err != nil {
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to create CommandValue", err)
	}

	// prepare CommandRequest
	reqs := make([]dsModels.CommandRequest, 1)
	reqs[0].DeviceResourceName = cv.DeviceResourceName
	reqs[0].Attributes = c.deviceResource.Attributes
	reqs[0].Type = cv.Type

	// transform write value
	configuration := container.ConfigurationFrom(c.dic.Get)
	if configuration.Device.DataTransform {
		err = transformer.TransformWriteParameter(cv, c.deviceResource.Properties.Value, lc)
		if err != nil {
			return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to transform write value", nil)
		}
	}

	// execute protocol-specific write operation
	driver := container.ProtocolDriverFrom(c.dic.Get)
	err = driver.HandleWriteCommands(c.device.Name, c.device.Protocols, reqs, []*dsModels.CommandValue{cv})
	if err != nil {
		errMsg := fmt.Sprintf("error writing DeviceResourece %s for %s: %v", c.deviceResource.Name, c.device.Name, err)
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, err)
	}

	return nil
}

func (c *CommandProcessor) WriteCommand() edgexErr.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	lc.Debug(fmt.Sprintf("Application - writeCmd: writting command: %s", c.cmd))

	// check SET ResourceOperation(s) exist for provided command
	ros, err := cache.Profiles().ResourceOperations(c.device.Profile.Name, c.cmd, sdkCommon.SetCmdMethod)
	if err != nil {
		errMsg := fmt.Sprintf("SET ResourceOperation(s) for %s command not found", c.cmd)
		return edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, err)
	}

	// check ResourceOperation count does not exceed MaxCmdOps defined in configuration
	configuration := container.ConfigurationFrom(c.dic.Get)
	if len(ros) > configuration.Device.MaxCmdOps {
		errMsg := fmt.Sprintf("PUT command %s exceed device %s MaxCmdOps (%d)", c.cmd, c.device.Name, configuration.Device.MaxCmdOps)
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
	}

	// parse request body
	paramMap, err := parseParams(c.params)
	if err != nil {
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to parse PUT parameters", err)
	}

	// create CommandValues
	cvs := make([]*dsModels.CommandValue, 0, len(paramMap))
	for _, ro := range ros {
		drName := ro.DeviceResource
		// check the deviceResource in ResourceOperation actually exist
		dr, ok := cache.Profiles().DeviceResource(c.device.Profile.Name, drName)
		if !ok {
			errMsg := fmt.Sprintf("deviceResource %s in PUT commnd %s for %s not defined", drName, c.cmd, c.device.Name)
			return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
		}

		// check the deviceResource isn't read-only
		if dr.Properties.Value.ReadWrite == "R" {
			errMsg := fmt.Sprintf("deviceResource %s in PUT command %s is marked as read-only", drName, c.cmd)
			return edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, nil)
		}

		// check request body contains the deviceResource
		value, ok := paramMap[ro.DeviceResource]
		if !ok {
			if ro.Parameter != "" {
				value = ro.Parameter
			} else if dr.Properties.Value.DefaultValue != "" {
				value = dr.Properties.Value.DefaultValue
			} else {
				errMsg := fmt.Sprintf("deviceResource %s not found in request body and no default value defined", dr.Name)
				return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
			}
		}

		// write value mapping
		if len(ro.Mappings) > 0 {
			newValue, ok := ro.Mappings[value]
			if ok {
				value = newValue
			} else {
				lc.Warn(fmt.Sprintf("ResourceOperation %s mapping value (%s) failed with the mapping table: %v", ro.DeviceResource, value, ro.Mappings))
			}
		}

		// create CommandValue
		cv, err := createCommandValueFromDeviceResource(&dr, value)
		if err == nil {
			cvs = append(cvs, cv)
		} else {
			return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to create CommandValue", err)
		}
	}

	// prepare CommandRequests
	reqs := make([]dsModels.CommandRequest, len(cvs))
	for i, cv := range cvs {
		dr, _ := cache.Profiles().DeviceResource(c.device.Profile.Name, cv.DeviceResourceName)

		reqs[i].DeviceResourceName = cv.DeviceResourceName
		reqs[i].Attributes = dr.Attributes
		reqs[i].Type = cv.Type

		// transform write value
		if configuration.Device.DataTransform {
			err = transformer.TransformWriteParameter(cv, dr.Properties.Value, lc)
			if err != nil {
				return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to transform write values", err)
			}
		}
	}

	// execute protocol-specific write operation
	driver := container.ProtocolDriverFrom(c.dic.Get)
	err = driver.HandleWriteCommands(c.device.Name, c.device.Protocols, reqs, cvs)
	if err != nil {
		errMsg := fmt.Sprintf("error writing DeviceResourece for %s: %v", c.device.Name, err)
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, err)
	}

	return nil
}

func parseParams(params string) (paramMap map[string]string, err error) {
	err = json.Unmarshal([]byte(params), &paramMap)
	if err != nil {
		return
	}

	if len(paramMap) == 0 {
		err = fmt.Errorf("no parameters specified")
		return
	}

	return
}

func commandValuesToEvent(
	device *contract.Device,
	cvs []*dsModels.CommandValue,
	cmd string,
	lc logger.LoggingClient,
	dc metadata.DeviceClient,
	configuration *sdkCommon.ConfigurationStruct) (*dsModels.Event, error) {
	var err error
	var transformsOK = true
	readings := make([]contract.Reading, 0, configuration.Device.MaxCmdOps)

	for _, cv := range cvs {
		// double check the CommandValue return from ProtocolDriver match device command
		dr, ok := cache.Profiles().DeviceResource(device.Profile.Name, cv.DeviceResourceName)
		if !ok {
			return nil, fmt.Errorf("no deviceResource %s for %s in CommandValue (%s)", cv.DeviceResourceName, device.Name, cv.String())
		}

		// perform data transformation
		if configuration.Device.DataTransform {
			err = transformer.TransformReadResult(cv, dr.Properties.Value, lc)
			if err != nil {
				lc.Error(fmt.Sprintf("failed to transform CommandValue (%s): %v", cv.String(), err))
				transformsOK = false
			}
		}

		// assertion
		err = transformer.CheckAssertion(cv, dr.Properties.Value.Assertion, device, lc, dc)
		if err != nil {
			cv = dsModels.NewStringValue(cv.DeviceResourceName, cv.Origin, fmt.Sprintf("Assertion failed for device resource: %s, with value: %s", cv.DeviceResourceName, cv.String()))
		}

		// ResourceOperation mapping
		ro, err := cache.Profiles().ResourceOperation(device.Profile.Name, cv.DeviceResourceName, sdkCommon.GetCmdMethod)
		if err != nil {
			// this allows SDK to directly read deviceResource without deviceCommands defined.
			lc.Debug(fmt.Sprintf("failed to read ResourceOperation: %v", err))
		} else if len(ro.Mappings) > 0 {
			newCV, ok := transformer.MapCommandValue(cv, ro.Mappings)
			if ok {
				cv = newCV
			} else {
				lc.Warn(fmt.Sprintf("ResourceOperation (%s) mapping value (%s) failed with the mapping table: %v", ro.DeviceResource, cv.String(), ro.Mappings))
				//transformsOK = false  // issue #89 will discuss how to handle there is no mapping matched
			}
		}

		reading := sdkCommon.CommandValueToReading(cv, device.Name, dr.Properties.Value.MediaType, dr.Properties.Value.FloatEncoding)
		readings = append(readings, *reading)

		if cv.Type == dsModels.Binary {
			lc.Debug(fmt.Sprintf("device: %s DeviceResource: %v reading: binary value", device.Name, cv.DeviceResourceName))
		} else {
			lc.Debug(fmt.Sprintf("device: %s DeviceResource: %v reading: %v", device.Name, cv.DeviceResourceName, reading))
		}
	}

	if !transformsOK {
		return nil, fmt.Errorf("GET command %s transform failed for %s", cmd, device.Name)
	}

	// TODO: update to v2 NewEventResponse
	cevent := contract.Event{Device: device.Name, Readings: readings}
	event := &dsModels.Event{Event: cevent}
	event.Origin = sdkCommon.GetUniqueOrigin()

	return event, nil
}

func createCommandValueFromDeviceResource(dr *contract.DeviceResource, v string) (*dsModels.CommandValue, error) {
	var err error
	var result *dsModels.CommandValue

	origin := time.Now().UnixNano()
	switch strings.ToLower(dr.Properties.Value.Type) {
	case "string":
		result = dsModels.NewStringValue(dr.Name, origin, v)
	case "bool":
		value, err := strconv.ParseBool(v)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewBoolValue(dr.Name, origin, value)
	case "boolarray":
		var arr []bool
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewBoolArrayValue(dr.Name, origin, arr)
	case "uint8":
		n, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint8Value(dr.Name, origin, uint8(n))
	case "uint8array":
		var arr []uint8
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 8)
			if err != nil {
				return result, err
			}
			arr = append(arr, uint8(n))
		}
		result, err = dsModels.NewUint8ArrayValue(dr.Name, origin, arr)
	case "uint16":
		n, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint16Value(dr.Name, origin, uint16(n))
	case "uint16array":
		var arr []uint16
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 16)
			if err != nil {
				return result, err
			}
			arr = append(arr, uint16(n))
		}
		result, err = dsModels.NewUint16ArrayValue(dr.Name, origin, arr)
	case "uint32":
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint32Value(dr.Name, origin, uint32(n))
	case "uint32array":
		var arr []uint32
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 32)
			if err != nil {
				return result, err
			}
			arr = append(arr, uint32(n))
		}
		result, err = dsModels.NewUint32ArrayValue(dr.Name, origin, arr)
	case "uint64":
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint64Value(dr.Name, origin, n)
	case "uint64array":
		var arr []uint64
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 64)
			if err != nil {
				return result, err
			}
			arr = append(arr, n)
		}
		result, err = dsModels.NewUint64ArrayValue(dr.Name, origin, arr)
	case "int8":
		n, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt8Value(dr.Name, origin, int8(n))
	case "int8array":
		var arr []int8
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt8ArrayValue(dr.Name, origin, arr)
	case "int16":
		n, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt16Value(dr.Name, origin, int16(n))
	case "int16array":
		var arr []int16
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt16ArrayValue(dr.Name, origin, arr)
	case "int32":
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt32Value(dr.Name, origin, int32(n))
	case "int32array":
		var arr []int32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt32ArrayValue(dr.Name, origin, arr)
	case "int64":
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt64Value(dr.Name, origin, n)
	case "int64array":
		var arr []int64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt64ArrayValue(dr.Name, origin, arr)
	case "float32":
		n, e := strconv.ParseFloat(v, 32)
		if e == nil {
			result, err = dsModels.NewFloat32Value(dr.Name, origin, float32(n))
			break
		}
		if numError, ok := e.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				err = e
				break
			}
		}
		var decodedToBytes []byte
		decodedToBytes, err = base64.StdEncoding.DecodeString(v)
		if err == nil {
			var val float32
			val, err = float32FromBytes(decodedToBytes)
			if err != nil {
				break
			} else if math.IsNaN(float64(val)) {
				err = fmt.Errorf("fail to parse %v to float32, unexpected result %v", v, val)
			} else {
				result, err = dsModels.NewFloat32Value(dr.Name, origin, val)
			}
		}
	case "float32array":
		var arr []float32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewFloat32ArrayValue(dr.Name, origin, arr)
	case "float64":
		var val float64
		val, err = strconv.ParseFloat(v, 64)
		if err == nil {
			result, err = dsModels.NewFloat64Value(dr.Name, origin, val)
			break
		}
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				break
			}
		}
		var decodedToBytes []byte
		decodedToBytes, err = base64.StdEncoding.DecodeString(v)
		if err == nil {
			val, err = float64FromBytes(decodedToBytes)
			if err != nil {
				break
			} else if math.IsNaN(val) {
				err = fmt.Errorf("fail to parse %v to float64, unexpected result %v", v, val)
			} else {
				result, err = dsModels.NewFloat64Value(dr.Name, origin, val)
			}
		}
	case "float64array":
		var arr []float64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewFloat64ArrayValue(dr.Name, origin, arr)
	default:
		err = errors.New("unsupported deviceResource value type")
	}

	if err != nil {
		return result, err
	}

	return result, err
}

func float32FromBytes(numericValue []byte) (res float32, err error) {
	reader := bytes.NewReader(numericValue)
	err = binary.Read(reader, binary.BigEndian, &res)
	return
}

func float64FromBytes(numericValue []byte) (res float64, err error) {
	reader := bytes.NewReader(numericValue)
	err = binary.Read(reader, binary.BigEndian, &res)
	return
}
