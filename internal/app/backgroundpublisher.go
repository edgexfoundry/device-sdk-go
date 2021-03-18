//
// Copyright (c) 2020 Technotects
// Copyright (c) 2021 Intel Corporation
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

package app

import (
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

type backgroundPublisher struct {
	output chan<- types.MessageEnvelope
}

// Publish provided message through the configured MessageBus output
func (pub *backgroundPublisher) Publish(payload []byte, correlationID string, contentType string) {
	outputEnvelope := types.MessageEnvelope{
		CorrelationID: correlationID,
		Payload:       payload,
		ContentType:   contentType,
	}

	pub.output <- outputEnvelope
}

func newBackgroundPublisher(capacity int) (<-chan types.MessageEnvelope, interfaces.BackgroundPublisher) {
	backgroundChannel := make(chan types.MessageEnvelope, capacity)
	return backgroundChannel, &backgroundPublisher{output: backgroundChannel}
}
