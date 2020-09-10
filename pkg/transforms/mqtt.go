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
	"sync"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// MqttConfig contains mqtt client parameters
type MqttConfig struct {
	Qos            byte
	Retain         bool
	AutoReconnect  bool
	SkipCertVerify bool
	User           string
	Password       string
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

type MQTTSender struct {
	lock           sync.Mutex
	client         MQTT.Client
	topic          string
	opts           MqttConfig
	persistOnError bool
}

// NewMQTTSender - create new mqtt sender
func NewMQTTSender(logging logger.LoggingClient, addr models.Addressable, keyCertPair *KeyCertPair,
	mqttConfig MqttConfig, persistOnError bool) *MQTTSender {
	protocol := strings.ToLower(addr.Protocol)

	opts := MQTT.NewClientOptions()
	broker := protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(mqttConfig.AutoReconnect)

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
			InsecureSkipVerify: mqttConfig.SkipCertVerify,
			Certificates:       []tls.Certificate{cert},
		}

		opts.SetTLSConfig(tlsConfig)

	}

	sender := &MQTTSender{
		client:         MQTT.NewClient(opts),
		topic:          addr.Topic,
		opts:           mqttConfig,
		persistOnError: persistOnError,
	}

	return sender
}

// MQTTSend sends data from the previous function to the specified MQTT broker.
// If no previous function exists, then the event that triggered the pipeline will be used.
func (sender *MQTTSender) MQTTSend(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}

	exportData, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}

	if !sender.client.IsConnected() {
		err = sender.connectToBroker(edgexcontext, exportData)
		if err != nil {
			return false, err
		}
	}

	token := sender.client.Publish(sender.topic, sender.opts.Qos, sender.opts.Retain, exportData)
	token.Wait()
	if token.Error() != nil {
		sender.setRetryData(edgexcontext, exportData)
		return false, token.Error()
	}

	edgexcontext.LoggingClient.Debug("Sent data to MQTT Broker")
	edgexcontext.LoggingClient.Trace("Data exported", "Transport", "MQTT", clients.CorrelationHeader, edgexcontext.CorrelationID)

	return true, nil
}

func (sender *MQTTSender) connectToBroker(edgexcontext *appcontext.Context, exportData []byte) error {
	sender.lock.Lock()
	defer sender.lock.Unlock()

	// If other thread made the connection while this one was waiting for the lock
	// then skip trying to connect
	if sender.client.IsConnected() {
		return nil
	}

	edgexcontext.LoggingClient.Info("Connecting to mqtt server")
	if token := sender.client.Connect(); token.Wait() && token.Error() != nil {
		sender.setRetryData(edgexcontext, exportData)
		subMessage := "drop event"
		if sender.persistOnError {
			subMessage = "persisting Event for later retry"
		}
		return fmt.Errorf("Could not connect to mqtt server, %s. Error: %s", subMessage, token.Error().Error())
	}
	edgexcontext.LoggingClient.Info("Connected to mqtt server")
	return nil
}

func (sender *MQTTSender) setRetryData(ctx *appcontext.Context, exportData []byte) {
	if sender.persistOnError {
		ctx.RetryData = exportData
	}
}
