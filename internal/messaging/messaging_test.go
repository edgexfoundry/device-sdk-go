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

package messaging

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	messaging2 "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/stretchr/testify/assert"
)

var lc logger.LoggingClient
var dic *di.Container
var usernameSecretData = map[string]string{
	messaging2.SecretUsernameKey: "username",
	messaging2.SecretPasswordKey: "password",
}

func TestMain(m *testing.M) {
	lc = logger.NewMockClient()

	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	os.Exit(m.Run())
}

func TestBootstrapHandler(t *testing.T) {
	validNotUsingMessageBus := config.ConfigurationStruct{
		Service: config.ServiceInfo{
			UseMessageBus: false,
		},
	}

	validCreateClient := config.ConfigurationStruct{
		Service: config.ServiceInfo{
			UseMessageBus: true,
		},
		MessageQueue: bootstrapConfig.MessageBusInfo{
			Type:               messaging.ZeroMQ, // Use ZMQ so no issue connecting.
			Protocol:           "http",
			Host:               "*",
			Port:               8765,
			PublishTopicPrefix: "edgex/events/#",
			AuthMode:           messaging2.AuthModeUsernamePassword,
			SecretName:         "redisdb",
		},
	}

	invalidSecrets := config.ConfigurationStruct{
		Service: config.ServiceInfo{
			UseMessageBus: true,
		},
		MessageQueue: bootstrapConfig.MessageBusInfo{
			AuthMode:   messaging2.AuthModeCert,
			SecretName: "redisdb",
		},
	}

	invalidNoConnect := config.ConfigurationStruct{
		Service: config.ServiceInfo{
			UseMessageBus: true,
		},
		MessageQueue: bootstrapConfig.MessageBusInfo{
			Type:       messaging.MQTT, // This will cause no connection since broker not available
			Protocol:   "tcp",
			Host:       "localhost",
			Port:       8765,
			AuthMode:   messaging2.AuthModeUsernamePassword,
			SecretName: "redisdb",
		},
	}

	tests := []struct {
		Name           string
		Config         *config.ConfigurationStruct
		ExpectedResult bool
		ExpectClient   bool
	}{
		{"Valid - Not using client", &validNotUsingMessageBus, true, false},
		{"Valid - creates client", &validCreateClient, true, true},
		{"Invalid - secrets error", &invalidSecrets, false, false},
		{"Invalid - can't connect", &invalidNoConnect, false, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			provider := &mocks.SecretProvider{}
			provider.On("GetSecret", test.Config.MessageQueue.SecretName).Return(usernameSecretData, nil)
			dic.Update(di.ServiceConstructorMap{
				container.ConfigurationName: func(get di.Get) interface{} {
					return test.Config
				},
				bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
					return provider
				},
				container.MessagingClientName: func(get di.Get) interface{} {
					return nil
				},
			})

			actual := BootstrapHandler(context.Background(), &sync.WaitGroup{}, startup.NewTimer(1, 1), dic)
			assert.Equal(t, test.ExpectedResult, actual)
			if test.ExpectClient {
				assert.NotNil(t, container.MessagingClientFrom(dic.Get))
			} else {
				assert.Nil(t, container.MessagingClientFrom(dic.Get))
			}
		})
	}
}
