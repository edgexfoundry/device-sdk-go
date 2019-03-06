// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
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

func CommandValueToReading(cv *ds_models.CommandValue, devName string) *models.Reading {
	reading := &models.Reading{Name: cv.RO.Parameter, Device: devName}
	reading.Value = cv.ValueToString()

	// if value has a non-zero Origin, use it
	if cv.Origin > 0 {
		reading.Origin = cv.Origin
	} else {
		reading.Origin = time.Now().UnixNano() / int64(time.Millisecond)
	}

	return reading
}

func SendEvent(event *models.Event) {
	ctx := context.WithValue(context.Background(), CorrelationHeader, uuid.New().String())
	_, err := EventClient.Add(event, ctx)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to push event for device %s: %v", event.Device, err))
	}
}

func CompareCommands(a []models.Command, b []models.Command) bool {
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

func CompareDevices(a models.Device, b models.Device) bool {
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

func CompareDeviceProfiles(a models.DeviceProfile, b models.DeviceProfile) bool {
	labelsOk := CompareStrings(a.Labels, b.Labels)
	cmdsOk := CompareCommands(a.Commands, b.Commands)
	devResourcesOk := CompareDeviceResources(a.DeviceResources, b.DeviceResources)
	resourcesOk := CompareResources(a.Resources, b.Resources)

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

func CompareDeviceResources(a []models.DeviceResource, b []models.DeviceResource) bool {
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

func CompareDeviceServices(a models.DeviceService, b models.DeviceService) bool {
	serviceOk := CompareServices(a.Service, b.Service)
	return a.AdminState == b.AdminState && serviceOk
}

func CompareResources(a []models.ProfileResource, b []models.ProfileResource) bool {
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

func CompareResourceOperations(a []models.ResourceOperation, b []models.ResourceOperation) bool {
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

func CompareServices(a models.Service, b models.Service) bool {
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

func MakeAddressable(name string, addr *models.Addressable) (*models.Addressable, error) {
	// check whether there has been an existing addressable
	ctx := context.WithValue(context.Background(), CorrelationHeader, uuid.New().String())
	addressable, err := AddressableClient.AddressableForName(name, ctx)
	if err != nil {
		if errsc, ok := err.(*types.ErrServiceClient); ok && (errsc.StatusCode == http.StatusNotFound) {
			LoggingClient.Debug(fmt.Sprintf("Addressable %s doesn't exist, creating a new one", addr.Name))
			millis := time.Now().UnixNano() / int64(time.Millisecond)
			addressable = *addr
			addressable.Name = name
			addressable.Origin = millis
			LoggingClient.Debug(fmt.Sprintf("Adding Addressable: %v", addressable))
			id, err := AddressableClient.Add(&addressable, ctx)
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("Add Addressable failed %v, error: %v", addr, err))
				return nil, err
			}
			if err = VerifyIdFormat(id, "Addressable"); err != nil {
				return nil, err
			}
			addressable.Id = id
		} else {
			LoggingClient.Error(fmt.Sprintf("AddressableForName failed: %v", err))
			return nil, err
		}
	} else {
		LoggingClient.Debug(fmt.Sprintf("Addressable %s exists, using the existing one", addressable.Name))
	}

	return &addressable, nil
}

func VerifyIdFormat(id string, drName string) error {
	if len(id) == 0 {
		errMsg := fmt.Sprintf("The Id of %s is empty string", drName)
		LoggingClient.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	return nil
}
