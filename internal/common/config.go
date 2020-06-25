//
// Copyright (c) 2020 Intel Corporation
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

package common

import (
	"time"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
)

// WritableInfo is used to hold configuration information that is considered "live" or can be changed on the fly without a restart of the service.
type WritableInfo struct {
	// Set level of logging to report
	//
	// example: TRACE
	// required: true
	// enum: TRACE,DEBUG,INFO,WARN,ERROR
	LogLevel        string
	Pipeline        PipelineInfo
	StoreAndForward StoreAndForwardInfo
	InsecureSecrets InsecureSecrets
}

// ConfigurationStruct
// swagger:model ConfigurationStruct
type ConfigurationStruct struct {
	// Writable
	Writable WritableInfo
	// Logging
	Logging bootstrapConfig.LoggingInfo
	// Registry
	Registry bootstrapConfig.RegistryInfo
	// Service
	Service ServiceInfo
	// MessageBus
	MessageBus types.MessageBusConfig
	// Binding
	Binding BindingInfo
	// ApplicationSettings
	ApplicationSettings map[string]string
	// Clients
	Clients map[string]bootstrapConfig.ClientInfo
	// Database
	Database db.DatabaseInfo
	// SecretStore
	SecretStore bootstrapConfig.SecretStoreInfo
	// SecretStoreExclusive
	SecretStoreExclusive bootstrapConfig.SecretStoreInfo
	// Startup
	Startup bootstrapConfig.StartupInfo
}

// ServiceInfo is used to hold and configure various settings related to the hosting of this service
type ServiceInfo struct {
	BootTimeout   string
	CheckInterval string
	Host          string
	HTTPSCert     string
	HTTPSKey      string
	Port          int
	Protocol      string
	StartupMsg    string
	ReadMaxLimit  int
	Timeout       string
}

// BindingInfo contains Metadata associated with each binding
type BindingInfo struct {
	// Type of trigger to start pipeline
	//
	// example: messagebus
	// required: true
	// enum: messagebus,http
	Type           string
	SubscribeTopic string
	PublishTopic   string
}

type PipelineInfo struct {
	ExecutionOrder           string
	UseTargetTypeOfByteArray bool
	Functions                map[string]PipelineFunction
}

type PipelineFunction struct {
	// Name	string
	Parameters  map[string]string
	Addressable models.Addressable
}

type StoreAndForwardInfo struct {
	Enabled       bool
	RetryInterval string
	MaxRetryCount int
}

// Credentials encapsulates username-password attributes.
type Credentials struct {
	Username string
	Password string
}

// InsecureSecrets is used to hold the secrets stored in the configuration
type InsecureSecrets map[string]InsecureSecretsInfo

// InsecureSecretsInfo encapsulates info used to retrieve insecure secrets
type InsecureSecretsInfo struct {
	Path    string
	Secrets map[string]string
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct which is
// then used to overwrite the service's existing configuration struct.
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ConfigurationStruct)
	if ok {
		// Check that information was successfully read from Registry
		if configuration.Service.Port == 0 {
			return false
		}
		*c = *configuration
	}
	return ok
}

// EmptyWritablePtr returns a pointer to an empty WritableInfo struct.  It is used by the bootstrap to
// provide the appropriate structure for Config Client's WatchForChanges().
func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return &WritableInfo{}
}

// UpdateWritableFromRaw updates the Writeable section of configuration from raw update received from Configuration Provider.
func (c *ConfigurationStruct) UpdateWritableFromRaw(rawWritable interface{}) bool {
	writable, ok := rawWritable.(*WritableInfo)
	if ok {
		c.Writable = *writable
	}
	return ok
}

// GetBootstrap returns the configuration elements required by the bootstrap.
func (c *ConfigurationStruct) GetBootstrap() interfaces.BootstrapConfiguration {
	return interfaces.BootstrapConfiguration{
		Clients:     c.Clients,
		Service:     c.transformToBootstrapServiceInfo(),
		Registry:    c.Registry,
		Logging:     c.Logging,
		SecretStore: c.SecretStore,
		Startup:     c.Startup,
	}
}

// GetLogLevel returns log level from the configuration
func (c *ConfigurationStruct) GetLogLevel() string {
	return c.Writable.LogLevel
}

// GetRegistryInfo returns the RegistryInfo section from the configuration
func (c *ConfigurationStruct) GetRegistryInfo() bootstrapConfig.RegistryInfo {
	return c.Registry
}

// transformToBootstrapServiceInfo transforms the SDK's ServiceInfo to the bootstrap's version of ServiceInfo
func (c *ConfigurationStruct) transformToBootstrapServiceInfo() bootstrapConfig.ServiceInfo {
	return bootstrapConfig.ServiceInfo{
		BootTimeout:    durationToMill(c.Service.BootTimeout),
		CheckInterval:  c.Service.CheckInterval,
		Host:           c.Service.Host,
		Port:           c.Service.Port,
		Protocol:       c.Service.Protocol,
		StartupMsg:     c.Service.StartupMsg,
		MaxResultCount: c.Service.ReadMaxLimit,
		Timeout:        durationToMill(c.Service.Timeout),
	}
}

// durationToMill converts a duration string to milliseconds integer value
func durationToMill(s string) int {
	v, _ := time.ParseDuration(s)
	return int(v.Milliseconds())
}
