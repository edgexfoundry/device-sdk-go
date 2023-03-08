//
// Copyright (C) 2022 Intel Corporation
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	sdkModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
)

// UpdatableConfig interface allows services to have custom configuration populated from configuration stored
// in the Configuration Provider (aka Consul). Services using custom configuration must implement this interface
// on their custom configuration, even if they do not use Configuration Provider. If they do not use the
// Configuration Provider they can have a dummy implementation of this interface.
// This wraps the actual interface from go-mod-bootstrap so device service code doesn't have to have the additional
// direct import of go-mod-bootstrap.
type UpdatableConfig interface {
	interfaces.UpdatableConfig
}

// DeviceServiceSDK defines the interface for an EdgeX Device Service SDK
type DeviceServiceSDK interface {
	// AddDevice adds a new Device to the Device Service and Core Metadata
	// Returns new Device id or non-nil error.
	AddDevice(device models.Device) (string, error)
	// Devices return all managed Devices from cache
	Devices() []models.Device
	// GetDeviceByName returns the Device by its name if it exists in the cache, or returns an error.
	GetDeviceByName(name string) (models.Device, error)
	// UpdateDevice updates the Device in the cache and ensures that the
	// copy in Core Metadata is also updated.
	UpdateDevice(device models.Device) error
	// RemoveDeviceByName removes the specified Device by name from the cache and ensures that the
	// instance in Core Metadata is also removed.
	RemoveDeviceByName(name string) error
	// AddDeviceProfile adds a new DeviceProfile to the Device Service and Core Metadata
	// Returns new DeviceProfile id or non-nil error.
	AddDeviceProfile(profile models.DeviceProfile) (string, error)
	// DeviceProfiles return all managed DeviceProfiles from cache
	DeviceProfiles() []models.DeviceProfile
	// GetProfileByName returns the Profile by its name if it exists in the cache, or returns an error.
	GetProfileByName(name string) (models.DeviceProfile, error)
	// UpdateDeviceProfile updates the DeviceProfile in the cache and ensures that the
	// copy in Core Metadata is also updated.
	UpdateDeviceProfile(profile models.DeviceProfile) error
	// RemoveDeviceProfileByName removes the specified DeviceProfile by name from the cache and ensures that the
	// instance in Core Metadata is also removed.
	RemoveDeviceProfileByName(name string) error
	// AddProvisionWatcher adds a new Watcher to the cache and Core Metadata
	// Returns new Watcher id or non-nil error.
	AddProvisionWatcher(watcher models.ProvisionWatcher) (string, error)
	// ProvisionWatchers return all managed Watchers from cache
	ProvisionWatchers() []models.ProvisionWatcher
	// GetProvisionWatcherByName returns the Watcher by its name if it exists in the cache, or returns an error.
	GetProvisionWatcherByName(name string) (models.ProvisionWatcher, error)
	// UpdateProvisionWatcher updates the Watcher in the cache and ensures that the
	// copy in Core Metadata is also updated.
	UpdateProvisionWatcher(watcher models.ProvisionWatcher) error
	// RemoveProvisionWatcher removes the specified Watcher by name from the cache and ensures that the
	// instance in Core Metadata is also removed.
	RemoveProvisionWatcher(name string) error
	// DeviceResource retrieves the specific DeviceResource instance from cache according to
	// the Device name and Device Resource name
	DeviceResource(deviceName string, deviceResource string) (models.DeviceResource, bool)
	// DeviceCommand retrieves the specific DeviceCommand instance from cache according to
	// the Device name and Command name
	DeviceCommand(deviceName string, commandName string) (models.DeviceCommand, bool)
	// AddDeviceAutoEvent adds a new AutoEvent to the Device with given name
	AddDeviceAutoEvent(deviceName string, event models.AutoEvent) error
	// RemoveDeviceAutoEvent removes an AutoEvent from the Device with given name
	RemoveDeviceAutoEvent(deviceName string, event models.AutoEvent) error
	// SetDeviceOpState sets the operating state of device
	SetDeviceOpState(name string, state models.OperatingState) error
	// UpdateDeviceOperatingState updates the Device's OperatingState with given name
	// in Core Metadata and device service cache.
	UpdateDeviceOperatingState(deviceName string, state string) error

	// Name returns the name of this Device Service
	Name() string

	// Version returns the version number of this Device Service
	Version() string

	// AsyncReadingsEnabled returns a bool value to indicate whether the asynchronous reading is enabled.
	AsyncReadingsEnabled() bool

	// AsyncValuesChannel returns a channel to allow developer send asynchronous reading back to SDK.
	AsyncValuesChannel() chan *sdkModels.AsyncValues

	// DiscoveredDeviceChannel returns a channel to allow developer send discovered devices back to SDK.
	DiscoveredDeviceChannel() chan []sdkModels.DiscoveredDevice

	// DeviceDiscoveryEnabled returns a bool value to indicate whether device discovery is enabled.
	DeviceDiscoveryEnabled() bool

	// DriverConfigs retrieves the driver specific configuration
	DriverConfigs() map[string]string

	// AddRoute allows leveraging the existing internal web server to add routes specific to Device Service.
	AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) error

	// LoadCustomConfig uses the Config Processor from go-mod-bootstrap to attempt to load service's
	// custom configuration. It uses the same command line flags to process the custom config in the same manner
	// as the standard configuration.
	LoadCustomConfig(customConfig UpdatableConfig, sectionName string) error

	// ListenForCustomConfigChanges uses the Config Processor from go-mod-bootstrap to attempt to listen for
	// changes to the specified custom configuration section. LoadCustomConfig must be called previously so that
	// the instance of sdkService.configProcessor has already been set.
	ListenForCustomConfigChanges(configToWatch interface{}, sectionName string, changedCallback func(interface{})) error

	// LoggingClient returns the logger.LoggingClient.
	LoggingClient() logger.LoggingClient

	// SecretProvider returns the interfaces.SecretProvider.
	SecretProvider() interfaces.SecretProvider

	// MetricsManager returns the Metrics Manager used to register counter, gauge, gaugeFloat64 or timer metric types from
	// github.com/rcrowley/go-metrics
	MetricsManager() interfaces.MetricsManager
}
