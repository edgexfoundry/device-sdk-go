//
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

// This test will only be executed if the tag brokerRunning is added when running
// the tests with a command like:
// go test -tags brokerRunning
package transforms

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetRetryDataPersistFalse(t *testing.T) {
	context.SetRetryData(nil)
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, false)
	sender.mqttConfig = MQTTSecretConfig{}
	sender.setRetryData(context, []byte("data"))
	assert.Nil(t, context.RetryData())
}

func TestSetRetryDataPersistTrue(t *testing.T) {
	context.SetRetryData(nil)
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, true)
	sender.mqttConfig = MQTTSecretConfig{}
	sender.setRetryData(context, []byte("data"))
	assert.Equal(t, []byte("data"), context.RetryData())
}

func TestMQTTSendNodata(t *testing.T) {
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, true)
	sender.mqttConfig = MQTTSecretConfig{}
	continuePipeline, result := sender.MQTTSend(context, nil)
	require.False(t, continuePipeline)
	require.Error(t, result.(error))
}
