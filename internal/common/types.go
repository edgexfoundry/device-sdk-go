// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	dsModels "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// WritableInfo is a struct which contains configuration settings that can be changed in the Registry .
type WritableInfo struct {
	// Level is the logging level of writing log message
	LogLevel string
}

// ServiceInfo is a struct which contains service related configuration
// settings.
type ServiceInfo struct {
	bootstrapConfig.ServiceInfo
	// ConnectRetries is the number of times the DS will try to connect to all dependent services.
	// If exceeded for even one dependent service, the DS will exit.
	ConnectRetries int
	// Labels are...
	Labels []string
	// EnableAsyncReadings to determine whether the Device Service would deal with the asynchronous readings
	EnableAsyncReadings bool
	// AsyncBufferSize defines the size of asynchronous channel
	AsyncBufferSize int
}

type RegistryInfo struct {
	bootstrapConfig.RegistryInfo
	// Timeout specifies a timeout (in milliseconds) for
	// processing REST calls from other services.
	Timeout int
	// Health check interval
	CheckInterval string
	// Maximum number of retries
	FailLimit int
	// Time to wait until next retry
	FailWaitTime int64
}

// DeviceInfo is a struct which contains device specific configuration settings.
type DeviceInfo struct {
	// DataTransform specifies whether or not the DS perform transformations
	// specified by valuedescriptor on a actuation or query command.
	DataTransform bool
	// InitCmd specifies a device resource command which is automatically
	// generated whenever a new device is added to the DS.
	InitCmd string
	// InitCmdArgs specify arguments to be used when building the InitCmd.
	InitCmdArgs string
	// MaxCmdOps defines the maximum number of resource operations that
	// can be sent to a Driver in a single command.
	MaxCmdOps int
	// MaxCmdValueLen is the maximum string length of a command parameter or
	// result (including the valuedescriptor name) that can be returned
	// by a Driver.
	MaxCmdValueLen int
	// InitCmd specifies a device resource command which is automatically
	// generated whenever a new device is removed from the DS.
	RemoveCmd string
	// RemoveCmdArgs specify arguments to be used when building the RemoveCmd.
	RemoveCmdArgs string
	// ProfilesDir specifies a directory which contains deviceprofile
	// files which should be imported on startup.
	ProfilesDir string
	// UpdateLastConnected specifies whether to update device's LastConnected
	// timestamp in metadata.
	UpdateLastConnected bool
}

// LoggingInfo is a struct which contains logging specific configuration settings.
type LoggingInfo struct {
	bootstrapConfig.LoggingInfo
}

// Config is a struct which contains all of a DS's configuration settings.
type Config struct {
	// WritableInfo contains configuration settings that can be changed in the Registry .
	Writable WritableInfo
	// Service contains RegistryService-specific settings.
	Service ServiceInfo
	// Registry contains registry-specific settings.
	Registry RegistryInfo
	// Clients is a map of services used by a DS.
	Clients map[string]ClientInfo
	// Device contains device-specific configuration settings.
	Device DeviceInfo
	// Logging contains logging-specific configuration settings.
	Logging LoggingInfo
	// DeviceList is the list of pre-define Devices
	DeviceList []DeviceConfig `consul:"-"`
	// Driver is a string map contains customized configuration for the protocol driver implemented based on Device SDK
	Driver map[string]string
}

// DeviceConfig is the definition of Devices which will be auto created when the Device Service starts up
type DeviceConfig struct {
	// Name is the Device name
	Name string
	// Profile is the profile name of the Device
	Profile string

	Description string
	// Other labels applied to the device to help with searching
	Labels []string
	// Protocols for the device - stores protocol properties
	Protocols map[string]dsModels.ProtocolProperties
	// AutoEvent supports auto-generated events sourced from a device service
	AutoEvents []dsModels.AutoEvent
}

// ClientInfo provides the host and port of another service in the eco-system.
type ClientInfo struct {
	bootstrapConfig.ClientInfo
	// Name is the client service name
	Name string
	// Timeout specifies a timeout (in milliseconds) for
	// processing REST calls from other services.
	Timeout int
}

func (c ClientInfo) Url() string {
	url := fmt.Sprintf("%s://%s:%v", c.Protocol, c.Host, c.Port)
	return url
}

// Telemetry provides metrics (on a given device service) to system management.
type Telemetry struct {
	Alloc,
	TotalAlloc,
	Sys,
	Mallocs,
	Frees,
	LiveObjects uint64
}
