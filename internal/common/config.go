//
// Copyright (c) 2019 Intel Corporation
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
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/pkg/providers/vault"
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
}

// ClientInfo provides the host and port of another service in the eco-system.
type ClientInfo struct {
	// Host is the hostname or IP address of a service.
	Host string
	// Port defines the port on which to access a given service
	Port int
	// Protocol indicates the protocol to use when accessing a given service
	Protocol string
}

func (c ClientInfo) Url() string {
	url := fmt.Sprintf("%s://%s:%v", c.Protocol, c.Host, c.Port)
	return url
}

// ConfigurationStruct
// swagger:model ConfigurationStruct
type ConfigurationStruct struct {
	// Writable
	Writable WritableInfo
	// Logging
	Logging LoggingInfo
	// Registry
	Registry RegistryInfo
	// Service
	Service ServiceInfo
	// MessageBus
	MessageBus types.MessageBusConfig
	// Binding
	Binding BindingInfo
	// ApplicationSettings
	ApplicationSettings map[string]string
	// Clients
	Clients map[string]ClientInfo
	// Database
	Database db.DatabaseInfo
	// SecretStore
	SecretStore SecretStoreInfo
}

// RegistryInfo is used for defining settings for connection to the registry.
type RegistryInfo struct {
	Host string
	Port int
	Type string
}

// LoggingInfo is used to indicate whether remote logging should be used or not. If not, File designates the location of the log file to output logs to
type LoggingInfo struct {
	EnableRemote bool
	File         string
}

// ServiceInfo is used to hold and configure various settings related to the hosting of this service
type ServiceInfo struct {
	BootTimeout   string
	CheckInterval string
	ClientMonitor string
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

// SecretStoreInfo encapsulates configuration properties used to create a SecretClient.
type SecretStoreInfo struct {
	vault.SecretConfig
	// TokenFile provides a location to a token file.
	TokenFile string
}

// Credentials encapsulates username-password attributes.
type Credentials struct {
	Username string
	Password string
}
