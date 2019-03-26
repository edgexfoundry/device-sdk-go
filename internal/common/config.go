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

import "github.com/edgexfoundry/go-mod-messaging/pkg/types"

// WritableInfo ...
type WritableInfo struct {
	LogLevel string
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
