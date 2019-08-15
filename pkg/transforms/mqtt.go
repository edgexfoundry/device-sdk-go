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

package transforms

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// MqttConfig contains mqtt client parameters
type MqttConfig struct {
	qos           byte
	retain        bool
	autoreconnect bool
	user          string
	password      string
}

// KeyCertPair is used to pass key/cert pair to NewMQTTSender
// KeyPEMBlock and CertPEMBlock will be used if they are not nil
// then it will fall back to KeyFile and CertFile
type KeyCertPair struct {
	KeyFile      string
	CertFile     string
	KeyPEMBlock  []byte
	CertPEMBlock []byte
}

// NewMqttConfig returns a new MqttConfig with default values. Use Setter functions to change specific values.
func NewMqttConfig() *MqttConfig {
	mqttConfig := &MqttConfig{}
	mqttConfig.qos = 0
	mqttConfig.retain = false
	mqttConfig.autoreconnect = false

	return mqttConfig
}

// SetRetain enables or disables mqtt retain option
func (mqttConfig MqttConfig) SetRetain(retain bool) {
	mqttConfig.retain = retain
}

// SetQos changes mqtt qos(0,1,2) for all messages
func (mqttConfig MqttConfig) SetQos(qos byte) {
	mqttConfig.qos = qos
}

// SetAutoreconnect enables or disables the automatic client reconnection to broker
func (mqttConfig MqttConfig) SetAutoreconnect(reconnect bool) {
	mqttConfig.autoreconnect = reconnect
}

type MQTTSender struct {
	client MQTT.Client
	topic  string
	opts   MqttConfig
}

// NewMQTTSender - create new mqtt sender
func NewMQTTSender(logging logger.LoggingClient, addr models.Addressable, keyCertPair *KeyCertPair, mqttConfig *MqttConfig) *MQTTSender {
	protocol := strings.ToLower(addr.Protocol)

	opts := MQTT.NewClientOptions()
	broker := protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(mqttConfig.autoreconnect)

	if (protocol == "tcps" || protocol == "ssl" || protocol == "tls") && keyCertPair != nil {
		var cert tls.Certificate
		var err error

		if keyCertPair.KeyPEMBlock != nil && keyCertPair.CertPEMBlock != nil {
			cert, err = tls.X509KeyPair(keyCertPair.CertPEMBlock, keyCertPair.KeyPEMBlock)
		} else {
			cert, err = tls.LoadX509KeyPair(keyCertPair.CertFile, keyCertPair.KeyFile)
		}

		if err != nil {
			logging.Error("Failed loading x509 data")
			return nil
		}

		tlsConfig := &tls.Config{
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		}

		opts.SetTLSConfig(tlsConfig)

	}

	sender := &MQTTSender{
		client: MQTT.NewClient(opts),
		topic:  addr.Topic,
		opts:   *mqttConfig,
	}

	return sender
}

// MQTTSend sends data from the previous function to the specified MQTT broker.
// If no previous function exists, then the event that triggered the pipeline will be used.
func (sender MQTTSender) MQTTSend(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	if !sender.client.IsConnected() {
		edgexcontext.LoggingClient.Info("Connecting to mqtt server")
		if token := sender.client.Connect(); token.Wait() && token.Error() != nil {
			return false, fmt.Errorf("Could not connect to mqtt server, drop event. Error: %s", token.Error().Error())
		}
		edgexcontext.LoggingClient.Info("Connected to mqtt server")
	}
	data, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}
	token := sender.client.Publish(sender.topic, sender.opts.qos, sender.opts.retain, data)
	// FIXME: could be removed? set of tokens?
	token.Wait()
	if token.Error() != nil {
		return false, token.Error()
	}
	edgexcontext.LoggingClient.Info("Sent data to MQTT Broker")
	edgexcontext.LoggingClient.Trace("Data exported", "Transport", "MQTT", clients.CorrelationHeader, edgexcontext.CorrelationID)

	return true, nil
}
