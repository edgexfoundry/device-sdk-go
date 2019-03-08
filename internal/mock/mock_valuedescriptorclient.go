// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var (
	ValueDescriptorEnableRandomization = models.ValueDescriptor{}
	ValueDescriptorBool                = models.ValueDescriptor{}
	ValueDescriptorInt8                = models.ValueDescriptor{}
	ValueDescriptorInt16               = models.ValueDescriptor{}
	ValueDescriptorInt32               = models.ValueDescriptor{}
	ValueDescriptorInt64               = models.ValueDescriptor{}
	ValueDescriptorUint8               = models.ValueDescriptor{}
	ValueDescriptorUint16              = models.ValueDescriptor{}
	ValueDescriptorUint32              = models.ValueDescriptor{}
	ValueDescriptorUint64              = models.ValueDescriptor{}
	ValueDescriptorFloat32             = models.ValueDescriptor{}
	ValueDescriptorFloat64             = models.ValueDescriptor{}
	//ValueDescriptorString              = models.ValueDescriptor{}
	NewValueDescriptor            = models.ValueDescriptor{}
	DuplicateValueDescriptorInt16 = models.ValueDescriptor{}
)

type ValueDescriptorMock struct{}

func (ValueDescriptorMock) ValueDescriptors(ctx context.Context) ([]models.ValueDescriptor, error) {
	populateValueDescriptorMock()
	return []models.ValueDescriptor{
		ValueDescriptorEnableRandomization,
		ValueDescriptorBool,
		ValueDescriptorInt8,
		ValueDescriptorInt16,
		ValueDescriptorInt32,
		ValueDescriptorInt64,
		ValueDescriptorUint8,
		ValueDescriptorUint16,
		ValueDescriptorUint32,
		ValueDescriptorUint64,
		ValueDescriptorFloat32,
		ValueDescriptorFloat64,
		//ValueDescriptorString,
	}, nil
}

func (ValueDescriptorMock) ValueDescriptor(id string, ctx context.Context) (models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorForName(name string, ctx context.Context) (models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsByLabel(label string, ctx context.Context) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsForDevice(deviceId string, ctx context.Context) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsForDeviceByName(deviceName string, ctx context.Context) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) ValueDescriptorsByUomLabel(uomLabel string, ctx context.Context) ([]models.ValueDescriptor, error) {
	panic("implement me")
}

func (ValueDescriptorMock) Add(vdr *models.ValueDescriptor, ctx context.Context) (string, error) {
	panic("implement me")
}

func (ValueDescriptorMock) Update(vdr *models.ValueDescriptor, ctx context.Context) error {
	panic("implement me")
}

func (ValueDescriptorMock) Delete(id string, ctx context.Context) error {
	panic("implement me")
}

func (ValueDescriptorMock) DeleteByName(name string, ctx context.Context) error {
	panic("implement me")
}

func populateValueDescriptorMock() {
	valueDescriptorDataEnableRandomization := `{"id":"4e8bed38-4db0-4d70-add1-aa0458b2bf61","created":1551711642684,"description":"used to decide whether to re-generate a random value","modified":0,"origin":0,"name":"Enable_Randomization","min":null,"max":null,"defaultValue":null,"type":"Bool","uomLabel":"Random","formatting":"%s","labels":null}`
	valueDescriptorDataBool := `{"id":"e6dcedf9-d6c2-47fe-b38d-d03e8f098846","created":1551711642678,"description":"Generate random boolean value","modified":0,"origin":0,"name":"RandomValue_Bool","min":"false","max":"true","defaultValue":"true","type":"Bool","uomLabel":"random bool value","formatting":"%s","labels":null}`
	valueDescriptorDataInt8 := `{"id":"d3a272a6-f625-42a2-9d40-50b8c2cd5a4c","created":1552274458525,"description":"Generate random int8 value","modified":0,"origin":0,"name":"RandomValue_Int8","min":"-100","max":"100","defaultValue":"0","type":"Int8","uomLabel":"random int8 value","formatting":"%s","labels":null}`
	valueDescriptorDataInt16 := `{"id":"2e69d813-172a-4aa3-8dd6-129e749676a2","created":1552274458525,"description":"Generate random int16 value","modified":0,"origin":0,"name":"RandomValue_Int16","min":"-100","max":"100","defaultValue":"0","type":"Int16","uomLabel":"random int16 value","formatting":"%s","labels":null}`
	valueDescriptorDataInt32 := `{"id":"1fb7e2b4-4efe-4ab4-9adc-1644789cbdc3","created":1552274458526,"description":"Generate random int32 value","modified":0,"origin":0,"name":"RandomValue_Int32","min":"-100","max":"100","defaultValue":"0","type":"Int32","uomLabel":"random int32 value","formatting":"%s","labels":null}`
	valueDescriptorDataInt64 := `{"id":"6d37c597-2399-49af-a7be-a94d7820ef94","created":1552274458526,"description":"Generate random int64 value","modified":0,"origin":0,"name":"RandomValue_Int64","min":"-100","max":"100","defaultValue":"0","type":"Int64","uomLabel":"random int64 value","formatting":"%s","labels":null}`
	valueDescriptorDataUint8 := `{"id":"553f8e3f-4e72-491b-af42-b62d87b24994","created":1552274458544,"description":"Generate random uint8 value","modified":0,"origin":0,"name":"RandomValue_Uint8","min":"0","max":"100","defaultValue":"0","type":"Uint8","uomLabel":"random uint8 value","formatting":"%s","labels":null}`
	valueDescriptorDataUint16 := `{"id":"6516a696-c91a-4c92-890f-8eeb17552447","created":1552274458547,"description":"Generate random uint16 value","modified":0,"origin":0,"name":"RandomValue_Uint16","min":"0","max":"100","defaultValue":"0","type":"Uint16","uomLabel":"random uint16 value","formatting":"%s","labels":null}`
	valueDescriptorDataUint32 := `{"id":"39eff728-c371-4385-ad1e-11cd4461785a","created":1552274458550,"description":"Generate random uint32 value","modified":0,"origin":0,"name":"RandomValue_Uint32","min":"0","max":"100","defaultValue":"0","type":"Uint32","uomLabel":"random uint32 value","formatting":"%s","labels":null}`
	valueDescriptorDataUint64 := `{"id":"c24b6bdf-ec75-475b-8c88-ff5c5b0b5663","created":1552274458551,"description":"Generate random uint64 value","modified":0,"origin":0,"name":"RandomValue_Uint64","min":"0","max":"100","defaultValue":"0","type":"Uint64","uomLabel":"random uint64 value","formatting":"%s","labels":null}`
	valueDescriptorDataFloat32 := `{"id":"26d4640e-ea08-4e9b-9411-5782edab996b","created":1551711642702,"description":"Generate random float32 value","modified":0,"origin":0,"name":"RandomValue_Float32","min":"0","max":"100","defaultValue":"3.14159","type":"Float32","uomLabel":"random float32 value","formatting":"%s","labels":null}`
	valueDescriptorDataFloat64 := `{"id":"42ca291b-389c-44f2-880f-a2033f1f887c","created":1551711642706,"description":"Generate random float64 value","modified":0,"origin":0,"name":"RandomValue_Float64","min":"0","max":"100","defaultValue":"3.14159265359","type":"Float64","uomLabel":"random float64 value","formatting":"%s","labels":null}`
	//valueDescriptorDataString := `{"id":"b7050351-148b-4a20-8468-e218c070a4e6","created":1552274458999,"description":"Random string","modified":0,"origin":0,"name":"RandomValue_String","min":"","max":"","defaultValue":"","type":"String","uomLabel":"random string value","formatting":"%s","labels":null}`
	newValueDescriptorData := `{"id":"eefafde4-16f8-4ef9-9d0a-94898aa4b513","created":1562274458123,"description":"test","modified":0,"origin":0,"name":"test","min":"","max":"","defaultValue":"","type":"String","uomLabel":"test","formatting":"%s","labels":null}`
	json.Unmarshal([]byte(valueDescriptorDataEnableRandomization), &ValueDescriptorEnableRandomization)
	json.Unmarshal([]byte(valueDescriptorDataBool), &ValueDescriptorBool)
	json.Unmarshal([]byte(valueDescriptorDataInt8), &ValueDescriptorInt8)
	json.Unmarshal([]byte(valueDescriptorDataInt16), &ValueDescriptorInt16)
	json.Unmarshal([]byte(valueDescriptorDataInt32), &ValueDescriptorInt32)
	json.Unmarshal([]byte(valueDescriptorDataInt64), &ValueDescriptorInt64)
	json.Unmarshal([]byte(valueDescriptorDataUint8), &ValueDescriptorUint8)
	json.Unmarshal([]byte(valueDescriptorDataUint16), &ValueDescriptorUint16)
	json.Unmarshal([]byte(valueDescriptorDataUint32), &ValueDescriptorUint32)
	json.Unmarshal([]byte(valueDescriptorDataUint64), &ValueDescriptorUint64)
	json.Unmarshal([]byte(valueDescriptorDataFloat32), &ValueDescriptorFloat32)
	json.Unmarshal([]byte(valueDescriptorDataFloat64), &ValueDescriptorFloat64)
	//json.Unmarshal([]byte(valueDescriptorDataString), &ValueDescriptorString)
	json.Unmarshal([]byte(valueDescriptorDataFloat64), &ValueDescriptorFloat64)
	json.Unmarshal([]byte(valueDescriptorDataInt16), &DuplicateValueDescriptorInt16)
	json.Unmarshal([]byte(newValueDescriptorData), &NewValueDescriptor)
}
