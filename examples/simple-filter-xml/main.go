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

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
)

const (
	serviceKey = "sampleFilterXml"
)

var counter int = 0

func main() {
	// 1) First thing to do is to create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// 2) Since our FilterByDeviceID Function requires the list of DeviceID's we would
	// like to search for, we'll go ahead and define that now.
	deviceIDs := []string{"GS1-AC-Drive01"}

	// 3) This is our pipeline configuration, the collection of functions to
	// execute every time an event is triggered.
	edgexSdk.SetPipeline(
		edgexSdk.FilterByDeviceID(deviceIDs),
		edgexSdk.TransformToXML(),
		printXMLToConsole,
	)

	// 4) shows how to access the application's specific configuration settings.
	appSettings := edgexSdk.ApplicationSettings()
	if appSettings != nil {
		appName, ok := appSettings["ApplicationName"]
		if ok {
			edgexSdk.LoggingClient.Info(fmt.Sprintf("%s now running...", appName))
		} else {
			edgexSdk.LoggingClient.Error("ApplicationName application setting not found")
			os.Exit(-1)
		}
	} else {
		edgexSdk.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	// 5) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	edgexSdk.MakeItRun()
}

func printXMLToConsole(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}
	fmt.Println(params[0].(string))
	edgexcontext.Complete([]byte(params[0].(string)))
	return false, nil
}
