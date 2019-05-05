// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/mock"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	profileNameRandomBooleanGenerator         = "Random-Boolean-Generator"
	profileNameRandomIntegerGenerator         = "Random-Integer-Generator"
	profileNameRandomUnsignedIntegerGenerator = "Random-UnsignedInteger-Generator"
	profileNameRandomFloatGenerator           = "Random-Float-Generator"
	methodGet                                 = "get"
	methodSet                                 = "set"
	typeBool                                  = "Bool"
	typeInt8                                  = "Int8"
	typeInt16                                 = "Int16"
	typeInt32                                 = "Int32"
	typeInt64                                 = "Int64"
	typeUint8                                 = "Uint8"
	typeUint16                                = "Uint16"
	typeUint32                                = "Uint32"
	typeUint64                                = "Uint64"
	typeFloat32                               = "Float32"
	typeFloat64                               = "Float64"
	resourceObjectBool                        = "RandomValue_Bool"
	resourceObjectInt8                        = "RandomValue_Int8"
	resourceObjectInt16                       = "RandomValue_Int16"
	resourceObjectInt32                       = "RandomValue_Int32"
	resourceObjectInt64                       = "RandomValue_Int64"
	resourceObjectUint8                       = "RandomValue_Uint8"
	resourceObjectUint16                      = "RandomValue_Uint16"
	resourceObjectUint32                      = "RandomValue_Uint32"
	resourceObjectUint64                      = "RandomValue_Uint64"
	resourceObjectFloat32                     = "RandomValue_Float32"
	resourceObjectFloat64                     = "RandomValue_Float64"
)

var (
	ds                                                                            []contract.Device
	pc                                                                            cache.ProfileCache
	deviceIntegerGenerator                                                        contract.Device
	operationSetBool                                                              contract.ResourceOperation
	operationSetInt8, operationSetInt16, operationSetInt32, operationSetInt64     contract.ResourceOperation
	operationSetUint8, operationSetUint16, operationSetUint32, operationSetUint64 contract.ResourceOperation
	operationSetFloat32, operationSetFloat64                                      contract.ResourceOperation
)

func init() {
	common.ValueDescriptorClient = &mock.ValueDescriptorMock{}
	common.DeviceClient = &mock.DeviceClientMock{}
	common.EventClient = &mock.EventClientMock{}
	common.LoggingClient = logger.MockLogger{}
	common.Driver = &mock.DriverMock{}
	cache.InitCache()
	pc = cache.Profiles()
	operationSetBool, _ = pc.ResourceOperation(profileNameRandomBooleanGenerator, resourceObjectBool, methodSet)
	operationSetInt8, _ = pc.ResourceOperation(profileNameRandomIntegerGenerator, resourceObjectInt8, methodSet)
	operationSetInt16, _ = pc.ResourceOperation(profileNameRandomIntegerGenerator, resourceObjectInt16, methodSet)
	operationSetInt32, _ = pc.ResourceOperation(profileNameRandomIntegerGenerator, resourceObjectInt32, methodSet)
	operationSetInt64, _ = pc.ResourceOperation(profileNameRandomIntegerGenerator, resourceObjectInt64, methodSet)
	operationSetUint8, _ = pc.ResourceOperation(profileNameRandomUnsignedIntegerGenerator, resourceObjectUint8, methodSet)
	operationSetUint16, _ = pc.ResourceOperation(profileNameRandomUnsignedIntegerGenerator, resourceObjectUint16, methodSet)
	operationSetUint32, _ = pc.ResourceOperation(profileNameRandomUnsignedIntegerGenerator, resourceObjectUint32, methodSet)
	operationSetUint64, _ = pc.ResourceOperation(profileNameRandomUnsignedIntegerGenerator, resourceObjectUint64, methodSet)
	operationSetFloat32, _ = pc.ResourceOperation(profileNameRandomFloatGenerator, resourceObjectFloat32, methodSet)
	operationSetFloat64, _ = pc.ResourceOperation(profileNameRandomFloatGenerator, resourceObjectFloat64, methodSet)

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	ds, _ = common.DeviceClient.DevicesForServiceByName(common.ServiceName, ctx)
	deviceIntegerGenerator = ds[1]

	deviceInfo := common.DeviceInfo{DataTransform: true, MaxCmdOps: 128}
	common.CurrentConfig = &common.Config{Device: deviceInfo}
}

func TestParseWriteParamsWrongParamName(t *testing.T) {
	profileName := "notFound"
	roMap := roSliceToMap([]contract.ResourceOperation{{Index: ""}})
	params := "{ \"key\": \"value\" }"

	_, err := parseWriteParams(profileName, roMap, params)

	if err == nil {
		t.Error("expected error")
	}
}

func TestParseWriteParamsNoParams(t *testing.T) {
	profileName := "notFound"
	roMap := roSliceToMap([]contract.ResourceOperation{{Index: ""}})
	params := "{ }"

	_, err := parseWriteParams(profileName, roMap, params)

	if err == nil {
		t.Error("expected error")
	}
}

func TestFilterOperationalDevices(t *testing.T) {
	var (
		devicesTotal2Unlocked2 = []contract.Device{{AdminState: contract.Unlocked}, {AdminState: contract.Unlocked}}
		devicesTotal2Unlocked1 = []contract.Device{{AdminState: contract.Unlocked}, {AdminState: contract.Locked}}
		devicesTotal2Enabled2  = []contract.Device{{OperatingState: contract.Enabled}, {OperatingState: contract.Enabled}}
		devicesTotal2Enabled1  = []contract.Device{{OperatingState: contract.Enabled}, {OperatingState: contract.Disabled}}
	)
	tests := []struct {
		testName                   string
		devices                    []contract.Device
		expectedOperationalDevices int
	}{
		{"Total2Unlocked2", devicesTotal2Unlocked2, 2},
		{"Total2Unlocked1", devicesTotal2Unlocked1, 1},
		{"Total2Enabled2", devicesTotal2Enabled2, 2},
		{"Total2Enabled1", devicesTotal2Enabled1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			operationalDevices := filterOperationalDevices(tt.devices)
			assert.Equal(t, len(operationalDevices), tt.expectedOperationalDevices)
		})
	}
}

func TestCreateCommandValueForParam(t *testing.T) {
	tests := []struct {
		testName    string
		profileName string
		valueType   string
		op          *contract.ResourceOperation
		v           string
		parseCheck  dsModels.ValueType
		expectErr   bool
	}{
		{"DeviceResourceNotFound", profileNameRandomBooleanGenerator, typeBool, &contract.ResourceOperation{}, "", dsModels.Bool, true},
		{"BoolTruePass", profileNameRandomBooleanGenerator, typeBool, &operationSetBool, "true", dsModels.Bool, false},
		{"BoolFalsePass", profileNameRandomBooleanGenerator, typeBool, &operationSetBool, "false", dsModels.Bool, false},
		{"BoolTrueFail", profileNameRandomBooleanGenerator, typeBool, &operationSetBool, "error", dsModels.Bool, true},
		{"Int8Pass", profileNameRandomIntegerGenerator, typeInt8, &operationSetInt8, "12", dsModels.Int8, false},
		{"Int8NegativePass", profileNameRandomIntegerGenerator, typeInt8, &operationSetInt8, "-12", dsModels.Int8, false},
		{"Int8WordFail", profileNameRandomIntegerGenerator, typeInt8, &operationSetInt8, "hello", dsModels.Int8, true},
		{"Int8OverflowFail", profileNameRandomIntegerGenerator, typeInt8, &operationSetInt8, "9999999999", dsModels.Int8, true},
		{"Int16Pass", profileNameRandomIntegerGenerator, typeInt16, &operationSetInt16, "12", dsModels.Int16, false},
		{"Int16NegativePass", profileNameRandomIntegerGenerator, typeInt16, &operationSetInt16, "-12", dsModels.Int16, false},
		{"Int16WordFail", profileNameRandomIntegerGenerator, typeInt16, &operationSetInt16, "hello", dsModels.Int16, true},
		{"Int16OverflowFail", profileNameRandomIntegerGenerator, typeInt16, &operationSetInt16, "9999999999", dsModels.Int16, true},
		{"Int32Pass", profileNameRandomIntegerGenerator, typeInt32, &operationSetInt32, "12", dsModels.Int32, false},
		{"Int32NegativePass", profileNameRandomIntegerGenerator, typeInt32, &operationSetInt32, "-12", dsModels.Int32, false},
		{"Int32WordFail", profileNameRandomIntegerGenerator, typeInt32, &operationSetInt32, "hello", dsModels.Int32, true},
		{"Int32OverflowFail", profileNameRandomIntegerGenerator, typeInt32, &operationSetInt32, "9999999999", dsModels.Int32, true},
		{"Int64Pass", profileNameRandomIntegerGenerator, typeInt64, &operationSetInt64, "12", dsModels.Int64, false},
		{"Int64NegativePass", profileNameRandomIntegerGenerator, typeInt64, &operationSetInt64, "-12", dsModels.Int64, false},
		{"Int64WordFail", profileNameRandomIntegerGenerator, typeInt64, &operationSetInt64, "hello", dsModels.Int64, true},
		{"Int64OverflowFail", profileNameRandomIntegerGenerator, typeInt64, &operationSetInt64, "99999999999999999999", dsModels.Int64, true},
		{"Uint8Pass", profileNameRandomUnsignedIntegerGenerator, typeUint8, &operationSetUint8, "12", dsModels.Uint8, false},
		{"Uint8NegativeFail", profileNameRandomUnsignedIntegerGenerator, typeUint8, &operationSetUint8, "-12", dsModels.Uint8, true},
		{"Uint8WordFail", profileNameRandomUnsignedIntegerGenerator, typeUint8, &operationSetUint8, "hello", dsModels.Uint8, true},
		{"Uint8OverflowFail", profileNameRandomUnsignedIntegerGenerator, typeUint8, &operationSetUint8, "9999999999", dsModels.Uint8, true},
		{"Uint16Pass", profileNameRandomUnsignedIntegerGenerator, typeUint16, &operationSetUint16, "12", dsModels.Uint16, false},
		{"Uint16NegativeFail", profileNameRandomUnsignedIntegerGenerator, typeUint16, &operationSetUint16, "-12", dsModels.Uint16, true},
		{"Uint16WordFail", profileNameRandomUnsignedIntegerGenerator, typeUint16, &operationSetUint16, "hello", dsModels.Uint16, true},
		{"Uint16OverflowFail", profileNameRandomUnsignedIntegerGenerator, typeUint16, &operationSetUint16, "9999999999", dsModels.Uint16, true},
		{"Uint32Pass", profileNameRandomUnsignedIntegerGenerator, typeUint32, &operationSetUint32, "12", dsModels.Uint32, false},
		{"Uint32NegativeFail", profileNameRandomUnsignedIntegerGenerator, typeUint32, &operationSetUint32, "-12", dsModels.Uint32, true},
		{"Uint32WordFail", profileNameRandomUnsignedIntegerGenerator, typeUint32, &operationSetUint32, "hello", dsModels.Uint32, true},
		{"Uint32OverflowFail", profileNameRandomUnsignedIntegerGenerator, typeUint32, &operationSetUint32, "9999999999", dsModels.Uint32, true},
		{"Uint64Pass", profileNameRandomUnsignedIntegerGenerator, typeUint64, &operationSetUint64, "12", dsModels.Uint64, false},
		{"Uint64NegativeFail", profileNameRandomUnsignedIntegerGenerator, typeUint64, &operationSetUint64, "-12", dsModels.Uint64, true},
		{"Uint64WordFail", profileNameRandomUnsignedIntegerGenerator, typeUint64, &operationSetUint64, "hello", dsModels.Uint64, true},
		{"Uint64OverflowFail", profileNameRandomUnsignedIntegerGenerator, typeUint64, &operationSetUint64, "99999999999999999999", dsModels.Uint64, true},
		{"Float32Pass", profileNameRandomFloatGenerator, typeFloat32, &operationSetFloat32, "12.000", dsModels.Float32, false},
		{"Float32PassNegativePass", profileNameRandomFloatGenerator, typeFloat32, &operationSetFloat32, "-12.000", dsModels.Float32, false},
		{"Float32PassWordFail", profileNameRandomFloatGenerator, typeFloat32, &operationSetFloat32, "hello", dsModels.Float32, true},
		{"Float32PassOverflowFail", profileNameRandomFloatGenerator, typeFloat32, &operationSetFloat32, "440282346638528859811704183484516925440.0000000000000000", dsModels.Float32, true},
		{"Float64Pass", profileNameRandomFloatGenerator, typeFloat64, &operationSetFloat64, "12.000", dsModels.Float64, false},
		{"Float64PassNegativePass", profileNameRandomFloatGenerator, typeFloat64, &operationSetFloat64, "-12.000", dsModels.Float64, false},
		{"Float64PassWordFail", profileNameRandomFloatGenerator, typeFloat64, &operationSetFloat64, "hello", dsModels.Float64, true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cv, err := createCommandValueFromRO(tt.profileName, tt.op, tt.v)
			if !tt.expectErr && err != nil {
				t.Errorf("%s expectErr:%v error:%v", tt.testName, tt.expectErr, err)
				return
			}
			if tt.expectErr && err == nil {
				t.Errorf("%s expectErr:%v no error thrown", tt.testName, tt.expectErr)
				return
			}
			if cv != nil {
				var check dsModels.ValueType
				switch strings.ToLower(tt.valueType) {
				case "bool":
					check = dsModels.Bool
				case "string":
					check = dsModels.String
				case "uint8":
					check = dsModels.Uint8
				case "uint16":
					check = dsModels.Uint16
				case "uint32":
					check = dsModels.Uint32
				case "uint64":
					check = dsModels.Uint64
				case "int8":
					check = dsModels.Int8
				case "int16":
					check = dsModels.Int16
				case "int32":
					check = dsModels.Int32
				case "int64":
					check = dsModels.Int64
				case "float32":
					check = dsModels.Float32
				case "float64":
					check = dsModels.Float64
				}
				if cv.Type != check {
					t.Errorf("%s incorrect parsing. valueType: %s result: %v", tt.testName, tt.valueType, cv.Type)
				}
			}
		})
	}
}

func TestResourceOpSliceToMap(t *testing.T) {
	var ops []contract.ResourceOperation
	ops = append(ops, contract.ResourceOperation{Object: "first"})
	ops = append(ops, contract.ResourceOperation{Object: "second"})
	ops = append(ops, contract.ResourceOperation{Object: "third"})

	mapped := roSliceToMap(ops)

	if len(mapped) != 3 {
		t.Errorf("unexpected map length. wanted 3, got %v", len(mapped))
		return
	}

	tests := []struct {
		testName  string
		key       string
		expectErr bool
	}{
		{"FindFirst", "first", false},
		{"FindSecond", "second", false},
		{"FindThird", "third", false},
		{"NotFoundKey", "fourth", true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, ok := mapped[tt.key]
			if !tt.expectErr && !ok {
				t.Errorf("expected entry %s not found in map.", tt.key)
				return
			}
			if tt.expectErr && ok {
				t.Errorf("test %s expected error not received.", tt.testName)
				return
			}
		})
	}
}

func TestParseWriteParams(t *testing.T) {
	profileName := profileNameRandomIntegerGenerator
	profile, ok := pc.ForName(profileName)
	if !ok {
		t.Errorf("device profile was not found, cannot continue")
		return
	}
	roMap := roSliceToMap(profile.DeviceCommands[0].Set)
	roMapTestMappingPass := roSliceToMap(profile.DeviceCommands[7].Set)
	roMapTestMappingFail := roSliceToMap(profile.DeviceCommands[8].Set)
	tests := []struct {
		testName    string
		profile     string
		resourceOps map[string]*contract.ResourceOperation
		params      string
		expectErr   bool
	}{
		{"ValidWriteParam", profileName, roMap, `{"RandomValue_Int8":"123"}`, false},
		{"InvalidWriteParam", profileName, roMap, `{"NotFound":"true"}`, true},
		{"InvalidWriteParamType", profileName, roMap, `{"RandomValue_Int8":"abc"}`, true},
		{"ValueMappingPass", profileName, roMapTestMappingPass, `{"ResourceTestMapping_Pass":"Pass"}`, false},
		//The expectErr on the test below is false because parseWriteParams does NOT throw an error when there is no mapping value matched
		{"ValueMappingFail", profileName, roMapTestMappingFail, `{"ResourceTestMapping_Fail":"123"}`, false},
		{"ParseParamsFail", profileName, roMap, ``, true},
		{"NoParams", profileName, roMap, `{}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := parseWriteParams(tt.profile, tt.resourceOps, tt.params)
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected parse error params:%s %s", tt.params, err.Error())
				return
			}
			if tt.expectErr && err == nil {
				t.Errorf("expected error was not received params:%s", tt.params)
				return
			}
		})
	}
}

func TestExecReadCmd(t *testing.T) {
	tests := []struct {
		testName  string
		device    *contract.Device
		cmd       string
		expectErr bool
	}{
		{"CmdExecutionPass", &deviceIntegerGenerator, "RandomValue_Int8", false},
		{"CmdNotFound", &deviceIntegerGenerator, "InexistentCmd", true},
		{"ValueTransformFail", &deviceIntegerGenerator, "ResourceTestTransform_Fail", true},
		{"ValueAssertionPass", &deviceIntegerGenerator, "ResourceTestAssertion_Pass", false},
		//The expectErr on the test below is false because execReadCmd does NOT throw an error when assertion failed
		{"ValueAssertionFail", &deviceIntegerGenerator, "ResourceTestAssertion_Fail", false},
		{"ValueMappingPass", &deviceIntegerGenerator, "ResourceTestMapping_Pass", false},
		//The expectErr on the test below is false because execReadCmd does NOT throw an error when there is no mapping value matched
		{"ValueMappingFail", &deviceIntegerGenerator, "ResourceTestMapping_Fail", false},
		{"NoDeviceResourceForOperation", &deviceIntegerGenerator, "NoDeviceResourceForOperation", true},
		{"NoDeviceResourceForResult", &deviceIntegerGenerator, "NoDeviceResourceForResult", true},
		{"MaxCmdOpsExceeded", &deviceIntegerGenerator, "Error", true},
		{"ErrorOccurredInDriver", &deviceIntegerGenerator, "Error", true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.testName == "MaxCmdOpsExceeded" {
				common.CurrentConfig.Device.MaxCmdOps = 1
				defer func() {
					common.CurrentConfig.Device.MaxCmdOps = 128
				}()
			}
			v, err := execReadCmd(tt.device, tt.cmd)
			if !tt.expectErr && err != nil {
				t.Errorf("%s expectErr:%v error:%v", tt.testName, tt.expectErr, err)
				return
			}
			if tt.expectErr && err == nil {
				t.Errorf("%s expectErr:%v no error thrown", tt.testName, tt.expectErr)
				return
			}
			//The way to determine whether the assertion passed or failed is to see the return value contains "Assertion failed" or not
			if tt.testName == "ValueAssertionPass" && strings.Contains(v.Readings[0].Value, "Assertion failed") {
				t.Errorf("%s expect data assertion pass", tt.testName)
			}
			if tt.testName == "ValueAssertionFail" && !strings.Contains(v.Readings[0].Value, "Assertion failed") {
				t.Errorf("%s expect data assertion failed", tt.testName)
			}
			// issue #89 will discuss how to handle there is no mapping matched
			if tt.testName == "ValueMappingPass" && v.Readings[0].Value == strconv.Itoa(int(mock.Int8Value)) {
				t.Errorf("%s expect data mapping pass", tt.testName)
			}
			if tt.testName == "ValueMappingFail" && v.Readings[0].Value != strconv.Itoa(int(mock.Int8Value)) {
				t.Errorf("%s expect data mapping failed", tt.testName)
			}
		})
	}
}

func TestExecWriteCmd(t *testing.T) {
	var (
		paramsInt8                      = `{"RandomValue_Int8":"123"}`
		paramsError                     = `{"Error":"error"}`
		paramsTransformFail             = `{"ResourceTestTransform_Fail":"123"}`
		paramsNoDeviceResourceForResult = `{"error":""}`
	)
	tests := []struct {
		testName  string
		device    *contract.Device
		cmd       string
		params    string
		expectErr bool
	}{
		{"CmdExecutionPass", &deviceIntegerGenerator, "RandomValue_Int8", paramsInt8, false},
		{"CmdNotFound", &deviceIntegerGenerator, "inexistentCmd", paramsInt8, true},
		{"MaxCmdOpsExceeded", &deviceIntegerGenerator, "Error", paramsInt8, true},
		{"NoDeviceResourceForOperation", &deviceIntegerGenerator, "NoDeviceResourceForOperation", paramsError, true},
		{"NoDeviceResourceForResult", &deviceIntegerGenerator, "NoDeviceResourceForResult", paramsNoDeviceResourceForResult, true},
		{"DataTransformFail", &deviceIntegerGenerator, "ResourceTestTransform_Fail", paramsTransformFail, true},
		{"ErrorOccurredInDriver", &deviceIntegerGenerator, "Error", paramsError, true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.testName == "MaxCmdOpsExceeded" {
				common.CurrentConfig.Device.MaxCmdOps = 1
				defer func() {
					common.CurrentConfig.Device.MaxCmdOps = 128
				}()
			}
			appErr := execWriteCmd(tt.device, tt.cmd, tt.params)
			if !tt.expectErr && appErr != nil {
				t.Errorf("%s expectErr:%v error:%v", tt.testName, tt.expectErr, appErr.Error())
				return
			}
			if tt.expectErr && appErr == nil {
				t.Errorf("%s expectErr:%v no error thrown", tt.testName, tt.expectErr)
				return
			}
		})
	}
}

func TestCommandAllHandler(t *testing.T) {
	tests := []struct {
		testName  string
		cmd       string
		body      string
		method    string
		expectErr bool
	}{
		{"PartOfReadCommandExecutionSuccess", "RandomValue_Uint8", "", methodGet, false},
		{"PartOfReadCommandExecutionFail", "error", "", methodGet, true},
		{"PartOfWriteCommandExecutionSuccess", "RandomValue_Uint8", `{"RandomValue_Uint8":"123"}`, methodSet, false},
		{"PartOfWriteCommandExecutionFail", "error", `{"RandomValue_Uint8":"123"}`, methodSet, true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, appErr := CommandAllHandler(tt.cmd, tt.body, tt.method)
			if !tt.expectErr && appErr != nil {
				t.Errorf("%s expectErr:%v error:%v", tt.testName, tt.expectErr, appErr.Error())
				return
			}
			if tt.expectErr && appErr == nil {
				t.Errorf("%s expectErr:%v no error thrown", tt.testName, tt.expectErr)
				return
			}
		})
	}
}

func TestCommandHandler(t *testing.T) {
	var (
		varsFindDeviceByValidId     = map[string]string{"id": mock.RandomIntegerGeneratorDeviceId, "command": "RandomValue_Int8"}
		varsFindDeviceByInvalidId   = map[string]string{"id": mock.InvalidDeviceId, "command": "RandomValue_Int8"}
		varsFindDeviceByValidName   = map[string]string{"name": "Random-Integer-Generator01", "command": "RandomValue_Int8"}
		varsFindDeviceByInvalidName = map[string]string{"name": "Random-Integer-Generator09", "command": "RandomValue_Int8"}
		varsAdminStateLocked        = map[string]string{"name": "Random-Float-Generator01", "command": "RandomValue_Float32"}
		varsProfileNotFound         = map[string]string{"name": "Random-Boolean-Generator01", "command": "error"}
		varsCmdNotFound             = map[string]string{"name": "Random-Integer-Generator01", "command": "error"}
		varsWriteInt8               = map[string]string{"name": "Random-Integer-Generator01", "command": "RandomValue_Int8"}
	)
	if err := cache.Devices().UpdateAdminState(mock.RandomFloatGeneratorDeviceId, contract.Locked); err != nil {
		t.Errorf("Fail to update adminState, error: %v", err)
	}
	mock.ValidDeviceRandomBoolGenerator.Profile.Name = "error"
	if err := cache.Devices().Update(mock.ValidDeviceRandomBoolGenerator); err != nil {
		t.Errorf("Fail to update device, error: %v", err)
	}

	tests := []struct {
		testName  string
		vars      map[string]string
		body      string
		method    string
		expectErr bool
	}{
		{"ValidDeviceId", varsFindDeviceByValidId, "", methodGet, false},
		{"InvalidDeviceId", varsFindDeviceByInvalidId, "", methodGet, true},
		{"ValidDeviceName", varsFindDeviceByValidName, "", methodGet, false},
		{"InvalidDeviceName", varsFindDeviceByInvalidName, "", methodGet, true},
		{"AdminStateLocked", varsAdminStateLocked, "", methodGet, true},
		{"ProfileNotFound", varsProfileNotFound, "", methodGet, true},
		{"CmdNotFound", varsCmdNotFound, "", methodGet, true},
		{"WriteCommand", varsWriteInt8, `{"RandomValue_Int8":"123"}`, methodSet, false},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, appErr := CommandHandler(tt.vars, tt.body, tt.method)
			if !tt.expectErr && appErr != nil {
				t.Errorf("%s expectErr:%v error:%v", tt.testName, tt.expectErr, appErr.Error())
				return
			}
			if tt.expectErr && appErr == nil {
				t.Errorf("%s expectErr:%v no error thrown", tt.testName, tt.expectErr)
				return
			}
		})
	}
}
