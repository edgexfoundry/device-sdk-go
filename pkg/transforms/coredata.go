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
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
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
		valueType:    common.ValueTypeBinary,
		mediaType:    mediaType,
	}
	return coreData
}

// NewCoreDataObjectReading Is provided to interact with CoreData to add a object reading type
func NewCoreDataObjectReading(profileName string, deviceName string, resourceName string) *CoreData {
	coreData := &CoreData{
		profileName:  profileName,
		deviceName:   deviceName,
		resourceName: resourceName,
		valueType:    common.ValueTypeObject,
	}
	return coreData
}

// PushToCoreData pushes the provided value as an event to CoreData using the device name and reading name that have been set. If validation is turned on in
// CoreServices then your deviceName and readingName must exist in the CoreMetadata and be properly registered in EdgeX.
func (cdc *CoreData) PushToCoreData(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	ctx.LoggingClient().Info("Pushing To CoreData...")

	if data == nil {
		return false, fmt.Errorf("function PushToCoreData in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	client := ctx.EventClient()
	if client == nil {
		return false, fmt.Errorf("function PushToCoreData in pipeline '%s': EventClient not initialized. Core Data is missing from clients configuration", ctx.PipelineId())
	}

	event := dtos.NewEvent(cdc.profileName, cdc.deviceName, cdc.resourceName)

	switch cdc.valueType {
	case common.ValueTypeBinary:
		reading, err := util.CoerceType(data)
		if err != nil {
			return false, err
		}
		event.AddBinaryReading(cdc.resourceName, reading, cdc.mediaType)

	case common.ValueTypeString:
		reading, err := util.CoerceType(data)
		if err != nil {
			return false, err
		}
		err = event.AddSimpleReading(cdc.resourceName, cdc.valueType, string(reading))
		if err != nil {
			return false, fmt.Errorf("error adding Reading in pipeline '%s': %s", ctx.PipelineId(), err.Error())
		}

	case common.ValueTypeObject:
		event.AddObjectReading(cdc.resourceName, data)

	default:
		err := event.AddSimpleReading(cdc.resourceName, cdc.valueType, data)
		if err != nil {
			return false, fmt.Errorf("error adding Reading in pipeline '%s': %s", ctx.PipelineId(), err.Error())
		}

	}

	request := requests.NewAddEventRequest(event)
	result, err := client.Add(context.Background(), request)
	if err != nil {
		return false, err
	}

	return true, result
}
