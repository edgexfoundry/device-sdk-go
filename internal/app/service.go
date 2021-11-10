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

package app

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/handlers"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/webserver"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/config"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clientInterfaces "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	commonConstants "github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"

	"github.com/gorilla/mux"
)

const (
	envProfile    = "EDGEX_PROFILE"
	envServiceKey = "EDGEX_SERVICE_KEY"
)

// NewService create, initializes and returns new instance of app.Service which implements the
// interfaces.ApplicationService interface
func NewService(serviceKey string, targetType interface{}, profileSuffixPlaceholder string) *Service {
	return &Service{
		serviceKey:               serviceKey,
		targetType:               targetType,
		profileSuffixPlaceholder: profileSuffixPlaceholder,
	}
}

// Service provides the necessary struct and functions to create an instance of the
// interfaces.ApplicationService interface.
type Service struct {
	dic                       *di.Container
	serviceKey                string
	targetType                interface{}
	config                    *common.ConfigurationStruct
	lc                        logger.LoggingClient
	usingConfigurablePipeline bool
	runtime                   *runtime.GolangRuntime
	webserver                 *webserver.WebServer
	ctx                       contextGroup
	deferredFunctions         []bootstrap.Deferred
	backgroundPublishChannel  <-chan interfaces.BackgroundMessage
	customTriggerFactories    map[string]func(sdk *Service) (interfaces.Trigger, error)
	profileSuffixPlaceholder  string
	commandLine               commandLineFlags
	flags                     *flags.Default
	configProcessor           *config.Processor
}

type commandLineFlags struct {
	skipVersionCheck   bool
	serviceKeyOverride string
}

type contextGroup struct {
	storeForwardWg        *sync.WaitGroup
	storeForwardCancelCtx context.CancelFunc
	appWg                 *sync.WaitGroup
	appCtx                context.Context
	appCancelCtx          context.CancelFunc
	stop                  context.CancelFunc
}

// AddRoute allows you to leverage the existing webserver to add routes.
func (svc *Service) AddRoute(route string, handler func(nethttp.ResponseWriter, *nethttp.Request), methods ...string) error {
	if route == commonConstants.ApiPingRoute ||
		route == commonConstants.ApiConfigRoute ||
		route == commonConstants.ApiMetricsRoute ||
		route == commonConstants.ApiVersionRoute ||
		route == internal.ApiTriggerRoute {
		return errors.New("route is reserved")
	}
	return svc.webserver.AddRoute(route, svc.addContext(handler), methods...)
}

// AddBackgroundPublisher will create a channel of provided capacity to be
// consumed by the MessageBus output and return a publisher that writes to it
func (svc *Service) AddBackgroundPublisher(capacity int) (interfaces.BackgroundPublisher, error) {
	topic := svc.config.Trigger.EdgexMessageBus.PublishHost.PublishTopic

	if topic == "" {
		return nil, errors.New("no publish topic configured for messagebus, background publishing unavailable")
	}
	return svc.AddBackgroundPublisherWithTopic(capacity, topic)
}

// AddBackgroundPublisherWithTopic will create a channel of provided capacity to be
// consumed by the MessageBus output and return a publisher that writes to it on a different
// topic than configured for messagebus output.
func (svc *Service) AddBackgroundPublisherWithTopic(capacity int, topic string) (interfaces.BackgroundPublisher, error) {
	// for custom triggers we don't know if background publishing available or not
	// but probably makes sense to trust the caller.
	if svc.config.Trigger.Type == TriggerTypeHTTP || svc.config.Trigger.Type == TriggerTypeMQTT {
		return nil, fmt.Errorf("background publishing not supported for %s trigger", svc.config.Trigger.Type)
	}

	bgChan, pub := newBackgroundPublisher(topic, capacity)
	svc.backgroundPublishChannel = bgChan
	return pub, nil
}

// MakeItStop will force the service loop to exit in the same fashion as SIGINT/SIGTERM received from the OS
func (svc *Service) MakeItStop() {
	if svc.ctx.stop != nil {
		svc.ctx.stop()
	} else {
		svc.lc.Warn("MakeItStop called but no stop handler set on SDK - is the service running?")
	}
}

// MakeItRun initializes and starts the trigger as specified in the
// configuration. It will also configure the webserver and start listening on
// the specified port.
func (svc *Service) MakeItRun() error {
	runCtx, stop := context.WithCancel(context.Background())

	svc.ctx.stop = stop

	// determine input type and create trigger for it
	t := svc.setupTrigger(svc.config, svc.runtime)
	if t == nil {
		return errors.New("failed to create Trigger")
	}

	// Initialize the trigger (i.e. start a web server, or connect to message bus)
	deferred, err := t.Initialize(svc.ctx.appWg, svc.ctx.appCtx, svc.backgroundPublishChannel)
	if err != nil {
		svc.lc.Error(err.Error())
		return errors.New("failed to initialize Trigger")
	}

	// deferred is a function that needs to be called when services exits.
	svc.addDeferred(deferred)

	if svc.config.Writable.StoreAndForward.Enabled {
		svc.startStoreForward()
	} else {
		svc.lc.Info("StoreAndForward disabled. Not running retry loop.")
	}

	svc.lc.Info(svc.config.Service.StartupMsg)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	httpErrors := make(chan error)
	defer close(httpErrors)

	svc.webserver.StartWebServer(httpErrors)

	select {
	case httpError := <-httpErrors:
		svc.lc.Info("Http error received: ", httpError.Error())
		err = httpError

	case signalReceived := <-signals:
		svc.lc.Info("Terminating signal received: " + signalReceived.String())

	case <-runCtx.Done():
		svc.lc.Info("Terminating: svc.MakeItStop called")
	}

	svc.ctx.stop = nil

	if svc.config.Writable.StoreAndForward.Enabled {
		svc.ctx.storeForwardCancelCtx()
		svc.ctx.storeForwardWg.Wait()
	}

	svc.ctx.appCancelCtx() // Cancel all long-running go funcs
	svc.ctx.appWg.Wait()
	// Call all the deferred funcs that need to happen when exiting.
	// These are things like un-register from the Registry, disconnect from the Message Bus, etc
	for _, deferredFunc := range svc.deferredFunctions {
		deferredFunc()
	}

	return err
}

// LoadConfigurablePipeline sets the function pipeline from configuration
// Note this API has been deprecated, replaced by LoadConfigurableFunctionPipelines and will be removed in a future release
// TODO: Remove this API in 3.0 release
func (svc *Service) LoadConfigurablePipeline() ([]interfaces.AppFunction, error) {
	pipelines, err := svc.LoadConfigurableFunctionPipelines()
	if err != nil {
		return nil, err
	}

	defaultPipeline, found := pipelines[interfaces.DefaultPipelineId]
	if !found {
		return nil, fmt.Errorf("default functions pipeline not configured")
	}

	return defaultPipeline.Transforms, nil
}

// LoadConfigurableFunctionPipelines return the configured function pipelines (default and per topic) from configuration.
func (svc *Service) LoadConfigurableFunctionPipelines() (map[string]interfaces.FunctionPipeline, error) {
	pipelines := make(map[string]interfaces.FunctionPipeline)

	svc.usingConfigurablePipeline = true

	svc.targetType = nil

	if svc.config.Writable.Pipeline.UseTargetTypeOfByteArray {
		svc.targetType = &[]byte{}
	}

	configurable := reflect.ValueOf(NewConfigurable(svc.lc))
	pipelineConfig := svc.config.Writable.Pipeline

	defaultExecutionOrder := strings.TrimSpace(pipelineConfig.ExecutionOrder)

	if len(defaultExecutionOrder) == 0 && len(pipelineConfig.PerTopicPipelines) == 0 {
		return nil, errors.New("default ExecutionOrder has 0 functions specified and PerTopicPipelines is empty")
	}

	if len(defaultExecutionOrder) > 0 {
		svc.lc.Debugf("Default Function Pipeline Execution Order: [%s]", pipelineConfig.ExecutionOrder)
		functionNames := util.DeleteEmptyAndTrim(strings.FieldsFunc(defaultExecutionOrder, util.SplitComma))

		transforms, err := svc.loadConfigurablePipelineTransforms(interfaces.DefaultPipelineId, functionNames, pipelineConfig.Functions, configurable)
		if err != nil {
			return nil, err
		}
		pipeline := interfaces.FunctionPipeline{
			Id:         interfaces.DefaultPipelineId,
			Transforms: transforms,
			Topics:     []string{runtime.TopicWildCard},
		}
		pipelines[pipeline.Id] = pipeline
	}

	if len(pipelineConfig.PerTopicPipelines) > 0 {
		for _, perTopicPipeline := range pipelineConfig.PerTopicPipelines {
			svc.lc.Debugf("'%s' Function Pipeline Execution Order: [%s]", perTopicPipeline.Id, perTopicPipeline.ExecutionOrder)

			functionNames := util.DeleteEmptyAndTrim(strings.FieldsFunc(perTopicPipeline.ExecutionOrder, util.SplitComma))

			transforms, err := svc.loadConfigurablePipelineTransforms(perTopicPipeline.Id, functionNames, pipelineConfig.Functions, configurable)
			if err != nil {
				return nil, err
			}

			pipeline := interfaces.FunctionPipeline{
				Id:         perTopicPipeline.Id,
				Transforms: transforms,
				Topics:     util.DeleteEmptyAndTrim(strings.FieldsFunc(perTopicPipeline.Topics, util.SplitComma)),
			}

			pipelines[pipeline.Id] = pipeline
		}
	}

	return pipelines, nil
}

func (svc *Service) loadConfigurablePipelineTransforms(
	pipelineId string,
	executionOrder []string,
	functions map[string]common.PipelineFunction,
	configurable reflect.Value) ([]interfaces.AppFunction, error) {
	var transforms []interfaces.AppFunction

	for _, functionName := range executionOrder {
		functionName = strings.TrimSpace(functionName)
		configuration, ok := functions[functionName]
		if !ok {
			return nil, fmt.Errorf("function '%s' configuration not found in Pipeline.Functions section for pipeline '%s'", functionName, pipelineId)
		}

		functionValue, functionType, err := svc.findMatchingFunction(configurable, functionName)
		if err != nil {
			return nil, fmt.Errorf("%s for pipeline '%s'", err.Error(), pipelineId)
		}

		// determine number of parameters required for function call
		inputParameters := make([]reflect.Value, functionType.NumIn())
		// set keys to be all lowercase to avoid casing issues from configuration
		for key := range configuration.Parameters {
			value := configuration.Parameters[key]
			delete(configuration.Parameters, key) // Make sure the old key has been removed so don't have multiples
			configuration.Parameters[strings.ToLower(key)] = value
		}
		for index := range inputParameters {
			parameter := functionType.In(index)

			switch parameter {
			case reflect.TypeOf(map[string]string{}):
				inputParameters[index] = reflect.ValueOf(configuration.Parameters)

			default:
				return nil, fmt.Errorf(
					"function %s for pipeline '%s' has an unsupported parameter type: %s",
					pipelineId,
					functionName,
					parameter.String(),
				)
			}
		}

		function, ok := functionValue.Call(inputParameters)[0].Interface().(interfaces.AppFunction)
		if !ok {
			return nil, fmt.Errorf("failed to cast function %s as AppFunction type for pipeline '%s'", functionName, pipelineId)
		}

		if function == nil {
			return nil, fmt.Errorf("%s from configuration failed for pipeline '%s'", functionName, pipelineId)
		}

		transforms = append(transforms, function)
		svc.lc.Debugf("%s function added to '%s' configurable pipeline with parameters: [%s]",
			functionName,
			pipelineId,
			listParameters(configuration.Parameters))
	}

	return transforms, nil
}

// SetFunctionsPipeline has been deprecated and replaced by SetDefaultFunctionsPipeline.
func (svc *Service) SetFunctionsPipeline(transforms ...interfaces.AppFunction) error {
	return svc.SetDefaultFunctionsPipeline(transforms...)
}

// SetDefaultFunctionsPipeline sets the default functions pipeline to the list of specified functions in the order provided.
func (svc *Service) SetDefaultFunctionsPipeline(transforms ...interfaces.AppFunction) error {
	if len(transforms) == 0 {
		return errors.New("no transforms provided to pipeline")
	}

	svc.runtime.TargetType = svc.targetType
	svc.runtime.SetDefaultFunctionsPipeline(transforms)

	svc.lc.Debugf("Default pipeline added with %d transform(s)", len(transforms))

	return nil
}

// AddFunctionsPipelineForTopics adds a functions pipeline for the specified for the specified id and topics
func (svc *Service) AddFunctionsPipelineForTopics(id string, topics []string, transforms ...interfaces.AppFunction) error {
	if len(transforms) == 0 {
		return errors.New("no transforms provided to pipeline")
	}

	if len(topics) == 0 {
		return errors.New("topics for pipeline can not be empty")
	}

	for _, t := range topics {
		if strings.TrimSpace(t) == "" {
			return errors.New("blank topic not allowed")
		}
	}
	err := svc.runtime.AddFunctionsPipeline(id, topics, transforms)
	if err != nil {
		return err
	}

	svc.lc.Debugf("Pipeline '%s' added for topics '%v' with %d transform(s)", id, topics, len(transforms))
	return nil
}

// ApplicationSettings returns the values specified in the custom configuration section.
func (svc *Service) ApplicationSettings() map[string]string {
	return svc.config.ApplicationSettings
}

// GetAppSetting returns the string for the specified App Setting.
func (svc *Service) GetAppSetting(setting string) (string, error) {
	if svc.config.ApplicationSettings == nil {
		return "", fmt.Errorf("%s setting not found: ApplicationSettings section is missing", setting)
	}

	settingValue, ok := svc.config.ApplicationSettings[setting]
	if !ok {
		return "", fmt.Errorf("%s setting not found in ApplicationSettings section", setting)
	}

	return settingValue, nil
}

// GetAppSettingStrings returns the strings slice for the specified App Setting.
func (svc *Service) GetAppSettingStrings(setting string) ([]string, error) {
	if svc.config.ApplicationSettings == nil {
		return nil, fmt.Errorf("%s setting not found: ApplicationSettings section is missing", setting)
	}

	settingValue, ok := svc.config.ApplicationSettings[setting]
	if !ok {
		return nil, fmt.Errorf("%s setting not found in ApplicationSettings section", setting)
	}

	valueStrings := util.DeleteEmptyAndTrim(strings.FieldsFunc(settingValue, util.SplitComma))

	return valueStrings, nil
}

// Initialize bootstraps the service making it ready to accept functions for the pipeline and to run the configured trigger.
func (svc *Service) Initialize() error {
	startupTimer := startup.NewStartUpTimer(svc.serviceKey)

	additionalUsage :=
		"    -s/--skipVersionCheck           Indicates the service should skip the Core Service's version compatibility check.\n" +
			"    -sk/--serviceKey                Overrides the service service key used with Registry and/or Configuration Providers.\n" +
			"                                    If the name provided contains the text `<profile>`, this text will be replaced with\n" +
			"                                    the name of the profile used."

	svc.flags = flags.NewWithUsage(additionalUsage)
	svc.flags.FlagSet.BoolVar(&svc.commandLine.skipVersionCheck, "skipVersionCheck", false, "")
	svc.flags.FlagSet.BoolVar(&svc.commandLine.skipVersionCheck, "s", false, "")
	svc.flags.FlagSet.StringVar(&svc.commandLine.serviceKeyOverride, "serviceKey", "", "")
	svc.flags.FlagSet.StringVar(&svc.commandLine.serviceKeyOverride, "sk", "", "")

	svc.flags.Parse(os.Args[1:])

	// Temporarily setup logging to STDOUT so the client can be used before bootstrapping is completed
	svc.lc = logger.NewClient(svc.serviceKey, models.InfoLog)

	svc.setServiceKey(svc.flags.Profile())

	svc.lc.Info(fmt.Sprintf("Starting %s %s ", svc.serviceKey, internal.ApplicationVersion))

	svc.config = &common.ConfigurationStruct{}
	svc.dic = di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return svc.config
		},
	})

	svc.ctx.appCtx, svc.ctx.appCancelCtx = context.WithCancel(context.Background())
	svc.ctx.appWg = &sync.WaitGroup{}

	var deferred bootstrap.Deferred
	var successful bool
	var configUpdated config.UpdatedStream = make(chan struct{})

	svc.ctx.appWg, deferred, successful = bootstrap.RunAndReturnWaitGroup(
		svc.ctx.appCtx,
		svc.ctx.appCancelCtx,
		svc.flags,
		svc.serviceKey,
		internal.ConfigRegistryStem,
		svc.config,
		configUpdated,
		startupTimer,
		svc.dic,
		true,
		[]bootstrapInterfaces.BootstrapHandler{
			handlers.NewDatabase().BootstrapHandler,
			handlers.NewClients().BootstrapHandler,
			handlers.NewTelemetry().BootstrapHandler,
			handlers.NewVersionValidator(svc.commandLine.skipVersionCheck, internal.SDKVersion).BootstrapHandler,
		},
	)

	// deferred is a function that needs to be called when services exits.
	svc.addDeferred(deferred)

	if !successful {
		return fmt.Errorf("boostrapping failed")
	}

	svc.runtime = runtime.NewGolangRuntime(svc.serviceKey, svc.targetType, svc.dic)

	// Bootstrapping is complete, so now need to retrieve the needed objects from the containers.
	svc.lc = bootstrapContainer.LoggingClientFrom(svc.dic.Get)

	// We do special processing when the writeable section of the configuration changes, so have
	// to wait to be signaled when the configuration has been updated and then process the changes
	NewConfigUpdateProcessor(svc).WaitForConfigUpdates(configUpdated)

	svc.webserver = webserver.NewWebServer(svc.dic, mux.NewRouter())
	svc.webserver.ConfigureStandardRoutes()

	svc.lc.Info("Service started in: " + startupTimer.SinceAsString())

	return nil
}

// LoadCustomConfig uses the Config Processor from go-mod-bootstrap to attempt to load service's
// custom configuration. It uses the same command line flags to process the custom config in the same manner
// as the standard configuration.
func (svc *Service) LoadCustomConfig(customConfig interfaces.UpdatableConfig, sectionName string) error {
	if svc.configProcessor == nil {
		svc.configProcessor = config.NewProcessorForCustomConfig(svc.flags, svc.ctx.appCtx, svc.ctx.appWg, svc.dic)
	}
	return svc.configProcessor.LoadCustomConfigSection(customConfig, sectionName)
}

// ListenForCustomConfigChanges uses the Config Processor from go-mod-bootstrap to attempt to listen for
// changes to the specified custom configuration section. LoadCustomConfig must be called previously so that
// the instance of svc.configProcessor has already been set.
func (svc *Service) ListenForCustomConfigChanges(configToWatch interface{}, sectionName string, changedCallback func(interface{})) error {
	if svc.configProcessor == nil {
		return fmt.Errorf(
			"custom configuration must be loaded before '%s' section can be watched for changes",
			sectionName)
	}

	svc.configProcessor.ListenForCustomConfigChanges(configToWatch, sectionName, changedCallback)
	return nil
}

// GetSecret retrieves secret data from the secret store at the specified path.
func (svc *Service) GetSecret(path string, keys ...string) (map[string]string, error) {
	secretProvider := bootstrapContainer.SecretProviderFrom(svc.dic.Get)
	return secretProvider.GetSecret(path, keys...)
}

// StoreSecret stores the secret data to a secret store at the specified path.
func (svc *Service) StoreSecret(path string, secretData map[string]string) error {
	secretProvider := bootstrapContainer.SecretProviderFrom(svc.dic.Get)
	return secretProvider.StoreSecret(path, secretData)
}

// LoggingClient returns the Logging client from the dependency injection container
func (svc *Service) LoggingClient() logger.LoggingClient {
	return svc.lc
}

// RegistryClient returns the Registry client, which may be nil, from the dependency injection container
func (svc *Service) RegistryClient() registry.Client {
	return bootstrapContainer.RegistryFrom(svc.dic.Get)
}

// EventClient returns the Event client, which may be nil, from the dependency injection container
func (svc *Service) EventClient() clientInterfaces.EventClient {
	return container.EventClientFrom(svc.dic.Get)
}

// CommandClient returns the Command client, which may be nil, from the dependency injection container
func (svc *Service) CommandClient() clientInterfaces.CommandClient {
	return container.CommandClientFrom(svc.dic.Get)
}

// DeviceServiceClient returns the DeviceService client, which may be nil, from the dependency injection container
func (svc *Service) DeviceServiceClient() clientInterfaces.DeviceServiceClient {
	return container.DeviceServiceClientFrom(svc.dic.Get)
}

// DeviceProfileClient returns the DeviceProfile client, which may be nil, from the dependency injection container
func (svc *Service) DeviceProfileClient() clientInterfaces.DeviceProfileClient {
	return container.DeviceProfileClientFrom(svc.dic.Get)
}

// DeviceClient returns the Device client, which may be nil, from the dependency injection container
func (svc *Service) DeviceClient() clientInterfaces.DeviceClient {
	return container.DeviceClientFrom(svc.dic.Get)
}

// NotificationClient returns the Notifications client, which may be nil, from the dependency injection container
func (svc *Service) NotificationClient() clientInterfaces.NotificationClient {
	return container.NotificationClientFrom(svc.dic.Get)
}

// SubscriptionClient returns the Subscription client, which may be nil, from the dependency injection container
func (svc *Service) SubscriptionClient() clientInterfaces.SubscriptionClient {
	return container.SubscriptionClientFrom(svc.dic.Get)
}

func listParameters(parameters map[string]string) string {
	result := ""
	first := true
	for key, value := range parameters {
		if first {
			result = fmt.Sprintf("%s='%s'", key, value)
			first = false
			continue
		}

		result += fmt.Sprintf(", %s='%s'", key, value)
	}

	return result
}

func (svc *Service) addContext(next func(nethttp.ResponseWriter, *nethttp.Request)) func(nethttp.ResponseWriter, *nethttp.Request) {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		ctx := context.WithValue(r.Context(), interfaces.AppServiceContextKey, svc) //nolint: staticcheck
		next(w, r.WithContext(ctx))
	}
}

func (svc *Service) addDeferred(deferred bootstrap.Deferred) {
	if deferred != nil {
		svc.deferredFunctions = append(svc.deferredFunctions, deferred)
	}
}

func (svc *Service) setServiceKey(profile string) {
	envValue := os.Getenv(envServiceKey)
	if len(envValue) > 0 {
		svc.commandLine.serviceKeyOverride = envValue
		svc.lc.Info(
			fmt.Sprintf("Environment profileOverride of '-n/--serviceName' by environment variable: %s=%s",
				envServiceKey,
				envValue))
	}

	// serviceKeyOverride may have been set by the -n/--serviceName command-line option and not the environment variable
	if len(svc.commandLine.serviceKeyOverride) > 0 {
		svc.serviceKey = svc.commandLine.serviceKeyOverride
	}

	if !strings.Contains(svc.serviceKey, svc.profileSuffixPlaceholder) {
		// No placeholder, so nothing to do here
		return
	}

	// Have to handle environment override here before common bootstrap is used, so it is passed the proper service key
	profileOverride := os.Getenv(envProfile)
	if len(profileOverride) > 0 {
		profile = profileOverride
	}

	if len(profile) > 0 {
		svc.serviceKey = strings.Replace(svc.serviceKey, svc.profileSuffixPlaceholder, profile, 1)
		return
	}

	// No profile specified so remove the placeholder text
	svc.serviceKey = strings.Replace(svc.serviceKey, svc.profileSuffixPlaceholder, "", 1)
}

// BuildContext allows external callers that may need a context (e.g. background publishers)
// to easily create one around the service's dic
func (svc *Service) BuildContext(correlationId string, contentType string) interfaces.AppFunctionContext {
	return appfunction.NewContext(correlationId, svc.dic, contentType)
}
