# Advanced Target Type  

This **advanced-target-type** Application Service demonstrates how to create an Application Service that expects a custom type to feed to the functions pipeline. For more detail refer to the SDK's main README's section on [TargetType](https://github.com/edgexfoundry/app-functions-sdk-go/blob/master/README.md#target-type)

To run this example:

1.  Clone **[app-functions-sdk-go](https://github.com/edgexfoundry/app-functions-sdk-go)** repo

2. run `make build`

3. cd `examples/advance-target-type`

4. run `./advance-target-type`

5. Start PostMan

6. Load `Post Person to Trgger.postman_collection.json` collection in PostMan

7. Run the `Person Trigger` request

   - The following XML will be printed to the console by the Application Service and will be returned as the trigger HTTP response in PostMan.

     ```
     <Person>
        <FirstName>Sam</FirstName>
        <LastName>Smith</LastName>
        <Phone>
           <CountryCode>1</CountryCode>
           <AreaCode>480</AreaCode>
           <LocalPrefix>970</LocalPrefix>
           <LocalNumber>3476</LocalNumber>
        </Phone>
        <PhoneDisplay>+01(480) 970-3476</PhoneDisplay>
     </Person>
     ```

   - Note that the PhoneDisplay field that is not present in the XML sent from PostMan is now present and filled out.


