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
package secure

import (
	"os"
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
)

var lc logger.LoggingClient
var dic *di.Container
var context *appfunction.Context

func TestMain(m *testing.M) {
	lc = logger.NewMockClient()
	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	context = appfunction.NewContext("123", dic, "")

	os.Exit(m.Run())
}

func TestNewMqttFactory(t *testing.T) {
	expectedMode := "none"
	expectedPath := "myPath"
	expectedSkipVerify := true
	target := NewMqttFactory(context, expectedMode, expectedPath, expectedSkipVerify)

	assert.NotNil(t, target.logger)
	assert.Equal(t, expectedMode, target.authMode)
	assert.Equal(t, expectedPath, target.secretPath)
	assert.Equal(t, expectedSkipVerify, target.skipCertVerify)
	assert.Nil(t, target.opts)

}

func TestConfigureMQTTClientForAuth(t *testing.T) {
	target := NewMqttFactory(context, "", "", false)
	target.opts = mqtt.NewClientOptions()
	tests := []struct {
		Name             string
		AuthMode         string
		secrets          messaging.SecretData
		ErrorExpectation bool
		ErrorMessage     string
	}{
		{"Username and Password should be set", messaging.AuthModeUsernamePassword, messaging.SecretData{
			Username: messaging.SecretUsernameKey,
			Password: messaging.SecretPasswordKey,
		},
			false, ""},
		{"No AuthMode", messaging.AuthModeNone, messaging.SecretData{}, false, ""},
		{"Invalid AuthMode", "", messaging.SecretData{}, false, ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			target.authMode = test.AuthMode
			result := target.configureMQTTClientForAuth(&test.secrets)
			if test.ErrorExpectation {
				assert.Error(t, result, "Result should be an error")
				assert.Equal(t, test.ErrorMessage, result.(error).Error())
			} else {
				assert.Nil(t, result, "Should be nil")
			}
		})
	}
}
func TestConfigureMQTTClientForAuthWithUsernamePassword(t *testing.T) {
	target := NewMqttFactory(context, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = messaging.AuthModeUsernamePassword
	err := target.configureMQTTClientForAuth(&messaging.SecretData{
		Username: "Username",
		Password: "Password",
	})
	require.NoError(t, err)
	assert.Equal(t, target.opts.Username, "Username")
	assert.Equal(t, target.opts.Password, "Password")
	assert.Nil(t, target.opts.TLSConfig.ClientCAs)
	assert.Nil(t, target.opts.TLSConfig.Certificates)

}
func TestConfigureMQTTClientForAuthWithUsernamePasswordAndCA(t *testing.T) {
	target := NewMqttFactory(context, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = messaging.AuthModeUsernamePassword
	err := target.configureMQTTClientForAuth(&messaging.SecretData{
		Username:   "Username",
		Password:   "Password",
		CaPemBlock: []byte(testCACert),
	})
	require.NoError(t, err)
	assert.Equal(t, target.opts.Username, "Username")
	assert.Equal(t, target.opts.Password, "Password")
	assert.Nil(t, target.opts.TLSConfig.Certificates)
	assert.NotNil(t, target.opts.TLSConfig.ClientCAs)
}

func TestConfigureMQTTClientForAuthWithCACert(t *testing.T) {
	target := NewMqttFactory(context, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = messaging.AuthModeCA
	err := target.configureMQTTClientForAuth(&messaging.SecretData{
		Username:   "Username",
		Password:   "Password",
		CaPemBlock: []byte(testCACert),
	})

	require.NoError(t, err)
	assert.NotNil(t, target.opts.TLSConfig.ClientCAs)
	assert.Empty(t, target.opts.Username)
	assert.Empty(t, target.opts.Password)
	assert.Nil(t, target.opts.TLSConfig.Certificates)
}
func TestConfigureMQTTClientForAuthWithClientCert(t *testing.T) {
	target := NewMqttFactory(context, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = messaging.AuthModeCert
	err := target.configureMQTTClientForAuth(&messaging.SecretData{
		Username:     "Username",
		Password:     "Password",
		CertPemBlock: []byte(testClientCert),
		KeyPemBlock:  []byte(testClientKey),
		CaPemBlock:   []byte(testCACert),
	})
	require.NoError(t, err)
	assert.Empty(t, target.opts.Username)
	assert.Empty(t, target.opts.Password)
	assert.NotNil(t, target.opts.TLSConfig.Certificates)
	assert.NotNil(t, target.opts.TLSConfig.ClientCAs)
}

func TestConfigureMQTTClientForAuthWithClientCertNoCA(t *testing.T) {
	target := NewMqttFactory(context, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = messaging.AuthModeCert
	err := target.configureMQTTClientForAuth(&messaging.SecretData{
		Username:     messaging.SecretUsernameKey,
		Password:     messaging.SecretPasswordKey,
		CertPemBlock: []byte(testClientCert),
		KeyPemBlock:  []byte(testClientKey),
	})

	require.NoError(t, err)
	assert.Empty(t, target.opts.Username)
	assert.Empty(t, target.opts.Password)
	assert.NotNil(t, target.opts.TLSConfig.Certificates)
	assert.Nil(t, target.opts.TLSConfig.ClientCAs)
}
func TestConfigureMQTTClientForAuthWithNone(t *testing.T) {
	target := NewMqttFactory(context, "", "", false)
	target.opts = mqtt.NewClientOptions()
	target.authMode = messaging.AuthModeNone
	err := target.configureMQTTClientForAuth(&messaging.SecretData{})

	require.NoError(t, err)
}
