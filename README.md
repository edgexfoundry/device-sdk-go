# Go Device Service SDK
## Overview
This repository is a set of Go packages which can be used to build a Go-based EdgeX Foundry Device Service.
## Usage
Developers could make their own Device Service by implementing the `ProtocolDriver  interface` for the specific IoT protocol and main function to start the Device Service.  To implement the main function, the startup package could be optional leveraged, or developers could write the customized bootstrap code by themselves.
Please see the build-in [Simple Device Service](https://github.com/edgexfoundry/device-sdk-go/tree/master/example) as an example.

## Community
- Chat: https://edgexfoundry.slack.com
- Mailing lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)

