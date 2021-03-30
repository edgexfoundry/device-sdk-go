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

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db"
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
	InsecureSecrets bootstrapConfig.InsecureSecrets
}

// ConfigurationStruct
// swagger:model ConfigurationStruct
type ConfigurationStruct struct {
	// Writable contains the configuration that change be change on the fly
	Writable WritableInfo
	// Registry contains the configuration for connecting the Registry service
	Registry bootstrapConfig.RegistryInfo
	// Service contains the standard 'service' configuration for the Application service
	Service ServiceInfo
	// Trigger contains the configuration for the Function Pipeline Trigger
	Trigger TriggerInfo
	// ApplicationSettings contains the custom configuration for the Application service
	ApplicationSettings map[string]string
	// Clients contains the configuration for connecting to the dependent Edgex clients
	Clients map[string]bootstrapConfig.ClientInfo
	// Database contains the configuration for connection to the Database
	Database db.DatabaseInfo
	// SecretStore contains the configuration for connection to the Secret Store when in secure mode
	SecretStore bootstrapConfig.SecretStoreInfo
}

// ServiceInfo is used to hold and configure various settings related to the hosting of this service
type ServiceInfo struct {
	BootTimeout           string
	CheckInterval         string
	Host                  string
	HTTPSCert             string
	HTTPSKey              string
	ServerBindAddr        string
	Port                  int
	Protocol              string
	StartupMsg            string
	ReadMaxLimit          int
	Timeout               string
	ConfigAccessTokenFile string
}

// TriggerInfo contains Metadata associated with each Trigger
type TriggerInfo struct {
	// Type of trigger to start pipeline
	// enum: http, edgex-messagebus, or external-mqtt
	Type string
	// SubscribeTopics is a comma separated list of topics
	// Used when Type=edgex-messagebus, or Type=external-mqtt
	SubscribeTopics string
	// PublishTopic is the topic to publish pipeline output (if any) to
	// Used when Type=edgex-messagebus, or Type=external-mqtt
	PublishTopic string
	// Used when Type=edgex-messagebus
	EdgexMessageBus types.MessageBusConfig
	// Used when Type=external-mqtt
	ExternalMqtt ExternalMqttConfig
}

// ExternalMqttConfig contains the MQTT broker configuration for MQTT Trigger
type ExternalMqttConfig struct {
	// Url contains the fully qualified URL to connect to the MQTT broker
	Url string
	// ClientId to connect to the broker with.
	ClientId string
	// ConnectTimeout is a time duration indicating how long to wait timing out on the broker connection
	ConnectTimeout string
	// AutoReconnect indicated whether or not to retry connection if disconnected
	AutoReconnect bool
	// KeepAlive is seconds between client ping when no active data flowing to avoid client being disconnected
	KeepAlive int64
	// QoS for MQTT Connection
	QoS byte
	// Retain setting for MQTT Connection
	Retain bool
	// SkipCertVerify indicates if the certificate verification should be skipped
	SkipCertVerify bool
	// SecretPath is the name of the path in secret provider to retrieve your secrets
	SecretPath string
	// AuthMode indicates what to use when connecting to the broker. Options are "none", "cacert" , "usernamepassword", "clientcert".
	// If a CA Cert exists in the SecretPath then it will be used for all modes except "none".
	AuthMode string
}

type PipelineInfo struct {
	ExecutionOrder           string
	UseTargetTypeOfByteArray bool
	Functions                map[string]PipelineFunction
}

type PipelineFunction struct {
	// Name	string
	Parameters map[string]string
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
func (c *ConfigurationStruct) GetBootstrap() bootstrapConfig.BootstrapConfiguration {
	return bootstrapConfig.BootstrapConfiguration{
		Clients:     c.Clients,
		Service:     c.transformToBootstrapServiceInfo(),
		Registry:    c.Registry,
		SecretStore: c.SecretStore,
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

// GetInsecureSecrets returns the service's InsecureSecrets.
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return c.Writable.InsecureSecrets
}

// transformToBootstrapServiceInfo transforms the SDK's ServiceInfo to the bootstrap's version of ServiceInfo
func (c *ConfigurationStruct) transformToBootstrapServiceInfo() bootstrapConfig.ServiceInfo {
	return bootstrapConfig.ServiceInfo{
		BootTimeout:           durationToMill(c.Service.BootTimeout),
		CheckInterval:         c.Service.CheckInterval,
		Host:                  c.Service.Host,
		Port:                  c.Service.Port,
		Protocol:              c.Service.Protocol,
		StartupMsg:            c.Service.StartupMsg,
		MaxResultCount:        c.Service.ReadMaxLimit,
		Timeout:               durationToMill(c.Service.Timeout),
		ConfigAccessTokenFile: c.Service.ConfigAccessTokenFile,
	}
}

// durationToMill converts a duration string to milliseconds integer value
func durationToMill(s string) int {
	v, _ := time.ParseDuration(s)
	return int(v.Milliseconds())
}
