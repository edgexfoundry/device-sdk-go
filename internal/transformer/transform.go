//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"errors"
	"fmt"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/v2/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

func CommandValuesToEventDTO(cvs []*models.CommandValue, deviceName string, dic *di.Container) (eventDTO dtos.Event, e edgexErr.EdgeX) {
	device, exist := cache.Devices().ForName(deviceName)
	if !exist {
		errMsg := fmt.Sprintf("failed to find device %s", deviceName)
		return eventDTO, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, errMsg, nil)
	}

	var transformsOK = true
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	readings := make([]dtos.BaseReading, 0, config.Device.MaxCmdOps)
	for _, cv := range cvs {
		// double check the CommandValue return from ProtocolDriver match device command
		dr, ok := cache.Profiles().DeviceResource(device.ProfileName, cv.DeviceResourceName)
		if !ok {
			msg := fmt.Sprintf("failed to find DeviceResource %s in Device %s for CommandValue (%s)", cv.DeviceResourceName, deviceName, cv.String())
			lc.Error(msg)
			return eventDTO, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, msg, nil)
		}

		// perform data transformation
		if config.Device.DataTransform {
			err := TransformReadResult(cv, dr.Properties, lc)
			if err != nil {
				lc.Errorf("failed to transform CommandValue (%s): %v", cv.String(), err)

				if errors.As(err, &OverflowError{}) {
					cv = models.NewStringValue(cv.DeviceResourceName, cv.Origin, Overflow)
				} else if errors.As(err, &NaNError{}) {
					cv = models.NewStringValue(cv.DeviceResourceName, cv.Origin, NaN)
				} else {
					transformsOK = false
				}
			}
		}

		// assertion
		dc := container.MetadataDeviceClientFrom(dic.Get)
		err := CheckAssertion(cv, dr.Properties.Assertion, device.Name, lc, dc)
		if err != nil {
			cv = models.NewStringValue(cv.DeviceResourceName, cv.Origin, fmt.Sprintf("Assertion failed for device resource: %s, with value: %s", cv.DeviceResourceName, cv.String()))
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

		reading := commandValueToReading(cv, device.Name, device.ProfileName, dr.Properties.MediaType, "")
		readings = append(readings, reading)

		if cv.Type == v2.ValueTypeBinary {
			lc.Debugf("device: %s DeviceResource: %v reading: binary value", device.Name, cv.DeviceResourceName)
		} else {
			lc.Debugf("device: %s DeviceResource: %v reading: %+v", device.Name, cv.DeviceResourceName, reading)
		}
	}

	if !transformsOK {
		return eventDTO, edgexErr.NewCommonEdgeX(edgexErr.KindServerError, fmt.Sprintf("failed to transform value for %s", deviceName), nil)
	}

	eventDTO = dtos.NewEvent(device.ProfileName, device.Name)
	eventDTO.Readings = readings

	return
}

func commandValueToReading(cv *models.CommandValue, deviceName, profileName, mediaType, encoding string) dtos.BaseReading {
	if encoding == "" {
		encoding = models.DefaultFloatEncoding
	}

	reading := dtos.BaseReading{
		Versionable: commonDTO.Versionable{
			ApiVersion: v2.ApiVersion,
		},
		Id:           uuid.NewString(),
		DeviceName:   deviceName,
		ProfileName:  profileName,
		ResourceName: cv.DeviceResourceName,
		ValueType:    cv.Type,
	}
	if cv.Type == v2.ValueTypeBinary {
		reading.BinaryValue = cv.BinValue
		reading.MediaType = mediaType
	} else {
		reading.Value = cv.ValueToString(encoding)
	}

	// if value has a non-zero Origin, use it
	if cv.Origin > 0 {
		reading.Origin = cv.Origin
	} else {
		reading.Origin = time.Now().UnixNano() / int64(time.Millisecond)
	}

	return reading
}
