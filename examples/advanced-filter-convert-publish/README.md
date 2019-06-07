# Example Advanced App Functions Service 

This **advanced-filter-convert-publish** Application Service depends on the new Device Virtual Go device service to be generating random number events. It uses the following functions in its pipeline:

- Built in **Filter by Value Descriptor** function to filter just for the random **float32** & **float64** values.
- Custom **Value Converter** function which converts the encoded float values to human readable string values.
- Custom **Print to Console** function which simply prints the human readable strings to the console.
- Custom **Publish** function which prepares the modified event with converted values and outputs it to be published back to the message bus using the configured publish host/topic.

The end result from this application service is random float values in human readable format are published back to the message bus for another App Service to consume.

#### End to end Edgex integration proof point for App Functions

Using the following setup, this example advanced **App Functions Service** can be used to demonstrate an EdgeX end to end proof point with **App Functions**.

1. Start **EdgeX Mongo**

   - [ ] clone **[developer-scripts](https://github.com/edgexfoundry/developer-scripts)** repo
   - [ ] cd **compose-files** folder
   - [ ] run "**docker-compose up mongo**"
        - This uses the default compose file to start the EdgeX Mongo service which exposes it's port to apps running on localhost

2. Run **EdgeX Core Services**

   - [ ] Clone **[edgex-go](https://github.com/edgexfoundry/edgex-go)** repo
   - [ ] run "**make build**"
   - [ ] run "**make run**"
     - This starts all the required Edgex core, export and support services 

3. Run **Advanced App Functions** example

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

5. Run **Device Virtual** service

   - [ ] Clone **<https://github.com/edgexfoundry/device-virtual-go>** repo

   - [ ] run "**make build**"

   - [ ] cd to **cmd**/res folder

   - [ ] Edit the `device.virtual.float.yaml` file
   
      This app functions example expects the float encoding for all random floats to be `Base64` and needs to restrict the range of values that are generated so they are easy to read. By default the device-virtual is using `Base64` for Float32 & `eNotation` for Float64 and doesn't set any range limits. Make the following changes to the `deviceResources` section to meet these needs.
   
      For `RandomValue_Float32` change the `value` property to:

      ```
      { type: "Float32", readWrite: "R", defaultValue: "0", floatEncoding: "Base64", minimum: "1.0", maximum: "1.9" }
      ```
   
      For `RandomValue_Float64` change the `value` property to:
   
      ```
      { type: "Float64", readWrite: "R", defaultValue: "0", floatEncoding: "Base64", minimum: "2.0", maximum: "2.9" }
      ```
   
   - [ ] If you previously ran the Device Virtual service, run the follow `curl` commands to clear the old profile so that the new changes are used when the device service is started.
   
      ```
      curl -X DELETE http://localhost:48081/api/v1/deviceservice/name/device-virtual
      curl -X DELETE http://localhost:48081/api/v1/deviceprofile/name/Random-Float-Device
      ```
   
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

   - In the terminal that you ran **advanced-filter-convert-publish** you will see the random float values printed.

        ```text
        RandomValue_Float64 readable value from Random-Float-Device is '2.1742
        RandomValue_Float32 readable value from Random-Float-Device is '1.3577
        ```

   - In the terminal that you ran **simple-filter-xml** you will see the xml representation of the events printed. Note the human readable float values in the event XML.
        ```xml
        <Event><ID></ID><Pushed>0</Pushed><Device>Random-Float-Device</Device><Created>0</Created><Modified>0</Modified><Origin>1555609767442</Origin><Readings><Id>835c5541-d4d2-42a8-8937-8b24b4308d3f</Id><Pushed>0</Pushed><Created>0</Created><Origin>1555609767411</Origin><Modified>0</Modified><Device>Random-Float-Device</Device><Name>RandomValue_Float64</Name><Value>2.1742</Value><BinaryValue></BinaryValue></Readings></Event>
        <Event><ID></ID><Pushed>0</Pushed><Device>Random-Float-Device</Device><Created>0</Created><Modified>0</Modified><Origin>1555609797452</Origin><Readings><Id>21c8ccdc-3438-4baa-8fab-23a63bf4fa18</Id><Pushed>0</Pushed><Created>0</Created><Origin>1555609797419</Origin><Modified>0</Modified><Device>Random-Float-Device</Device><Name>RandomValue_Float32</Name><Value>1.3577</Value><BinaryValue></BinaryValue></Readings></Event>
        ```
