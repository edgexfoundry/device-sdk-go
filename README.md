# Go Device Service SDK
## Overview
This repository is a set of Go packages which can be used to build a Go-based EdgeX Foundry Device Service.
## Usage
Developers could make their own Device Service by implementing the `ProtocolDriver  interface` for the specific IoT protocol and main function to start the Device Service.  To implement the main function, the startup package could be optional leveraged, or developers could write the customized bootstrap code by themselves.
Please see the build-in [Simple Device Service](https://github.com/edgexfoundry/device-sdk-go/tree/master/example) as an example.

## Command Line Options
The following command line options are available
```
  -c=<path>
  --confdir=<path>
        Specify an alternate configuration directory.
  -p=<profile>
  --profile=<profile>
        Specify a profile other than default.
  -f=<file>
  --file=<file>
        Indicates name of the local configuration file.
  -i=<instace>
  --instance=<instance>             
        Provides a service name suffix which allows unique instance to be created.
        If the option is provided, service name will be replaced with "<name>_<instance>"
  -o    
  --overwrite
        Overwrite configuration in the Registry with local values.
  -r    
  --registry
        Indicates the service should use the registry.
  -cp    
  --configProvider
        Indicates to use Configuration Provider service at specified URL.
        URL Format: {type}.{protocol}://{host}:{port} ex: consul.http://localhost:8500

```

## Float value encodeing

In EdgeX, float value has two kinds of encoding, Base64, and eNotation.

> When EdgeX is given (or returns) a float32 or float64 value as a string, the format of the string is by default a base64 encoded little-endian of the float32 or float64 value, but the “floatEncoding” attribute relating to the value may instead specify “eNotation” in which case the representation is a decimal with exponent (eg “1.234e-5”)

https://docs.google.com/document/d/1aMIQ0kb46VE5eeCpDlaTg8PP29-DBSBTlgeWrv6LuYk/edit

### base64
The SDK should convert the float value to little-endian binary and then encode the binary as base64 string.
Currently, C SDK converts float value to little-endian binary. GO SDK encodes convert float value to big-endian binary because SDK spec changed in fuji release, we will modify GO SDK to make the behavior consistent in the future.

Usage:
```
-
  name: "Temperature"
  description: "Temperature value"
  properties:
    value:
      { type: "FLOAT64", readWrite: "RW", floatEncoding: "Base64"}
    units:
      { type: "String", readWrite: "R", defaultValue: "degrees Celsius"}
```

### eNoation
The SDK should convert the float value to string with eNotaion representation.

Usage:
```
-
  name: "Temperature"
  description: "Temperature value"
  properties:
    value:
      { type: "FLOAT64", readWrite: "RW", floatEncoding: "eNotation"}
    units:
      { type: "String", readWrite: "R", defaultValue: "degrees Celsius"}
```

## Community
- Chat: https://edgexfoundry.slack.com
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)

