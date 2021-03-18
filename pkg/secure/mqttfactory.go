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

package secure

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

type mqttSecrets struct {
	username     string
	password     string
	keyPemBlock  []byte
	certPemBlock []byte
	caPemBlock   []byte
}

const (
	AuthModeNone             = "none"
	AuthModeUsernamePassword = "usernamepassword"
	AuthModeCert             = "clientcert"
	AuthModeCA               = "cacert"
	// Name of the keys to look for in secret provider
	MQTTSecretUsername   = "username"
	MQTTSecretPassword   = "password"
	MQTTSecretClientKey  = "clientkey"
	MQTTSecretClientCert = AuthModeCert
	MQTTSecretCACert     = AuthModeCA
)

type MqttFactory struct {
	appContext     interfaces.AppFunctionContext
	logger         logger.LoggingClient
	authMode       string
	secretPath     string
	opts           *mqtt.ClientOptions
	skipCertVerify bool
}

func NewMqttFactory(appContext interfaces.AppFunctionContext, mode string, path string, skipVerify bool) MqttFactory {
	return MqttFactory{
		appContext:     appContext,
		authMode:       mode,
		secretPath:     path,
		skipCertVerify: skipVerify,
	}
}

func (factory MqttFactory) Create(opts *mqtt.ClientOptions) (mqtt.Client, error) {
	if factory.authMode == "" {
		factory.authMode = AuthModeNone
		factory.logger.Warn("AuthMode not set, defaulting to \"" + AuthModeNone + "\"")
	}

	factory.opts = opts

	//get the secrets from the secret provider and populate the struct
	secrets, err := factory.getSecrets()
	if err != nil {
		return nil, err
	}
	//ensure that the authmode selected has the required secret values
	if secrets != nil {
		err = factory.validateSecrets(*secrets)
		if err != nil {
			return nil, err
		}
		// configure the mqtt client with the retrieved secret values
		err = factory.configureMQTTClientForAuth(*secrets)
		if err != nil {
			return nil, err
		}
	}

	return mqtt.NewClient(factory.opts), nil
}

func (factory MqttFactory) getSecrets() (*mqttSecrets, error) {
	// No Auth? No Problem!...No secrets required.
	if factory.authMode == AuthModeNone {
		return nil, nil
	}

	secrets, err := factory.appContext.GetSecret(factory.secretPath)
	if err != nil {
		return nil, err
	}
	mqttSecrets := &mqttSecrets{
		username:     secrets[MQTTSecretUsername],
		password:     secrets[MQTTSecretPassword],
		keyPemBlock:  []byte(secrets[MQTTSecretClientKey]),
		certPemBlock: []byte(secrets[MQTTSecretClientCert]),
		caPemBlock:   []byte(secrets[MQTTSecretCACert]),
	}

	return mqttSecrets, nil
}

func (factory MqttFactory) validateSecrets(secrets mqttSecrets) error {
	caCertPool := x509.NewCertPool()
	if factory.authMode == AuthModeUsernamePassword {
		if secrets.username == "" || secrets.password == "" {
			return errors.New("AuthModeUsernamePassword selected however Username or Password was not found at secret path")
		}

	} else if factory.authMode == AuthModeCert {
		// need both to make a successful connection
		if len(secrets.keyPemBlock) <= 0 || len(secrets.certPemBlock) <= 0 {
			return errors.New("AuthModeCert selected however the key or cert PEM block was not found at secret path")
		}
	} else if factory.authMode == AuthModeCA {
		if len(secrets.caPemBlock) <= 0 {
			return errors.New("AuthModeCA selected however no PEM Block was found at secret path")
		}
	} else if factory.authMode != AuthModeNone {
		return errors.New("Invalid AuthMode selected")
	}

	if len(secrets.caPemBlock) > 0 {
		ok := caCertPool.AppendCertsFromPEM(secrets.caPemBlock)
		if !ok {
			return errors.New("Error parsing CA Certificate")
		}
	}

	return nil
}

func (factory MqttFactory) configureMQTTClientForAuth(secrets mqttSecrets) error {
	var cert tls.Certificate
	var err error
	caCertPool := x509.NewCertPool()
	tlsConfig := &tls.Config{
		InsecureSkipVerify: factory.skipCertVerify,
	}
	switch factory.authMode {
	case AuthModeUsernamePassword:
		factory.opts.SetUsername(secrets.username)
		factory.opts.SetPassword(secrets.password)
	case AuthModeCert:
		cert, err = tls.X509KeyPair(secrets.certPemBlock, secrets.keyPemBlock)
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	case AuthModeCA:
		break
	case AuthModeNone:
		return nil
	}

	if len(secrets.caPemBlock) > 0 {
		ok := caCertPool.AppendCertsFromPEM(secrets.caPemBlock)
		if !ok {
			return errors.New("Error parsing CA PEM block")
		}
		tlsConfig.ClientCAs = caCertPool
	}

	factory.opts.SetTLSConfig(tlsConfig)

	return nil
}
