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
	"errors"
	"fmt"
	// can't use pkg.NewAppFuncContextForTest due circular reference
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/google/uuid"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMQTTSecretSender_setRetryDataPersistFalse(t *testing.T) {
	context.SetRetryData(nil)
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, false)
	sender.mqttConfig = MQTTSecretConfig{}
	sender.setRetryData(context, []byte("data"))
	assert.Nil(t, context.RetryData())
}

func TestMQTTSecretSender_setRetryDataPersistTrue(t *testing.T) {
	context.SetRetryData(nil)
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, true)
	sender.mqttConfig = MQTTSecretConfig{}
	sender.setRetryData(context, []byte("data"))
	assert.Equal(t, []byte("data"), context.RetryData())
}

func TestMQTTSecretSender_formatTopic_Default(t *testing.T) {
	configuredTopic := uuid.NewString()

	sender := NewMQTTSecretSender(MQTTSecretConfig{Topic: configuredTopic}, false)

	ctx := appfunction.NewContext(uuid.NewString(), nil, "")

	formattedTopic, err := sender.formatTopic(ctx, nil)

	require.NoError(t, err)
	require.Equal(t, configuredTopic, formattedTopic)
}

func TestMQTTSecretSender_formatTopic_Default_ContextKeyUsed(t *testing.T) {
	configuredTopic := uuid.NewString() + "/{test}"

	sender := NewMQTTSecretSender(MQTTSecretConfig{Topic: configuredTopic}, false)

	ctx := appfunction.NewContext(uuid.NewString(), nil, "")

	ctx.AddValue("test", "testreplacement")

	formattedTopic, err := sender.formatTopic(ctx, nil)

	expectedTopic := configuredTopic

	for k, v := range ctx.GetAllValues() {
		expectedTopic = strings.Replace(expectedTopic, fmt.Sprintf("{%s}", k), v, -1)
	}

	require.NoError(t, err)
	require.Equal(t, expectedTopic, formattedTopic)
}

func TestMQTTSecretSender_formatTopic_Default_MissingContextKeyUsed(t *testing.T) {
	configuredTopic := uuid.NewString() + "/{test}"

	sender := NewMQTTSecretSender(MQTTSecretConfig{Topic: configuredTopic}, false)

	ctx := appfunction.NewContext(uuid.NewString(), nil, "")

	formattedTopic, err := sender.formatTopic(ctx, nil)

	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("failed to replace all context placeholders in configured topic ('%s' after replacements)", configuredTopic), err.Error())
	require.Equal(t, "", formattedTopic)
}

func TestMQTTSecretSender_WithTopicFormatter_formatTopic(t *testing.T) {
	configuredTopic := uuid.NewString()
	subtopic := uuid.NewString()

	sender := NewMQTTSecretSenderWithTopicFormatter(MQTTSecretConfig{Topic: configuredTopic}, false, func(s string, functionContext interfaces.AppFunctionContext, i interface{}) (string, error) {
		return configuredTopic + "/" + subtopic, nil
	})

	ctx := appfunction.NewContext(uuid.NewString(), nil, "")

	formattedTopic, err := sender.formatTopic(ctx, nil)

	require.NoError(t, err)
	require.Equal(t, configuredTopic+"/"+subtopic, formattedTopic)
}

func TestMQTTSecretSender_WithTopicFormatter_formatTopic_Error(t *testing.T) {
	configuredTopic := uuid.NewString()

	sender := NewMQTTSecretSenderWithTopicFormatter(MQTTSecretConfig{Topic: configuredTopic}, false, func(s string, functionContext interfaces.AppFunctionContext, i interface{}) (string, error) {
		return "", errors.New("returned error from formatter")
	})

	ctx := appfunction.NewContext(uuid.NewString(), nil, "")

	formattedTopic, err := sender.formatTopic(ctx, nil)

	require.Error(t, err)
	require.Equal(t, "", formattedTopic)
}

func TestMQTTSecretSender_MQTTSendNodata(t *testing.T) {
	sender := NewMQTTSecretSender(MQTTSecretConfig{}, true)
	sender.mqttConfig = MQTTSecretConfig{}
	continuePipeline, result := sender.MQTTSend(context, nil)
	require.False(t, continuePipeline)
	require.Error(t, result.(error))
}
