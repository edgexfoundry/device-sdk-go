# App Functions SDK (Golang) - WORK IN PROGRESS

Welcome the App Functions SDK for EdgeX. This sdk is meant to provide all the plumbing necessary for developers to get started in processing/transforming/exporting data out of EdgeX. 

## Getting Started

The SDK is built around the idea of a "Pipeline". A pipeline is a collection of various functions that process the data in the order that you've specified. Each pipeline is triggered by every EdgeX CoreData Event provided. The first function of each pipeline is called with the event that triggered the pipeline (ex. `events.Model`). Each successive call in the pipeline is called with the return result of the previous function. Let's take a look at a simple example that creates a pipeline to filter particular device ids and subsequently transform the data to XML:
```golang
package main

import (
	"fmt"
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
	)
	// 4) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	edgexsdk.MakeItRun()
}
```

The above example is meant to merely demonstrate the structure of your application. Notice that the output of the last function is not available anywhere inside this application. You must provide a function to in order to work with the data from the previous function. Let's go ahead and add the following function that prints the output to the console.

```golang
func printXMLToConsole(params ...interface{}) interface{} {
	if len(params) < 1 {
		// We didn't receive a result
		return nil
	}
	println(params[0].(string))
	return nil
}
```
After placing the above function in your code, the next step is to modify the pipeline to call this function:
```golang
    edgexsdk.SetPipeline(
	edgexsdk.FilterByDeviceID(deviceIDs),
        edgexsdk.TransformToXML(),
        printXMLToConsole //notice this is not a function call, but simply a function pointer. 
    )
```
After making the above modifications you should now see data printing out to the console in XML.

## Built-In Transforms/Functions 

## Filtering
There are two basic types of filtering included in the SDK to add to your pipeline.
 - `FilterByDeviceId([]string deviceIDs)`
 - `FilterByValueDescriptor`

## Conversion
 There are two primary conversions available in the SDK that can be added to your pipeline. 
 
 - `TransformToXML()`: 
 - `TransformToJSON()`:

