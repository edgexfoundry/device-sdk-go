//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"errors"
	"fmt"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

func CommandValuesToEventDTO(cvs []*models.CommandValue, deviceName string, sourceName string, dic *di.Container) (*dtos.Event, edgexErr.EdgeX) {
	// in some case device service driver implementation would generate no readings
	// in this case no event would be created. Based on the implementation there would be 2 scenarios:
	// 1. uninitialized *CommandValue slices, i.e. nil
	// 2. initialized *CommandValue slice with no value in it.
	if cvs == nil {
		return nil, nil
	}

	device, exist := cache.Devices().ForName(deviceName)
	if !exist {
		errMsg := fmt.Sprintf("failed to find device %s", deviceName)
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, errMsg, nil)
	}

	var transformsOK = true
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	readings := make([]dtos.BaseReading, 0, config.Device.MaxCmdOps)
	for _, cv := range cvs {
		if cv == nil {
			continue
		}
		// double check the CommandValue return from ProtocolDriver match device command
		dr, ok := cache.Profiles().DeviceResource(device.ProfileName, cv.DeviceResourceName)
		if !ok {
			msg := fmt.Sprintf("failed to find DeviceResource %s in Device %s for CommandValue (%s)", cv.DeviceResourceName, deviceName, cv.String())
			lc.Error(msg)
			return nil, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, msg, nil)
		}

		// perform data transformation
		if config.Device.DataTransform {
			err := TransformReadResult(cv, dr.Properties, lc)
			if err != nil {
				lc.Errorf("failed to transform CommandValue (%s): %v", cv.String(), err)

				var edgexErr edgexErr.EdgeX
				if errors.As(err, &OverflowError{}) {
					cv, edgexErr = models.NewCommandValue(cv.DeviceResourceName, v2.ValueTypeString, Overflow)
					if edgexErr != nil {
						return nil, edgexErr
					}
				} else if errors.As(err, &NaNError{}) {
					cv, edgexErr = models.NewCommandValue(cv.DeviceResourceName, v2.ValueTypeString, NaN)
					if edgexErr != nil {
						return nil, edgexErr
					}
				} else {
					transformsOK = false
				}
			}
		}

		// assertion
		dc := container.MetadataDeviceClientFrom(dic.Get)
		err := CheckAssertion(cv, dr.Properties.Assertion, device.Name, lc, dc)
		if err != nil {
			return nil, err
		}

		// ResourceOperation mapping
		ro, err := cache.Profiles().ResourceOperation(device.ProfileName, cv.DeviceResourceName, common.GetCmdMethod)
		if err != nil {
			// this allows SDK to directly read deviceResource without deviceCommands defined.
			lc.Debugf("failed to read ResourceOperation: %v", err)
		} else if len(ro.Mappings) > 0 {
			newCV, ok := MapCommandValue(cv, ro.Mappings)
			if ok {
				cv = newCV
			} else {
				lc.Warnf("ResourceOperation (%s) mapping value (%s) failed with the mapping table: %v", ro.DeviceResource, cv.String(), ro.Mappings)
			}
		}

		reading, edgexErr := commandValueToReading(cv, device.Name, device.ProfileName, dr.Properties.MediaType)
		if err != nil {
			return nil, edgexErr
		}
		readings = append(readings, reading)

		if cv.Type == v2.ValueTypeBinary {
			lc.Debugf("device: %s DeviceResource: %v reading: binary value", device.Name, cv.DeviceResourceName)
		} else {
			lc.Debugf("device: %s DeviceResource: %v reading: %+v", device.Name, cv.DeviceResourceName, reading)
		}
	}

	if !transformsOK {
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, fmt.Sprintf("failed to transform value for %s", deviceName), nil)
	}

	if len(readings) > 0 {
		eventDTO := dtos.NewEvent(device.ProfileName, device.Name, sourceName)
		eventDTO.Readings = readings

		return &eventDTO, nil
	} else {
		return nil, nil
	}
}

func commandValueToReading(cv *models.CommandValue, deviceName, profileName, mediaType string) (dtos.BaseReading, edgexErr.EdgeX) {
	var err edgexErr.EdgeX
	var reading dtos.BaseReading

	if cv.Type == v2.ValueTypeBinary {
		var binary []byte
		binary, err = cv.BinaryValue()
		reading = dtos.NewBinaryReading(profileName, deviceName, cv.DeviceResourceName, binary, mediaType)
	} else {
		var e error
		reading, e = dtos.NewSimpleReading(profileName, deviceName, cv.DeviceResourceName, cv.Type, cv.Value)
		if e != nil {
			err = edgexErr.NewCommonEdgeX(edgexErr.KindServerError, fmt.Sprintf("failed to transform CommandValue (%s) to Reading", cv.String()), err)
		}
	}

	return reading, err
}
