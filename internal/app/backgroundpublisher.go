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
	"fmt"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

type backgroundPublisher struct {
	topic  string
	output chan<- interfaces.BackgroundMessage
}

// Publish provided message through the configured MessageBus output
func (pub *backgroundPublisher) Publish(payload []byte, context interfaces.AppFunctionContext) error {
	outputEnvelope := types.MessageEnvelope{
		CorrelationID: context.CorrelationID(),
		Payload:       payload,
		ContentType:   context.InputContentType(),
	}

	topic, err := context.ApplyValues(pub.topic)

	if err != nil {
		return fmt.Errorf("Failed to prepare topic for publishing: %s", err.Error())
	}

	pub.output <- BackgroundMessage{
		Payload:      outputEnvelope,
		PublishTopic: topic,
	}

	return nil
}

func newBackgroundPublisher(baseTopic string, capacity int) (<-chan interfaces.BackgroundMessage, interfaces.BackgroundPublisher) {
	backgroundChannel := make(chan interfaces.BackgroundMessage, capacity)
	return backgroundChannel, &backgroundPublisher{topic: baseTopic, output: backgroundChannel}
}
