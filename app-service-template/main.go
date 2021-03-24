// TODO: Change Copyright to your company if open sourcing or remove header
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

package main

import (
	"os"
	"reflect"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	"new-app-service/config"

	"new-app-service/functions"
)

const (
	serviceKey = "new-app-service"
)

// TODO: Define your app's struct
type myApp struct {
	service       interfaces.ApplicationService
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
	configChanged chan bool
}

func main() {
	// TODO: See https://docs.edgexfoundry.org/2.0/microservices/application/ApplicationServices/
	//       for documentation on application services.
	app := myApp{}
	code := app.CreateAndRunAppService(serviceKey, pkg.NewAppService)
	os.Exit(code)
}

// CreateAndRunAppService wraps what would normally be in main() so that it can be unit tested
// TODO: Remove and just use regular main() if unit tests of main logic not needed.
func (app *myApp) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = newServiceFactory(serviceKey)
	if !ok {
		return -1
	}

	app.lc = app.service.LoggingClient()

	// TODO: Replace with retrieving your custom ApplicationSettings from configuration or
	//       remove if not using AppSetting configuration section.
	// For more details see https://docs.edgexfoundry.org/2.0/microservices/application/GeneralAppServiceConfig/#application-settings
	deviceNames, err := app.service.GetAppSettingStrings("DeviceNames")
	if err != nil {
		app.lc.Errorf("failed to retrieve DeviceNames from configuration: %s", err.Error())
		return -1
	}

	// More advance custom structured configuration can be defined and loaded as in this example.
	// For more details see https://docs.edgexfoundry.org/2.0/microservices/application/GeneralAppServiceConfig/#custom-configuration
	// TODO: Change to use your service's custom configuration struct
	//       or remove if not using custom configuration capability
	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "AppCustom"); err != nil {
		app.lc.Errorf("failed load custom configuration: %s", err.Error())
		return -1
	}

	// Optionally validate the custom configuration after it is loaded.
	// TODO: remove if you don't have custom configuration or don't need to validate it
	if err := app.serviceConfig.AppCustom.Validate(); err != nil {
		app.lc.Errorf("custom configuration failed validation: %s", err.Error())
		return -1
	}

	// Custom configuration can be 'writable' or a section of the configuration can be 'writable' when using
	// the Configuration Provider, aka Consul.
	// For more details see https://docs.edgexfoundry.org/2.0/microservices/application/GeneralAppServiceConfig/#writable-custom-configuration
	// TODO: Remove if not using writable custom configuration
	if err := app.service.ListenForCustomConfigChanges(&app.serviceConfig.AppCustom, "AppCustom", app.ProcessConfigUpdates); err != nil {
		app.lc.Errorf("unable to watch custom writable configuration: %s", err.Error())
		return -1
	}

	// TODO: Replace below functions with built in and/or your custom functions for your use case.
	//       See https://docs.edgexfoundry.org/2.0/microservices/application/BuiltIn/ for list of built-in functions
	sample := functions.NewSample()
	err = app.service.SetFunctionsPipeline(
		transforms.NewFilterFor(deviceNames).FilterByDeviceName,
		sample.LogEventDetails,
		sample.ConvertEventToXML,
		sample.OutputXML)
	if err != nil {
		app.lc.Errorf("SetFunctionsPipeline returned error: %s", err.Error())
		return -1
	}

	if err := app.service.MakeItRun(); err != nil {
		app.lc.Errorf("MakeItRun returned error: %s", err.Error())
		return -1
	}

	// TODO: Do any required cleanup here, if needed

	return 0
}

// TODO: Update using your Custom configuration 'writeable' type or remove if not using ListenForCustomConfigChanges
// ProcessConfigUpdates processes the updated configuration for the service's writable configuration.
// At a minimum it must copy the updated configuration into the service's current configuration. Then it can
// do any special processing for changes that require more.
func (app *myApp) ProcessConfigUpdates(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*config.AppCustomConfig)
	if !ok {
		app.lc.Error("unable to process config updates: Can not cast raw config to type 'AppCustomConfig'")
		return
	}

	previous := app.serviceConfig.AppCustom
	app.serviceConfig.AppCustom = *updated

	if reflect.DeepEqual(previous, updated) {
		app.lc.Info("No changes detected")
		return
	}

	if previous.SomeValue != updated.SomeValue {
		app.lc.Infof("AppCustom.SomeValue changed to: %d", updated.SomeValue)
	}
	if previous.ResourceNames != updated.ResourceNames {
		app.lc.Infof("AppCustom.ResourceNames changed to: %s", updated.ResourceNames)
	}
	if !reflect.DeepEqual(previous.SomeService, updated.SomeService) {
		app.lc.Infof("AppCustom.SomeService changed to: %v", updated.SomeService)
	}
}
