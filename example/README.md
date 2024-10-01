# device-simple

The included device-simple example device service demonstrates basic usage of device-sdk-go.

## Protocol Driver

To make a functional Device Service, developers must implement the [ProtocolDriver](../pkg/interfaces/protocoldriver.go) interface. 
`ProtocolDriver` interface provides abstraction logic of how to interact with Device through specific protocol. See [simpledriver.go](driver/simpledriver.go) for example.

## Device Discovery

Some device protocols allow for devices to be discovered automatically.
A Device Service may include a capability for discovering devices and creating the corresponding Device objects within EdgeX.  

To enable device discovery, developers need to implement the `Discover` method which is used to trigger protocol-specific device discovery.
Any devices found as a result of discovery being triggered are returned to the SDK via a go channel, passed to the implementation as a parameter during Initialization.
New discovery attempts may be started as soon as a slice of devices is submitted, so in oreder to avoid the service being congested by concurrent discovery.
  
The SDK will then filter these devices against pre-defined acceptance criteria (i.e. Provision Watchers), and add any devices which match (excluding existing devices).

A Provision Watcher contains the following fields:

`Identifiers`: A set of name-value pairs against which a new device's ProtocolProperties are matched  
`BlockingIdentifiers`: An additional set of name-value pairs that, if matched, will block the addition of a newly discovered device  
`ProfileName`: The name of a DeviceProfile which should be assigned to new devices which meet the given criteria  
`ServiceName`: The name of a DeviceService which the ProvisionWatcher should be applied on  
`AdminState`: The initial Administrative State for new devices which meet the given criteria  
`AutoEvents`: A list of AutoEvent associated with the newly discovered device 
 
A candidate new device passes a ProvisionWatcher if all the Identifiers match, and none of the Blocking Identifiers match.
For devices with multiple `Device.Protocol`, each `Device.Protocol` is considered separately. A match on any of the protocols results in the device being added.

Finally, A boolean configuration value `Device/Discovery/Enabled` defaults to false. If it is set true, and the DS implementation supports discovery, discovery is enabled.
Dynamic Device Discovery is triggered either by internal timer(see `Device/Discovery/Interval` in [configuration.yaml](cmd/device-simple/res/configuration.yaml)) or by a call to the device service's `/discovery` REST endpoint.

The following steps show how to trigger discovery on device-simple:
1. Set `Device/Discovery/Enabled` to true in [configuration file](cmd/device-simple/res/configuration.yaml)
2. Trigger discovery by sending POST request to DS endpoint: http://edgex-device-simple:59999/api/v3/discovery
3. `Simple-Device02` will be discovered and added to EdgeX, while `Simple-Device03` will be blocked by the [Provision Watcher `BlockingIdentifiers`](cmd/device-simple/res/provisionwatchers/Simple-Provision-Watcher.yml)

## Extended Protocol Driver
### ProfileScan
Some device protocols allow for devices to discover profiles automatically.
A Device Service may include a capability for discovering device profiles and creating the associated Device Profile objects for devices within EdgeX.

To enable profile scan, developers need to implement the [ProfileScan](../pkg/interfaces/protocolprofile.go) interface.
The `ExtendedProtocolDriver` interface defines a `ProfileScan` method which is used to trigger device-specific profile scan.
The device profile found as a result of profile scan is returned to the SDK, and the SDK will then create the device profile and update the device to use the profile.

The following steps show how to trigger profile scan on device-simple:
1. A pre-define device `ProfileScan-Simple-Device` is created without an associated device profile.
2. Trigger profile scan by sending POST request to DS endpoint: http://edgex-device-simple:59999/api/v3/profilescan with payload:
   ```json
   {
     "apiVersion": "v3",
     "deviceName": "ProfileScan-Simple-Device",
     "profileName": "ProfileScan-Test-Profile"
   }
   ```
3. Device Profile `ProfileScan-Test-Profile` will be added to EdgeX and `ProfileScan-Simple-Device` will be updated to use the device profile.

### StopDeviceDiscovery and StopProfileScan
The `ExtendedProtocolDriver` interface defines a `StopDeviceDiscovery` to stop the device discovery and `StopProfileScan` to stop the profile scanning.
