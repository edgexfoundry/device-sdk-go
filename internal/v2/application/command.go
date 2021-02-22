// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
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

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/transformer"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/cache"
	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

type CommandProcessor struct {
	device         models.Device
	deviceResource models.DeviceResource
	correlationID  string
	cmd            string
	body           string
	attributes     string
	dic            *di.Container
}

func NewCommandProcessor(device models.Device, dr models.DeviceResource, correlationID string, cmd string, body string, attributes string, dic *di.Container) *CommandProcessor {
	return &CommandProcessor{
		device:         device,
		deviceResource: dr,
		correlationID:  correlationID,
		cmd:            cmd,
		body:           body,
		attributes:     attributes,
		dic:            dic,
	}
}

func CommandHandler(isRead bool, sendEvent bool, correlationID string, vars map[string]string, body string, attributes string, dic *di.Container) (res dtos.Event, err edgexErr.EdgeX) {
	var exist bool
	var device models.Device
	var deviceResource models.DeviceResource
	deviceKey := vars[v2.Name]

	// check device service's AdminState
	ds := container.DeviceServiceFrom(dic.Get)
	if ds.AdminState == models.Locked {
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindServiceLocked, "service locked", nil)
	}

	// check provided device exists
	device, exist = cache.Devices().ForName(deviceKey)
	if !exist {
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, fmt.Sprintf("device %s not found", deviceKey), nil)
	}

	// check device's AdminState
	if device.AdminState == models.Locked {
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindServiceLocked, fmt.Sprintf("device %s locked", device.Name), nil)
	}
	// the device service will perform some operations(e.g. update LastConnected timestamp,
	// push returning event to core-data) after a device is successfully interacted with if
	// it has been configured to do so, and those operation apply to every protocol and
	// need to be finished in the end of application layer before returning to protocol layer.
	defer func() {
		if err != nil {
			return
		}
		config := container.ConfigurationFrom(dic.Get)
		if config.Device.UpdateLastConnected {
			go common.UpdateLastConnected(device.Name, bootstrapContainer.LoggingClientFrom(dic.Get), container.MetadataDeviceClientFrom(dic.Get))
		}

		if sendEvent {
			ec := container.CoredataEventClientFrom(dic.Get)
			lc := bootstrapContainer.LoggingClientFrom(dic.Get)
			go common.SendEvent(res, correlationID, lc, ec)
		}
	}()

	var method string
	if isRead {
		method = common.GetCmdMethod
	} else {
		method = common.SetCmdMethod
	}
	cmd := vars[v2.Command]
	cmdExists, e := cache.Profiles().CommandExists(device.ProfileName, cmd, method)
	if e != nil {
		errMsg := fmt.Sprintf("failed to identify command %s in cache", cmd)
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, e)
	}

	helper := NewCommandProcessor(device, deviceResource, correlationID, cmd, body, attributes, dic)
	if cmdExists {
		if isRead {
			return helper.ReadCommand()
		} else {
			return res, helper.WriteCommand()
		}
	} else {
		deviceResource, exist = cache.Profiles().DeviceResource(device.ProfileName, cmd)
		if !exist {
			return res, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "command not found", nil)
		}

		helper = NewCommandProcessor(device, deviceResource, correlationID, cmd, body, attributes, dic)
		if isRead {
			return helper.ReadDeviceResource()
		} else {
			return res, helper.WriteDeviceResource()
		}
	}
}

func (c *CommandProcessor) ReadDeviceResource() (res dtos.Event, e edgexErr.EdgeX) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	lc.Debugf("Application - readDeviceResource: reading deviceResource: %s; %s: %s", c.deviceResource.Name, common.CorrelationHeader, c.correlationID)

	// check provided deviceResource is not write-only
	if c.deviceResource.Properties.ReadWrite == common.DeviceResourceWriteOnly {
		errMsg := fmt.Sprintf("deviceResource %s is marked as write-only", c.deviceResource.Name)
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, nil)
	}

	var req dsModels.CommandRequest
	var reqs []dsModels.CommandRequest

	// prepare CommandRequest
	req.DeviceResourceName = c.deviceResource.Name
	req.Attributes = c.deviceResource.Attributes
	if c.attributes != "" {
		if len(req.Attributes) <= 0 {
			req.Attributes = make(map[string]string)
		}
		req.Attributes[common.URLRawQuery] = c.attributes
	}
	req.Type = c.deviceResource.Properties.ValueType
	reqs = append(reqs, req)

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(c.dic.Get)
	results, err := driver.HandleReadCommands(c.device.Name, c.device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading DeviceResourece %s for %s: %v", c.deviceResource.Name, c.device.Name, err)
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	res, e = transformer.CommandValuesToEventDTO(results, c.device.Name, c.dic)
	if e != nil {
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to convert CommandValue to Event", e)
	}

	return
}

func (c *CommandProcessor) ReadCommand() (res dtos.Event, e edgexErr.EdgeX) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	lc.Debugf("Application - readCmd: reading cmd: %s; %s: %s", c.cmd, common.CorrelationHeader, c.correlationID)

	// check GET ResourceOperation(s) exist for provided command
	ros, e := cache.Profiles().ResourceOperations(c.device.ProfileName, c.cmd, common.GetCmdMethod)
	if e != nil {
		errMsg := fmt.Sprintf("GET ResourceOperation(s) for %s command not found", c.cmd)
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, e)
	}

	// check ResourceOperation count does not exceed MaxCmdOps defined in configuration
	configuration := container.ConfigurationFrom(c.dic.Get)
	if len(ros) > configuration.Device.MaxCmdOps {
		errMsg := fmt.Sprintf("GET command %s exceed device %s MaxCmdOps (%d)", c.cmd, c.device.Name, configuration.Device.MaxCmdOps)
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
	}

	// prepare CommandRequests
	reqs := make([]dsModels.CommandRequest, len(ros))
	for i, op := range ros {
		drName := op.DeviceResource
		// check the deviceResource in ResourceOperation actually exist
		dr, ok := cache.Profiles().DeviceResource(c.device.ProfileName, drName)
		if !ok {
			errMsg := fmt.Sprintf("deviceResource %s in GET commnd %s for %s not defined", drName, c.cmd, c.device.Name)
			return res, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
		}

		// check the deviceResource isn't write-only
		if dr.Properties.ReadWrite == common.DeviceResourceWriteOnly {
			errMsg := fmt.Sprintf("deviceResource %s in GET command %s is marked as write-only", drName, c.cmd)
			return res, edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, nil)
		}

		reqs[i].DeviceResourceName = dr.Name
		reqs[i].Attributes = dr.Attributes
		if c.attributes != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]string)
			}
			reqs[i].Attributes[common.URLRawQuery] = c.attributes
		}
		reqs[i].Type = dr.Properties.ValueType
	}

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(c.dic.Get)
	results, err := driver.HandleReadCommands(c.device.Name, c.device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading DeviceCommand %s for %s: %v", c.cmd, c.device.Name, err)
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	res, e = transformer.CommandValuesToEventDTO(results, c.device.Name, c.dic)
	if e != nil {
		return res, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to transform CommandValue to Event", e)
	}

	return
}

func (c *CommandProcessor) WriteDeviceResource() edgexErr.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	lc.Debugf("Application - writeDeviceResource: writing deviceResource: %s; %s: %s", c.deviceResource.Name, common.CorrelationHeader, c.correlationID)

	// check provided deviceResource is not read-only
	if c.deviceResource.Properties.ReadWrite == common.DeviceResourceReadOnly {
		errMsg := fmt.Sprintf("deviceResource %s is marked as read-only", c.deviceResource.Name)
		return edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, nil)
	}

	// parse request body string
	paramMap, err := parseParams(c.body)
	if err != nil {
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to parse SET command parameters", err)
	}

	// check request body contains provided deviceResource
	v, ok := paramMap[c.deviceResource.Name]
	if !ok {
		if c.deviceResource.Properties.DefaultValue != "" {
			v = c.deviceResource.Properties.DefaultValue
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
	if c.attributes != "" {
		if len(reqs[0].Attributes) <= 0 {
			reqs[0].Attributes = make(map[string]string)
		}
		reqs[0].Attributes[common.URLRawQuery] = c.attributes
	}
	reqs[0].Type = cv.Type

	// transform write value
	configuration := container.ConfigurationFrom(c.dic.Get)
	if configuration.Device.DataTransform {
		err = transformer.TransformWriteParameter(cv, c.deviceResource.Properties, lc)
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
	lc.Debugf("Application - writeCmd: writing command: %s; %s: %s", c.cmd, common.CorrelationHeader, c.correlationID)

	// check SET ResourceOperation(s) exist for provided command
	ros, e := cache.Profiles().ResourceOperations(c.device.ProfileName, c.cmd, common.SetCmdMethod)
	if e != nil {
		errMsg := fmt.Sprintf("SET ResourceOperation(s) for %s command not found", c.cmd)
		return edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, e)
	}

	// check ResourceOperation count does not exceed MaxCmdOps defined in configuration
	configuration := container.ConfigurationFrom(c.dic.Get)
	if len(ros) > configuration.Device.MaxCmdOps {
		errMsg := fmt.Sprintf("SET command %s exceed device %s MaxCmdOps (%d)", c.cmd, c.device.Name, configuration.Device.MaxCmdOps)
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
	}

	// parse request body
	paramMap, err := parseParams(c.body)
	if err != nil {
		return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to parse SET command parameters", err)
	}

	// create CommandValues
	cvs := make([]*dsModels.CommandValue, 0, len(paramMap))
	for _, ro := range ros {
		drName := ro.DeviceResource
		// check the deviceResource in ResourceOperation actually exist
		dr, ok := cache.Profiles().DeviceResource(c.device.ProfileName, drName)
		if !ok {
			errMsg := fmt.Sprintf("deviceResource %s in SET commnd %s for %s not defined", drName, c.cmd, c.device.Name)
			return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, errMsg, nil)
		}

		// check the deviceResource isn't read-only
		if dr.Properties.ReadWrite == common.DeviceResourceReadOnly {
			errMsg := fmt.Sprintf("deviceResource %s in SET command %s is marked as read-only", drName, c.cmd)
			return edgexErr.NewCommonEdgeX(edgexErr.KindNotAllowed, errMsg, nil)
		}

		// check request body contains the deviceResource
		value, ok := paramMap[ro.DeviceResource]
		if !ok {
			if ro.Parameter != "" {
				value = ro.Parameter
			} else if dr.Properties.DefaultValue != "" {
				value = dr.Properties.DefaultValue
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
		cv, err := createCommandValueFromDeviceResource(dr, value)
		if err == nil {
			cvs = append(cvs, cv)
		} else {
			return edgexErr.NewCommonEdgeX(edgexErr.KindServerError, "failed to create CommandValue", err)
		}
	}

	// prepare CommandRequests
	reqs := make([]dsModels.CommandRequest, len(cvs))
	for i, cv := range cvs {
		dr, _ := cache.Profiles().DeviceResource(c.device.ProfileName, cv.DeviceResourceName)

		reqs[i].DeviceResourceName = cv.DeviceResourceName
		reqs[i].Attributes = dr.Attributes
		if c.attributes != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]string)
			}
			reqs[i].Attributes[common.URLRawQuery] = c.attributes
		}
		reqs[i].Type = cv.Type

		// transform write value
		if configuration.Device.DataTransform {
			err = transformer.TransformWriteParameter(cv, dr.Properties, lc)
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

func createCommandValueFromDeviceResource(dr models.DeviceResource, v string) (*dsModels.CommandValue, error) {
	var err error
	var result *dsModels.CommandValue

	origin := time.Now().UnixNano()
	switch strings.ToLower(dr.Properties.ValueType) {
	case strings.ToLower(v2.ValueTypeString):
		result = dsModels.NewStringValue(dr.Name, origin, v)
	case strings.ToLower(v2.ValueTypeBool):
		value, err := strconv.ParseBool(v)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewBoolValue(dr.Name, origin, value)
	case strings.ToLower(v2.ValueTypeBoolArray):
		var arr []bool
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewBoolArrayValue(dr.Name, origin, arr)
	case strings.ToLower(v2.ValueTypeUint8):
		n, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint8Value(dr.Name, origin, uint8(n))
	case strings.ToLower(v2.ValueTypeUint8Array):
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
	case strings.ToLower(v2.ValueTypeUint16):
		n, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint16Value(dr.Name, origin, uint16(n))
	case strings.ToLower(v2.ValueTypeUint16Array):
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
	case strings.ToLower(v2.ValueTypeUint32):
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint32Value(dr.Name, origin, uint32(n))
	case strings.ToLower(v2.ValueTypeUint32Array):
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
	case strings.ToLower(v2.ValueTypeUint64):
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewUint64Value(dr.Name, origin, n)
	case strings.ToLower(v2.ValueTypeUint64Array):
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
	case strings.ToLower(v2.ValueTypeInt8):
		n, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt8Value(dr.Name, origin, int8(n))
	case strings.ToLower(v2.ValueTypeInt8Array):
		var arr []int8
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt8ArrayValue(dr.Name, origin, arr)
	case strings.ToLower(v2.ValueTypeInt16):
		n, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt16Value(dr.Name, origin, int16(n))
	case strings.ToLower(v2.ValueTypeInt16Array):
		var arr []int16
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt16ArrayValue(dr.Name, origin, arr)
	case strings.ToLower(v2.ValueTypeInt32):
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt32Value(dr.Name, origin, int32(n))
	case strings.ToLower(v2.ValueTypeInt32Array):
		var arr []int32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt32ArrayValue(dr.Name, origin, arr)
	case strings.ToLower(v2.ValueTypeInt64):
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt64Value(dr.Name, origin, n)
	case strings.ToLower(v2.ValueTypeInt64Array):
		var arr []int64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewInt64ArrayValue(dr.Name, origin, arr)
	case strings.ToLower(v2.ValueTypeFloat32):
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
	case strings.ToLower(v2.ValueTypeFloat32Array):
		var arr []float32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			return result, err
		}
		result, err = dsModels.NewFloat32ArrayValue(dr.Name, origin, arr)
	case strings.ToLower(v2.ValueTypeFloat64):
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
	case strings.ToLower(v2.ValueTypeFloat64Array):
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
