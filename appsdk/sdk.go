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

package appsdk

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-registry/registry"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/config"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/handlers/message"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/bootstrap/handlers"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/webserver"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"
)

const (
	// ProfileSuffixPlaceholder is used to create unique names for profiles
	ProfileSuffixPlaceholder = "<profile>"
	envV1Profile             = "edgex_profile" // TODO: Remove for release v2.0.0
	envProfile               = "EDGEX_PROFILE"
	envServiceKey            = "EDGEX_SERVICE_KEY"
	envV1Service             = "edgex_service"    // deprecated TODO: Remove for release v2.0.0
	envServiceProtocol       = "Service_Protocol" // Used for envV1Service processing TODO: Remove for release v2.0.0
	envServiceHost           = "Service_Host"     // Used for envV1Service processing TODO: Remove for release v2.0.0
	envServicePort           = "Service_Port"     // Used for envV1Service processing TODO: Remove for release v2.0.0
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// SDKKey is the context key for getting the sdk context.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const SDKKey key = 0

// AppFunctionsSDK provides the necessary struct to create an instance of the Application Functions SDK. Be sure and provide a ServiceKey
// when creating an instance of the SDK. After creating an instance, you'll first want to call .Initialize(), to start up the SDK. Secondly,
// provide the desired transforms for your pipeline by calling .SetFunctionsPipeline(). Lastly, call .MakeItRun() to start listening for events based on
// your configured trigger.
type AppFunctionsSDK struct {
	// ServiceKey is the application services's key used for Configuration and Registration when the Registry is enabled
	ServiceKey string
	// LoggingClient is the EdgeX logger client used to log messages
	LoggingClient logger.LoggingClient
	// TargetType is the expected type of the incoming data. Must be set to a pointer to an instance of the type.
	// Defaults to &models.Event{} if nil. The income data is unmarshaled (JSON or CBOR) in to the type,
	// except when &[]byte{} is specified. In this case the []byte data is pass to the first function in the Pipeline.
	TargetType                interface{}
	transforms                []appcontext.AppFunction
	skipVersionCheck          bool
	usingConfigurablePipeline bool
	httpErrors                chan error
	runtime                   *runtime.GolangRuntime
	webserver                 *webserver.WebServer
	edgexClients              common.EdgeXClients
	registryClient            registry.Client
	config                    *common.ConfigurationStruct
	storeClient               interfaces.StoreClient
	secretProvider            *security.SecretProvider
	storeForwardWg            *sync.WaitGroup
	storeForwardCancelCtx     context.CancelFunc
	appWg                     *sync.WaitGroup
	appCtx                    context.Context
	appCancelCtx              context.CancelFunc
	deferredFunctions         []bootstrap.Deferred
	serviceKeyOverride        string
}

// AddRoute allows you to leverage the existing webserver to add routes.
func (sdk *AppFunctionsSDK) AddRoute(route string, handler func(nethttp.ResponseWriter, *nethttp.Request), methods ...string) error {
	if route == clients.ApiPingRoute ||
		route == clients.ApiConfigRoute ||
		route == clients.ApiMetricsRoute ||
		route == clients.ApiVersionRoute ||
		route == internal.ApiTriggerRoute {
		return errors.New("route is reserved")
	}
	return sdk.webserver.AddRoute(route, sdk.addContext(handler), methods...)
}

// MakeItRun will initialize and start the trigger as specifed in the
// configuration. It will also configure the webserver and start listening on
// the specified port.
func (sdk *AppFunctionsSDK) MakeItRun() error {
	httpErrors := make(chan error)
	defer close(httpErrors)

	sdk.runtime = &runtime.GolangRuntime{
		TargetType: sdk.TargetType,
		ServiceKey: sdk.ServiceKey,
	}

	sdk.runtime.Initialize(sdk.storeClient, sdk.secretProvider)
	sdk.runtime.SetTransforms(sdk.transforms)
	// determine input type and create trigger for it
	t := sdk.setupTrigger(sdk.config, sdk.runtime)

	// Initialize the trigger (i.e. start a web server, or connect to message bus)
	deferred, err := t.Initialize(sdk.appWg, sdk.appCtx)
	if err != nil {
		sdk.LoggingClient.Error(err.Error())
	}

	// deferred is a a function that needs to be called when services exits.
	sdk.addDeferred(deferred)

	if sdk.config.Writable.StoreAndForward.Enabled {
		sdk.startStoreForward()
	} else {
		sdk.LoggingClient.Info("StoreAndForward disabled. Not running retry loop.")
	}

	sdk.LoggingClient.Info(sdk.config.Service.StartupMsg)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sdk.webserver.StartWebServer(sdk.httpErrors)

	select {
	case httpError := <-sdk.httpErrors:
		sdk.LoggingClient.Info("Terminating: ", httpError.Error())
		err = httpError

	case signalReceived := <-signals:
		sdk.LoggingClient.Info("Terminating: " + signalReceived.String())
	}

	if sdk.config.Writable.StoreAndForward.Enabled {
		sdk.storeForwardCancelCtx()
		sdk.storeForwardWg.Wait()
	}

	sdk.appCancelCtx() // Cancel all long running go funcs
	sdk.appWg.Wait()

	// Call all the deferred funcs that need to happen when exiting.
	// These are things like un-register from the Registry, disconnect from the Message Bus, etc
	for _, deferredFunc := range sdk.deferredFunctions {
		deferredFunc()
	}

	return err
}

// LoadConfigurablePipeline ...
func (sdk *AppFunctionsSDK) LoadConfigurablePipeline() ([]appcontext.AppFunction, error) {
	var pipeline []appcontext.AppFunction

	sdk.usingConfigurablePipeline = true

	sdk.TargetType = nil
	if sdk.config.Writable.Pipeline.UseTargetTypeOfByteArray {
		sdk.TargetType = &[]byte{}
	}

	configurable := AppFunctionsSDKConfigurable{
		Sdk: sdk,
	}
	valueOfType := reflect.ValueOf(configurable)
	pipelineConfig := sdk.config.Writable.Pipeline
	executionOrder := util.DeleteEmptyAndTrim(strings.FieldsFunc(pipelineConfig.ExecutionOrder, util.SplitComma))

	if len(executionOrder) <= 0 {
		return nil, errors.New(
			"execution Order has 0 functions specified. You must have a least one function in the pipeline")
	}
	sdk.LoggingClient.Debug("Execution Order", "Functions", strings.Join(executionOrder, ","))

	for _, functionName := range executionOrder {
		functionName = strings.TrimSpace(functionName)
		configuration, ok := pipelineConfig.Functions[functionName]
		if !ok {
			return nil, fmt.Errorf("function %s configuration not found in Pipeline.Functions section", functionName)
		}

		result := valueOfType.MethodByName(functionName)
		if result.Kind() == reflect.Invalid {
			return nil, fmt.Errorf("function %s is not a built in SDK function", functionName)
		} else if result.IsNil() {
			return nil, fmt.Errorf("invalid/missing configuration for %s", functionName)
		}

		// determine number of parameters required for function call
		inputParameters := make([]reflect.Value, result.Type().NumIn())
		// set keys to be all lowercase to avoid casing issues from configuration
		for key := range configuration.Parameters {
			configuration.Parameters[strings.ToLower(key)] = configuration.Parameters[key]
		}
		for index := range inputParameters {
			parameter := result.Type().In(index)

			switch parameter {
			case reflect.TypeOf(map[string]string{}):
				inputParameters[index] = reflect.ValueOf(configuration.Parameters)

			case reflect.TypeOf(models.Addressable{}):
				inputParameters[index] = reflect.ValueOf(configuration.Addressable)

			default:
				return nil, fmt.Errorf(
					"function %s has an unsupported parameter type: %s",
					functionName,
					parameter.String(),
				)
			}
		}

		function, ok := result.Call(inputParameters)[0].Interface().(appcontext.AppFunction)
		if !ok {
			return nil, fmt.Errorf("failed to cast function %s as AppFunction type", functionName)
		}
		pipeline = append(pipeline, function)
		configurable.Sdk.LoggingClient.Debug(fmt.Sprintf("%s function added to configurable pipeline", functionName))
	}

	return pipeline, nil
}

// SetFunctionsPipeline allows you to define each fgitunction to execute and the order in which each function
// will be called as each event comes in.
func (sdk *AppFunctionsSDK) SetFunctionsPipeline(transforms ...appcontext.AppFunction) error {
	if len(transforms) == 0 {
		return errors.New("no transforms provided to pipeline")
	}

	sdk.transforms = transforms

	if sdk.runtime != nil {
		sdk.runtime.SetTransforms(transforms)
		sdk.runtime.TargetType = sdk.TargetType
	}

	return nil
}

// ApplicationSettings returns the values specifed in the custom configuration section.
func (sdk *AppFunctionsSDK) ApplicationSettings() map[string]string {
	return sdk.config.ApplicationSettings
}

// GetAppSettingStrings returns the strings slice for the specified App Setting.
func (sdk *AppFunctionsSDK) GetAppSettingStrings(setting string) ([]string, error) {
	if sdk.config.ApplicationSettings == nil {
		return nil, fmt.Errorf("%s setting not found: ApplicationSettings section is missing", setting)
	}

	settingValue, ok := sdk.config.ApplicationSettings[setting]
	if !ok {
		return nil, fmt.Errorf("%s setting not found in ApplicationSettings", setting)
	}

	valueStrings := util.DeleteEmptyAndTrim(strings.FieldsFunc(settingValue, util.SplitComma))

	return valueStrings, nil
}

// Initialize will parse command line flags, register for interrupts,
// initialize the logging system, and ingest configuration.
func (sdk *AppFunctionsSDK) Initialize() error {
	startupTimer := startup.NewStartUpTimer(internal.BootRetrySecondsDefault, internal.BootTimeoutSecondsDefault)

	additionalUsage :=
		"    -s/--skipVersionCheck           Indicates the service should skip the Core Service's version compatibility check.\n" +
			"    -sk/--serviceKey                Overrides the service service key used with Registry and/or Configuration Providers.\n" +
			"                                    If the name provided contains the text `<profile>`, this text will be replaced with\n" +
			"                                    the name of the profile used."

	sdkFlags := flags.NewWithUsage(additionalUsage)
	sdkFlags.FlagSet.BoolVar(&sdk.skipVersionCheck, "skipVersionCheck", false, "")
	sdkFlags.FlagSet.BoolVar(&sdk.skipVersionCheck, "s", false, "")
	sdkFlags.FlagSet.StringVar(&sdk.serviceKeyOverride, "serviceKey", "", "")
	sdkFlags.FlagSet.StringVar(&sdk.serviceKeyOverride, "sk", "", "")

	sdkFlags.Parse(os.Args[1:])

	// Temporarily setup logging to STDOUT so the client can be used before bootstrapping is completed
	sdk.LoggingClient = logger.NewClientStdOut(sdk.ServiceKey, false, "INFO")

	sdk.setServiceKey(sdkFlags.Profile())

	// The use of the edgex_service environment variable (only used for App Services) has been deprecated
	// and not included in the common bootstrap. Have to be handle here before calling into the common bootstrap
	// so proper overrides are set.
	// TODO: Remove for release v2.0.0
	if err := sdk.handleEdgexService(); err != nil {
		return err
	}

	sdk.config = &common.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return sdk.config
		},
	})

	sdk.appCtx, sdk.appCancelCtx = context.WithCancel(context.Background())
	sdk.appWg = &sync.WaitGroup{}

	var deferred bootstrap.Deferred
	var successful bool
	var configUpdated config.UpdatedStream = make(chan struct{})

	sdk.appWg, deferred, successful = bootstrap.RunAndReturnWaitGroup(
		sdk.appCtx,
		sdk.appCancelCtx,
		sdkFlags,
		sdk.ServiceKey,
		internal.ConfigRegistryStem,
		sdk.config,
		configUpdated,
		startupTimer,
		dic,
		[]bootstrapInterfaces.BootstrapHandler{
			handlers.NewSecrets().BootstrapHandler,
			handlers.NewDatabase().BootstrapHandler,
			handlers.NewClients().BootstrapHandler,
			handlers.NewTelemetry().BootstrapHandler,
			handlers.NewVersionValidator(sdk.skipVersionCheck, internal.SDKVersion).BootstrapHandler,
			message.NewBootstrap(sdk.ServiceKey, internal.ApplicationVersion).BootstrapHandler,
		},
	)

	// deferred is a a function that needs to be called when services exits.
	sdk.addDeferred(deferred)

	if !successful {
		return fmt.Errorf("boostrapping failed")
	}

	// Bootstrapping is complete, so now need to retrieve the needed objects from the containers.
	sdk.secretProvider = container.SecretProviderFrom(dic.Get)
	sdk.storeClient = container.StoreClientFrom(dic.Get)
	sdk.LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)
	sdk.edgexClients.LoggingClient = sdk.LoggingClient
	sdk.edgexClients.EventClient = container.EventClientFrom(dic.Get)
	sdk.edgexClients.ValueDescriptorClient = container.ValueDescriptorClientFrom(dic.Get)
	sdk.edgexClients.NotificationsClient = container.NotificationsClientFrom(dic.Get)
	sdk.edgexClients.CommandClient = container.CommandClientFrom(dic.Get)

	// We do special processing when the writeable section of the configuration changes, so have
	// to wait to be signaled when the configuration has been updated and then process the changes
	NewConfigUpdateProcessor(sdk).WaitForConfigUpdates(configUpdated)

	sdk.webserver = webserver.NewWebServer(sdk.config, sdk.secretProvider, sdk.LoggingClient, mux.NewRouter())
	sdk.webserver.ConfigureStandardRoutes()

	return nil
}

// GetSecrets retrieves secrets from a secret store.
// path specifies the type or location of the secrets to retrieve. If specified it is appended
// to the base path from the SecretConfig
// keys specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the
// specified path will be returned.
func (sdk *AppFunctionsSDK) GetSecrets(path string, keys ...string) (map[string]string, error) {
	return sdk.secretProvider.GetSecrets(path, keys...)
}

// StoreSecrets stores the secrets to a secret store.
// it sets the values requested at provided keys
// path specifies the type or location of the secrets to store. If specified it is appended
// to the base path from the SecretConfig
// secrets map specifies the "key": "value" pairs of secrets to store
func (sdk *AppFunctionsSDK) StoreSecrets(path string, secrets map[string]string) error {
	return sdk.secretProvider.StoreSecrets(path, secrets)
}

// setupTrigger configures the appropriate trigger as specified by configuration.
func (sdk *AppFunctionsSDK) setupTrigger(configuration *common.ConfigurationStruct, runtime *runtime.GolangRuntime) trigger.Trigger {
	var t trigger.Trigger
	// Need to make dynamic, search for the binding that is input

	switch strings.ToUpper(configuration.Binding.Type) {
	case "HTTP":
		sdk.LoggingClient.Info("HTTP trigger selected")
		t = &http.Trigger{Configuration: configuration, Runtime: runtime, Webserver: sdk.webserver, EdgeXClients: sdk.edgexClients}
	case "MESSAGEBUS":
		sdk.LoggingClient.Info("MessageBus trigger selected")
		t = &messagebus.Trigger{Configuration: configuration, Runtime: runtime, EdgeXClients: sdk.edgexClients}
	}

	return t
}

func (sdk *AppFunctionsSDK) addContext(next func(nethttp.ResponseWriter, *nethttp.Request)) func(nethttp.ResponseWriter, *nethttp.Request) {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		ctx := context.WithValue(r.Context(), SDKKey, sdk)
		next(w, r.WithContext(ctx))
	}
}

func (sdk *AppFunctionsSDK) addDeferred(deferred bootstrap.Deferred) {
	if deferred != nil {
		sdk.deferredFunctions = append(sdk.deferredFunctions, deferred)
	}
}

// setServiceKey creates the service's service key with profile name if the original service key has the
// appropriate profile placeholder, otherwise it leaves the original service key unchanged
func (sdk *AppFunctionsSDK) setServiceKey(profile string) {
	envValue := os.Getenv(envServiceKey)
	if len(envValue) > 0 {
		sdk.serviceKeyOverride = envValue
		sdk.LoggingClient.Info(
			fmt.Sprintf("Environment profileOverride of '-n/--serviceName' by environment variable: %s=%s",
				envServiceKey,
				envValue))
	}

	// serviceKeyOverride may have been set by the -n/--serviceName command-line option and not the environment variable
	if len(sdk.serviceKeyOverride) > 0 {
		sdk.ServiceKey = sdk.serviceKeyOverride
	}

	if !strings.Contains(sdk.ServiceKey, ProfileSuffixPlaceholder) {
		// No placeholder, so nothing to do here
		return
	}

	// Have to handle environment override here before common bootstrap is used so it is passed the proper service key
	profileOverride := os.Getenv(envProfile)
	if len(profileOverride) == 0 {
		// V2 not set so try V1
		profileOverride = os.Getenv(envV1Profile) // TODO: Remove for release v2.0.0:
	}

	if len(profileOverride) > 0 {
		profile = profileOverride
	}

	if len(profile) > 0 {
		sdk.ServiceKey = strings.Replace(sdk.ServiceKey, ProfileSuffixPlaceholder, profile, 1)
		return
	}

	// No profile specified so remove the placeholder text
	sdk.ServiceKey = strings.Replace(sdk.ServiceKey, ProfileSuffixPlaceholder, "", 1)
}

// handleEdgexService checks to see if the "edgex_service" environment variable is set and if so creates appropriate config
// overrides from the URL parts.
// TODO: Remove for release v2.0.0
func (sdk *AppFunctionsSDK) handleEdgexService() error {
	if envValue := os.Getenv(envV1Service); envValue != "" {
		u, err := url.Parse(envValue)
		if err != nil {
			return fmt.Errorf(
				"failed to parse 'edgex_service' environment value '%s' as a URL: %s",
				envValue,
				err.Error())
		}

		_, err = strconv.ParseInt(u.Port(), 10, 0)
		if err != nil {
			return fmt.Errorf(
				"failed to parse port from 'edgex_service' environment value '%s' as an integer: %s",
				envValue,
				err.Error())
		}

		os.Setenv(envServiceProtocol, u.Scheme)
		os.Setenv(envServiceHost, u.Hostname())
		os.Setenv(envServicePort, u.Port())
	}

	return nil
}
