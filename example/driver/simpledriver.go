// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// Package driver provides a simple example implementation of
// ProtocolDriver interface.
package driver

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	gometrics "github.com/rcrowley/go-metrics"

	"github.com/edgexfoundry/device-sdk-go/v4/example/config"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

const readCommandsExecutedName = "ReadCommandsExecuted"

type SimpleDriver struct {
	sdk                  interfaces.DeviceServiceSDK
	lc                   logger.LoggingClient
	stopDiscovery        stopDiscoveryState
	stopProfileScan      stopProfileScanState
	asyncCh              chan<- *sdkModels.AsyncValues
	deviceCh             chan<- []sdkModels.DiscoveredDevice
	switchButton         bool
	xRotation            int32
	yRotation            int32
	zRotation            int32
	counter              interface{}
	stringArray          []string
	readCommandsExecuted gometrics.Counter
	serviceConfig        *config.ServiceConfig
}

type stopDiscoveryState struct {
	stop   bool
	locker sync.RWMutex
}

type stopProfileScanState struct {
	stop   map[string]bool
	locker sync.RWMutex
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
	_, err = img.Seek(0, 0)
	if err != nil {
		return err
	}
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
func (s *SimpleDriver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	s.sdk = sdk
	s.lc = sdk.LoggingClient()
	s.asyncCh = sdk.AsyncValuesChannel()
	s.deviceCh = sdk.DiscoveredDeviceChannel()
	s.serviceConfig = &config.ServiceConfig{}
	s.counter = map[string]interface{}{
		"f1": "ABC",
		"f2": 123,
	}
	s.stringArray = []string{"foo", "bar"}
	s.stopProfileScan = stopProfileScanState{stop: make(map[string]bool)}

	if err := sdk.LoadCustomConfig(s.serviceConfig, "SimpleCustom"); err != nil {
		return fmt.Errorf("unable to load 'SimpleCustom' custom configuration: %s", err.Error())
	}

	s.lc.Infof("Custom config is: %v", s.serviceConfig.SimpleCustom)

	if err := s.serviceConfig.SimpleCustom.Validate(); err != nil {
		return fmt.Errorf("'SimpleCustom' custom configuration validation failed: %s", err.Error())
	}

	if err := sdk.ListenForCustomConfigChanges(
		&s.serviceConfig.SimpleCustom.Writable,
		"SimpleCustom/Writable", s.ProcessCustomConfigChanges); err != nil {
		return fmt.Errorf("unable to listen for changes for 'SimpleCustom.Writable' custom configuration: %s", err.Error())
	}

	s.readCommandsExecuted = gometrics.NewCounter()

	var err error
	metricsManger := sdk.MetricsManager()
	if metricsManger != nil {
		err = metricsManger.Register(readCommandsExecutedName, s.readCommandsExecuted, nil)
	} else {
		err = errors.New("metrics manager not available")
	}

	if err != nil {
		return fmt.Errorf("unable to register metric %s: %s", readCommandsExecutedName, err.Error())
	}

	s.lc.Infof("Registered %s metric for collection when enabled", readCommandsExecutedName)

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
func (s *SimpleDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest) (res []*sdkModels.CommandValue, err error) {
	s.lc.Debugf("SimpleDriver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes)

	res = make([]*sdkModels.CommandValue, 0)
	for _, req := range reqs {
		var cv *sdkModels.CommandValue
		switch req.DeviceResourceName {
		case "SwitchButton":
			cv, _ = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeBool, s.switchButton)
		case "Xrotation":
			cv, _ = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, s.xRotation)
		case "Yrotation":
			cv, _ = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, s.yRotation)
		case "Zrotation":
			cv, _ = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, s.zRotation)
		case "Image":
			// Show a binary/image representation of the switch's on/off value
			buf := new(bytes.Buffer)
			if s.switchButton {
				err = getImageBytes(s.serviceConfig.SimpleCustom.OnImageLocation, buf)
			} else {
				err = getImageBytes(s.serviceConfig.SimpleCustom.OffImageLocation, buf)
			}
			cv, _ = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeBinary, buf.Bytes())
		case "StringArray":
			cv, _ = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeStringArray, s.stringArray)
		case "Uint8Array":
			cv, _ = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint8Array, []uint8{0, 1, 2})
		case "Counter":
			cv, _ = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, s.counter)
		}

		res = append(res, cv)
	}

	s.readCommandsExecuted.Inc(1)

	return
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *SimpleDriver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest,
	params []*sdkModels.CommandValue) error {
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
		case "StringArray":
			if s.stringArray, err = params[i].StringArrayValue(); err != nil {
				err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the data type of parameter should be string array, parameter: %s", params[i].String())
				return err
			}
		case "Uint8Array":
			v, err := params[i].Uint8ArrayValue()
			if err == nil {
				s.lc.Debugf("Uint8 array value from write command: ", v)
			} else {
				return err
			}
		case "Counter":
			if s.counter, err = params[i].ObjectValue(); err != nil {
				err := fmt.Errorf("SimpleDriver.HandleWriteCommands; the data type of parameter should be Object, parameter: %s", params[i].String())
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

func (s *SimpleDriver) Start() error {
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *SimpleDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf("a new Device is added: %s", deviceName)
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *SimpleDriver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf("Device %s is updated", deviceName)
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *SimpleDriver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	s.lc.Debugf("Device %s is removed", deviceName)
	return nil
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (s *SimpleDriver) Discover() error {
	proto := make(map[string]models.ProtocolProperties)
	proto["other"] = map[string]any{"Address": "simple02", "Port": 301}

	device2 := sdkModels.DiscoveredDevice{
		Name:        "Simple-Device02",
		Protocols:   proto,
		Description: "found by discovery",
		Labels:      []string{"auto-discovery"},
	}

	proto = make(map[string]models.ProtocolProperties)
	proto["other"] = map[string]any{"Address": "simple03", "Port": 399}

	device3 := sdkModels.DiscoveredDevice{
		Name:        "Simple-Device03",
		Protocols:   proto,
		Description: "found by discovery",
		Labels:      []string{"auto-discovery"},
	}

	res := []sdkModels.DiscoveredDevice{device2}
	time.Sleep(time.Duration(s.serviceConfig.SimpleCustom.Writable.DiscoverSleepDurationSecs) * time.Second)
	s.sdk.PublishDeviceDiscoveryProgressSystemEvent(50, len(res), "")
	defer s.setStopDeviceDiscovery(false)
	if !s.getStopDeviceDiscovery() {
		time.Sleep(time.Duration(s.serviceConfig.SimpleCustom.Writable.DiscoverSleepDurationSecs) * time.Second)
		res = append(res, device3)
	}
	s.deviceCh <- res
	return nil
}

func (s *SimpleDriver) ValidateDevice(device models.Device) error {
	protocol, ok := device.Protocols["other"]
	if !ok {
		return errors.New("missing 'other' protocols")
	}

	addr, ok := protocol["Address"]
	if !ok {
		return errors.New("missing 'Address' information")
	} else if addr == "" {
		return errors.New("address must not empty")
	}

	port, ok := protocol["Port"]
	if !ok {
		return errors.New("missing 'Port' information")
	} else {
		portString := fmt.Sprintf("%v", port)
		_, err := strconv.ParseUint(portString, 10, 64)
		if err != nil {
			return errors.New("port must be a number")
		}
	}

	return nil
}

func (s *SimpleDriver) ProfileScan(payload requests.ProfileScanRequest) (models.DeviceProfile, error) {
	time.Sleep(time.Duration(s.serviceConfig.SimpleCustom.Writable.DiscoverSleepDurationSecs) * time.Second)
	s.sdk.PublishProfileScanProgressSystemEvent(payload.RequestId, 50, "")
	if s.getStopProfileScan(payload.DeviceName) {
		return models.DeviceProfile{}, fmt.Errorf("profile scanning is stopped")
	}
	time.Sleep(time.Duration(s.serviceConfig.SimpleCustom.Writable.DiscoverSleepDurationSecs) * time.Second)
	return models.DeviceProfile{Name: payload.ProfileName}, nil
}

func (s *SimpleDriver) StopDeviceDiscovery(options map[string]any) {
	s.lc.Debugf("StopDeviceDiscovery called: options=%v", options)
	s.setStopDeviceDiscovery(true)
}

func (s *SimpleDriver) StopProfileScan(device string, options map[string]any) {
	s.lc.Debugf("StopProfileScan called: options=%v", options)
	s.setStopProfileScan(device, true)
}

func (s *SimpleDriver) getStopDeviceDiscovery() bool {
	s.stopDiscovery.locker.RLock()
	defer s.stopDiscovery.locker.RUnlock()
	return s.stopDiscovery.stop
}

func (s *SimpleDriver) setStopDeviceDiscovery(stop bool) {
	s.stopDiscovery.locker.Lock()
	defer s.stopDiscovery.locker.Unlock()
	s.stopDiscovery.stop = stop
	s.lc.Debugf("set stopDeviceDiscovery to %v", stop)
}

func (s *SimpleDriver) getStopProfileScan(device string) bool {
	s.stopProfileScan.locker.RLock()
	defer s.stopProfileScan.locker.RUnlock()
	return s.stopProfileScan.stop[device]
}

func (s *SimpleDriver) setStopProfileScan(device string, stop bool) {
	s.stopProfileScan.locker.Lock()
	defer s.stopProfileScan.locker.Unlock()
	s.stopProfileScan.stop[device] = stop
	s.lc.Debugf("set stopProfileScan to %v", stop)
}
