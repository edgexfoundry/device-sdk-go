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

package mqtt

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/secure"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	pahoMqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

// Trigger implements Trigger to support Triggers
type Trigger struct {
	dic          *di.Container
	lc           logger.LoggingClient
	mqttClient   pahoMqtt.Client
	runtime      *runtime.GolangRuntime
	qos          byte
	retain       bool
	publishTopic string
}

func NewTrigger(dic *di.Container, runtime *runtime.GolangRuntime) *Trigger {
	return &Trigger{
		dic:     dic,
		runtime: runtime,
		lc:      bootstrapContainer.LoggingClientFrom(dic.Get),
	}
}

// Initialize initializes the Trigger for an external MQTT broker
func (trigger *Trigger) Initialize(_ *sync.WaitGroup, _ context.Context, background <-chan interfaces.BackgroundMessage) (bootstrap.Deferred, error) {
	// Convenience short cuts
	lc := trigger.lc
	config := container.ConfigurationFrom(trigger.dic.Get)
	brokerConfig := config.Trigger.ExternalMqtt
	topics := config.Trigger.ExternalMqtt.SubscribeTopics

	trigger.qos = brokerConfig.QoS
	trigger.retain = brokerConfig.Retain
	trigger.publishTopic = config.Trigger.ExternalMqtt.PublishTopic

	lc.Info("Initializing MQTT Trigger")

	if background != nil {
		return nil, errors.New("background publishing not supported for services using MQTT trigger")
	}

	if len(strings.TrimSpace(topics)) == 0 {
		return nil, fmt.Errorf("missing SubscribeTopics for MQTT Trigger. Must be present in [Trigger.ExternalMqtt] section")
	}

	brokerUrl, err := url.Parse(brokerConfig.Url)
	if err != nil {
		return nil, fmt.Errorf("invalid MQTT Broker Url '%s': %s", config.Trigger.ExternalMqtt.Url, err.Error())
	}

	opts := pahoMqtt.NewClientOptions()
	opts.AutoReconnect = brokerConfig.AutoReconnect
	opts.OnConnect = trigger.onConnectHandler
	opts.ClientID = brokerConfig.ClientId
	if len(brokerConfig.ConnectTimeout) > 0 {
		duration, err := time.ParseDuration(brokerConfig.ConnectTimeout)
		if err != nil {
			return nil, fmt.Errorf("invalid MQTT ConnectTimeout '%s': %s", brokerConfig.ConnectTimeout, err.Error())
		}
		opts.ConnectTimeout = duration
	}
	opts.KeepAlive = brokerConfig.KeepAlive
	opts.Servers = []*url.URL{brokerUrl}

	// Since this factory is shared between the MQTT pipeline function and this trigger we must provide
	// a dummy AppFunctionContext which will provide access to GetSecret
	mqttFactory := secure.NewMqttFactory(
		appfunction.NewContext("", trigger.dic, ""),
		brokerConfig.AuthMode,
		brokerConfig.SecretPath,
		brokerConfig.SkipCertVerify,
	)

	mqttClient, err := mqttFactory.Create(opts)
	if err != nil {
		return nil, fmt.Errorf("unable to create secure MQTT Client: %s", err.Error())
	}

	lc.Infof("Connecting to mqtt broker for MQTT trigger at: %s", brokerUrl)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("could not connect to broker for MQTT trigger: %s", token.Error().Error())
	}

	lc.Info("Connected to mqtt server for MQTT trigger")

	deferred := func() {
		lc.Info("Disconnecting from broker for MQTT trigger")
		trigger.mqttClient.Disconnect(0)
	}

	trigger.mqttClient = mqttClient

	return deferred, nil
}

func (trigger *Trigger) onConnectHandler(mqttClient pahoMqtt.Client) {
	// Convenience short cuts
	lc := trigger.lc
	config := container.ConfigurationFrom(trigger.dic.Get)
	topics := util.DeleteEmptyAndTrim(strings.FieldsFunc(config.Trigger.ExternalMqtt.SubscribeTopics, util.SplitComma))
	qos := config.Trigger.ExternalMqtt.QoS

	for _, topic := range topics {
		if token := mqttClient.Subscribe(topic, qos, trigger.messageHandler); token.Wait() && token.Error() != nil {
			mqttClient.Disconnect(0)
			lc.Errorf("could not subscribe to topic '%s' for MQTT trigger: %s",
				topic, token.Error().Error())
			return
		}
	}

	lc.Infof("Subscribed to topic(s) '%s' for MQTT trigger", config.Trigger.ExternalMqtt.SubscribeTopics)
}

func (trigger *Trigger) messageHandler(_ pahoMqtt.Client, mqttMessage pahoMqtt.Message) {
	// Convenience short cuts
	lc := trigger.lc

	data := mqttMessage.Payload()
	contentType := common.ContentTypeJSON
	if data[0] != byte('{') && data[0] != byte('[') {
		// If not JSON then assume it is CBOR
		contentType = common.ContentTypeCBOR
	}

	correlationID := uuid.New().String()

	message := types.MessageEnvelope{
		CorrelationID: correlationID,
		ContentType:   contentType,
		Payload:       data,
		ReceivedTopic: mqttMessage.Topic(),
	}

	lc.Debugf("MQTT Trigger: Received message with %d bytes on topic '%s'. Content-Type=%s",
		len(message.Payload),
		message.ReceivedTopic,
		message.ContentType)
	lc.Tracef("%s=%s", common.CorrelationHeader, correlationID)

	pipelines := trigger.runtime.GetMatchingPipelines(message.ReceivedTopic)
	lc.Debugf("MQTT Trigger found %d pipeline(s) that match the incoming topic '%s'", len(pipelines), message.ReceivedTopic)
	for _, pipeline := range pipelines {
		go trigger.processMessageWithPipeline(message, pipeline)
	}
}

func (trigger *Trigger) processMessageWithPipeline(envelope types.MessageEnvelope, pipeline *interfaces.FunctionPipeline) {
	appContext := appfunction.NewContext(envelope.CorrelationID, trigger.dic, envelope.ContentType)

	messageError := trigger.runtime.ProcessMessage(appContext, envelope, pipeline)
	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		// ToDo: Do we want to publish the error back to the Broker?
		return
	}

	if len(appContext.ResponseData()) > 0 && len(trigger.publishTopic) > 0 {
		formattedTopic, err := appContext.ApplyValues(trigger.publishTopic)

		if err != nil {
			trigger.lc.Errorf("MQTT trigger: Unable to format topic '%s' for pipeline '%s': %s",
				trigger.publishTopic,
				pipeline.Id,
				err.Error())
		}

		if token := trigger.mqttClient.Publish(formattedTopic, trigger.qos, trigger.retain, appContext.ResponseData()); token.Wait() && token.Error() != nil {
			trigger.lc.Errorf("MQTT trigger: Could not publish to topic '%s' for pipeline '%s': %s",
				formattedTopic,
				pipeline.Id,
				token.Error().Error())
		} else {
			trigger.lc.Debugf("MQTT Trigger: Published response message for pipeline '%s' on topic '%s' with %d bytes",
				pipeline.Id,
				formattedTopic,
				len(appContext.ResponseData()))
			trigger.lc.Tracef("MQTT Trigger published message: %s=%s", common.CorrelationHeader, envelope.CorrelationID)
		}
	}
}
