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
	syscontext "context"
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// Context ...
type Context struct {
	// ID of the EdgeX Event -- will be filled for a received JSON Event
	EventID string
	// Checksum of the EdgeX Event -- will be filled for a received CBOR Event
	EventChecksum string
	// This is the ID used to track the EdgeX event through entire EdgeX framework.
	CorrelationID string
	// OutputData is used for specifying the data that is to be outputted. Leverage the .Complete() function to set.
	OutputData []byte
	// This holds the configuration for your service. This is the preferred way to access your custom application settings that have been set in the configuration.
	Configuration common.ConfigurationStruct
	// This is exposed to allow logging following the preferred logging strategy within EdgeX.
	LoggingClient logger.LoggingClient
	EventClient   coredata.EventClient
}

// Complete is optional and provides a way to return the specified data.
// In the case of an HTTP Trigger, the data will be returned as the http response.
// In the case of the message bus trigger, the data will be placed on the specifed
// message bus publish topic and host in the configuration.
func (context *Context) Complete(output []byte) {
	context.OutputData = output
}

// MarkAsPushed ...
func (context *Context) MarkAsPushed() error {
	if context.EventID != "" {
		return context.EventClient.MarkPushed(context.EventID, syscontext.WithValue(syscontext.Background(), clients.CorrelationHeader, context.CorrelationID))
	} else if context.EventChecksum != "" {
		return context.EventClient.MarkPushedByChecksum(context.EventChecksum, syscontext.WithValue(syscontext.Background(), clients.CorrelationHeader, context.CorrelationID))
	} else {
		return errors.New("No EventID or EventChecksum Provided")
	}
}
