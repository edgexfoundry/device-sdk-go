//
// Copyright (c) 2020 Intel Corporation
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

package transforms

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

// MQTTSecretSender ...
type MQTTSecretSender struct {
	client               MQTT.Client
	mqttConfig           MQTTSecretConfig
	persistOnError       bool
	opts                 *MQTT.ClientOptions
	secretsLastRetrieved time.Time
}

// MQTTSecretConfig ...
type MQTTSecretConfig struct {
	// BrokerAddress should be set to the complete broker address i.e. mqtts://mosquitto:8883/mybroker
	BrokerAddress string
	// ClientId to connect with the broker with.
	ClientId string
	// The name of the path in secret provider to retrieve your secrets
	SecretPath string
	// AutoReconnect indicated whether or not to retry connection if disconnected
	AutoReconnect bool
	// Topic that you wish to publish to
	Topic string
	// QoS for MQTT Connection
	QoS byte
	// Retain setting for MQTT Connection
	Retain bool
	// SkipCertVerify
	SkipCertVerify bool
	// AuthMode indicates what to use when connecting to the broker. Options are "none", "cacert" , "usernamepassword", "clientcert".
	// If a CA Cert exists in the SecretPath then it will be used for all modes except "none".
	AuthMode string
}
type mqttSecrets struct {
	username     string
	password     string
	keypemblock  []byte
	certpemblock []byte
	capemblock   []byte
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

// NewMQTTSecretSender ...
func NewMQTTSecretSender(mqttConfig MQTTSecretConfig, persistOnError bool) *MQTTSecretSender {
	opts := MQTT.NewClientOptions()

	opts.AddBroker(mqttConfig.BrokerAddress)
	opts.SetClientID(mqttConfig.ClientId)
	opts.SetAutoReconnect(mqttConfig.AutoReconnect)
	//avoid casing issues
	mqttConfig.AuthMode = strings.ToLower(mqttConfig.AuthMode)
	sender := &MQTTSecretSender{
		client:         nil,
		mqttConfig:     mqttConfig,
		persistOnError: persistOnError,
		opts:           opts,
	}

	return sender
}
func (sender *MQTTSecretSender) getSecrets(edgexcontext *appcontext.Context) (*mqttSecrets, error) {
	// No Auth? No Problem!...No secrets required.
	if sender.mqttConfig.AuthMode == AuthModeNone {
		return nil, nil
	}

	secrets, err := edgexcontext.GetSecrets(sender.mqttConfig.SecretPath)
	if err != nil {
		return nil, err
	}
	mqttSecrets := &mqttSecrets{
		username:     secrets[MQTTSecretUsername],
		password:     secrets[MQTTSecretPassword],
		keypemblock:  []byte(secrets[MQTTSecretClientKey]),
		certpemblock: []byte(secrets[MQTTSecretClientCert]),
		capemblock:   []byte(secrets[MQTTSecretCACert]),
	}

	return mqttSecrets, nil
}
func (sender *MQTTSecretSender) validateSecrets(secrets mqttSecrets) error {
	caCertPool := x509.NewCertPool()
	if sender.mqttConfig.AuthMode == AuthModeUsernamePassword {
		if secrets.username == "" || secrets.password == "" {
			return errors.New("AuthModeUsernamePassword selected however username or password was not found at secret path")
		}

	} else if sender.mqttConfig.AuthMode == AuthModeCert {
		// need both to make a successful connection
		if len(secrets.keypemblock) <= 0 || len(secrets.certpemblock) <= 0 {
			return errors.New("AuthModeCert selected however the key or cert PEM block was not found at secret path")
		}
	} else if sender.mqttConfig.AuthMode == AuthModeCA && len(secrets.capemblock) <= 0 {
		return errors.New("AuthModeCA selected however no PEM Block was found at secret path")
	} else if sender.mqttConfig.AuthMode != AuthModeNone {
		return errors.New("Invalid AuthMode selected")
	}

	if len(secrets.capemblock) > 0 {
		ok := caCertPool.AppendCertsFromPEM([]byte(secrets.capemblock))
		if !ok {
			return errors.New("Error parsing CA Certificate")
		}
	}
	return nil
}
func (sender *MQTTSecretSender) configureMQTTClientForAuth(secrets mqttSecrets) error {
	var cert tls.Certificate
	var err error
	caCertPool := x509.NewCertPool()
	tlsConfig := &tls.Config{
		InsecureSkipVerify: sender.mqttConfig.SkipCertVerify,
	}
	switch sender.mqttConfig.AuthMode {
	case AuthModeUsernamePassword:
		sender.opts.SetUsername(secrets.username)
		sender.opts.SetPassword(secrets.password)
	case AuthModeCert:
		cert, err = tls.X509KeyPair(secrets.certpemblock, secrets.keypemblock)
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	case AuthModeCA:
		break
	case AuthModeNone:
		return nil
	}

	if len(secrets.capemblock) > 0 {
		ok := caCertPool.AppendCertsFromPEM(secrets.capemblock)
		if !ok {
			return errors.New("Error parsing CA PEM block")
		}
		tlsConfig.ClientCAs = caCertPool
	}

	sender.opts.SetTLSConfig(tlsConfig)

	return nil
}
func (sender *MQTTSecretSender) initializeMQTTClient(edgexcontext *appcontext.Context) error {
	if sender.mqttConfig.AuthMode == "" {
		sender.mqttConfig.AuthMode = AuthModeNone
		edgexcontext.LoggingClient.Warn("AuthMode not set, defaulting to \"" + AuthModeNone + "\"")
	}

	//get the secrets from the secret provider and populate the struct
	secrets, err := sender.getSecrets(edgexcontext)
	if err != nil {
		return err
	}
	//ensure that the authmode selected has the required secret values
	if secrets != nil {
		err = sender.validateSecrets(*secrets)
		if err != nil {
			return err
		}
		// configure the mqtt client with the retrieved secret values
		err = sender.configureMQTTClientForAuth(*secrets)
		if err != nil {
			return err
		}
	}

	sender.secretsLastRetrieved = time.Now()
	sender.client = MQTT.NewClient(sender.opts)
	return nil
}

func (sender *MQTTSecretSender) connectToBroker(edgexcontext *appcontext.Context, exportData []byte) error {
	edgexcontext.LoggingClient.Info("Connecting to mqtt server for export")
	if token := sender.client.Connect(); token.Wait() && token.Error() != nil {
		sender.setRetryData(edgexcontext, exportData)
		subMessage := "dropping event"
		if sender.persistOnError {
			subMessage = "persisting Event for later retry"
		}
		return fmt.Errorf("Could not connect to mqtt server for export, %s. Error: %s", subMessage, token.Error().Error())
	}
	edgexcontext.LoggingClient.Info("Connected to mqtt server for export")
	return nil
}

// MQTTSend sends data from the previous function to the specified MQTT broker.
// If no previous function exists, then the event that triggered the pipeline will be used.
func (sender *MQTTSecretSender) MQTTSend(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}

	exportData, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}
	// if we havent initialized the client yet OR the cache has been invalidated (due to new/updated secrets) we need to (re)initialize the client
	if sender.client == nil || sender.secretsLastRetrieved.Before(edgexcontext.SecretProvider.SecretsLastUpdated()) {
		err := sender.initializeMQTTClient(edgexcontext)
		if err != nil {
			return false, err
		}
	}
	if !sender.client.IsConnected() {
		err := sender.connectToBroker(edgexcontext, exportData)
		if err != nil {
			return false, err
		}
	}

	token := sender.client.Publish(sender.mqttConfig.Topic, sender.mqttConfig.QoS, sender.mqttConfig.Retain, exportData)
	token.Wait()
	if token.Error() != nil {
		sender.setRetryData(edgexcontext, exportData)
		return false, token.Error()
	}

	edgexcontext.LoggingClient.Debug("Sent data to MQTT Broker")
	edgexcontext.LoggingClient.Trace("Data exported", "Transport", "MQTT", clients.CorrelationHeader, edgexcontext.CorrelationID)

	return true, nil
}

func (sender *MQTTSecretSender) setRetryData(ctx *appcontext.Context, exportData []byte) {
	if sender.persistOnError {
		ctx.RetryData = exportData
	}
}
