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
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPublish_Plain_Topic(t *testing.T) {
	topic := uuid.NewString()

	background, pub := newBackgroundPublisher(topic, 1)

	payload := []byte("something")
	correlationId := "id"
	contentType := "type"

	pub.Publish(payload, appfunction.NewContext(correlationId, nil, contentType))

	waiting := true

	for waiting {
		select {
		case msgs := <-background:
			msg := msgs.Message()
			assert.Equal(t, correlationId, msg.CorrelationID)
			assert.Equal(t, contentType, msg.ContentType)
			assert.Equal(t, payload, msg.Payload)
			assert.Equal(t, topic, msgs.Topic())
			waiting = false
		case <-time.After(1 * time.Second):
			assert.Fail(t, "message timed out, background channel likely not configured correctly")
			waiting = false
		}
	}
}

func TestPublish_Formatted_Topic(t *testing.T) {
	topic := "{context-key}/topic"

	background, pub := newBackgroundPublisher(topic, 1)

	payload := []byte("something")
	correlationId := "id"
	contentType := "type"

	appCtx := appfunction.NewContext(correlationId, nil, contentType)

	appCtx.AddValue("context-key", "replaced")
	err := pub.Publish(payload, appCtx)

	require.NoError(t, err)

	waiting := true

	for waiting {
		select {
		case msgs := <-background:
			msg := msgs.Message()
			assert.Equal(t, correlationId, msg.CorrelationID)
			assert.Equal(t, contentType, msg.ContentType)
			assert.Equal(t, payload, msg.Payload)
			assert.Equal(t, "replaced/topic", msgs.Topic())
			waiting = false
		case <-time.After(1 * time.Second):
			assert.Fail(t, "message timed out, background channel likely not configured correctly")
			waiting = false
		}
	}
}

func TestPublish_Topic_Formatting_Error(t *testing.T) {
	topic := "{context-key}/topic"

	_, pub := newBackgroundPublisher(topic, 1)

	payload := []byte("something")
	correlationId := "id"
	contentType := "type"

	err := pub.Publish(payload, appfunction.NewContext(correlationId, nil, contentType))

	require.Error(t, err)

	require.Equal(t, fmt.Sprintf("Failed to prepare topic for publishing: failed to replace all context placeholders in input ('%s' after replacements)", topic), err.Error())
}
