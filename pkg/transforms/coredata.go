//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package transforms

import (
	"context"
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
)

type CoreData struct {
	profileName  string
	deviceName   string
	resourceName string
	valueType    string
	mediaType    string
}

// NewCoreDataSimpleReading Is provided to interact with CoreData to add a simple reading
func NewCoreDataSimpleReading(profileName string, deviceName string, resourceName string, valueType string) *CoreData {
	coreData := &CoreData{
		profileName:  profileName,
		deviceName:   deviceName,
		resourceName: resourceName,
		valueType:    valueType,
	}
	return coreData
}

// NewCoreDataBinaryReading Is provided to interact with CoreData to add a binary reading
func NewCoreDataBinaryReading(profileName string, deviceName string, resourceName string, mediaType string) *CoreData {
	coreData := &CoreData{
		profileName:  profileName,
		deviceName:   deviceName,
		resourceName: resourceName,
		valueType:    v2.ValueTypeBinary,
		mediaType:    mediaType,
	}
	return coreData
}

// PushToCoreData pushes the provided value as an event to CoreData using the device name and reading name that have been set. If validation is turned on in
// CoreServices then your deviceName and readingName must exist in the CoreMetadata and be properly registered in EdgeX.
func (cdc *CoreData) PushToCoreData(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	ctx.LoggingClient().Info("Pushing To CoreData...")

	if data == nil {
		return false, errors.New("PushToCoreData - No Data Received")
	}

	client := ctx.EventClient()
	if client == nil {
		return false, errors.New("EventClient not initialized. Core Data is missing from clients configuration")
	}

	event := dtos.NewEvent(cdc.profileName, cdc.deviceName, cdc.resourceName)
	if cdc.valueType == v2.ValueTypeBinary {
		reading, err := util.CoerceType(data)
		if err != nil {
			return false, err
		}
		event.AddBinaryReading(cdc.resourceName, reading, cdc.mediaType)
	} else if cdc.valueType == v2.ValueTypeString {
		reading, err := util.CoerceType(data)
		if err != nil {
			return false, err
		}
		event.AddSimpleReading(cdc.resourceName, cdc.valueType, string(reading))
	} else {
		event.AddSimpleReading(cdc.resourceName, cdc.valueType, data)
	}

	request := requests.NewAddEventRequest(event)
	result, err := client.Add(context.Background(), request)
	if err != nil {
		return false, err
	}

	return true, result
}
