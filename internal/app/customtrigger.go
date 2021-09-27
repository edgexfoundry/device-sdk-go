//
// Copyright (c) 2020 Technotects
// Copyright (c) 2021 Intel Corporation
// Copyright (c) 2021 One Track Consulting
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

package app

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/hashicorp/go-multierror"
	"strings"
	"sync"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger/mqtt"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

const (
	// Valid types of App Service triggers
	TriggerTypeMessageBus = "EDGEX-MESSAGEBUS"
	TriggerTypeMQTT       = "EXTERNAL-MQTT"
	TriggerTypeHTTP       = "HTTP"
)

type customTriggerBinding struct {
	TriggerServiceBinding
	dic *di.Container
	log logger.LoggingClient
}

func newCustomTriggerBinding(svc *Service) *customTriggerBinding {
	return &customTriggerBinding{
		TriggerServiceBinding: simpleTriggerServiceBinding{
			svc,
			svc.runtime,
		},
		log: svc.LoggingClient(),
		dic: svc.dic,
	}
}

func (bnd *customTriggerBinding) buildContext(env types.MessageEnvelope) interfaces.AppFunctionContext {
	return appfunction.NewContext(env.CorrelationID, bnd.dic, env.ContentType)
}

// processMessage provides an interface for custom triggers to execute service's configured pipelines through the runtime.
func (bnd *customTriggerBinding) processMessage(ctx interfaces.AppFunctionContext, envelope types.MessageEnvelope) error {
	bnd.log.Debugf("custom trigger attempting to find pipeline(s) for topic %s", envelope.ReceivedTopic)

	// ensure we have a context established that we can safely cast to *appfunction.Context to pass to runtime
	if _, ok := ctx.(*appfunction.Context); ctx == nil || !ok {
		ctx = bnd.buildContext(envelope)
	}

	pipelines := bnd.GetMatchingPipelines(envelope.ReceivedTopic)

	bnd.log.Debugf("custom trigger found %d pipeline(s) that match the incoming topic '%s'", len(pipelines), envelope.ReceivedTopic)

	var finalErr error

	pipelinesWaitGroup := sync.WaitGroup{}

	for _, pipeline := range pipelines {
		pipelinesWaitGroup.Add(1)
		go func(p *interfaces.FunctionPipeline, wg *sync.WaitGroup, errCollector func(error)) {
			defer wg.Done()

			bnd.log.Tracef("custom trigger sending message to pipeline %s (%s)", p.Id, envelope.CorrelationID)

			if msgErr := bnd.ProcessMessage(ctx.Clone().(*appfunction.Context), envelope, p); msgErr != nil {
				bnd.log.Tracef("custom trigger message error in pipeline %s (%s) %s", p.Id, envelope.CorrelationID, msgErr.Err.Error())
				errCollector(msgErr.Err)
			}
		}(pipeline, &pipelinesWaitGroup, func(e error) { finalErr = multierror.Append(finalErr, e) })
	}

	pipelinesWaitGroup.Wait()

	return finalErr
}

func (svc *Service) setupTrigger(configuration *common.ConfigurationStruct, runtime *runtime.GolangRuntime) interfaces.Trigger {
	var t interfaces.Trigger
	// Need to make dynamic, search for the trigger that is input

	switch triggerType := strings.ToUpper(configuration.Trigger.Type); triggerType {
	case TriggerTypeHTTP:
		svc.LoggingClient().Info("HTTP trigger selected")
		t = http.NewTrigger(svc.dic, runtime, svc.webserver)

	case TriggerTypeMessageBus:
		svc.LoggingClient().Info("EdgeX MessageBus trigger selected")
		t = messagebus.NewTrigger(svc.dic, runtime)

	case TriggerTypeMQTT:
		svc.LoggingClient().Info("External MQTT trigger selected")
		t = mqtt.NewTrigger(svc.dic, runtime)

	default:
		if factory, found := svc.customTriggerFactories[triggerType]; found {
			var err error
			t, err = factory(svc)
			if err != nil {
				svc.LoggingClient().Errorf("failed to initialize custom trigger [%s]: %s", triggerType, err.Error())
				return nil
			}
		} else {
			svc.LoggingClient().Errorf("Invalid Trigger type of '%s' specified", configuration.Trigger.Type)
		}
	}

	return t
}

// RegisterCustomTriggerFactory allows users to register builders for custom trigger types
func (svc *Service) RegisterCustomTriggerFactory(name string,
	factory func(interfaces.TriggerConfig) (interfaces.Trigger, error)) error {
	nu := strings.ToUpper(name)

	if nu == TriggerTypeMessageBus ||
		nu == TriggerTypeHTTP ||
		nu == TriggerTypeMQTT {
		return fmt.Errorf("cannot register custom trigger for builtin type (%s)", name)
	}

	if svc.customTriggerFactories == nil {
		svc.customTriggerFactories = make(map[string]func(sdk *Service) (interfaces.Trigger, error), 1)
	}

	svc.customTriggerFactories[nu] = func(sdk *Service) (interfaces.Trigger, error) {
		binding := newCustomTriggerBinding(sdk)

		cfg := interfaces.TriggerConfig{
			Logger:           sdk.lc,
			ContextBuilder:   binding.buildContext,
			MessageProcessor: binding.processMessage,
			ConfigLoader:     binding.LoadCustomConfig,
		}

		return factory(cfg)
	}

	return nil
}
