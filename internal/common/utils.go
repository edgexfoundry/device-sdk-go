// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func BuildAddr(host string, port string) string {
	var buffer bytes.Buffer

	buffer.WriteString(HttpScheme)
	buffer.WriteString(host)
	buffer.WriteString(Colon)
	buffer.WriteString(port)

	return buffer.String()
}

func CommandValueToReading(cv *dsModels.CommandValue, devName string, encoding string) *contract.Reading {
	reading := &contract.Reading{Name: cv.DeviceResourceName, Device: devName}
	if cv.Type == dsModels.Binary {
		reading.BinaryValue = cv.BinValue
	} else {
		reading.Value = cv.ValueToString(encoding)
	}

	// if value has a non-zero Origin, use it
	if cv.Origin > 0 {
		reading.Origin = cv.Origin
	} else {
		reading.Origin = time.Now().UnixNano()
	}

	return reading
}

func SendEvent(event *dsModels.Event) {
	correlation := uuid.New().String()
	ctx := context.WithValue(context.Background(), CorrelationHeader, correlation)
	if event.HasBinaryValue() {
		ctx = context.WithValue(ctx, clients.ContentType, clients.ContentTypeCBOR)
	} else {
		ctx = context.WithValue(ctx, clients.ContentType, clients.ContentTypeJSON)
	}
	// Call MarshalEvent to encode as byte array whether event contains binary or JSON readings
	var err error
	if len(event.EncodedEvent) <= 0 {
		event.EncodedEvent, err = EventClient.MarshalEvent(event.Event)
		if err != nil {
			LoggingClient.Error("SendEvent: Error encoding event", "device", event.Device, clients.CorrelationHeader, correlation, "error", err)
		} else {
			LoggingClient.Debug("SendEvent: EventClient.MarshalEvent encoded event", clients.CorrelationHeader, correlation)
		}
	} else {
		LoggingClient.Debug("SendEvent: EventClient.MarshalEvent passed through encoded event", clients.CorrelationHeader, correlation)
	}
	// Call AddBytes to post event to core data
	responseBody, errPost := EventClient.AddBytes(event.EncodedEvent, ctx)
	if errPost != nil {
		LoggingClient.Error("SendEvent Failed to push event", "device", event.Device, "response", responseBody, "error", errPost)
	} else {
		LoggingClient.Info("SendEvent: Pushed event to core data", clients.ContentType, clients.FromContext(clients.ContentType, ctx), clients.CorrelationHeader, correlation)
		LoggingClient.Trace("SendEvent: Pushed this event to core data", clients.ContentType, clients.FromContext(clients.ContentType, ctx), clients.CorrelationHeader, correlation, "event", event)
	}
}

func CompareCoreCommands(a []contract.Command, b []contract.Command) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}

	return true
}

func CompareDevices(a contract.Device, b contract.Device) bool {
	labelsOk := CompareStrings(a.Labels, b.Labels)
	profileOk := CompareDeviceProfiles(a.Profile, b.Profile)
	serviceOk := CompareDeviceServices(a.Service, b.Service)

	return reflect.DeepEqual(a.Protocols, b.Protocols) &&
		a.AdminState == b.AdminState &&
		a.Description == b.Description &&
		a.Id == b.Id &&
		a.Location == b.Location &&
		a.Name == b.Name &&
		a.OperatingState == b.OperatingState &&
		labelsOk &&
		profileOk &&
		serviceOk
}

func CompareDeviceProfiles(a contract.DeviceProfile, b contract.DeviceProfile) bool {
	labelsOk := CompareStrings(a.Labels, b.Labels)
	cmdsOk := CompareCoreCommands(a.CoreCommands, b.CoreCommands)
	devResourcesOk := CompareDeviceResources(a.DeviceResources, b.DeviceResources)
	resourcesOk := CompareDeviceCommands(a.DeviceCommands, b.DeviceCommands)

	// TODO: Objects fields aren't compared as to dr properly
	// requires introspection as Obects is a slice of interface{}

	return a.DescribedObject == b.DescribedObject &&
		a.Id == b.Id &&
		a.Name == b.Name &&
		a.Manufacturer == b.Manufacturer &&
		a.Model == b.Model &&
		labelsOk &&
		cmdsOk &&
		devResourcesOk &&
		resourcesOk
}

func CompareDeviceResources(a []contract.DeviceResource, b []contract.DeviceResource) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		// TODO: Attributes aren't compared, as to dr properly
		// requires introspection as Attributes is an interface{}

		if a[i].Description != b[i].Description ||
			a[i].Name != b[i].Name ||
			a[i].Tag != b[i].Tag ||
			a[i].Properties != b[i].Properties {
			return false
		}
	}

	return true
}

func CompareDeviceServices(a contract.DeviceService, b contract.DeviceService) bool {
	serviceOk := CompareServices(a, b)
	return a.AdminState == b.AdminState && serviceOk
}

func CompareDeviceCommands(a []contract.ProfileResource, b []contract.ProfileResource) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		getOk := CompareResourceOperations(a[i].Get, b[i].Set)
		setOk := CompareResourceOperations(a[i].Get, b[i].Set)

		if a[i].Name != b[i].Name && !getOk && !setOk {
			return false
		}
	}

	return true
}

func CompareResourceOperations(a []contract.ResourceOperation, b []contract.ResourceOperation) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		secondaryOk := CompareStrings(a[i].Secondary, b[i].Secondary)
		mappingsOk := CompareStrStrMap(a[i].Mappings, b[i].Mappings)

		if a[i].Index != b[i].Index ||
			a[i].Operation != b[i].Operation ||
			a[i].Object != b[i].Object ||
			a[i].Parameter != b[i].Parameter ||
			a[i].Resource != b[i].Resource ||
			!secondaryOk ||
			!mappingsOk {
			return false
		}
	}

	return true
}

func CompareServices(a contract.DeviceService, b contract.DeviceService) bool {
	labelsOk := CompareStrings(a.Labels, b.Labels)

	return a.DescribedObject == b.DescribedObject &&
		a.Id == b.Id &&
		a.Name == b.Name &&
		a.LastConnected == b.LastConnected &&
		a.LastReported == b.LastReported &&
		a.OperatingState == b.OperatingState &&
		a.Addressable == b.Addressable &&
		labelsOk
}

func CompareStrings(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func CompareStrStrMap(a map[string]string, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, av := range a {
		if bv, ok := b[k]; !ok || av != bv {
			return false
		}
	}

	return true
}

func VerifyIdFormat(id string, drName string) error {
	if len(id) == 0 {
		errMsg := fmt.Sprintf("The Id of %s is empty string", drName)
		LoggingClient.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	return nil
}
