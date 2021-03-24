// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a simple example implementation of
// ProtocolDriver interface.
//
package driver

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"reflect"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/example/config"
	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
)

type SimpleDriver struct {
	lc            logger.LoggingClient
	asyncCh       chan<- *dsModels.AsyncValues
	deviceCh      chan<- []dsModels.DiscoveredDevice
	switchButton  bool
	xRotation     int32
	yRotation     int32
	zRotation     int32
	serviceConfig *config.ServiceConfig
}

func getImageBytes(imgFile string, buf *bytes.Buffer) error {
	// Read existing image from file
	img, err := os.Open(imgFile)
	if err != nil {
		return err
	}
	defer img.Close()

	// TODO: Attach MediaType property, determine if decoding
	//  early is required (to optimize edge processing)

	// Expect "png" or "jpeg" image type
	imageData, imageType, err := image.Decode(img)
	if err != nil {
		return err
	}
	// Finished with file. Reset file pointer
	img.Seek(0, 0)
	if imageType == "jpeg" {
		err = jpeg.Encode(buf, imageData, nil)
		if err != nil {
			return err
		}
	} else if imageType == "png" {
		err = png.Encode(buf, imageData)
		if err != nil {
			return err
		}
	}
	return nil
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *SimpleDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceCh chan<- []dsModels.DiscoveredDevice) error {
	s.lc = lc
	s.asyncCh = asyncCh
	s.deviceCh = deviceCh
	s.serviceConfig = &config.ServiceConfig{}

	ds := service.RunningService()

	if err := ds.LoadCustomConfig(s.serviceConfig, "SimpleCustom"); err != nil {
		return fmt.Errorf("unable to load 'SimpleCustom' custom configuration: %s", err.Error())
	}

	lc.Infof("Custom config is: %v", s.serviceConfig.SimpleCustom)

	if err := s.serviceConfig.SimpleCustom.Validate(); err != nil {
		return fmt.Errorf("'SimpleCustom' custom configuration validation failed: %s", err.Error())
	}

	if err := ds.ListenForCustomConfigChanges(
		&s.serviceConfig.SimpleCustom.Writable,
		"SimpleCustom/Writable", s.ProcessCustomConfigChanges); err != nil {
		return fmt.Errorf("unable to listen for changes for 'SimpleCustom.Writable' custom configuration: %s", err.Error())
	}

	return nil
}

// ProcessCustomConfigChanges ...
func (s *SimpleDriver) ProcessCustomConfigChanges(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*config.SimpleWritable)
	if !ok {
		s.lc.Error("unable to process custom config updates: Can not cast raw config to type 'SimpleWritable'")
		return
	}

	s.lc.Info("Received configuration updates for 'SimpleCustom.Writable' section")

	previous := s.serviceConfig.SimpleCustom.Writable
	s.serviceConfig.SimpleCustom.Writable = *updated

	if reflect.DeepEqual(previous, *updated) {
		s.lc.Info("No changes detected")
		return
	}

	// Now check to determine what changed.
	// In this example we only have the one writable setting,
	// so the check is not really need but left here as an example.
	// Since this setting is pulled from configuration each time it is need, no extra processing is required.
	// This may not be true for all settings, such as external host connection info, which
	// may require re-establishing the connection to the external host for example.
	if previous.DiscoverSleepDurationSecs != updated.DiscoverSleepDurationSecs {
		s.lc.Infof("DiscoverSleepDurationSecs changed to: %d", updated.DiscoverSleepDurationSecs)
	}
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *SimpleDriver) HandleReadCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	s.lc.Debugf("SimpleDriver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes)

	if len(reqs) == 1 {
		res = make([]*dsModels.CommandValue, 1)
		if reqs[0].DeviceResourceName == "SwitchButton" {
			cv, _ := dsModels.NewCommandValue(reqs[0].DeviceResourceName, v2.ValueTypeBool, s.switchButton)
			res[0] = cv
		} else if reqs[0].DeviceResourceName == "Xrotation" {
			cv, _ := dsModels.NewCommandValue(reqs[0].DeviceResourceName, v2.ValueTypeInt32, s.xRotation)
			res[0] = cv
		} else if reqs[0].DeviceResourceName == "Yrotation" {
			cv, _ := dsModels.NewCommandValue(reqs[0].DeviceResourceName, v2.ValueTypeInt32, s.yRotation)
			res[0] = cv
		} else if reqs[0].DeviceResourceName == "Zrotation" {
			cv, _ := dsModels.NewCommandValue(reqs[0].DeviceResourceName, v2.ValueTypeInt32, s.zRotation)
			res[0] = cv
		} else if reqs[0].DeviceResourceName == "Image" {
			// Show a binary/image representation of the switch's on/off value
			buf := new(bytes.Buffer)
			if s.switchButton == true {
				err = getImageBytes(s.serviceConfig.SimpleCustom.OnImageLocation, buf)
			} else {
				err = getImageBytes(s.serviceConfig.SimpleCustom.OffImageLocation, buf)
			}
			cvb, _ := dsModels.NewCommandValue(reqs[0].DeviceResourceName, v2.ValueTypeBinary, buf.Bytes())
			res[0] = cvb
		} else if reqs[0].DeviceResourceName == "Uint8Array" {
			cv, _ := dsModels.NewCommandValue(reqs[0].DeviceResourceName, v2.ValueTypeUint8Array, []uint8{0, 1, 2})
			res[0] = cv
		}
	} else if len(reqs) == 3 {
		res = make([]*dsModels.CommandValue, 3)
		for i, r := range reqs {
			var cv *dsModels.CommandValue
			switch r.DeviceResourceName {
			case "Xrotation":
				cv, _ = dsModels.NewCommandValue(r.DeviceResourceName, v2.ValueTypeInt32, s.xRotation)
			case "Yrotation":
				cv, _ = dsModels.NewCommandValue(r.DeviceResourceName, v2.ValueTypeInt32, s.yRotation)
			case "Zrotation":
				cv, _ = dsModels.NewCommandValue(r.DeviceResourceName, v2.ValueTypeInt32, s.zRotation)
			}
			res[i] = cv
		}
	}

	return
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *SimpleDriver) HandleWriteCommands(deviceName string, protocols map[string]contract.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {
	var err error

	for i, r := range reqs {
		s.lc.Debugf("SimpleDriver.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v, attributes: %v", protocols, reqs[i].DeviceResourceName, params[i], reqs[i].Attributes)
		switch r.DeviceResourceName {
		case "SwitchButton":
			if s.switchButton, err = params[i].BoolValue(); err != nil {
				err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the data type of parameter should be Boolean, parameter: %s", params[0].String())
				return err
			}
		case "Xrotation":
			if s.xRotation, err = params[i].Int32Value(); err != nil {
				err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the data type of parameter should be Int32, parameter: %s", params[i].String())
				return err
			}
		case "Yrotation":
			if s.yRotation, err = params[i].Int32Value(); err != nil {
				err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the data type of parameter should be Int32, parameter: %s", params[i].String())
				return err
			}
		case "Zrotation":
			if s.zRotation, err = params[i].Int32Value(); err != nil {
				err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the data type of parameter should be Int32, parameter: %s", params[i].String())
				return err
			}
		case "Uint8Array":
			v, err := params[i].Uint8ArrayValue()
			if err == nil {
				s.lc.Debugf("Uint8 array value from write command: ", v)
			} else {
				return err
			}
		}
	}

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *SimpleDriver) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if s.lc != nil {
		s.lc.Debugf("SimpleDriver.Stop called: force=%v", force)
	}
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *SimpleDriver) AddDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	s.lc.Debugf("a new Device is added: %s", deviceName)
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *SimpleDriver) UpdateDevice(deviceName string, protocols map[string]contract.ProtocolProperties, adminState contract.AdminState) error {
	s.lc.Debugf("Device %s is updated", deviceName)
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *SimpleDriver) RemoveDevice(deviceName string, protocols map[string]contract.ProtocolProperties) error {
	s.lc.Debugf("Device %s is removed", deviceName)
	return nil
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (s *SimpleDriver) Discover() {
	proto := make(map[string]contract.ProtocolProperties)
	proto["other"] = map[string]string{"Address": "simple02", "Port": "301"}

	device2 := dsModels.DiscoveredDevice{
		Name:        "Simple-Device02",
		Protocols:   proto,
		Description: "found by discovery",
		Labels:      []string{"auto-discovery"},
	}

	proto = make(map[string]contract.ProtocolProperties)
	proto["other"] = map[string]string{"Address": "simple03", "Port": "399"}

	device3 := dsModels.DiscoveredDevice{
		Name:        "Simple-Device03",
		Protocols:   proto,
		Description: "found by discovery",
		Labels:      []string{"auto-discovery"},
	}

	res := []dsModels.DiscoveredDevice{device2, device3}

	time.Sleep(time.Duration(s.serviceConfig.SimpleCustom.Writable.DiscoverSleepDurationSecs) * time.Second)
	s.deviceCh <- res
}
