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

	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// MQTTSender ...
type MQTTSender struct {
	client MQTT.Client
	topic  string
}

// MQTTSend ...
func (sender MQTTSender) MQTTSend(edgexcontext excontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	if !sender.client.IsConnected() {
		edgexcontext.LoggingClient.Info("Connecting to mqtt server")
		if token := sender.client.Connect(); token.Wait() && token.Error() != nil {
			edgexcontext.LoggingClient.Error(fmt.Sprintf("Could not connect to mqtt server, drop event. Error: %s", token.Error().Error()))
			return false, nil
		}
		edgexcontext.LoggingClient.Info("Connected to mqtt server")
	}
	if data, ok := params[0].(string); ok {
		token := sender.client.Publish(sender.topic, 0, false, ([]byte)(data))
		// FIXME: could be removed? set of tokens?
		token.Wait()
		if token.Error() != nil {
			return false, token.Error()
		} else {
			edgexcontext.LoggingClient.Info("Sent data to MQTT Broker")
			edgexcontext.LoggingClient.Debug(fmt.Sprintf("Sent data to MQTT Broker: %X", ([]byte)(data)))
			return true, nil
		}
	}
	return false, errors.New("Unexpected type received")
}

// NewMQTTSender - create new mqtt sender
func NewMQTTSender(logging logger.LoggingClient, addr models.Addressable, cert string, key string) *MQTTSender {
	protocol := strings.ToLower(addr.Protocol)

	opts := MQTT.NewClientOptions()
	broker := protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(false)

	if protocol == "tcps" || protocol == "ssl" || protocol == "tls" {
		cert, err := tls.LoadX509KeyPair(cert, key)

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
	}

	return sender
}
