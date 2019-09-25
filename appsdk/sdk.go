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

package appsdk

import (
	syscontext "context"
	"errors"
	"flag"
	"fmt"
	nethttp "net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pelletier/go-toml"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/config"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/telemetry"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/webserver"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/startup"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/command"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
	coreTypes "github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

// ProfileSuffixPlaceholder is used to create unique names for profiles
const ProfileSuffixPlaceholder = "<profile>"

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
	configProfile             string
	configDir                 string
	useRegistry               bool
	usingConfigurablePipeline bool
	httpErrors                chan error
	runtime                   *runtime.GolangRuntime
	webserver                 *webserver.WebServer
	edgexClients              common.EdgeXClients
	registryClient            registry.Client
	config                    common.ConfigurationStruct
}

// AddRoute allows you to leverage the existing webserver to add routes.
func (sdk *AppFunctionsSDK) AddRoute(route string, handler func(nethttp.ResponseWriter, *nethttp.Request), methods ...string) error {
	if route == clients.ApiPingRoute ||
		route == clients.ApiConfigRoute ||
		route == clients.ApiMetricsRoute ||
		route == clients.ApiVersionRoute ||
		route == internal.ApiTriggerRoute {
		return errors.New("Route is reserved")
	}
	sdk.webserver.AddRoute(route, sdk.addContext(handler), methods...)
	return nil
}
func (sdk *AppFunctionsSDK) addContext(next func(nethttp.ResponseWriter, *nethttp.Request)) func(nethttp.ResponseWriter, *nethttp.Request) {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		ctx := syscontext.WithValue(r.Context(), SDKKey, sdk)
		next(w, r.WithContext(ctx))
	})
}

// MakeItRun will initialize and start the trigger as specifed in the
// configuration. It will also configure the webserver and start listening on
// the specified port.
func (sdk *AppFunctionsSDK) MakeItRun() error {
	httpErrors := make(chan error)
	defer close(httpErrors)

	sdk.runtime = &runtime.GolangRuntime{TargetType: sdk.TargetType} //Transforms: sdk.transforms
	sdk.runtime.SetTransforms(sdk.transforms)

	// determine input type and create trigger for it
	trigger := sdk.setupTrigger(sdk.config, sdk.runtime)

	// Initialize the trigger (i.e. start a web server, or connect to message bus)
	err := trigger.Initialize()
	if err != nil {
		sdk.LoggingClient.Error(err.Error())
	}

	sdk.LoggingClient.Info(sdk.config.Service.StartupMsg)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sdk.webserver.StartHTTPServer(sdk.httpErrors)

	select {
	case httpError := <-sdk.httpErrors:
		sdk.LoggingClient.Info("Terminating: ", httpError.Error())
		return httpError

	case signalReceived := <-signals:
		sdk.LoggingClient.Info("Terminating: " + signalReceived.String())

	}

	return nil
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
		return nil, errors.New("Execution Order has 0 functions specified. You must have a least one function in the pipeline")
	}
	sdk.LoggingClient.Debug("Execution Order", "Functions", strings.Join(executionOrder, ","))

	for _, functionName := range executionOrder {
		functionName = strings.TrimSpace(functionName)
		configuration, ok := pipelineConfig.Functions[functionName]
		if !ok {
			return nil, fmt.Errorf("Function %s configuration not found in Pipeline.Functions section", functionName)
		}

		result := valueOfType.MethodByName(functionName)
		if result.Kind() == reflect.Invalid {
			return nil, fmt.Errorf("Function %s is not a built in SDK function", functionName)
		} else if result.IsNil() {
			return nil, fmt.Errorf("Invalid/Missing configuration for %s", functionName)
		}

		//determine number of parameters required for function call
		inputParameters := make([]reflect.Value, result.Type().NumIn())
		//set keys to be all lowercase to avoid casing issues from configuration
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
				return nil, fmt.Errorf("Function %s has an unsupported parameter type: %s", functionName, parameter.String())
			}
		}

		function, ok := result.Call(inputParameters)[0].Interface().(appcontext.AppFunction)
		if !ok {
			return nil, fmt.Errorf("Failed to cast function %s as AppFunction type", functionName)
		}
		pipeline = append(pipeline, function)
		configurable.Sdk.LoggingClient.Debug(fmt.Sprintf("%s function added to configurable pipeline", functionName))
	}

	return pipeline, nil
}

//
// SetFunctionsPipeline allows you to define each fgitunction to execute and the order in which each function
// will be called as each event comes in.
func (sdk *AppFunctionsSDK) SetFunctionsPipeline(transforms ...appcontext.AppFunction) error {
	if len(transforms) == 0 {
		return errors.New("No transforms provided to pipeline")
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

// setupTrigger configures the appropriate trigger as specified by configuration.
func (sdk *AppFunctionsSDK) setupTrigger(configuration common.ConfigurationStruct, runtime *runtime.GolangRuntime) trigger.Trigger {
	var trigger trigger.Trigger
	// Need to make dynamic, search for the binding that is input

	switch strings.ToUpper(configuration.Binding.Type) {
	case "HTTP":
		sdk.LoggingClient.Info("HTTP trigger selected")
		trigger = &http.Trigger{Configuration: configuration, Runtime: runtime, Webserver: sdk.webserver, EdgeXClients: sdk.edgexClients}
	case "MESSAGEBUS":
		sdk.LoggingClient.Info("MessageBus trigger selected")
		trigger = &messagebus.Trigger{Configuration: configuration, Runtime: runtime, EdgeXClients: sdk.edgexClients}
	}

	return trigger
}

// Initialize will parse command line flags, register for interrupts,
// initialize the logging system, and ingest configuration.
func (sdk *AppFunctionsSDK) Initialize() error {

	flag.BoolVar(&sdk.useRegistry, "registry", false, "Indicates the service should use the registry.")
	flag.BoolVar(&sdk.useRegistry, "r", false, "Indicates the service should use registry.")

	flag.StringVar(&sdk.configProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&sdk.configProfile, "p", "", "Specify a profile other than default.")

	flag.StringVar(&sdk.configDir, "confdir", "", "Specify an alternate configuration directory.")
	flag.StringVar(&sdk.configDir, "c", "", "Specify an alternate configuration directory.")

	flag.Parse()

	// Service keys must be unique. If an executable is run multiple times, it must have a different
	// profile for each instance, thus adding the profile to the base key will make it unique.
	// This requires services that are expected to have multiple instances running, such as the Configurable App Service,
	// add the ProfileSuffixPlaceholder placeholder in the service key.
	//
	// The Dockerfile must also take this into account and set the profile appropriately, i.e. not just "docker"
	//

	if strings.Contains(sdk.ServiceKey, ProfileSuffixPlaceholder) {
		if sdk.configProfile == "" {
			sdk.ServiceKey = strings.Replace(sdk.ServiceKey, ProfileSuffixPlaceholder, "", 1)
		} else {
			sdk.ServiceKey = strings.Replace(sdk.ServiceKey, ProfileSuffixPlaceholder, "-"+sdk.configProfile, 1)
		}
	}

	loggerInitialized := false
	configurationInitialized := false
	bootstrapComplete := false

	// Bootstrap retry loop to ensure all dependencies are ready before continuing.
	until := time.Now().Add(time.Millisecond * time.Duration(internal.BootTimeoutDefault))
	for time.Now().Before(until) {
		if !configurationInitialized {
			err := sdk.initializeConfiguration()
			if err != nil {
				fmt.Printf("failed to initialize Registry: %v\n", err)
				goto ContinueWithSleep
			}
			configurationInitialized = true
			fmt.Printf("Configuration & Registry initialized")
		}

		if !loggerInitialized {
			loggingTarget, err := sdk.setLoggingTarget()
			if err != nil {
				fmt.Printf("logger initialization failed: %v", err)
				goto ContinueWithSleep
			}

			sdk.LoggingClient = logger.NewClient(sdk.ServiceKey, sdk.config.Logging.EnableRemote, loggingTarget, sdk.config.Writable.LogLevel)
			sdk.LoggingClient.Info("Configuration and logger successfully initialized")
			sdk.edgexClients.LoggingClient = sdk.LoggingClient
			loggerInitialized = true
		}

		sdk.initializeClients()
		sdk.LoggingClient.Info("Clients initialized")
		bootstrapComplete = true
		break

	ContinueWithSleep:
		time.Sleep(time.Second * time.Duration(1))
	}

	if !bootstrapComplete {
		return fmt.Errorf("bootstrap retry timed out")
	}

	if sdk.useRegistry {
		go sdk.listenForConfigChanges()
	}

	sdk.initializeClients()

	go telemetry.StartCpuUsageAverage()

	sdk.webserver = webserver.NewWebServer(&sdk.config, sdk.LoggingClient, mux.NewRouter())
	sdk.webserver.ConfigureStandardRoutes()

	return nil
}

func (sdk *AppFunctionsSDK) initializeClients() {
	// Need when passing all Clients to other components
	sdk.edgexClients.LoggingClient = sdk.LoggingClient

	// Use of these client interfaces is optional, so they are not required to be configured. For instance if not
	// sending commands, then don't need to have the Command client in the configuration.
	if _, ok := sdk.config.Clients[common.CoreDataClientName]; ok {
		params := sdk.getClientParams(clients.CoreDataServiceKey, common.CoreDataClientName, clients.ApiEventRoute)
		sdk.edgexClients.EventClient = coredata.NewEventClient(params, startup.Endpoint{RegistryClient: &sdk.registryClient})

		params = sdk.getClientParams(clients.CoreDataServiceKey, common.CoreDataClientName, clients.ApiValueDescriptorRoute)
		sdk.edgexClients.ValueDescriptorClient = coredata.NewValueDescriptorClient(params, startup.Endpoint{RegistryClient: &sdk.registryClient})
	}

	if _, ok := sdk.config.Clients[common.CoreCommandClientName]; ok {
		params := sdk.getClientParams(clients.CoreCommandServiceKey, common.CoreCommandClientName, clients.ApiDeviceRoute)
		sdk.edgexClients.CommandClient = command.NewCommandClient(params, startup.Endpoint{RegistryClient: &sdk.registryClient})
	}

	if _, ok := sdk.config.Clients[common.NotificationsClientName]; ok {
		params := sdk.getClientParams(clients.SupportNotificationsServiceKey, common.NotificationsClientName, clients.ApiNotificationRoute)
		sdk.edgexClients.NotificationsClient = notifications.NewNotificationsClient(params, startup.Endpoint{RegistryClient: &sdk.registryClient})
	}
}

func (sdk *AppFunctionsSDK) getClientParams(serviceKey string, clientName string, route string) coreTypes.EndpointParams {
	return coreTypes.EndpointParams{
		ServiceKey:  serviceKey,
		Path:        route,
		UseRegistry: sdk.useRegistry,
		Url:         sdk.config.Clients[clientName].Url() + route,
		Interval:    sdk.config.Service.ClientMonitor,
	}
}

func (sdk *AppFunctionsSDK) initializeConfiguration() error {

	// Currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	configuration, err := common.LoadFromFile(sdk.configProfile, sdk.configDir)
	if err != nil {
		return err
	}

	if sdk.useRegistry {
		e := config.NewEnvironment()
		configuration.Registry = e.OverrideRegistryInfoFromEnvironment(configuration.Registry)
		configuration.Service = e.OverrideServiceInfoFromEnvironment(configuration.Service)

		registryConfig := registryTypes.Config{
			Host:            configuration.Registry.Host,
			Port:            configuration.Registry.Port,
			Type:            configuration.Registry.Type,
			Stem:            internal.ConfigRegistryStem,
			CheckInterval:   "1s",
			CheckRoute:      clients.ApiPingRoute,
			ServiceKey:      sdk.ServiceKey,
			ServiceHost:     configuration.Service.Host,
			ServicePort:     configuration.Service.Port,
			ServiceProtocol: configuration.Service.Protocol,
		}

		client, err := registry.NewRegistryClient(registryConfig)
		if err != nil {
			return fmt.Errorf("Connection to Registry could not be made: %v", err)
		}

		//set registryClient
		sdk.registryClient = client

		if !sdk.registryClient.IsAlive() {
			return fmt.Errorf("Registry (%s) is not running", registryConfig.Type)
		}

		hasConfig, err := sdk.registryClient.HasConfiguration()
		if err != nil {
			return fmt.Errorf("Could not determine if registry has configuration: %v", err)
		}

		if hasConfig {
			rawConfig, err := sdk.registryClient.GetConfiguration(configuration)
			if err != nil {
				return fmt.Errorf("Could not get configuration from Registry: %v", err)
			}

			actual, ok := rawConfig.(*common.ConfigurationStruct)
			if !ok {
				return fmt.Errorf("Configuration from Registry failed type check")
			}
			configuration = actual

			//Check that information was successfully read from Consul
			if configuration.Service.Port == 0 {
				sdk.LoggingClient.Error("Error reading from registry")
			}

			fmt.Println("Configuration loaded from registry with service key: " + sdk.ServiceKey)
		} else {
			// Marshal into a toml Tree for overriding with environment variables.
			contents, err := toml.Marshal(*configuration)
			if err != nil {
				return err
			}
			configTree, err := toml.LoadBytes(contents)
			if err != nil {
				return err
			}

			err = sdk.registryClient.PutConfigurationToml(e.OverrideFromEnvironment(configTree), true)
			if err != nil {
				return fmt.Errorf("could not push configuration into registry: %v", err)
			}
			err = configTree.Unmarshal(configuration)
			if err != nil {
				return fmt.Errorf("could not marshal configTree to configuration: %v", err.Error())
			}
			fmt.Println("Configuration pushed to registry with service key: " + sdk.ServiceKey)
		}

		// Register the service with Registry
		err = sdk.registryClient.Register()
		if err != nil {
			return fmt.Errorf("Could not register service with Registry: %v", err)
		}
	}

	sdk.config = *configuration
	return nil
}

func (sdk *AppFunctionsSDK) listenForConfigChanges() {

	updates := make(chan interface{})
	registryErrors := make(chan error)

	defer close(updates)

	sdk.LoggingClient.Info("Listening for changes from registry")
	sdk.registryClient.WatchForChanges(updates, registryErrors, &common.WritableInfo{}, internal.WritableKey)

	for {
		select {
		case err := <-registryErrors:
			sdk.LoggingClient.Error(err.Error())

		case raw, ok := <-updates:
			if !ok {
				sdk.LoggingClient.Error("Failed to receive changes from update channel")
				return
			}

			actual, ok := raw.(*common.WritableInfo)
			if !ok {
				sdk.LoggingClient.Error("listenForConfigChanges() type check failed")
				return
			}

			previousLogLevel := sdk.config.Writable.LogLevel

			sdk.config.Writable = *actual
			sdk.LoggingClient.SetLogLevel(sdk.config.Writable.LogLevel)
			sdk.LoggingClient.Info("Writable configuration has been updated from Registry")

			if previousLogLevel != sdk.config.Writable.LogLevel {
				// Log level changed, not Pipeline, so skip updating the pipeline
				continue
			}

			if sdk.usingConfigurablePipeline {
				transforms, err := sdk.LoadConfigurablePipeline()
				if err != nil {
					sdk.LoggingClient.Error("unable to reload Configurable Pipeline from Registry: " + err.Error())
					continue
				}
				err = sdk.SetFunctionsPipeline(transforms...)
				if err != nil {
					sdk.LoggingClient.Error("unable to set Configurable Pipeline from Registry: " + err.Error())
					continue
				}

				sdk.LoggingClient.Info("ReLoaded Configurable Pipeline from Registry")
			}

			// TODO: Deal with pub/sub topics may have changed. Save copy of writeable so that we can determine what if anything changed?
		}
	}

}

func (sdk *AppFunctionsSDK) setLoggingTarget() (string, error) {
	if sdk.config.Logging.EnableRemote {
		logging, ok := sdk.config.Clients[common.LoggingClientName]
		if !ok {
			return "", errors.New("logging client configuration is missing")
		}

		return logging.Url() + clients.ApiLoggingRoute, nil
	}

	return sdk.config.Logging.File, nil
}
