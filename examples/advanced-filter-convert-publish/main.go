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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/examples/advanced-filter-convert-publish/functions"
)

const (
	serviceKey = "advancedFilterConvertPublish"
)

func main() {
	// 1) First thing to do is to create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// 2) shows how to access the application's specific configuration settings.
	appSettings := edgexSdk.ApplicationSettings()
	if appSettings == nil {
		edgexSdk.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	appName, ok := appSettings["ApplicationName"]
	if ok {
		edgexSdk.LoggingClient.Info(fmt.Sprintf("%s now running...", appName))
	} else {
		edgexSdk.LoggingClient.Error("ApplicationName application setting not found")
		os.Exit(-1)
	}

	valueDescriptorList, ok := appSettings["ValueDescriptors"]
	if !ok {
		edgexSdk.LoggingClient.Error("ValueDescriptors application setting not found")
		os.Exit(-1)
	}

	// 3) Since our FilterByValueDescriptor Function requires the list of ValueDescriptor's we would
	// like to search for, we'll go ahead create that list from the corresponding configuration setting.
	valueDescriptorList = strings.Replace(valueDescriptorList, " ", "", -1)
	valueDescriptors := strings.Split(valueDescriptorList, ",")
	edgexSdk.LoggingClient.Info(fmt.Sprintf("Filtering for %v value descriptors...", valueDescriptors))

	// 4) This is our functions pipeline configuration, the collection of functions to
	// execute every time an event is triggered.
	err := edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(valueDescriptors).FilterByValueDescriptor,
		functions.ConvertToReadableFloatValues,
		functions.PrintFloatValuesToConsole,
		functions.Publish,
	)

	if err != nil {
		edgexSdk.LoggingClient.Error("SDK initialization failed: " + err.Error())
		os.Exit(-1)
	}

	// 5) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}
