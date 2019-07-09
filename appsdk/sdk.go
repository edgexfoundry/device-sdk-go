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
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/runtime"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/telemetry"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/http"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/trigger/messagebus"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/webserver"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	coreTypes "github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	registryTypes "github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

// AppFunctionsSDK provides the necessary struct to create an instance of the Application Functions SDK. Be sure and provide a ServiceKey
// when creating an instance of the SDK. After creating an instance, you'll first want to call .Initialize(), to start up the SDK. Secondly,
// provide the desired transforms for your pipeline by calling .SetFunctionsPipeline(). Lastly, call .MakeItRun() to start listening for events based on
// your configured trigger.
type AppFunctionsSDK struct {
	transforms     []func(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{})
	ServiceKey     string
	configProfile  string
	configDir      string
	useRegistry    bool
	httpErrors     chan error
	webserver      *webserver.WebServer
	registryClient registry.Client
	eventClient    coredata.EventClient
	config         common.ConfigurationStruct
	LoggingClient  logger.LoggingClient
}

// MakeItRun will initialize and start the trigger as specifed in the
// configuration. It will also configure the webserver and start listening on
// the specified port.
func (sdk *AppFunctionsSDK) MakeItRun() error {
	httpErrors := make(chan error)
	defer close(httpErrors)

	runtime := runtime.GolangRuntime{Transforms: sdk.transforms}

	sdk.webserver = &webserver.WebServer{
		Config:        &sdk.config,
		LoggingClient: sdk.LoggingClient,
	}
	sdk.webserver.ConfigureStandardRoutes()

	// determine input type and create trigger for it
	trigger := sdk.setupTrigger(sdk.config, runtime)

	// Initialize the trigger (i.e. start a web server, or connect to message bus)
	err := trigger.Initialize(sdk.LoggingClient)
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

// ApplicationSettings returns the values specifed in the custom configuration section.
func (sdk *AppFunctionsSDK) ApplicationSettings() map[string]string {
	return sdk.config.ApplicationSettings
}

// setupTrigger configures the appropriate trigger as specified by configuration.
func (sdk *AppFunctionsSDK) setupTrigger(configuration common.ConfigurationStruct, runtime runtime.GolangRuntime) trigger.Trigger {
	var trigger trigger.Trigger
	// Need to make dynamic, search for the binding that is input

	switch strings.ToUpper(configuration.Binding.Type) {
	case "HTTP":
		sdk.LoggingClient.Info("HTTP trigger selected")
		trigger = &http.Trigger{Configuration: configuration, Runtime: runtime, Webserver: sdk.webserver, EventClient: sdk.eventClient}
	case "MESSAGEBUS":
		sdk.LoggingClient.Info("MessageBus trigger selected")
		trigger = &messagebus.Trigger{Configuration: configuration, Runtime: runtime, EventClient: sdk.eventClient}
	}

	return trigger
}

// Initialize will parse command line flags, register for interrupts,
// initalize the logging system, and ingest configuration.
func (sdk *AppFunctionsSDK) Initialize() error {

	flag.BoolVar(&sdk.useRegistry, "registry", false, "Indicates the service should use the registry.")
	flag.BoolVar(&sdk.useRegistry, "r", false, "Indicates the service should use registry.")

	flag.StringVar(&sdk.configProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&sdk.configProfile, "p", "", "Specify a profile other than default.")

	flag.StringVar(&sdk.configDir, "confdir", "", "Specify an alternate configuration directory.")
	flag.StringVar(&sdk.configDir, "c", "", "Specify an alternate configuration directory.")

	flag.Parse()

	now := time.Now()
	until := now.Add(time.Millisecond * time.Duration(internal.BootTimeoutDefault))
	for now.Before(until) {
		err := sdk.initializeConfiguration()
		if err != nil {
			fmt.Printf("failed to initialize Registry: %v\n", err)
		} else {
			//initialize logger
			sdk.LoggingClient = logger.NewClient("AppFunctionsSDK", sdk.config.Logging.EnableRemote, sdk.config.Logging.File, sdk.config.Writable.LogLevel)
			sdk.LoggingClient.Info("Configuration and logger successfully initialized")
			break
		}

		time.Sleep(time.Second * time.Duration(1))
	}

	if sdk.useRegistry {
		go sdk.listenForConfigChanges()
	}
	//Setup eventClient
	params := coreTypes.EndpointParams{
		ServiceKey:  clients.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: sdk.useRegistry,
		Url:         sdk.config.Clients["CoreData"].Url() + clients.ApiEventRoute,
		Interval:    sdk.config.Service.ClientMonitor,
	}
	sdk.eventClient = coredata.NewEventClient(params, startup.Endpoint{RegistryClient: &sdk.registryClient})

	go telemetry.StartCpuUsageAverage()

	return nil
}

func (sdk *AppFunctionsSDK) initializeConfiguration() error {

	// Currently have to load configuration from filesystem first in order to obtain Registry Host/Port
	configuration := &common.ConfigurationStruct{}
	err := common.LoadFromFile(sdk.configProfile, sdk.configDir, configuration)
	if err != nil {
		return err
	}
	sdk.config = *configuration

	if sdk.useRegistry {
		registryConfig := registryTypes.Config{
			Host:          sdk.config.Registry.Host,
			Port:          sdk.config.Registry.Port,
			Type:          sdk.config.Registry.Type,
			Stem:          internal.ConfigRegistryStem,
			CheckInterval: "1s",
			CheckRoute:    internal.ApiPingRoute,
			ServiceKey:    sdk.ServiceKey,
			ServiceHost:   sdk.config.Service.Host,
			ServicePort:   sdk.config.Service.Port,
		}

		client, err := registry.NewRegistryClient(registryConfig)
		if err != nil {
			return fmt.Errorf("connection to Registry could not be made: %v", err)
		}
		//set registryClient
		sdk.registryClient = client

		if !sdk.registryClient.IsAlive() {
			return fmt.Errorf("registry (%s) is not running", registryConfig.Type)
		}

		// Register the service with Registry
		err = sdk.registryClient.Register()
		if err != nil {
			return fmt.Errorf("could not register service with Registry: %v", err)
		}

		hasConfig, err := sdk.registryClient.HasConfiguration()
		if err != nil {
			return fmt.Errorf("could not determine if registry has configuration: %v", err)
		}

		if hasConfig {
			rawConfig, err := sdk.registryClient.GetConfiguration(configuration)
			if err != nil {
				return fmt.Errorf("could not get configuration from Registry: %v", err)
			}

			actual, ok := rawConfig.(*common.ConfigurationStruct)
			if !ok {
				return fmt.Errorf("configuration from Registry failed type check")
			}

			sdk.config = *actual
			//Check that information was successfully read from Consul
			if sdk.config.Service.Port == 0 {
				sdk.LoggingClient.Error("Error reading from registry")
			}

			fmt.Println("Configuration loaded from registry")
		} else {
			err := sdk.registryClient.PutConfiguration(sdk.config, true)
			if err != nil {
				return fmt.Errorf("could not push configuration into registry: %v", err)
			}
			fmt.Println("Configuration pushed to registry")
		}

	}

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

			sdk.config.Writable = *actual

			sdk.LoggingClient.Info("Writeable configuration has been updated from Registry")
			sdk.LoggingClient.SetLogLevel(sdk.config.Writable.LogLevel)

			// TODO: Deal with pub/sub topics may have changed. Save copy of writeable so that we can determine what if anything changed?
		}
	}
}
