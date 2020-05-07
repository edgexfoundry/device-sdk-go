# App Functions SDK (Golang) - Beta

Welcome the App Functions SDK for EdgeX. This sdk is meant to provide all the plumbing necessary for developers to get started in processing/transforming/exporting data out of EdgeX. 

Table of contents
=================

<!--ts-->
   * [Getting Started](#getting-started)
   * [Triggers](#triggers)
   * [Context API](#context-api)
   * [Built-In Functions](#built-in-transformsfunctions)
      * [Filtering](#filtering)
      * [Encryption](#encryption)
      * [Batch](#batch)
      * [Conversion](#conversion)
      * [Compressions](#compressions)
      * [Core Data](#CoreData-Functions)
      * [Export Functions](#export-functions)    
   * [Configuration](#configuration)
   * [Error Handling](#error-handling)
   * [Advanced Topics](#advanced-topics)
     * [Using The Webserver](#using-the-webserver)
     * [Target Type](#target-type)
     * [Command Line Options](#command_line_options)
     * [Environment Variable Overrides](#environment_variable_overrides)
     * [Store and Forward](#store-and-forward)
     * [Secrets](#secrets)

 <!--te-->

## Getting Started

### Build Prerequisites

Please see the [edgex-go README](https://github.com/edgexfoundry/edgex-go/blob/master/README.md).

### The SDK

The SDK is built around the idea of a "Functions Pipeline". A functions pipeline is a collection of various functions that process the data in the order that you've specified. The functions pipeline is executed by the specified [trigger](#triggers) in the `configuration.toml` . The first function in the pipeline is called with the event that triggered the pipeline (ex. `events.Model`). Each successive call in the pipeline is called with the return result of the previous function. Let's take a look at a simple example that creates a pipeline to filter particular device ids and subsequently transform the data to XML:
```go
package main

import (
	"fmt"
	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"os"
)

func main() {

	// 1) First thing to do is to create an instance of the EdgeX SDK, giving it a service key
	edgexSdk := &appsdk.AppFunctionsSDK{
		ServiceKey: "SimpleFilterXMLApp", // Key used by Registry (Aka Consul)
	}

	// 2) Next, we need to initialize the SDK
	if err := edgexSdk.Initialize(); err != nil {
		message := fmt.Sprintf("SDK initialization failed: %v\n", err)
		if edgexSdk.LoggingClient != nil {
			edgexSdk.LoggingClient.Error(message)
		} else {
			fmt.Println(message)
		}
		os.Exit(-1)
	}

	// 3) Shows how to access the application's specific configuration settings.
	deviceNames, err := edgexSdk.GetAppSettingStrings("DeviceNames")
	if err != nil {
		edgexSdk.LoggingClient.Error(err.Error())
		os.Exit(-1)
	}    

	// 4) This is our pipeline configuration, the collection of functions to
	// execute every time an event is triggered.
	if err = edgexSdk.SetFunctionsPipeline(
			transforms.NewFilter(deviceNames).FilterByDeviceName, 
			transforms.NewConversion().TransformToXML,
		); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK SetPipeline failed: %v\n", err))
		os.Exit(-1)
	}

	// 5) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events to trigger the pipeline.
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}
```

The above example is meant to merely demonstrate the structure of your application. Notice that the output of the last function is not available anywhere inside this application. You must provide a function in order to work with the data from the previous function. Let's go ahead and add the following function that prints the output to the console.

```go
func printXMLToConsole(edgexcontext *appcontext.Context, params ...interface{}) (bool,interface{}) {
  if len(params) < 1 { 
  	// We didn't receive a result
  	return false, errors.New("No Data Received")
  }
  println(params[0].(string))
  return true, nil
}
```
After placing the above function in your code, the next step is to modify the pipeline to call this function:

```go
edgexSdk.SetFunctionsPipeline(
  transforms.NewFilter(deviceNames).FilterByDeviceName, 
  transforms.NewConversion().TransformToXML,
  printXMLToConsole //notice this is not a function call, but simply a function pointer. 
)
```
After making the above modifications, you should now see data printing out to the console in XML when an event is triggered.
> You can find this example in the `/examples` directory located in this repository. You can also use the provided `EdgeX Applications Function SDK.postman_collection.json" file to load into postman to trigger the sample pipeline.

Up until this point, the pipeline has been [triggered](#triggers) by an event over HTTP and the data at the end of that pipeline lands in the last function specified. In the example, data ends up printed to the console. Perhaps we'd like to send the data back to where it came from. In the case of an HTTP trigger, this would be the HTTP response. In the case of a message bus, this could be a new topic to send the data back to for other applications that wish to receive it. To do this, simply call `edgexcontext.Complete([]byte outputData)` passing in the data you wish to "respond" with. In the above `printXMLToConsole(...)` function, replace `println(params[0].(string))` with `edgexcontext.Complete([]byte(params[0].(string)))`. You should now see the response in your postman window when testing the pipeline.

## Examples

The [App Service Examples](https://github.com/edgexfoundry-holding/app-service-examples) repo contains a variety of simple to advanced example **Application Services** built upon the **App Functions SDK**. Examples that once were in the examples folder of the SDK have been moved to this `Examples` repo.

## Triggers

Triggers determine how the app functions pipeline begins execution. In the simple example provided above, an HTTP trigger is used. The trigger is determine by the `configuration.toml` file located in the `/res` directory under a section called `[Binding]`. Check out the [Configuration Section](#configuration) for more information about the toml file.

### Message Bus Trigger

A message bus trigger will execute the pipeline every time data is received off of the configured topic.  

#### Type and Topic configuration 
Here's an example:
```toml
Type="messagebus" 
SubscribeTopic="events"
PublishTopic=""
```
The `Type=` is set to "messagebus". [EdgeX Core Data]() is publishing data to the `events` topic. So to receive data from core data, you can set your `SubscribeTopic=` either to `""` or `"events"`. You may also designate a `PublishTopic=` if you wish to publish data back to the message bus.
`edgexcontext.Complete([]byte outputData)` - Will send data back to back to the message bus with the topic specified in the `PublishTopic=` property
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
>**Important Note:** Publish Host **MUST** be different for every topic you wish to publish to since the SDK will bind to the specific port. 5563 for example cannot be used to publish since `EdgeX Core Data` has bound to that port. Similarly, you cannot have two separate instances of the app functions SDK running publishing to the same port. 

### HTTP Trigger

Designating an HTTP trigger will allow the pipeline to be triggered by a RESTful `POST` call to `http://[host]:[port]/trigger/`. The body of the POST must be an EdgeX event. 

`edgexcontext.Complete([]byte outputData)` - Will send the specified data as the response to the request that originally triggered the HTTP Request. 

## Context API

The context parameter passed to each function/transform provides operations and data associated with each execution of the pipeline. Let's take a look at a few of the properties that are available:
```golang
type Context struct {
	// ID of the EdgeX Event -- will be filled for a received JSON Event
	EventID string
	
	// Checksum of the EdgeX Event -- will be filled for a received CBOR Event
	EventChecksum string
	
	// This is the ID used to track the EdgeX event through entire EdgeX framework.
	CorrelationID string
	
	// OutputData is used for specifying the data that is to be outputted. Leverage the .Complete() function to set.
	OutputData []byte
	
	// This holds the configuration for your service. This is the preferred way to access your custom application settings that have been set in the configuration.	
	Configuration common.ConfigurationStruct
	
	// LoggingClient is exposed to allow logging following the preferred logging strategy within EdgeX.
	LoggingClient logger.LoggingClient
	
	// EventClient exposes Core Data's EventClient API
	EventClient coredata.EventClient
	
	// ValueDescriptorClient exposes Core Data's ValueDescriptor API
	ValueDescriptorClient coredata.ValueDescriptorClient
	
	// CommandClient exposes Core Commands's Command API
	CommandClient command.CommandClient
	
	// NotificationsClient exposes Support Notification's Notifications API
	NotificationsClient notifications.NotificationsClient
	
	// RetryData holds the data to be stored for later retry when the pipeline function returns an error
	RetryData []byte
	
	// SecretProvider exposes the support for getting and storing secrets
	SecretProvider *security.SecretProvider
}
```

### LoggingClient

The `LoggingClient` exposed on the context is available to leverage logging libraries/service utilized throughout the EdgeX framework. The SDK has initialized everything so it can be used to log `Trace`, `Debug`, `Warn`, `Info`, and `Error` messages as appropriate. See `examples/simple-filter-xml/main.go` for an example of how to use the `LoggingClient`.

### EventClient 

The `EventClient ` exposed on the context is available to leverage Core Data's `Event` API. See [interface definition](https://github.com/edgexfoundry/go-mod-core-contracts/blob/master/clients/coredata/event.go#L35) for more details. This client is useful for querying events and is used by the [MarkAsPushed](#markaspushed) convenience API described below.

### ValueDescriptorClient

The `ValueDescriptorClient ` exposed on the context is available to leverage Core Data's `ValueDescriptor` API. See [interface definition](https://github.com/edgexfoundry/go-mod-core-contracts/blob/master/clients/coredata/value_descriptor.go#L29) for more details. Useful for looking up the value descriptor for a reading received.

### CommandClient 

The `CommandClient ` exposed on the context is available to leverage Core Command's `Command` API. See [interface definition](https://github.com/edgexfoundry/go-mod-core-contracts/blob/master/clients/command/client.go#L28) for more details. Useful for sending commands to devices.

### NotificationsClient

The `CommandClient` exposed on the context is available to leverage Support Notifications' `Notifications` API. See [README](https://github.com/edgexfoundry/go-mod-core-contracts/blob/master/clients/notifications/README.md) for more details. Useful for sending notifications. 

### Note about Clients

Each of the clients above is only initialized if the Clients section of the configuration contains an entry for the service associated with the Client API. If it isn't in the configuration the client will be `nil`. Your code must check for `nil` to avoid panic in case it is missing from the configuration. Only add the clients to your configuration that your Application Service will actually be using. All application services need the `Logging` and many will need `Core-Data`. The following is an example `Clients` section of a configuration.toml with all supported clients specified:

```
[Clients]
  [Clients.Logging]
  Protocol = "http"
  Host = "localhost"
  Port = 48061

  [Clients.CoreData]
  Protocol = 'http'
  Host = 'localhost'
  Port = 48080

  [Clients.Command]
  Protocol = 'http'
  Host = 'localhost'
  Port = 48082

  [Clients.Notifications]
  Protocol = 'http'
  Host = 'localhost'
  Port = 48060
```

### .MarkAsPushed()

`.MarkAsPushed()` is used to indicate to EdgeX Core Data that an event has been "pushed" and is no longer required to be stored. The scheduler service will purge all events that have been marked as pushed based on the configured schedule. By default, it is once daily at midnight. If you leverage the built in export functions (i.e. HTTP Export, or MQTT Export), then the event will automatically be marked as pushed upon a successful export. 
### .PushToCore()
`.PushToCore(string deviceName, string readingName, byte[] value)` is used to push data to EdgeX Core Data so that it can be shared with other applications that are subscribed to the message bus that core-data publishes to. `deviceName` can be set as you like along with the `readingName` which will be set on the EdgeX event sent to CoreData. This function will return the new EdgeX Event with the ID populated, however the CorrelationId will not be available.

 > NOTE: If validation is turned on in CoreServices then your `deviceName` and `readingName` must exist in the CoreMetadata and be properly registered in EdgeX. 

 > WARNING: Be aware that without a filter in your pipeline, it is possible to create an infinite loop when the messagebus trigger is used. Choose your device-name and reading name appropriately.
### .Complete()
`.Complete([]byte outputData)` can be used to return data back to the configured trigger. In the case of an HTTP trigger, this would be an HTTP Response to the caller. In the case of a message bus trigger, this is how data can be published to a new topic per the configuration. 

### .SetRetryData()

`.SetRetryData(payload []byte)` can be used to store data for later retry. This is useful when creating a custom export function that needs to retry on failure sending the data. The payload data will be stored for later retry based on `Store and Forward` configuration. When the retry is triggered, the function pipeline will be re-executed starting with the function that called this API. That function will be passed the stored data, so it is important that all transformations occur in functions prior to the export function. The `Context` will also be restored to the state when the function called this API. See [Store and Forward](#store-and-forward) for more details.

> NOTE: `Store and Forward` be must enabled when calling this API. 

### .GetSecrets()

`.GetSecrets(path string, keys ...string)` is used to retrieve secrets from the secret store. `path` specifies the type or location of the secrets to retrieve. If specified it is appended to the base path from the exclusive secret store configuration. `keys` specifies the secrets which to retrieve. If no keys are provided then all the keys associated with the specified path will be returned.

## Built-In Transforms/Functions 

All transforms define a type and a `New` function which is used to initialize an instance of the type with the  required parameters. These instances returned by these `New` functions give access to their appropriate pipeline function pointers when build  the function pipeline.

```
E.G. NewFilter([] {"Device1", "Device2"}).FilterByDeviceName
```

### Filtering

There are two basic types of filtering included in the SDK to add to your pipeline. Theses provided Filter functions return a type of events.Model. If filtering results in no remaining data, the pipeline execution for that pass is terminated. If no values are provided for filtering, then data flows through unfiltered.
 - `NewFilter([]string filterValues)` - This function returns a `Filter` instance initialized with the passed in filter values. This `Filter` instance is used to access the following filter functions that will operate using the specified filter values.
    - `FilterByDeviceName` - This function will filter the event data down to the specified device names and return the filtered data to the pipeline.
    - `FilterByValueDescriptor` - This function will filter the event data down to the specified device value descriptor and return the filtered data to the pipeline.

#### JSON Logic
  - `NewJSONLogic(rule string)` - This function returns a `JSONLogic` instance initialized with the passed in JSON rule. The rule passed in should be a JSON string conforming to the specification here: http://jsonlogic.com/operations.html. 
  > NOTE: Only simple logic/filtering operators are supported. Manipulation of data via JSONLogic rules are not yet supported. For more advanced scenarios checkout [EMQ X Kuiper](https://github.com/emqx/kuiper).

   - `Evaluate` - This is the function that will be used in the pipeline to apply the JSON rule to data coming in on the pipeline. If the condition of your rule is met, then the pipeline will continue and the data will continue to flow to the next function in the pipeline. If the condition of your rule is NOT met, then pipeline execution stops. 


### Encryption
There is one encryption transform included in the SDK that can be added to your pipeline. 

- `NewEncryption(key string, initializationVector string)` - This function returns a `Encryption` instance initialized with the passed in key and initialization vector. This `Encryption` instance is used to access the following encryption function that will use the specified key and initialization vector.
  - `EncryptWithAES` - This function receives a either a `string`, `[]byte`, or `json.Marshaller` type and encrypts it using AES encryption and returns a `[]byte` to the pipeline.

### Batch
Included in the SDK is an in-memory batch function that will hold on to your data before continuing the pipeline. There are three functions provided for batching each with their own strategy.
- `NewBatchByTime(timeInterval string)` - This function returns a `BatchConfig` instance with time being the strategy that is used for determining when to release the batched data and continue the pipeline. `timeInterval` is the duration to wait (i.e. `10s`). The time begins after the first piece of data is received. If no data has been received no data will be sent forward. 
- `NewBatchByCount(batchThreshold int)` - This function returns a `BatchConfig` instance with count being the strategy that is used for determining when to release the batched data and continue the pipeline. `batchThreshold` is how many events to hold on to (i.e. `25`). The count begins after the first piece of data is received and once the threshold is met, the batched data will continue forward and the counter will be reset.
- `NewBatchByTimeAndCount(timeInterval string, batchThreshold int)` - This function returns a `BatchConfig` instance with a combination of both time and count being the strategy that is used for determining when to release the batched data and continue the pipeline. Whichever occurs first will trigger the data to continue and be reset.
  - `Batch` - This function will apply the selected strategy in your pipeline.
  
### Conversion
There are two conversions included in the SDK that can be added to your pipeline. These transforms return a `string`.

 - `NewConversion()` - This function returns a `Conversion` instance that is used to access the following conversion functions: 
    - `TransformToXML`  - This function receives an `events.Model` type, converts it to XML format and returns the XML string to the pipeline. 
    - `TransformToJSON` - This function receives an `events.Model` type and converts it to JSON format and returns the JSON string to the pipeline.

### Compressions
There are two compression types included in the SDK that can be added to your pipeline. These transforms return a `[]byte`.

 - `NewCompression()` - This function returns a `Compression` instance that is used to access the following compression functions:
    - `CompressWithGZIP`  - This function receives either a `string`,`[]byte`, or `json.Marshaler` type, GZIP compresses the data, converts result to base64 encoded string, which is returned as a `[]byte` to the pipeline.
    - `CompressWithZLIB` - This function receives either a `string`,`[]byte`, or `json.Marshaler` type, ZLIB compresses the data, converts result to base64 encoded string, which is returned as a `[]byte` to the pipeline.

### CoreData Functions
These are functions that enable interactions with the CoreData REST API. 
- `NewCoreData()` - This function returns a `CoreData` instance. This `CoreData` instance is used to access the following function(s).
  - `MarkAsPushed` - This function provides the MarkAsPushed function from the context as a First-Class Transform that can be called in your pipeline. [See Definition Above](#.MarkAsPushed()). The data passed into this function from the pipeline is passed along unmodifed since all required information is provided on the context (EventId, CorrelationId,etc.. )
  - `PushToCore` - This function provides the PushToCore function from the context as a First-Class Transform that can be called in your pipeline. [See Definition Above](#.PushToCore()). The data passed into this function from the pipeline is wrapped in an EdgeX event with the `deviceName` and `readingName` that were set upon instantiation and then sent to CoreData to be added as an event. Returns the new EdgeX event with ID populated.
    
    > NOTE: If validation is turned on in CoreServices then your `deviceName` and `readingName` must exist in the CoreMetadata and be properly registered in EdgeX. 

### Export Functions
There are few export functions included in the SDK that can be added to your pipeline. 
- `NewHTTPSender(url string, mimeType string, persistOnError bool)` - This function returns a `HTTPSender` instance initialized with the passed in url, mime type and persistOnError values. 
- `NewHTTPSenderWithSecretHeader(url string, mimeType string, persistOnError bool, httpHeaderSecretName string, secretPath string)` - This function returns a `HTTPSender` instance similar to the above function however will set up the `HTTPSender` to add a header to the HTTP request using the `httpHeaderSecretName` as both the header key  and the key to search for in the secret provider at `secretPath` leveraging secure storage of secrets. 
This `HTTPSender` instance is used to access the following functions that will use the required url and optional mime type and persistOnError:
  
  - `HTTPPost` - This function receives either a `string`,`[]byte`, or `json.Marshaler` type from the previous function in the pipeline and posts it to the configured endpoint. If no previous function exists, then the event that triggered the pipeline, marshaled to json, will be used. Currently, only unauthenticated endpoints are supported. Authenticated endpoints will be supported in the future. If the post fails and `persistOnError`is `true` and `Store and Forward` is enabled, the data will be stored for later retry. See [Store and Forward](#store-and-forward) for more details
- `NewMQTTSecretSender(mqttConfig MQTTSecretConfig, persistOnError bool)` - This function returns a `MQTTSecretSender` instance intialized with the options specfied in the `MQTTSecretConfig`.
```golang
type MQTTSecretConfig struct {
    // BrokerAddress should be set to the complete broker address i.e. mqtts://mosquitto:8883/mybroker
    BrokerAddress string
    // ClientId to connect with the broker with.
    ClientId string
    // The name of the path in secret provider to retrieve your secrets
    SecretPath string
    // AutoReconnect indicated whether or not to retry connection if disconnected
    AutoReconnect bool
    // Topic that you wish to publish to
    Topic string
    // QoS for MQTT Connection
    QoS byte
    // Retain setting for MQTT Connection
    Retain bool
    // SkipCertVerify
    SkipCertVerify bool
    // AuthMode indicates what to use when connecting to the broker. Options are "none", "cacert" , "usernamepassword", "clientcert".
    // If a CA Cert exists in the SecretPath then it will be used for all modes except "none". 
    AuthMode string
}
```
Secrets in the secret provider may be located at any path however they must have the follow keys at the specified `SecretPath`. What `AuthMode` you choose depends on what values are used. For example, if "none" is specified as auth mode all keys will be ignored. Similarily, if `AuthMode` is set to "clientcert" username and password will be ignored.
  `username` - username to connect to the broker
  `password` - password used to connect to the broker
  `clientkey`- client private key in PEM format
  `clientcert` - client cert in PEM format
  `cacert` - ca cert in PEM format

- **DEPRECATED**`NewMQTTSender(logging logger.LoggingClient, addr models.Addressable, keyCertPair *KeyCertPair, mqttConfig MqttConfig, persistOnError bool)` - This function returns a `MQTTSender` instance initialized with the passed in MQTT configuration . This `MQTTSender` instance is used to access the following  function that will use the specified MQTT configuration
  
  - `KeyCertPair` - This structure holds the Key and Certificate information for when using secure **TLS** connection to the broker. Can be `nil` if not using secure **TLS** connection. 
  
  - `MqttConfig` - This structure holds addition MQTT configuration settings. 
  
    ```
    	Qos            byte
    	Retain         bool
    	AutoReconnect  bool
    	SkipCertVerify bool
    	User           string
    	Password       string
    ```
  
    The `GO` complier will default these to `0`, `false` and `""`, so you only need to set the fields that your usage requires that differ from the default.
  
  - `MQTTSend` - This function receives either a `string`,`[]byte`, or `json.Marshaler` type from the previous function in the pipeline and sends it to the specified MQTT broker. If no previous function exists, then the event that triggered the pipeline, marshaled to json, will be used. If the send fails and `persistOnError`is `true` and `Store and Forward` is enabled, the data will be stored for later retry. See [Store and Forward](#store-and-forward) for more details

### Output Functions

There is one output function included in the SDK that can be added to your pipeline. 

- NewOutput() - This function returns a `Output` instance that is used to access the following output function: 
  - `SetOutput` - This function receives either a `string`,`[]byte`, or `json.Marshaler` type from the previous function in the pipeline and sets it as the output data for the pipeline to return to the configured trigger. If configured to use message bus, the data will be published to the message bus as determined by the `MessageBus` and `Binding` configuration. If configured to use HTTP trigger the data is returned as the HTTP response. Note that calling Complete() from the Context API in a custom function can be used in place of adding this function to your pipeline

## Configuration

Similar to other EdgeX services, configuration is first determined by the `configuration.toml` file in the `/res` folder. If `-r` is passed to the application on startup, the SDK will leverage the provided registry (i.e Consul) to push configuration from the file into the registry and monitor configuration from there. You will find the configuration under the `edgex/appservices/1.0/` key. There are two primary sections in the `configuration.toml` file that will need to be set that are specific to the AppFunctionsSDK. 
  1) `[Binding]` - This specifies the [trigger](#triggers) type and associated data required to configure a trigger. 

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
 - The SDK will return control back to main when receiving a SIGTERM/SIGINT event to allow for custom clean up.


## Advanced Topics

The following items discuss topics that are a bit beyond the basic use cases of the Application Functions SDK when interacting with EdgeX.

### Configurable Functions Pipeline

This SDK provides the capability to define the functions pipeline via configuration rather than code using the **app-service-configurable** application service. See **app-service-configurable** [README](https://github.com/edgexfoundry/app-service-configurable/blob/master/README.md) for more details.

### Using The Webserver

It is not uncommon to require your own API endpoints when building an app service. Rather than spin up your own webserver inside of your app (alongside the already existing running webserver), we've exposed a method that allows you add your own routes to the existing webserver. A few routes are reserved and cannot be used:
- /api/version
- /api/v1/ping
- /api/v1/metrics
- /api/v1/config
- /api/v1/trigger
To add your own route, use the `AddRoute(route string, handler func(nethttp.ResponseWriter, *nethttp.Request), methods ...string) error` function provided on the sdk. Here's an example:
```golang
edgexSdk.AddRoute("/myroute", func(writer http.ResponseWriter, req *http.Request) {
    context := req.Context().Value(appsdk.SDKKey).(*appsdk.AppFunctionsSDK) 
		context.LoggingClient.Info("TEST") // alternative to edgexSdk.LoggingClient.Info("TEST")
		writer.Header().Set("Content-Type", "text/plain")
		writer.Write([]byte("hello"))
		writer.WriteHeader(200)
}, "GET")
```
Under the hood, this simply adds the provided route, handler, and method to the gorilla `mux.Router` we use in the SDK. For more information you can check out the github repo [here](https://github.com/gorilla/mux). 
You can access the resources such as the logging client by accessing the context as shown above -- this is useful for when your routes might not be defined in your main.go where you have access to the `edgexSdk` instance.

### Target Type

The target type is the object type of the incoming data that is sent to the first function in the function pipeline. By default this is an EdgeX `Event` since typical usage is receiving `events` from Core Data via Message Bus. 

For other usages where the data is not `events` coming from Core Data, the `TargetType` of the accepted incoming data can be set when the SDK instance is created. 
There are scenarios where the incoming data is not an EdgeX `Event`. One example scenario is 2 application services are chained via the Message Bus. The output of the first service back to the Messages Bus is inference data from analyzing the original input `Event`data.  The second service needs to be able to let the SDK know the target type of the input data it is expecting.

For usages where the incoming data is not `events`, the `TargetType` of the excepted incoming data can the set when the SDK instance is created. 

Example:

```
type Person struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

edgexSdk := &appsdk.AppFunctionsSDK {
	ServiceKey: serviceKey, 
	TargetType: &Person{},
}
```

Note that `TargetType` must be set to a pointer to an instance of your target type such as `&Person{}` . The first function in your function pipeline will be passed an instance of your target type, not a pointer to it. In the example above the first function in the pipeline would start something like:

```
func MyPersonFunction(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {

	edgexcontext.LoggingClient.Debug("MyPersonFunction")

	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	person, ok := params[0].(Person)
	if !ok {
        return false, errors.New("type received is not a Person")
	}
	
	....
```

The SDK supports unmarshaling JSON or CBOR encoded data into an instance of the target type. If your incoming data is not JSON or CBOR encoded, you then need to set the `TargetType` to  `&[]byte`.

If the target type is set to `&[]byte` the incoming data will not be unmarshaled.  The content type, if set, will be passed as the second parameter to the first function in your pipeline.  Your first function will be responsible for decoding the data or not.

### Command Line Options

The following command line options are available

```
  -c=<path>
  --confdir=<path>
        Specify an alternate configuration directory.
  -p=<profile>
  --profile=<profile>
        Specify a profile other than default.
  -r    
  --registry
        Indicates the service should use the registry.
  -o    
  -overwrite
        Overwrite configuration in the Registry with local values.
  -s    
  -skipVersionCheck
        Indicates the service should skip the Core Service's version compatibility check.
  -n
  --serviceName                
        Overrides the service name, aka service key, to be used with Registry and/or 
        Configuration Providers. If the name provided contains the text `<profile>`, 
        this text will be replaced with the name of the profile used..
```

Examples:

```
simple-filter-xml -r -c=./res -p=http-export
```

or

```
simple-filter-xml --registry --confdir=./res --profile=mqtt-export
```

### Environment Variable Overrides

All the configuration settings from the configuration.toml file can be overridden by environment variables.  

The environment variable names have the following format:

```
<TOML KEY>
<TOML SECTION>_<TOML KEY>
<TOML SECTION>_<TOML SUB-SECTION>_<TOML KEY>
```

> *Note: With the Geneva release CamelCased environment variable names are deprecated. Instead use all upper case environment variable names as in the example below.*

Examples:

```
TOML   : FailLimit = 30
ENVVAR : FAILLIMIT=100

TOML   : [Logging]
		 EnableRemote = false
ENVVAR : LOGGING.ENABLEREMOTE=true

TOML   : [Clients]
  			[Clients.CoreData]
  			Host = 'localhost'
ENVVAR : CLIENTS_COREDATA_HOST=edgex-core-data
```

#### EDGEX_SERVICE_NAME

This environment variable overrides the service name, aka service key, to be used with Registry and/or Configuration Provider.

If the name provided contains the text `<profile>` , this text will be replaced with the name of the profile used.

Example:

```
EDGEX_SERVICE_NAME=AppService-<profile>-mycloud
and profile used is http-export 
then the service name will be:

   AppService-http-export-mycloud
```

#### edgex_registry

**[Deprecated]** This environment variable overrides the Registry connection information and occurs every time the application service starts. The value is in the format of a URL.

> *Note: This environment variable override has been deprecated in the Geneva Release. Instead, use configuration overrides of **REGISTRY_PROTOCOL** and/or **REGISTRY_HOST** and/or **REGISTRY_PORT***

```
edgex_registry=consul://edgex-core-consul:8500

This sets the Registry information fields as follows:
    Type: consul
    Host: edgex-core-consul
    Port: 8500
```

#### edgex_service

**[Deprecated]** This environment variable overrides the Service connection information and occurs every time the application service starts. The value is in the format of a URL.

> *Note: This environment variable override has been deprecated in the Geneva Release. Instead, use configuration overrides of **SERVICE_PROTOCOL** and/or **SERVICE_HOST** and/or **SERVICE_PORT***

```
edgex_service=http://192.168.1.2:4903

This sets the Service information fields as follows:
    Protocol: http
    Host: 192.168.1.2
    Port: 4903
```

#### edgex_profile / EDGEX_PROFILE

This environment variable overrides the command line `profile` argument. It will replace the current value passed via the `-p` or `--profile`, if one exists. If not specified it will add the `--profile` argument. This is useful when running the service via docker-compose.

> *Note: The lower case version has been deprecated* in the Geneva release. Instead use upper case version **EDGEX_PROFILE**

Using docker-compose:

```
  app-service-configurable-rules:
    image: edgexfoundry/docker-app-service-configurable:1.1.0
    environment: 
      - EDGEX_PROFILE : "rules-engine"
    ports:
      - "48095:48095"
    container_name: edgex-app-service-configurable
    hostname: edgex-app-service-configurable
    networks:
      edgex-network:
        aliases:
          - edgex-app-service-configurable
    depends_on:
      - data
      - command
```

This sets the `--profile=docker-rules-engine` command line argument so that the application service uses the `docker-rules-engine` configuration profile which resides at `/res/docker-rules-engine/configuration.toml`

> *Note that Application Services no longer use docker profiles. They use Environment Overrides in the docker compose file to make the necessary changes to the configuration for running in Docker. See the **Environment Variable Overrides For Docker** section in [App Service Configurable's README](https://github.com/edgexfoundry/app-service-configurable/blob/master/README.md#environment-variable-overrides-for-docker)* for more details and an example. 

### Store and Forward

The Store and Forward capability allows for export functions to persist data on failure and for the export of the data to be retried at a later time. 

> *Note: The order the data exported via this retry mechanism is not guaranteed to be the same order in which the data was initial received from Core Data*

#### Configuration

Two sections of configuration have been added for Store and Forward.

`Writable.StoreAndForward` allows enabling, setting the interval between retries and the max number of retries. If running with Registry, these setting can be changed on the fly without having to restart the service.

```toml
  [Writable.StoreAndForward]
    Enabled = false
    RetryInterval = '5m'
    MaxRetryCount = 10
```

> *Note: RetryInterval should be at least 1 second (eg. '1s') or greater. If a value less than 1 second is specified, 1 second will be used.*

> *Note: Endless retries will occur when MaxRetryCount is set to 0.*

> *Note: If MaxRetryCount is set to less than 0, a default of 1 retry will be used.*

Database describes which database type to use, `mongodb` or `redisdb`, and the information required to connect to the database. This section is required if Store and Forward is enabled, otherwise it is currently optional.

```toml
[Database]
Type = "mongodb"
Host = "localhost"
Port = 27017
Timeout = '5s'
Username = ""
Password = ""
```

#### How it works

When an export function encounters an error sending data it can call `SetRetryData(payload []byte)` on the Context. This will store the data for later retry. If the application service is stop and then restarted while stored data hasn't been successfully exported, the export retry will resume once the service is up and running again.

> *Note: It is important that export functions return an error and stop pipeline execution* after the call to `SetRetryData`. See [HTTPPost](https://github.com/edgexfoundry/app-functions-sdk-go/blob/master/pkg/transforms/http.go) function in SDK as an example

When the `RetryInterval` expires, the function pipeline will be re-executed starting with the export function that saved the data. The saved data will be passed to the export function which can then attempt to resend the data. 

> *NOTE: The export function will receive the data as it was stored, so it is important that any transformation of the data occur in functions prior to the export function. The export function should only export the data that it receives.*

One of three out comes can occur after the export retried has completed. 

1. Export retry was successful

   In this case the stored data is removed from the database and the execution of the pipeline functions after the export function, if any, continues. 

2. Export retry fails and retry count `has not been` exceeded

   In this case the store data is updated in the database with the incremented retry count

3. Export retry fails and retry count `has been` exceeded

   In this case the store data is removed from the database and never retried again.

> *NOTE: Changing Writable.Pipeline.ExecutionOrder will invalidate all currently stored data and result in it all being removed from the database on the next retry.* This is because the position of the export function can no longer be guaranteed and no way to ensure it is properly executed on the retry.

### Secrets

#### Configuration

All instances of App Services share the same database and database credentials. However, there are secrets for each App Service that are exclusive to the instance running. As a result, two separate configuration for secret store clients are used to manage shared and exclusive application service secrets.

The GetSecrets() and StoreSecrets() calls  use the exclusive secret store client to manage application secrets.

An example of configuration settings for each secret store client is below:

```toml
# Shared Secret Store
[SecretStore]
    Host = 'localhost'
    Port = 8200
    Path = '/v1/secret/edgex/appservice/'
    Protocol = 'https'
    RootCaCertPath = '/tmp/edgex/secrets/ca/ca.pem'
    ServerName = 'edgex-vault'
    TokenFile = '/tmp/edgex/secrets/edgex-appservice/secrets-token.json'
    # Number of attempts to retry retrieving secrets before failing to start the service.
    AdditionalRetryAttempts = 10
    # Amount of time to wait before attempting another retry
    RetryWaitPeriod = "1s"

	[SecretStore.Authentication]
		AuthType = 'X-Vault-Token'	

# Exclusive Secret Store
[SecretStoreExclusive]
    Host = 'localhost'
    Port = 8200
    Path = '/v1/secret/edgex/<app service key>/'
    Protocol = 'https'
    ServerName = 'edgex-vault'
    TokenFile = '/tmp/edgex/secrets/<app service key>/secrets-token.json'
    # Number of attempts to retry retrieving secrets before failing to start the service.
    AdditionalRetryAttempts = 10
    # Amount of time to wait before attempting another retry
    RetryWaitPeriod = "1s"

    [SecretStoreExclusive.Authentication]
    	AuthType = 'X-Vault-Token'
```

#### Storing Secrets

##### Secure Mode

When running an application service in secure mode, secrets can be stored in the secret store (Vault) by making an HTTP `POST` call to the secrets API route in the application service, `http://[host]:[port]/api/v1/secrets`. The secrets are stored and retrieved from the secret store based on values in the *SecretStoreExclusive* section of the configuration file. Once a secret is stored, only the service that added the secret will be able to retrieve it.  For secret retrieval see [Getting Secrets](#getting-secrets).

An example of the message body JSON is below.  

```json
{
  "path" : "/MyPath",
  "secrets" : [
    {
      "key" : "MySecretKey",
      "value" : "MySecretValue"
    }
  ]
}
```

`NOTE: path specifies the type or location of the secrets to store. It is appended to the base path from the SecretStoreExclusive configuration. An empty path is a valid configuration for a secret's location.`

##### Insecure Mode

When running in insecure mode, the secrets are stored and retrieved from the *Writable.InsecureSecrets* section of the service's configuration toml file. Insecure secrets and their paths can be configured as below.

```toml
   [Writable.InsecureSecrets]    
      [Writable.InsecureSecrets.AWS]
        Path = 'aws'
        [Writable.InsecureSecrets.AWS.Secrets]
          username = 'aws-user'
          password = 'aws-pw'
      
      [Writable.InsecureSecrets.MongoDB]
        Path = ''
        [Writable.InsecureSecrets.MongoDB.Secrets]
          username = 'mongo-user'
          password = 'mongo-pw'
```

`NOTE: An empty path is a valid configuration for a secret's location  `

#### Getting Secrets

Application Services can retrieve their secrets from the underlying secret store using the [GetSecrets()](#.GetSecrets()) API in the SDK. 

If in secure mode, the secrets are retrieved from the secret store based on the *SecretStoreExclusive* configuration values. 

If running in insecure mode, the secrets are retrieved from the *Writable.InsecureSecrets* configuration.

