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

package internal

import (
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

const (
	BootTimeoutDefault   = time.Duration(30 * time.Second)
	ClientMonitorDefault = time.Duration(15 * time.Second)
	ConfigFileName       = "configuration.toml"
	ConfigRegistryStem   = "edgex/appservices/1.0/"
	WritableKey          = "/Writable"
	ApiTriggerRoute      = "/api/v1/trigger"
	DatabaseName         = "application-service"
)

// SDKVersion indicates the version of the SDK - will be overwritten by build
var SDKVersion string = "0.0.0"

// ApplicationVersion indicates the version of the application itself, not the SDK - will be overwritten by build
var ApplicationVersion string = "0.0.0"

// SecretsAPIRoute api route for posting secrets
var SecretsAPIRoute = clients.ApiBase + "/secrets"
