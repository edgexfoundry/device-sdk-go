// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/transformer"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/mock"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"

	"github.com/stretchr/testify/require"
)

const (
	methodGet   = "get"
	methodSet   = "set"
	typeBool    = "Bool"
	typeInt8    = "Int8"
	typeInt16   = "Int16"
	typeInt32   = "Int32"
	typeInt64   = "Int64"
	typeUint8   = "Uint8"
	typeUint16  = "Uint16"
	typeUint32  = "Uint32"
	typeUint64  = "Uint64"
	typeFloat32 = "Float32"
	typeFloat64 = "Float64"
)

var (
	ds                                                                            []contract.Device
	pc                                                                            cache.ProfileCache
	deviceIntegerGenerator                                                        contract.Device
	deviceFloatGenerator                                                          contract.Device
	operationSetBool                                                              contract.ResourceOperation
	operationSetInt8, operationSetInt16, operationSetInt32, operationSetInt64     contract.ResourceOperation
	operationSetUint8, operationSetUint16, operationSetUint32, operationSetUint64 contract.ResourceOperation
	operationSetFloat32, operationSetFloat64                                      contract.ResourceOperation
	vdc                                                                           *mock.ValueDescriptorMock
	pwc                                                                           *mock.ProvisionWatcherClientMock
	dc                                                                            *mock.DeviceClientMock
	ec                                                                            *mock.EventClientMock
	lc                                                                            logger.LoggingClient
	driver                                                                        *mock.DriverMock
	configuration                                                                 *common.ConfigurationStruct
	dic                                                                           *di.Container
)

func init() {
	vdc = &mock.ValueDescriptorMock{}
	pwc = &mock.ProvisionWatcherClientMock{}
	dc = &mock.DeviceClientMock{}
	ec = &mock.EventClientMock{}
	lc = logger.NewClientStdOut("device-sdk-test", false, "DEBUG")
	driver = &mock.DriverMock{}
	deviceInfo := common.DeviceInfo{DataTransform: true, MaxCmdOps: 128}
	configuration = &common.ConfigurationStruct{Device: deviceInfo}
	dic = di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
		container.CoredataValueDescriptorClientName: func(get di.Get) interface{} {
			return vdc
		},
		container.MetadataDeviceClientName: func(get di.Get) interface{} {
			return dc
		},
		container.CoredataEventClientName: func(get di.Get) interface{} {
			return ec
		},
		container.ProtocolDriverName: func(get di.Get) interface{} {
			return driver
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	cache.InitCache("device-sdk-test", lc, vdc, dc, pwc)
	pc = cache.Profiles()
	operationSetBool, _ = pc.ResourceOperation(mock.ProfileBool, mock.ResourceObjectBool, methodSet)
	operationSetInt8, _ = pc.ResourceOperation(mock.ProfileInt, mock.ResourceObjectInt8, methodSet)
	operationSetInt16, _ = pc.ResourceOperation(mock.ProfileInt, mock.ResourceObjectInt16, methodSet)
	operationSetInt32, _ = pc.ResourceOperation(mock.ProfileInt, mock.ResourceObjectInt32, methodSet)
	operationSetInt64, _ = pc.ResourceOperation(mock.ProfileInt, mock.ResourceObjectInt64, methodSet)
	operationSetUint8, _ = pc.ResourceOperation(mock.ProfileUint, mock.ResourceObjectUint8, methodSet)
	operationSetUint16, _ = pc.ResourceOperation(mock.ProfileUint, mock.ResourceObjectUint16, methodSet)
	operationSetUint32, _ = pc.ResourceOperation(mock.ProfileUint, mock.ResourceObjectUint32, methodSet)
	operationSetUint64, _ = pc.ResourceOperation(mock.ProfileUint, mock.ResourceObjectUint64, methodSet)
	operationSetFloat32, _ = pc.ResourceOperation(mock.ProfileFloat, mock.ResourceObjectFloat32, methodSet)
	operationSetFloat64, _ = pc.ResourceOperation(mock.ProfileFloat, mock.ResourceObjectFloat64, methodSet)

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	ds, _ = dc.DevicesForServiceByName(ctx, "device-sdk-test")
	deviceIntegerGenerator = ds[1]
	deviceFloatGenerator = ds[3]
}

func TestParseWriteParamsWrongParamName(t *testing.T) {
	profileName := "notFound"
	ro := []contract.ResourceOperation{{Index: ""}}
	params := "{ \"key\": \"value\" }"

	_, err := parseWriteParams(profileName, ro, params, lc)

	if err == nil {
		t.Error("expected error")
	}
}

func TestParseWriteParamsNoParams(t *testing.T) {
	profileName := "notFound"
	ro := []contract.ResourceOperation{{Index: ""}}
	params := "{ }"

	_, err := parseWriteParams(profileName, ro, params, lc)

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
		{"DeviceResourceNotFound", mock.ProfileBool, typeBool, &contract.ResourceOperation{}, "", dsModels.Bool, true},
		{"BoolTruePass", mock.ProfileBool, typeBool, &operationSetBool, "true", dsModels.Bool, false},
		{"BoolFalsePass", mock.ProfileBool, typeBool, &operationSetBool, "false", dsModels.Bool, false},
		{"BoolTrueFail", mock.ProfileBool, typeBool, &operationSetBool, "error", dsModels.Bool, true},
		{"Int8Pass", mock.ProfileInt, typeInt8, &operationSetInt8, "12", dsModels.Int8, false},
		{"Int8NegativePass", mock.ProfileInt, typeInt8, &operationSetInt8, "-12", dsModels.Int8, false},
		{"Int8WordFail", mock.ProfileInt, typeInt8, &operationSetInt8, "hello", dsModels.Int8, true},
		{"Int8OverflowFail", mock.ProfileInt, typeInt8, &operationSetInt8, "9999999999", dsModels.Int8, true},
		{"Int16Pass", mock.ProfileInt, typeInt16, &operationSetInt16, "12", dsModels.Int16, false},
		{"Int16NegativePass", mock.ProfileInt, typeInt16, &operationSetInt16, "-12", dsModels.Int16, false},
		{"Int16WordFail", mock.ProfileInt, typeInt16, &operationSetInt16, "hello", dsModels.Int16, true},
		{"Int16OverflowFail", mock.ProfileInt, typeInt16, &operationSetInt16, "9999999999", dsModels.Int16, true},
		{"Int32Pass", mock.ProfileInt, typeInt32, &operationSetInt32, "12", dsModels.Int32, false},
		{"Int32NegativePass", mock.ProfileInt, typeInt32, &operationSetInt32, "-12", dsModels.Int32, false},
		{"Int32WordFail", mock.ProfileInt, typeInt32, &operationSetInt32, "hello", dsModels.Int32, true},
		{"Int32OverflowFail", mock.ProfileInt, typeInt32, &operationSetInt32, "9999999999", dsModels.Int32, true},
		{"Int64Pass", mock.ProfileInt, typeInt64, &operationSetInt64, "12", dsModels.Int64, false},
		{"Int64NegativePass", mock.ProfileInt, typeInt64, &operationSetInt64, "-12", dsModels.Int64, false},
		{"Int64WordFail", mock.ProfileInt, typeInt64, &operationSetInt64, "hello", dsModels.Int64, true},
		{"Int64OverflowFail", mock.ProfileInt, typeInt64, &operationSetInt64, "99999999999999999999", dsModels.Int64, true},
		{"Uint8Pass", mock.ProfileUint, typeUint8, &operationSetUint8, "12", dsModels.Uint8, false},
		{"Uint8NegativeFail", mock.ProfileUint, typeUint8, &operationSetUint8, "-12", dsModels.Uint8, true},
		{"Uint8WordFail", mock.ProfileUint, typeUint8, &operationSetUint8, "hello", dsModels.Uint8, true},
		{"Uint8OverflowFail", mock.ProfileUint, typeUint8, &operationSetUint8, "9999999999", dsModels.Uint8, true},
		{"Uint16Pass", mock.ProfileUint, typeUint16, &operationSetUint16, "12", dsModels.Uint16, false},
		{"Uint16NegativeFail", mock.ProfileUint, typeUint16, &operationSetUint16, "-12", dsModels.Uint16, true},
		{"Uint16WordFail", mock.ProfileUint, typeUint16, &operationSetUint16, "hello", dsModels.Uint16, true},
		{"Uint16OverflowFail", mock.ProfileUint, typeUint16, &operationSetUint16, "9999999999", dsModels.Uint16, true},
		{"Uint32Pass", mock.ProfileUint, typeUint32, &operationSetUint32, "12", dsModels.Uint32, false},
		{"Uint32NegativeFail", mock.ProfileUint, typeUint32, &operationSetUint32, "-12", dsModels.Uint32, true},
		{"Uint32WordFail", mock.ProfileUint, typeUint32, &operationSetUint32, "hello", dsModels.Uint32, true},
		{"Uint32OverflowFail", mock.ProfileUint, typeUint32, &operationSetUint32, "9999999999", dsModels.Uint32, true},
		{"Uint64Pass", mock.ProfileUint, typeUint64, &operationSetUint64, "12", dsModels.Uint64, false},
		{"Uint64NegativeFail", mock.ProfileUint, typeUint64, &operationSetUint64, "-12", dsModels.Uint64, true},
		{"Uint64WordFail", mock.ProfileUint, typeUint64, &operationSetUint64, "hello", dsModels.Uint64, true},
		{"Uint64OverflowFail", mock.ProfileUint, typeUint64, &operationSetUint64, "99999999999999999999", dsModels.Uint64, true},
		{"Float32Pass", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "12.000", dsModels.Float32, false},
		{"Float32PassWithBase64String", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "QUAAAA==", dsModels.Float32, false},
		{"Float32PassWithENotation", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "0.123123e-05", dsModels.Float32, false},
		{"Float32PassNegativePass", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "-12.000", dsModels.Float32, false},
		{"Float32PassNegativePassWithBase64String", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "wUAAAA==", dsModels.Float32, false},
		{"Float32PassNegativePassWithENotation", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "-0.123123e-05", dsModels.Float32, false},
		{"Float32PassWordFail", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "hello", dsModels.Float32, true},
		{"Float32PassOverflowFail", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "440282346638528859811704183484516925440.0000000000000000", dsModels.Float32, true},
		{"Float32PassOverflowFailWithBase64String", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "f+////////8=", dsModels.Float32, true},
		{"Float32PassOverflowFailWithENotation", mock.ProfileFloat, typeFloat32, &operationSetFloat32, "12.000e+38", dsModels.Float32, true},
		{"Float64Pass", mock.ProfileFloat, typeFloat64, &operationSetFloat64, "12.000", dsModels.Float64, false},
		{"Float64PassWithBase64String", mock.ProfileFloat, typeFloat64, &operationSetFloat64, "QCgAAAAAAAA=", dsModels.Float64, false},
		{"Float64PassWithENotation", mock.ProfileFloat, typeFloat64, &operationSetFloat64, "0.12345e-16", dsModels.Float64, false},
		{"Float64PassNegativePass", mock.ProfileFloat, typeFloat64, &operationSetFloat64, "-12.000", dsModels.Float64, false},
		{"Float64PassNegativePassWithBase64String", mock.ProfileFloat, typeFloat64, &operationSetFloat64, "wCgAAAAAAAA=", dsModels.Float64, false},
		{"Float64PassNegativePassWithENotation", mock.ProfileFloat, typeFloat64, &operationSetFloat64, "-0.12345e-16", dsModels.Float64, false},
		{"Float64PassWordFail", mock.ProfileFloat, typeFloat64, &operationSetFloat64, "hello", dsModels.Float64, true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cv, err := createCommandValueFromRO(tt.profileName, tt.op, tt.v, lc)
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

func TestParseWriteParams(t *testing.T) {
	profileName := mock.ProfileInt

	ros, _ := cache.Profiles().ResourceOperations(profileName, "RandomValue_Int8", common.SetCmdMethod)
	rosTestDefaultParam, _ := cache.Profiles().ResourceOperations(profileName, "RandomValue_Int16", common.SetCmdMethod)
	rosTestDefaultValue, _ := cache.Profiles().ResourceOperations(profileName, "RandomValue_Int32", common.SetCmdMethod)
	rosTestMappingPass, _ := cache.Profiles().ResourceOperations(profileName, "ResourceTestMapping_Pass", common.SetCmdMethod)
	rosTestMappingFail, _ := cache.Profiles().ResourceOperations(profileName, "ResourceTestMapping_Fail", common.SetCmdMethod)

	tests := []struct {
		testName    string
		profile     string
		resourceOps []contract.ResourceOperation
		params      string
		expectErr   bool
	}{
		{"ValidWriteParam", profileName, ros, `{"RandomValue_Int8":"123"}`, false},
		{"InvalidWriteParam", profileName, ros, `{"NotFound":"true"}`, true},
		{"InvalidWriteParamType", profileName, ros, `{"RandomValue_Int8":"abc"}`, true},
		{"ValueMappingPass", profileName, rosTestMappingPass, `{"ResourceTestMapping_Pass":"Pass"}`, false},
		// The expectErr on the test below is false because parseWriteParams does NOT throw an error when there is no mapping value matched
		{"ValueMappingFail", profileName, rosTestMappingFail, `{"ResourceTestMapping_Fail":"123"}`, false},
		{"ParseParamsFail", profileName, ros, ``, true},
		{"NoRequestParameter", profileName, ros, `{}`, true},
		{"DefaultParameter", profileName, rosTestDefaultParam, `{"NotMatchedResourceName":"value"}`, false},
		{"DefaultValue", profileName, rosTestDefaultValue, `{"NotMatchedResourceName":"value"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := parseWriteParams(tt.profile, tt.resourceOps, tt.params, lc)
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
		testName    string
		device      *contract.Device
		cmd         string
		queryParams string
		expectErr   bool
	}{
		{"CmdExecutionPass", &deviceIntegerGenerator, "RandomValue_Int8", "", false},
		{"CmdNotFound", &deviceIntegerGenerator, "InexistentCmd", "", true},
		{"ValueTransformFail", &deviceIntegerGenerator, "ResourceTestTransform_Fail", "", true},
		{"ValueAssertionPass", &deviceIntegerGenerator, "ResourceTestAssertion_Pass", "", false},
		// The expectErr on the test below is false because execReadCmd does NOT throw an error when assertion failed
		{"ValueAssertionFail", &deviceIntegerGenerator, "ResourceTestAssertion_Fail", "", false},
		{"ValueMappingPass", &deviceIntegerGenerator, "ResourceTestMapping_Pass", "", false},
		// The expectErr on the test below is false because execReadCmd does NOT throw an error when there is no mapping value matched
		{"ValueMappingFail", &deviceIntegerGenerator, "ResourceTestMapping_Fail", "", false},
		{"NoDeviceResourceForOperation", &deviceIntegerGenerator, "NoDeviceResourceForOperation", "", true},
		{"NoDeviceResourceForResult", &deviceIntegerGenerator, "NoDeviceResourceForResult", "", true},
		{"MaxCmdOpsExceeded", &deviceIntegerGenerator, "Error", "", true},
		{"ErrorOccurredInDriver", &deviceIntegerGenerator, "Error", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.testName == "MaxCmdOpsExceeded" {
				configuration.Device.MaxCmdOps = 1
				defer func() {
					configuration.Device.MaxCmdOps = 128
				}()
			}
			v, err := execReadCmd(tt.device, tt.cmd, tt.queryParams, driver, lc, dc, configuration)
			if !tt.expectErr && err != nil {
				t.Errorf("%s expectErr:%v error:%v", tt.testName, tt.expectErr, err)
				return
			}
			if tt.expectErr && err == nil {
				t.Errorf("%s expectErr:%v no error thrown", tt.testName, tt.expectErr)
				return
			}
			// The way to determine whether the assertion passed or failed is to see the return value contains "Assertion failed" or not
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
				configuration.Device.MaxCmdOps = 1
				defer func() {
					configuration.Device.MaxCmdOps = 128
				}()
			}
			appErr := execWriteCmd(tt.device, tt.cmd, tt.params, driver, lc, configuration)
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
		testName    string
		cmd         string
		body        string
		queryParams string
		method      string
		expectErr   bool
	}{
		{"PartOfReadCommandExecutionSuccess", "RandomValue_Uint8", "", "", methodGet, false},
		{"PartOfReadCommandExecutionSuccessWithQueryParams", "RandomValue_Uint8", "", "test=test&test2=test2", methodGet, false},
		{"PartOfReadCommandExecutionFail", "error", "", "", methodGet, true},
		{"PartOfWriteCommandExecutionSuccess", "RandomValue_Uint8", `{"RandomValue_Uint8":"123"}`, "", methodSet, false},
		{"PartOfWriteCommandExecutionFail", "error", `{"RandomValue_Uint8":"123"}`, "", methodSet, true},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, appErr := CommandAllHandler(tt.cmd, tt.body, tt.method, tt.queryParams, dic)
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
		varsFindDeviceByValidId     = map[string]string{"id": mock.ValidDeviceRandomUnsignedIntegerGenerator.Id, "command": "RandomValue_Uint8"}
		varsFindDeviceByInvalidId   = map[string]string{"id": mock.InvalidDeviceId, "command": "RandomValue_Int8"}
		varsFindDeviceByValidName   = map[string]string{"name": "Random-UnsignedInteger-Generator01", "command": "RandomValue_Uint8"}
		varsFindDeviceByInvalidName = map[string]string{"name": "Random-Integer-Generator09", "command": "RandomValue_Int8"}
		varsAdminStateLocked        = map[string]string{"name": "Random-Float-Generator01", "command": "RandomValue_Float32"}
		varsOperatingStateDisabled  = map[string]string{"name": mock.OperatingStateDisabled.Name, "command": "testrandfloat32"}
		varsProfileNotFound         = map[string]string{"name": "Random-Boolean-Generator01", "command": "error"}
		varsCmdNotFound             = map[string]string{"name": "Random-Integer-Generator01", "command": "error"}
		varsWriteUint8              = map[string]string{"name": "Random-UnsignedInteger-Generator01", "command": "RandomValue_Uint8"}
	)
	if err := cache.Devices().UpdateAdminState(mock.ValidDeviceRandomFloatGenerator.Id, contract.Locked); err != nil {
		t.Errorf("Fail to update adminState, error: %v", err)
	}
	mock.ValidDeviceRandomBoolGenerator.Profile.Name = "error"
	if err := cache.Devices().Update(mock.ValidDeviceRandomBoolGenerator); err != nil {
		t.Errorf("Fail to update device, error: %v", err)
	}

	tests := []struct {
		testName    string
		vars        map[string]string
		body        string
		method      string
		queryParams string
		expectErr   bool
	}{
		{"ValidDeviceId", varsFindDeviceByValidId, "", methodGet, "test=test&test2=test2", false},
		{"InvalidDeviceId", varsFindDeviceByInvalidId, "", methodGet, "", true},
		{"ValidDeviceName", varsFindDeviceByValidName, "", methodGet, "", false},
		{"InvalidDeviceName", varsFindDeviceByInvalidName, "", methodGet, "", true},
		{"AdminStateLocked", varsAdminStateLocked, "", methodGet, "", true},
		{"OperatingStateDisabled", varsOperatingStateDisabled, "", methodGet, "", true},
		{"ProfileNotFound", varsProfileNotFound, "", methodGet, "", true},
		{"CmdNotFound", varsCmdNotFound, "", methodGet, "", true},
		{"WriteCommand", varsWriteUint8, `{"RandomValue_Uint8":"123"}`, methodSet, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, appErr := CommandHandler(tt.vars, tt.body, tt.method, tt.queryParams, dic)
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

func TestExecReadCmd_NaN(t *testing.T) {
	tests := []struct {
		testName    string
		device      *contract.Device
		cmd         string
		queryParams string
		expectErr   bool
	}{
		{"Float32 NaN", &deviceFloatGenerator, mock.ResourceObjectNaNFloat32, "", false},
		{"Float64 NaN", &deviceFloatGenerator, mock.ResourceObjectNaNFloat64, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			v, err := execReadCmd(tt.device, tt.cmd, tt.queryParams, driver, lc, dc, configuration)
			require.Nil(t, err)
			assert.Equal(t, transformer.NaN, v.Readings[0].Value)
		})
	}
}
