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
)

// WritableInfo ...
type WritableInfo struct {
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

// ConfigurationStruct ...
type ConfigurationStruct struct {
	Writable            WritableInfo
	Logging             LoggingInfo
	Registry            RegistryInfo
	Service             ServiceInfo
	MessageBus          types.MessageBusConfig
	Binding             BindingInfo
	ApplicationSettings map[string]string
	Clients             map[string]ClientInfo
	Database            db.DatabaseInfo
}

// RegistryInfo ...
type RegistryInfo struct {
	Host string
	Port int
	Type string
}

// LoggingInfo ...
type LoggingInfo struct {
	EnableRemote bool
	File         string
}

// ServiceInfo ...
type ServiceInfo struct {
	BootTimeout   int
	CheckInterval string
	ClientMonitor int
	Host          string
	Port          int
	Protocol      string
	StartupMsg    string
	ReadMaxLimit  int
	Timeout       int
}

// BindingInfo contains Metadata associated with each binding
type BindingInfo struct {
	Type           string
	Name           string
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
	RetryInterval int
	MaxRetryCount int
}
