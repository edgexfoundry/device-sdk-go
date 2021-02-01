// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// WritableInfo is a struct which contains configuration settings that can be changed in the Registry .
type WritableInfo struct {
	// Level is the logging level of writing log message
	LogLevel        string
	InsecureSecrets bootstrapConfig.InsecureSecrets
}

// ServiceInfo is a struct which contains service related configuration
// settings.
type ServiceInfo struct {
	// BootTimeout indicates, in milliseconds, how long the service will retry connecting to upstream dependencies
	// before giving up. Default is 30,000.
	BootTimeout int
	// Health check interval
	CheckInterval string
	// Host is the hostname or IP address of the service.
	Host string
	// Port is the HTTP port of the service.
	Port int
	// ServerBindAddr specifies an IP address or hostname
	// for ListenAndServe to bind to, such as 0.0.0.0
	ServerBindAddr string
	// The protocol that should be used to call this service
	Protocol string
	// StartupMsg specifies a string to log once service
	// initialization and startup is completed.
	StartupMsg string
	// MaxResultCount specifies the maximum size list supported
	// in response to REST calls to other services.
	MaxResultCount int
	// Timeout (in milliseconds) specifies both
	// - timeout for processing REST calls and
	// - interval time the DS will wait between each retry call.
	Timeout int
	// Labels are properties applied to the device service to help with searching
	Labels []string
	// EnableAsyncReadings to determine whether the Device Service would deal with the asynchronous readings
	EnableAsyncReadings bool
	// AsyncBufferSize defines the size of asynchronous channel
	AsyncBufferSize int
}

// DeviceInfo is a struct which contains device specific configuration settings.
type DeviceInfo struct {
	// DataTransform specifies whether or not the DS perform transformations
	// specified by value descriptor on a actuation or query command.
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
	// result (including the value descriptor name) that can be returned
	// by a Driver.
	MaxCmdValueLen int
	// InitCmd specifies a device resource command which is automatically
	// generated whenever a new device is removed from the DS.
	RemoveCmd string
	// RemoveCmdArgs specify arguments to be used when building the RemoveCmd.
	RemoveCmdArgs string
	// ProfilesDir specifies a directory which contains device profiles
	// files which should be imported on startup.
	ProfilesDir string
	// UpdateLastConnected specifies whether to update device's LastConnected
	// timestamp in metadata.
	UpdateLastConnected bool

	Discovery DiscoveryInfo
}

// DiscoveryInfo is a struct which contains configuration of device auto discovery.
type DiscoveryInfo struct {
	// Enabled controls whether or not device discovery is enabled.
	Enabled bool
	// Interval indicates how often the discovery process will be triggered.
	// It represents as a duration string.
	Interval string
}

// DeviceConfig is the definition of Devices which will be auto created when the Device Service starts up
type DeviceConfig struct {
	// Name is the Device name
	Name string
	// Profile is the profile name of the Device
	Profile string
	// Description describes the device
	Description string
	// Other labels applied to the device to help with searching
	Labels []string
	// Protocols for the device - stores protocol properties
	Protocols map[string]models.ProtocolProperties
	// AutoEvent supports auto-generated events sourced from a device service
	AutoEvents []models.AutoEvent
}

func (s ServiceInfo) GetBootstrapServiceInfo() bootstrapConfig.ServiceInfo {
	return bootstrapConfig.ServiceInfo{
		BootTimeout:    s.BootTimeout,
		CheckInterval:  s.CheckInterval,
		Host:           s.Host,
		Port:           s.Port,
		ServerBindAddr: s.ServerBindAddr,
		Protocol:       s.Protocol,
		StartupMsg:     s.StartupMsg,
		MaxResultCount: s.MaxResultCount,
		Timeout:        s.Timeout,
	}
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
