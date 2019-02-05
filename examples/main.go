package main

import (
	"github.com/edgexfoundry-holdings/app-functions-sdk-go/pkg/edgexsdk"
)

func main() {
	// 1) First thing to do is to create an instance of the EdgeX SDK.
	edgexsdk := &edgexsdk.AppFunctionsSDK{}

	// 2) Since our FilterByDeviceID Function requires the list of DeviceID's we would
	// like to search for, we'll go ahead and define that now.
	deviceIDs := []string{"GS1-AC-Drive01"}
	// 3) This is our pipeline configuration, the collection of functions to
	// execute everytime an event is triggered.
	edgexsdk.SetPipeline(
		edgexsdk.FilterByDeviceID(deviceIDs),
		edgexsdk.TransformToXML(),
		printXMLToConsole,
	)
	// 4) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	edgexsdk.MakeItRun()
}

func printXMLToConsole(params ...interface{}) interface{} {
	if len(params) < 1 {
		// We didn't receive a result
		return nil
	}
	println(params[0].(string))
	return nil
}
