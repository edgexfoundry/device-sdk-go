// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http/utils"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/transformer"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

func GetCommand(ctx context.Context, deviceName string, commandName string, queryParams string, dic *di.Container) (*dtos.Event, errors.EdgeX) {
	if deviceName == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if commandName == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "command is empty", nil)
	}

	device, err := validateServiceAndDeviceState(deviceName, dic)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	var res *dtos.Event
	_, cmdExist := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if cmdExist {
		res, err = readDeviceCommand(device, commandName, queryParams, dic)
	} else {
		res, err = readDeviceResource(device, commandName, queryParams, dic)
	}

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	configuration := container.ConfigurationFrom(dic.Get)
	if configuration.Device.UpdateLastConnected {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		dc := bootstrapContainer.DeviceClientFrom(dic.Get)
		go sdkCommon.UpdateLastConnected(device.Name, lc, dc)
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	lc.Debugf("GET Device Command successfully. Device: %s, Source: %s, %s: %s", deviceName, commandName, common.CorrelationHeader, utils.FromContext(ctx, common.CorrelationHeader))
	return res, nil
}

func SetCommand(ctx context.Context, deviceName string, commandName string, queryParams string, requests map[string]any, dic *di.Container) errors.EdgeX {
	if deviceName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if commandName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "command is empty", nil)
	}

	device, err := validateServiceAndDeviceState(deviceName, dic)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	_, cmdExist := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if cmdExist {
		err = writeDeviceCommand(device, commandName, queryParams, requests, dic)
	} else {
		err = writeDeviceResource(device, commandName, queryParams, requests, dic)
	}

	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	lc.Debugf("SET Device Command successfully. Device: %s, Source: %s, %s: %s", deviceName, commandName, common.CorrelationHeader, utils.FromContext(ctx, common.CorrelationHeader))
	return nil
}

func readDeviceResource(device models.Device, resourceName string, attributes string, dic *di.Container) (res *dtos.Event, edgexErr errors.EdgeX) {
	dr, ok := cache.Profiles().DeviceResource(device.ProfileName, resourceName)
	if !ok {
		errMsg := fmt.Sprintf("deviceResource %s not found", resourceName)
		return res, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// check deviceResource is not write-only
	if dr.Properties.ReadWrite == common.ReadWrite_W {
		errMsg := fmt.Sprintf("deviceResource %s is marked as write-only", dr.Name)
		return res, errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}

	var req sdkModels.CommandRequest
	var reqs []sdkModels.CommandRequest

	// prepare CommandRequest
	req.DeviceResourceName = dr.Name
	req.Attributes = dr.Attributes
	if attributes != "" {
		if len(req.Attributes) <= 0 {
			req.Attributes = make(map[string]interface{})
		}
		req.Attributes[sdkCommon.URLRawQuery] = attributes
	}
	req.Type = dr.Properties.ValueType
	reqs = append(reqs, req)

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(dic.Get)
	results, err := driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading DeviceResourece %s for %s", dr.Name, device.Name)
		return res, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	res, edgexErr = transformer.CommandValuesToEventDTO(results, device.Name, dr.Name, dic)
	if edgexErr != nil {
		return res, errors.NewCommonEdgeX(errors.KindServerError, "failed to convert CommandValue to Event", err)
	}

	return res, nil
}

func readDeviceCommand(device models.Device, commandName string, attributes string, dic *di.Container) (res *dtos.Event, edgexErr errors.EdgeX) {
	dc, ok := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if !ok {
		errMsg := fmt.Sprintf("deviceCommand %s not found", commandName)
		return res, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// check deviceCommand is not write-only
	if dc.ReadWrite == common.ReadWrite_W {
		errMsg := fmt.Sprintf("deviceCommand %s is marked as write-only", dc.Name)
		return res, errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}
	// check ResourceOperation count does not exceed MaxCmdOps defined in configuration
	configuration := container.ConfigurationFrom(dic.Get)
	if len(dc.ResourceOperations) > configuration.Device.MaxCmdOps {
		errMsg := fmt.Sprintf("GET command %s exceed device %s MaxCmdOps (%d)", dc.Name, device.Name, configuration.Device.MaxCmdOps)
		return res, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}

	// prepare CommandRequests
	reqs := make([]sdkModels.CommandRequest, len(dc.ResourceOperations))
	for i, op := range dc.ResourceOperations {
		drName := op.DeviceResource
		// check the deviceResource in ResourceOperation actually exist
		dr, ok := cache.Profiles().DeviceResource(device.ProfileName, drName)
		if !ok {
			errMsg := fmt.Sprintf("deviceResource %s in GET commnd %s for %s not defined", drName, dc.Name, device.Name)
			return res, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
		}

		reqs[i].DeviceResourceName = dr.Name
		reqs[i].Attributes = dr.Attributes
		if attributes != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]interface{})
			}
			reqs[i].Attributes[sdkCommon.URLRawQuery] = attributes
		}
		reqs[i].Type = dr.Properties.ValueType
	}

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(dic.Get)
	results, err := driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading DeviceCommand %s for %s", dc.Name, device.Name)
		return res, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	res, edgexErr = transformer.CommandValuesToEventDTO(results, device.Name, dc.Name, dic)
	if edgexErr != nil {
		return res, errors.NewCommonEdgeX(errors.KindServerError, "failed to transform CommandValue to Event", edgexErr)
	}

	return res, nil
}

func writeDeviceResource(device models.Device, resourceName string, attributes string, requests map[string]any, dic *di.Container) errors.EdgeX {
	dr, ok := cache.Profiles().DeviceResource(device.ProfileName, resourceName)
	if !ok {
		errMsg := fmt.Sprintf("deviceResource %s not found", resourceName)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// check deviceResource is not read-only
	if dr.Properties.ReadWrite == common.ReadWrite_R {
		errMsg := fmt.Sprintf("deviceResource %s is marked as read-only", dr.Name)
		return errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}

	// check set parameters contains provided deviceResource
	v, ok := requests[dr.Name]
	if !ok {
		if dr.Properties.DefaultValue != "" {
			v = dr.Properties.DefaultValue
		} else {
			errMsg := fmt.Sprintf("deviceResource %s not found in request body and no default value defined", dr.Name)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
		}
	}

	// create CommandValue
	cv, edgexErr := createCommandValueFromDeviceResource(dr, v)
	if edgexErr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgexErr), "failed to create CommandValue", edgexErr)
	}

	// prepare CommandRequest
	reqs := make([]sdkModels.CommandRequest, 1)
	reqs[0].DeviceResourceName = cv.DeviceResourceName
	reqs[0].Attributes = dr.Attributes
	if attributes != "" {
		if len(reqs[0].Attributes) <= 0 {
			reqs[0].Attributes = make(map[string]any)
		}
		reqs[0].Attributes[sdkCommon.URLRawQuery] = attributes
	}
	reqs[0].Type = cv.Type

	// transform write value
	configuration := container.ConfigurationFrom(dic.Get)
	if configuration.Device.DataTransform {
		edgexErr = transformer.TransformWriteParameter(cv, dr.Properties)
		if edgexErr != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to transform set parameter", edgexErr)
		}
	}

	// execute protocol-specific write operation
	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.HandleWriteCommands(device.Name, device.Protocols, reqs, []*sdkModels.CommandValue{cv})
	if err != nil {
		errMsg := fmt.Sprintf("error writing DeviceResourece %s for %s", dr.Name, device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	return nil
}

func writeDeviceCommand(device models.Device, commandName string, attributes string, requests map[string]any, dic *di.Container) errors.EdgeX {
	dc, ok := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if !ok {
		errMsg := fmt.Sprintf("deviceCommand %s not found", commandName)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// check deviceCommand is not read-only
	if dc.ReadWrite == common.ReadWrite_R {
		errMsg := fmt.Sprintf("deviceCommand %s is marked as read-only", dc.Name)
		return errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}
	// check ResourceOperation count does not exceed MaxCmdOps defined in configuration
	configuration := container.ConfigurationFrom(dic.Get)
	if len(dc.ResourceOperations) > configuration.Device.MaxCmdOps {
		errMsg := fmt.Sprintf("SET command %s exceed device %s MaxCmdOps (%d)", dc.Name, device.Name, configuration.Device.MaxCmdOps)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}

	// create CommandValues
	cvs := make([]*sdkModels.CommandValue, 0, len(requests))
	for _, ro := range dc.ResourceOperations {
		drName := ro.DeviceResource
		// check the deviceResource in ResourceOperation actually exist
		dr, ok := cache.Profiles().DeviceResource(device.ProfileName, drName)
		if !ok {
			errMsg := fmt.Sprintf("deviceResource %s in SET commnd %s for %s not defined", drName, dc.Name, device.Name)
			return errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
		}

		// check request body contains the deviceResource
		value, ok := requests[ro.DeviceResource]
		if !ok {
			if ro.DefaultValue != "" {
				value = ro.DefaultValue
			} else if dr.Properties.DefaultValue != "" {
				value = dr.Properties.DefaultValue
			} else {
				errMsg := fmt.Sprintf("deviceResource %s not found in request body and no default value defined", dr.Name)
				return errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
			}
		}

		// ResourceOperation mapping, notice that the order is opposite to get command mapping
		// i.e. the mapping value is actually the key for set command.
		if len(ro.Mappings) > 0 {
			for k, v := range ro.Mappings {
				if v == value {
					value = k
					break
				}
			}
		}

		// create CommandValue
		cv, err := createCommandValueFromDeviceResource(dr, value)
		if err == nil {
			cvs = append(cvs, cv)
		} else {
			return errors.NewCommonEdgeX(errors.Kind(err), "failed to create CommandValue", err)
		}
	}

	// prepare CommandRequests
	reqs := make([]sdkModels.CommandRequest, len(cvs))
	for i, cv := range cvs {
		dr, _ := cache.Profiles().DeviceResource(device.ProfileName, cv.DeviceResourceName)

		reqs[i].DeviceResourceName = cv.DeviceResourceName
		reqs[i].Attributes = dr.Attributes
		if attributes != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]interface{})
			}
			reqs[i].Attributes[sdkCommon.URLRawQuery] = attributes
		}
		reqs[i].Type = cv.Type

		// transform write value
		if configuration.Device.DataTransform {
			err := transformer.TransformWriteParameter(cv, dr.Properties)
			if err != nil {
				return errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to transform set parameter", err)
			}
		}
	}

	// execute protocol-specific write operation
	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.HandleWriteCommands(device.Name, device.Protocols, reqs, cvs)
	if err != nil {
		errMsg := fmt.Sprintf("error writing DeviceCommand %s for %s", dc.Name, device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	return nil
}

func validateServiceAndDeviceState(deviceName string, dic *di.Container) (models.Device, errors.EdgeX) {
	// check device service AdminState
	ds := container.DeviceServiceFrom(dic.Get)
	if ds.AdminState == models.Locked {
		return models.Device{}, errors.NewCommonEdgeX(errors.KindServiceLocked, "service locked", nil)
	}

	// check requested device exists
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		return models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device %s not found", deviceName), nil)
	}

	// check device's AdminState
	if device.AdminState == models.Locked {
		return models.Device{}, errors.NewCommonEdgeX(errors.KindServiceLocked, fmt.Sprintf("device %s locked", device.Name), nil)
	}
	// check device's OperatingState
	if device.OperatingState == models.Down {
		return models.Device{}, errors.NewCommonEdgeX(errors.KindServiceLocked, fmt.Sprintf("device %s OperatingState is DOWN", device.Name), nil)
	}

	return device, nil
}

func createCommandValueFromDeviceResource(dr models.DeviceResource, value interface{}) (*sdkModels.CommandValue, errors.EdgeX) {
	var err error
	var result *sdkModels.CommandValue

	v := fmt.Sprint(value)

	if dr.Properties.ValueType != common.ValueTypeString && strings.TrimSpace(v) == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("empty string is invalid for %v value type", dr.Properties.ValueType), nil)
	}

	switch dr.Properties.ValueType {
	case common.ValueTypeString:
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeString, v)
	case common.ValueTypeStringArray:
		var arr []string
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeStringArray, arr)
	case common.ValueTypeBool:
		var value bool
		value, err = strconv.ParseBool(v)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeBool, value)
	case common.ValueTypeBoolArray:
		var arr []bool
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeBoolArray, arr)
	case common.ValueTypeUint8:
		var n uint64
		n, err = strconv.ParseUint(v, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint8, uint8(n))
	case common.ValueTypeUint8Array:
		var arr []uint8
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 8)
			if err != nil {
				errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
				return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
			}
			arr = append(arr, uint8(n))
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint8Array, arr)
	case common.ValueTypeUint16:
		var n uint64
		n, err = strconv.ParseUint(v, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint16, uint16(n))
	case common.ValueTypeUint16Array:
		var arr []uint16
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 16)
			if err != nil {
				errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
				return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
			}
			arr = append(arr, uint16(n))
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint16Array, arr)
	case common.ValueTypeUint32:
		var n uint64
		n, err = strconv.ParseUint(v, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint32, uint32(n))
	case common.ValueTypeUint32Array:
		var arr []uint32
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 32)
			if err != nil {
				errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
				return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
			}
			arr = append(arr, uint32(n))
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint32Array, arr)
	case common.ValueTypeUint64:
		var n uint64
		n, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint64, n)
	case common.ValueTypeUint64Array:
		var arr []uint64
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 64)
			if err != nil {
				errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
				return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
			}
			arr = append(arr, n)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint64Array, arr)
	case common.ValueTypeInt8:
		var n int64
		n, err = strconv.ParseInt(v, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt8, int8(n))
	case common.ValueTypeInt8Array:
		var arr []int8
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt8Array, arr)
	case common.ValueTypeInt16:
		var n int64
		n, err = strconv.ParseInt(v, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt16, int16(n))
	case common.ValueTypeInt16Array:
		var arr []int16
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt16Array, arr)
	case common.ValueTypeInt32:
		var n int64
		n, err = strconv.ParseInt(v, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt32, int32(n))
	case common.ValueTypeInt32Array:
		var arr []int32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt32Array, arr)
	case common.ValueTypeInt64:
		var n int64
		n, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt64, n)
	case common.ValueTypeInt64Array:
		var arr []int64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt64Array, arr)
	case common.ValueTypeFloat32:
		var val float64
		val, err = strconv.ParseFloat(v, 32)
		if err == nil {
			result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat32, float32(val))
			break
		}
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				err = errors.NewCommonEdgeX(errors.KindServerError, "NumError", err)
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
				result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat32, val)
			}
		}
	case common.ValueTypeFloat32Array:
		var arr []float32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat32Array, arr)
	case common.ValueTypeFloat64:
		var val float64
		val, err = strconv.ParseFloat(v, 64)
		if err == nil {
			result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat64, val)
			break
		}
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				err = errors.NewCommonEdgeX(errors.KindServerError, "NumError", err)
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
				result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat64, val)
			}
		}
	case common.ValueTypeFloat64Array:
		var arr []float64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat64Array, arr)
	case common.ValueTypeObject:
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeObject, value)
	default:
		err = errors.NewCommonEdgeX(errors.KindServerError, "unrecognized value type", nil)
	}

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return result, nil
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
