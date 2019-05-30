# Simple CBOR Filter Application Service 

This **simple-cbor-filter** Application Service demonstrates end to end `CBOR` integration. It depends on the **device-simple** example device service from **Device SDK Go** to generate `CBOR` encode events. 

This **simple-cbor-filter** Application Service uses two application functions:

- Built in **Filter by Value Descriptor** function to filter just for the **Image** values.
- Custom **Process Images** function which re-encodes the `binary value` as an Image and prints stats about the image to the console.

The end result from this application service is that it shows that the Application Functions SDK is un-marshaling `CBOR` encode events sent from the **device-simple** device service. These event can be processed by functions similar to `JSON` encoded events. The only difference is the `CBOR` encode events have the `BinaryValue` field set, while the `JSON` encoded events have the `Value` field set.

#### Follow these steps to run the end to end CBOR demonstration

1. Start EdgeX Mongo

   - [ ] clone **[developer-scripts](https://github.com/edgexfoundry/developer-scripts)** repo
   - [ ] cd **compose-files** folder
   - [ ] run "**docker-compose up mongo**"
     - This uses the default compose file to start the EdgeX Mongo service which exposes it's port to apps running on localhost

2. Run EdgeX cores services

   - [ ] Clone **[edgex-go](https://github.com/edgexfoundry/edgex-go)** repo
   - [ ] run "**make build**"
   - [ ] run "**make run**"
     - This starts all the required EdgeX services 

3. Run **simple-cbor-filter** example

   - [ ] Clone **[app-functions-sdk-go](https://github.com/edgexfoundry/app-functions-sdk-go)** repo
   - [ ] run "**make build**"
   - [ ] cd to **examples/simple-cbor-filter** folder
   - [ ] run "./**simple-cbor-filter**"

4. Run **device-simple** device service

   - [ ] Clone **<https://github.com/edgexfoundry/device-sdk-go>** repo

   - [ ] run "**make build**"

   - [ ] cd to **example/cmd/device-simple** folder

   - [ ] run "./**device-simple**"

     This sample device service will send a `png` (light bulb on) or `jpeg` (light bulb off) image every 30 seconds. The image it sends depends on the value of its `switch` resource, which is `off` (false) by default.  

5. Now data will be flowing due to auto-events configured in **device-simple**.

   - In the terminal that you ran **simple-cbor-filter** you will see the messages like this:

     ```text
     Received Image from Device: Simple-Device01, ReadingName: Image, Image Type: jpeg, Image Size: (1000,1307), Color in middle: {0 128 128}
     ```

     Note that the image received is a jpeg since the `switch` resource in **device-simple** is set to `off ` (false)

   - The `switch` resource can be queried and changed using commands sent via PostMan by doing the following:

     1. Start PostMan

     2. Load the postman collection from the **simple-cbor-filter** example

        `Device Simple Switch commands.postman_collection.json`  

     3. This collection contains 3 commands 

        - `Get Switch status`
        - `Turn Switch on`
        - `Turn Switch off`

     4. Run  `Turn Switch on` 

   -  Now see how the **simple-cbor-filter** output has changed

         ```
         Received Image from Device: Simple-Device01, ReadingName: Image, Image Type: png, Image Size: (1000,1307), Color in middle: {255 246 0 255}
         ```

   
