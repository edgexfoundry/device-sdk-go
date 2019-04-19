# Example Advanced App Functions Service 

This **advanced-filter-convert-publish** Application Service depends on the new Device Virtual Go device service to be generating random number events. It uses the following functions in its pipeline:

- Built in **Filter by Value Descriptor** function to filter just for the random **float32** & **float64** values.
- Custom **Value Converter** function which converts the encoded float values to human readable string values.
- Custom **Print to Console** function which simply prints the human readable strings to the console.
- Custom **Publish** function which prepares the modified event with converted values and outputs it to be published back to the message bus using the configured publish host/topic.

The end result from this application service is random float values in human readable format are published back to the message bus for another App Service to consume.

#### End to end Edgex integration proof point for App Functions

Using the following setup, this example advanced **App Functions Service** can be used to demonstrate an EdgeX end to end proof point with **App Functions**.

1. Start EdgeX Mongo

   - [ ] clone **[developer-scripts](https://github.com/edgexfoundry/developer-scripts)** repo
   - [ ] cd **compose-files** folder
   - [ ] run "**docker-compose up mongo**"
        - This uses the default compose file to start the EdgeX Mongo service which exposes it's port to apps running on localhost

2. Run EdgeX cores services

   - [ ] Clone **[edgex-go](https://github.com/edgexfoundry/edgex-go)** repo
   - [ ] run "**make build**"
   - [ ] run "**make run**"
     - This starts all the required Edgex core, export and support services 

3. Run Advanced App Functions example

   - [ ] Clone **[app-functions-sdk-go](https://github.com/edgexfoundry/app-functions-sdk-go)** repo
   - [ ] run "**make build**"
   - [ ] cd to **examples/advanced-filter-convert-publish** folder
   - [ ] run "./**advanced-filter-convert-publish**"

4. Configure and Run **Simple Filter XML** App Functions example

   - [ ] cd **examples/simple-filter-xml** folder

   - [ ] edit **res/configuration.toml** so the **MessageBus** and **Binding** sections are as follows:

     ```toml
             [MessageBus]
             Type = 'zero'
                 [MessageBus.PublishHost]
                     Host = '*'
                     Port = 5565
                     Protocol = 'tcp'
                 [MessageBus.SubscribeHost]
                     Host = 'localhost'
                     Port = 5564
                     Protocol = 'tcp'
                 
             [Binding]
             Type="messagebus"
             SubscribeTopic="converted"
             PublishTopic="xml"
     ```

   - [ ] Run "**./simple-filter-xml**"

5. Run Device Virtual service

   - [ ] Clone **<https://github.com/edgexfoundry/device-virtual-go>** repo

   - [ ] run "**make build**"

   - [ ] cd to **cmd** folder

   - [ ] run "./**device-virtual**"

      first time this is run, the output will have these messages :
        ```text
        level=INFO ts=2019-04-17T22:42:08.238390389Z app=device-virtual source=service.go:138 msg="**Device Service  doesn't exist, creating a new one**"
        level=INFO ts=2019-04-17T22:42:08.277064025Z app=device-virtual source=service.go:196 msg="**Addressable device-virtual doesn't exist, creating a new one**"
        ```

      One subsequent runs you will see this message:
     
        ```text
        level=INFO ts=2019-04-18T17:37:18.304805374Z app=device-virtual source=service.go:145 msg="Device Service device-virtual exists"
        ```
6. Now data will be flowing due to the auto-events configured in Device Virtual Go.

   - In the terminal that you ran **EdgeX services** you will see the logs like this. Note the encoded float values in the event JSON:
     
        ```text
        level=INFO ts=2019-04-18T17:49:27.460849685Z app=edgex-core-data source=router.go:225 msg="Posting Event: {\"device\":\"Random-Float-Generator01\",\"origin\":1555609767442,\"readings\":[{\"origin\":1555609767411,\"device\":\"Random-Float-Generator01\",\"name\":\"RandomValue_Float64\",\"value\":\"QAFk2HxRUOo=\"}]}"
        level=INFO ts=2019-04-18T17:49:27.472665597Z app=edgex-core-data source=event.go:211 msg="Putting event on message queue"
        level=INFO ts=2019-04-18T17:49:27.480548953Z app=edgex-core-data source=event.go:229 msg="Event Published on message queue. Topic: events, Correlation-id: 66f41fb9-5c57-4b6e-9bea-9c0d914d62d1 "
        
        level=INFO ts=2019-04-18T17:49:57.471868785Z app=edgex-core-data source=router.go:225 msg="Posting Event: {\"device\":\"Random-Float-Generator01\",\"origin\":1555609797452,\"readings\":[{\"origin\":1555609797419,\"device\":\"Random-Float-Generator01\",\"name\":\"RandomValue_Float32\",\"value\":\"P63Kqg==\"}]}"
        level=INFO ts=2019-04-18T17:49:57.484907249Z app=edgex-core-data source=event.go:211 msg="Putting event on message queue"
        level=INFO ts=2019-04-18T17:49:57.495394989Z app=edgex-core-data source=event.go:229 msg="Event Published on message queue. Topic: events, Correlation-id: 96fc9255-8062-4db0-83c8-5a0e5174449b "
        ```

   - In the terminal that you ran **advanced-filter-convert-publish** you will see the random float values printed.

        ```text
        RandomValue_Float64 readable value from Random-Float-Generator01 is '2.1742
        RandomValue_Float32 readable value from Random-Float-Generator01 is '1.3577
        ```

   - In the terminal that you ran **simple-filter-xml** you will see the xml representation of the events printed. Note the human readable float values in the event XML.
        ```xml
        <Event><ID></ID><Pushed>0</Pushed><Device>Random-Float-Generator01</Device><Created>0</Created><Modified>0</Modified><Origin>1555609767442</Origin><Readings><Id>835c5541-d4d2-42a8-8937-8b24b4308d3f</Id><Pushed>0</Pushed><Created>0</Created><Origin>1555609767411</Origin><Modified>0</Modified><Device>Random-Float-Generator01</Device><Name>RandomValue_Float64</Name><Value>2.1742</Value><BinaryValue></BinaryValue></Readings></Event>
        <Event><ID></ID><Pushed>0</Pushed><Device>Random-Float-Generator01</Device><Created>0</Created><Modified>0</Modified><Origin>1555609797452</Origin><Readings><Id>21c8ccdc-3438-4baa-8fab-23a63bf4fa18</Id><Pushed>0</Pushed><Created>0</Created><Origin>1555609797419</Origin><Modified>0</Modified><Device>Random-Float-Generator01</Device><Name>RandomValue_Float32</Name><Value>1.3577</Value><BinaryValue></BinaryValue></Readings></Event>
        ```
