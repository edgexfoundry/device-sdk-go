# device-simple

The included device-simple example device service demonstrates basic usage of device-sdk-go.

## Protocol Driver

To make a functional Device Service, developers must implement the [ProtocolDriver](../pkg/models/protocoldriver.go) interface. 
`ProtocolDriver` interface provides abstraction logic of how to interact with Device through specific protocol. See [simpledriver.go](driver/simpledriver.go) for example.

## Protocol Discovery

Some device protocols allow for devices to be discovered automatically.
A Device Service may include a capability for discovering devices and creating the corresponding Device objects within EdgeX.  

To enable device discovery, developers need to implement the [ProtocolDiscovery](../pkg/models/protocoldiscovery.go) interface.
The `ProtocolDiscovery` interface defines a single `Discover` method which is used to trigger protocol-specific device discovery.
Any devices found as a result of discovery being triggered are returned to the SDK via a go channel, passed to the implementation as a parameter during Initialization.
New discovery attempts may be started as soon as a slice of devices is submitted, so in oreder to avoid the service being congested by concurrent discovery.
  
The SDK will then filter these devices against pre-defined acceptance criteria (i.e. Provision Watchers), and add any devices which match (excluding existing devices).

A Provision Watcher contains the following fields:

`Identifiers`: A set of name-value pairs against which a new device's ProtocolProperties are matched  
`BlockingIdentifiers`: An additional set of name-value pairs which if matched, will block the addition of a newly discovered device  
`ProfileName`: The name of a DeviceProfile which should be assigned to new devices which meet the given criteria  
`ServiceName`: The name of a DeviceService which the ProvisionWatcher should be applied on  
`AdminState`: The initial Administrative State for new devices which meet the given criteria  
`AutoEvents`: A list of AutoEvent associated with the newly discovered device 
 
A candidate new device passes a ProvisionWatcher if all the Identifiers match, and none of the Blocking Identifiers match.
For devices with multiple `Device.Protocol`, each `Device.Protocol` is considered separately. A match on any of the protocols results in the device being added.

Finally, A boolean configuration value `Device/Discovery/Enabled` defaults to false. If it is set true, and the DS implementation supports discovery, discovery is enabled.
Dynamic Device Discovery is triggered either by internal timer(see `Device/Discovery/Interval` in [configuration.toml](cmd/device-simple/res/configuration.toml)) or by a call to the device service's `/discovery` REST endpoint.

The following steps show how to trigger discovery on device-simple:
1. Set `Device/Discovery/Enabled` to true in [configuration file](cmd/device-simple/res/configuration.toml)
2. Post the [provided provisionwatcher](cmd/device-simple/res/provisionwatcher.json) into core-metadata endpoint: http://edgex-core-metadata:59881/api/v2/provisionwatcher
3. Trigger discovery by sending POST request to DS endpoint: http://edgex-device-simple:59999/api/v2/discovery
4. `Simple-Device02` will be discovered and added to EdgeX.