// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package device

import (
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	ClientData     = "Data"
	ClientMetadata = "Metadata"
)

// ServiceInfo is a struct which contains service related configuration
// settings.
type ServiceInfo struct {
	// Host is the hostname or IP address of the service.
	Host string
	// Port is the HTTP port of the service.
	Port int
	// ConnectRetries is the number of times the DS will try
	// to connect to Core Metadata to either register itself
	// and provision it's objects, or use existing data. If
	// this limit is exceeded, the DS will exit.
	ConnectRetries int
	// HealthCheck is a URL specifying a healthcheck REST
	// endpoint used by the Registry to determine if the
	// service is available.
	HealthCheck string
	// Labels are...
	Labels []string
	// OpenMsg specifies a string logged on DS startup.
	OpenMsg string
	// ReadMaxLimit specifies the maximum size list supported
	// in response to REST calls to other services.
	ReadMaxLimit int
	// Timeout specifies a timeout (in milliseconds) for
	// processing REST calls from other services.
	Timeout int
}

type service struct {
	// Name is service's register name
	Name string
	// Host is the hostname or IP address of a service.
	Host string
	// Port is the HTTP port of a service.
	Port int
	// Timeout specifies a timeout (in milliseconds) for
	// processing REST calls from other services.
	Timeout int
	// Health check url
	CheckPath string
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
	// can be sent to a ProtocolDriver in a single command.
	MaxCmdOps int
	// MaxCmdValueLen is the maximum string length of a command parameter or
	// result (including the valuedescriptor name) that can be returned
	// by a ProtocolDriver.
	MaxCmdValueLen int
	// InitCmd specifies a device resource command which is automatically
	// generated whenever a new device is removed from the DS.
	RemoveCmd string
	// RemoveCmdArgs specify arguments to be used when building the RemoveCmd.
	RemoveCmdArgs string
	// ProfilesDir specifies a directory which contains deviceprofile
	// files which should be imported on startup.
	ProfilesDir string
	// SendReaingsOnChanged can be used to cause a DS to only send readings
	// to Core Data when the reading has changed (based on comparison to an
	// existing reading in the cache, if present).
	SendReadingsOnChanged bool
}

// LoggingInfo is a struct which contains logging specific configuration settings.
type LoggingInfo struct {
	// File is the pathname of a local log file to be created.
	File string
	// RemoteURL is the URL of the support logging service.
	// TODO: make this just another client!
	RemoteURL string
}

// ScheduleEventInfo is a struct which contains event schedule specific
// configuration settings.
type ScheduleEventInfo struct {
	// Schedule is the name of the associated schedule.
	Schedule string
	// Path is the endpoint of the DS to be called when the
	// ScheduleEvent is triggered.
	Path string
	// Service is the DS service name.
	Service string
}

// WatcherInfo is a struct which contains provisionwatcher configuration settings.
type WatcherInfo struct {
	Profile     string
	Key         string
	MatchString string
}

// Config is a struct which contains all of a DS's configuration settings.
type Config struct {
	// Service contains service-specific settings.
	Service ServiceInfo
	// Registry contains registry-specific settings.
	Registry service
	// Clients is a map of services used by a DS.
	Clients map[string]service
	// Device contains device-specific coniguration settings.
	Device DeviceInfo
	// Logging contains logging-specific configuration settings.
	Logging LoggingInfo
	// Schedules is created on startup.
	Schedules []models.Schedule
	// SchedulesEvents is created on startup.
	ScheduleEvents []models.ScheduleEvent
	// Watchers is a map provisionwatchers to be created on startup.
	Watchers map[string]WatcherInfo
}
