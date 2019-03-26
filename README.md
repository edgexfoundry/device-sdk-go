# App Functions SDK (Golang) - WORK IN PROGRESS

Welcome the App Functions SDK for EdgeX. This sdk is meant to provide all the plumbing necessary for developers to get started in processing/transforming/exporting data out of EdgeX. 

## Getting Started

The SDK is built around the idea of a "Pipeline". A pipeline is a collection of various functions that process the data in the order that you've specified. The pipeline is executed by the specified [trigger](#triggers) in the `configuration.toml` . The first function of each pipeline is called with the event that triggered the pipeline (ex. `events.Model`). Each successive call in the pipeline is called with the return result of the previous function. Let's take a look at a simple example that creates a pipeline to filter particular device ids and subsequently transform the data to XML:
```golang
package main

import (
  "fmt"
  "github.com/edgexfoundry/app-functions-sdk-go/edgexsdk"
  "github.com/edgexfoundry/app-functions-sdk-go/excontext"
)
func main() {

  // 1) First thing to do is to create an instance of the EdgeX SDK, giving it a service key
  edgexsdk := &edgexsdk.AppFunctionsSDK{
    ServiceKey: "SimpleFilterXMLApp" // Key used by Consul
  }
  // 2) Next, we need to Initilize the SDK
  if err := edgexSdk.Initialize(); err != nil {
		// TODO: Log rather than print
		fmt.Printf("SDK initialization failed: %v\n", err)
		os.Exit(-1)
	}
  // 2) Since our FilterByDeviceID Function requires the list of DeviceID's we would
  // like to search for, we'll go ahead and define that now.
  deviceIDs := []string{"GS1-AC-Drive01"}
  // 3) This is our pipeline configuration, the collection of functions to
  // execute everytime an event is triggered.
  edgexsdk.SetPipeline(
  	edgexsdk.FilterByDeviceID(deviceIDs),
  	edgexsdk.TransformToXML(),
  )
  
  // 4) shows how to access the application's specific configuration settings.
  appSettings:= edgexSdk.ApplicationSettings()
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
  
  // 5) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events  // to trigger the pipeline.
  edgexsdk.MakeItRun()
}
```

The above example is meant to merely demonstrate the structure of your application. Notice that the output of the last function is not available anywhere inside this application. You must provide a function in order to work with the data from the previous function. Let's go ahead and add the following function that prints the output to the console.

```golang
func printXMLToConsole(edgexcontext *excontext.Context, params ...interface{}) (bool,interface{}) {
  if len(params) < 1 { 
  	// We didn't receive a result
  	return false, errors.New("No Data Received")
  }
  println(params[0].(string))
  return true, nil
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
After making the above modifications, you should now see data printing out to the console in XML when an event is triggered.
> You can find this example in the `/examples` directory located in this repository. You can also use the provided `EdgeX Applications Function SDK.postman_collection.json" file to load into postman to trigger the sample pipeline.

Up until this point, the pipeline has been [triggered](#triggers) by an event over HTTP and the data at the end of that pipeline lands in the last function specified. In the example, data ends up printed to the console. Perhaps we'd like to send the data back to where it came from. In the case of an HTTP trigger, this would be the HTTP response. In the case of a message bus, this could be a new topic to send the data back to for other applications that wish to receive it. To do this, simply call `edgexcontext.Complete([]byte outputData)` passing in the data you wish to "respond" with. In the above `printXMLToConsole(...)` function, replace `println(params[0].(string))` with `edgexcontext.Complete([]byte(params[0].(string)))`. You should now see the response in your postman window when testing the pipeline.

## Triggers

Triggers determine how the the app functions pipeline begins execution. In the simple example provided above, an HTTP trigger is used. The trigger is determine by the `configuration.toml` file located in the `/res` directory under a section called `[Binding]`. Check out the [Configuration Section](#configuration) for more information about the toml file.

### Message Bus Trigger

A message bus trigger will execute the pipeline everytime data is received off of the configured topic.  

#### Type and Topic configuration 
Here's an example:
```toml
Type="messagebus" 
SubscribeTopic="events"
PublishTopic=""
```
The `Type=` is set to "messagebus". [EdgeX Core Data]() is publishing data to the `events` topic. So to receive data from core data, you can set your `SubscribeTopic=` either to `""` or `"events"`. You may also designate a `PublishTopic=` if you wish to publish data back to the message bus.
`edgexcontext.complete([]byte outputData)` - Will send data back to back to the message bus with the topic specified in the `PublishTopic=` property
#### Message bus connection configuration
The other piece of configuration required are the connection settings:
```toml
[MessageBus]
Type = 'zero' #specifies of message bus (i.e zero for ZMQ)
    [MessageBus.PublishHost]
        Host = '*'
        Port = 5564
        Protocol = 'tcp'
    [MessageBus.SubscribeHost]
        Host = 'localhost'
        Port = 5563
        Protocol = 'tcp'
```
By default, `EdgeX Core Data` publishes data to the `events`  topic on port 5563. The publish host is used if publishing data back to the message bus. 
>**Important Note:** Publish Host **MUST** be different for every topic you wish to publish to since the SDK will bind to the specific port. 5563 for example cannot be used to publish since `EdgeX Core Data` has bound to that port. Similiarly, you cannot have two separate instances of the app functions SDK running publishing to the same port. 

### HTTP Trigger

Designating an HTTP trigger will allow the pipeline to be triggered by a RESTful `POST` call to `http://[host]:[port]/trigger/`. The body of the POST must be an EdgeX event. 

`edgexcontext.complete([]byte outputData)` - Will send the specified data as the response to the request that originally triggered the HTTP Request. 

### Timer Based Trigger
Coming soon...

## Built-In Transforms/Functions 

### Filtering
There are two basic types of filtering included in the SDK to add to your pipeline. The provided Filter functions return a type of `events.Model`.
 - `FilterByDeviceId([]string deviceIDs)` - This function will filter the event data down to the specified deviceIDs before calling the next function. 
 - `FilterByValueDescriptor([]string valueDescriptors)` - This function will filter the event data down to the specified device value descriptor before calling the next function. 

### Conversion
There are two conversions included in the SDK that can be added to your pipeline. These transforms return a `string`.
 
 - `TransformToXML()`  - This function received an `events.Model` type and converts it to XML format. 
 - `TransformToJSON()` - This function received an `events.Model` type and converts it to JSON format. 

### Export Functions
There are two export functions included in the SDK that can be added to your pipeline. 
	
- `HTTPPost(string url)` - This function requires an endpoint be passed in order to configure the URL to `POST` data to. Currently, only unauthenticated endpoints are supported. Authenticated endpoints will be supported in the future. 
- `MQTTPublish()` - Coming Soon


## Configuration

Similar to other EdgeX services, configuration is first determined by the `configuration.toml` file in the `/res` folder. If `-r` is passed to the application on startup, the SDK will leverage the provided registry (i.e Consul) to push configuration from the file into the registry and monitor configuration from there. There are two primary sections in the `configuration.toml` file that will need to be set that are specific to the AppFunctionsSDK.
  1) `[Binding]` - This specifies the [trigger](#triggers) type and associated data required to configurate a trigger. 
  ```toml
  [Binding]
  Type=""
  SubscribeTopic=""
  PublishTopic=""
  ```
  2) `[ApplicationSettings]` - Is used for custom application settings and is accessed via the ApplicationSettings() API. The ApplicationSettings API returns a `map[string] string` containing the contents on the ApplicationSetting section of the `configuration.toml` file.
 ```toml
 [ApplicationSettings]
 ApplicationName = "My Application Service"
 ``` 
 
## Error Handling
 - Each transform returns a `true` or `false` as part of the return signature. This is called the `continuePipeline` flag and indicates whether the SDK should continue calling successive transforms in the pipeline.
 - `return false, nil` will stop the pipeline and stop processing the event. This is useful for example when filtering on values and nothing matches the criteria you've filtered on. 
 - `return false, error`, will stop the pipeline as well and the SDK will log the errorString you have returned.
- Returning `true` tells the SDK to continue, and will call the next function in the pipeline with your result.

- The SDK will handle exiting when receiving a SIGTERM/SIGINT event. 




