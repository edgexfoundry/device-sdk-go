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

package appcontext

import (
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// Context ...
type Context struct {
	CorrelationID string
	OutputData    []byte
	Trigger       trigger.Trigger
	Configuration common.ConfigurationStruct
	LoggingClient logger.LoggingClient
}

// Complete called when ready to send output and function is finished
func (context *Context) Complete(output []byte) {
	context.OutputData = output
}
