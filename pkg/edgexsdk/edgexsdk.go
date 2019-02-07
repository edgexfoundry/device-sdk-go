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

package edgexsdk

import (
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/configuration"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/runtime"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger"
	httptrigger "github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger/http"
	messagebustrigger "github.com/edgexfoundry/app-functions-sdk-go/pkg/trigger/messagebus"
)

// AppFunctionsSDK ...
type AppFunctionsSDK struct {
	transforms []func(params ...interface{}) interface{}
}

// SetPipeline defines the order in which each function will be called as each event comes in.
func (afsdk *AppFunctionsSDK) SetPipeline(transforms ...func(params ...interface{}) interface{}) {
	afsdk.transforms = transforms
}

// FilterByDeviceID ...
func (afsdk *AppFunctionsSDK) FilterByDeviceID(deviceIDs []string) func(...interface{}) interface{} {
	transforms := transforms.Filter{
		DeviceIDs: deviceIDs,
	}
	return transforms.FilterByDeviceID
}

// FilterByValueDescriptor ...
func (afsdk *AppFunctionsSDK) FilterByValueDescriptor() {

}

// TransformToXML ...
func (afsdk *AppFunctionsSDK) TransformToXML() func(...interface{}) interface{} {
	transforms := transforms.Conversion{}
	return transforms.TransformToXML
}

//MakeItRun the SDK
func (afsdk *AppFunctionsSDK) MakeItRun() {

	// load the configuration
	configuration := configuration.LoadConfiguration()

	// a little telemetry where?

	//determine which runtime to load
	runtime := runtime.GolangRuntime{Transforms: afsdk.transforms}

	// determine input type and create trigger for it
	trigger := afsdk.setupTrigger(configuration, runtime)

	// Initialize the trigger (i.e. start a web server, or connect to message bus)
	trigger.Initialize()

}

func (afsdk *AppFunctionsSDK) setupTrigger(configuration configuration.Configuration, runtime runtime.GolangRuntime) trigger.ITrigger {
	var trigger trigger.ITrigger
	// Need to make dynamic, search for the binding that is input
	switch configuration.Bindings[0].Type {
	case "http":
		trigger = &httptrigger.HTTPTrigger{Configuration: configuration, Runtime: runtime}
	case "messageBus":
		trigger = &messagebustrigger.MessageBusTrigger{Configuration: configuration, Runtime: runtime}
	}
	return trigger
}
