//
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
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/trigger"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
	"github.com/hashicorp/go-multierror"
	"sync"
)

type simpleTriggerServiceBinding struct {
	*Service
	*runtime.GolangRuntime
}

func (b *simpleTriggerServiceBinding) SecretProvider() messaging.SecretDataProvider {
	return container.SecretProviderFrom(b.Service.dic.Get)
}

func NewTriggerServiceBinding(svc *Service) trigger.ServiceBinding {
	return &simpleTriggerServiceBinding{
		svc,
		svc.runtime,
	}
}

func (b *simpleTriggerServiceBinding) DIC() *di.Container {
	return b.Service.dic
}

func (b *simpleTriggerServiceBinding) BuildContext(env types.MessageEnvelope) interfaces.AppFunctionContext {
	return appfunction.NewContext(env.CorrelationID, b.Service.dic, env.ContentType)
}

func (b *simpleTriggerServiceBinding) Config() *common.ConfigurationStruct {
	return b.Service.config
}

// customTriggerBinding wraps the CustomTriggerServiceBinding interface so that we can attach methods
type triggerMessageProcessor struct {
	bnd trigger.ServiceBinding
}

// Process provides runtime orchestration to pass the envelope / context to the pipeline.
// Deprecated: This does NOT support multi-pipeline usage.  Will send a message to the default pipeline ONLY and throw if not configured.  Use MessageReceived.
func (mp *triggerMessageProcessor) Process(ctx interfaces.AppFunctionContext, envelope types.MessageEnvelope) error {
	context, ok := ctx.(*appfunction.Context)
	if !ok {
		return fmt.Errorf("App Context must be an instance of internal appfunction.Context. Use NewAppContext to create instance.")
	}

	defaultPipelines := mp.bnd.GetMatchingPipelines(interfaces.DefaultPipelineId)

	if len(defaultPipelines) != 1 {
		return fmt.Errorf("TriggerMessageProcessor is deprecated and does not support non-default or multiple pipelines.  Please use TriggerMessageHandler.")
	}

	messageError := mp.bnd.ProcessMessage(context, envelope, defaultPipelines[0])
	if messageError != nil {
		// ProcessMessage logs the error, so no need to log it here.
		return messageError.Err
	}

	return nil
}

// MessageReceived provides runtime orchestration to pass the envelope / context to configured pipeline(s) along with a response callback to execute on each completion.
func (mp *triggerMessageProcessor) MessageReceived(ctx interfaces.AppFunctionContext, envelope types.MessageEnvelope, responseHandler interfaces.PipelineResponseHandler) error {
	lc := mp.bnd.LoggingClient()

	lc.Debugf("custom trigger attempting to find pipeline(s) for topic %s", envelope.ReceivedTopic)

	// ensure we have a context established that we can safely cast to *appfunction.Context to pass to runtime
	if _, ok := ctx.(*appfunction.Context); ctx == nil || !ok {
		ctx = mp.bnd.BuildContext(envelope)
	}

	pipelines := mp.bnd.GetMatchingPipelines(envelope.ReceivedTopic)

	lc.Debugf("trigger found %d pipeline(s) that match the incoming topic '%s'", len(pipelines), envelope.ReceivedTopic)

	var finalErr error

	pipelinesWaitGroup := sync.WaitGroup{}

	for _, pipeline := range pipelines {
		pipelinesWaitGroup.Add(1)
		go func(p *interfaces.FunctionPipeline, wg *sync.WaitGroup, errCollector func(error)) {
			defer wg.Done()

			lc.Debugf("trigger sending message to pipeline %s (%s)", p.Id, envelope.CorrelationID)

			if msgErr := mp.bnd.ProcessMessage(ctx.Clone().(*appfunction.Context), envelope, p); msgErr != nil {
				lc.Errorf("message error in pipeline %s (%s): %s", p.Id, envelope.CorrelationID, msgErr.Err.Error())
				errCollector(msgErr.Err)
			} else {
				if responseHandler != nil {
					if outputErr := responseHandler(ctx, p); outputErr != nil {
						lc.Errorf("failed to process output for message '%s' on pipeline %s: %s", ctx.CorrelationID(), p.Id, outputErr.Error())
						errCollector(outputErr)
						return
					}
				}
				lc.Debugf("trigger successfully processed message '%s' in pipeline %s", p.Id, envelope.CorrelationID)
			}
		}(pipeline, &pipelinesWaitGroup, func(e error) { finalErr = multierror.Append(finalErr, e) })
	}

	pipelinesWaitGroup.Wait()

	return finalErr
}
