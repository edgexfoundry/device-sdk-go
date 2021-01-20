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

package internal

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	v2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
)

const (
	ConfigRegistryStem   = "edgex/appservices/1.0/"
	DatabaseName         = "application-service"
	CorrelationHeaderKey = "X-Correlation-ID"

	ApiTriggerRoute     = clients.ApiBase + "/trigger"
	ApiV2TriggerRoute   = v2.ApiBase + "/trigger"
	ApiSecretsRoute     = clients.ApiBase + "/secrets"
	ApiV2AddSecretRoute = v2.ApiBase + "/secret"
)

// SDKVersion indicates the version of the SDK - will be overwritten by build
var SDKVersion string = "0.0.0"

// ApplicationVersion indicates the version of the application itself, not the SDK - will be overwritten by build
var ApplicationVersion string = "0.0.0"
