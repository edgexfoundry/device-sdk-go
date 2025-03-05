
<a name="EdgeX Device SDK Go (found in device-sdk-go) Changelog"></a>
## EdgeX Device SDK Go
[Github repository](https://github.com/edgexfoundry/device-sdk-go)

### Change Logs for EdgeX Dependencies

- [go-mod-bootstrap](https://github.com/edgexfoundry/go-mod-bootstrap/blob/main/CHANGELOG.md)
- [go-mod-core-contracts](https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/CHANGELOG.md)
- [go-mod-messaging](https://github.com/edgexfoundry/go-mod-messaging/blob/main/CHANGELOG.md)
- [go-mod-registry](https://github.com/edgexfoundry/go-mod-registry/blob/main/CHANGELOG.md) 
- [go-mod-secrets](https://github.com/edgexfoundry/go-mod-secrets/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-configuration](https://github.com/edgexfoundry/go-mod-configuration/blob/main/CHANGELOG.md) (indirect dependency)

## [4.0.0] Odessa - 2025-03-12 (Only compatible with the 4.x releases)

### ✨  Features

- Update to use the new message envelope in go-mod-messaging ([77f7962…](https://github.com/edgexfoundry/device-sdk-go/commit/77f7962edb3356bbdb68222c2d3d6cabc66f655b))
```text

BREAKING CHANGE: Change MessageEnvelope payload from a byte array to a generic type

```
- Ensure AutoEvent execution happens at the configured interval ([#1680](https://github.com/edgexfoundry/device-sdk-go/issues/1680)) ([9af0642…](https://github.com/edgexfoundry/device-sdk-go/commit/9af0642e0454afe5e7d0bd4928537ff3f2da2e70))
- Add device up/down detection ([4e68cb5…](https://github.com/edgexfoundry/device-sdk-go/commit/4e68cb59f209a13292d5305933a0070493bfeed9))
- Implement AutoEvent onChangeThreshold ([44a1270…](https://github.com/edgexfoundry/device-sdk-go/commit/44a1270054e23efbb0c017a84b6120f6b6fc29cb))
- Support null value in reading instead of nil reading value ([#1639](https://github.com/edgexfoundry/device-sdk-go/issues/1639)) ([791e380…](https://github.com/edgexfoundry/device-sdk-go/commit/791e3808a6f22236528886c5f30bc95897a4d2de))
- Enable PIE support for ASLR and full RELRO ([4175f78…](https://github.com/edgexfoundry/device-sdk-go/commit/4175f787d4d7dfbfc0240ea4dfcfa8e563a46958))
- Allow reading value to be null ([888d014…](https://github.com/edgexfoundry/device-sdk-go/commit/888d0147f37d1b0cb5e12db1661dbf13e1e18e03))
- Add new APIs to stop the device discovery/profile scan ([bbccac4…](https://github.com/edgexfoundry/device-sdk-go/commit/bbccac45e796a95e0126349a1e5980b980a572b6))
- Publish System Events for device discovery and profile scan progress ([6d5cd89…](https://github.com/edgexfoundry/device-sdk-go/commit/6d5cd8995dc981926f14347bf033eca9bc1eeb5e))
- Add /profilescan API ([e6ed876…](https://github.com/edgexfoundry/device-sdk-go/commit/e6ed876124f4173f783b12bcf038a65d9eaf4a49))
- Update auto discovery interval parsing error logs ([3f018f6…](https://github.com/edgexfoundry/device-sdk-go/commit/3f018f60d3c906f8579fab7dd580bd58f7339674))
- Reduce numbers of inactive goroutine to optimize performance ([#1580](https://github.com/edgexfoundry/device-sdk-go/issues/1580)) ([50e812d…](https://github.com/edgexfoundry/device-sdk-go/commit/50e812dd0fc65b097297637d35e073c818e89983))
- Allow empty profileName in Device ([b7a2f1e…](https://github.com/edgexfoundry/device-sdk-go/commit/b7a2f1edf3de431b6849887a9fcc8db68db620bc))
- Add openziti support ([#1569](https://github.com/edgexfoundry/device-sdk-go/issues/1569)) ([decf7a4…](https://github.com/edgexfoundry/device-sdk-go/commit/decf7a4867f03271291a6465384c5ba130c9fec8))

### ♻ Code Refactoring

- Update module to v4 ([6bfdd3f…](https://github.com/edgexfoundry/device-rest-go/commit/6bfdd3ff5f3e792eafad2b4c95b01495d5837e2e))
```text
BREAKING CHANGE: update go module to v4
```
- Refine the discovery and scan logs ([c1d0121…](https://github.com/edgexfoundry/device-sdk-go/commit/c1d0121d89c9148a4827eb1b9062e4b7ffaa6073))
- Remove the version number in Init logs ([#1541](https://github.com/edgexfoundry/device-sdk-go/issues/1541)) ([50dd63d…](https://github.com/edgexfoundry/device-sdk-go/commit/50dd63d392f370c86a33393d79501a04f0565adc))

### 🐛 Bug Fixes

- Add core-metadata dependency check ([f602b1d…](https://github.com/edgexfoundry/device-sdk-go/commit/f602b1d3b279f1300250a834fc65f11994324a69))
- Initialize DeviceService.Properties with an empty map ([89b0421…](https://github.com/edgexfoundry/device-sdk-go/commit/89b0421df83343ee133e8ef4f4d4a7ca251848a0))
- Only one ldflags flag is allowed ([2b424e5…](https://github.com/edgexfoundry/device-sdk-go/commit/2b424e540a2c3b8f091c6d5734c04c5071b69322))
- Update the status code to 400 for write commands that contain invalid values([44ddd41…](https://github.com/edgexfoundry/device-sdk-go/commit/44ddd41cb5e6b40211c543a30b3ec4dbb1ba819f))
- Remove device profile from the cache properly ([7eb1c14…](https://github.com/edgexfoundry/device-sdk-go/commit/7eb1c14b0483b0ae235a397d5607dc5eaf6f1543))
- Avoid running auto events for empty profile ([f35974c…](https://github.com/edgexfoundry/device-sdk-go/commit/f35974ce7fdb486d71758995b1726d45ec7c2ad0))
- Improve error handling and response for ProfileScan ([2c0f4ed…](https://github.com/edgexfoundry/device-sdk-go/commit/2c0f4ed6bb705a30f5bf719c67df6f419fc12bf2))
- Use atomic operation to provision profile ([36fb107…](https://github.com/edgexfoundry/device-sdk-go/commit/36fb1076ec3d34dc2fd4e33abbddf84d4499d5aa))
- Convert the Device Labels value from string to string slice ([0d5b88d…](https://github.com/edgexfoundry/device-sdk-go/commit/0d5b88dd80a75ef9b43031f4cf9724640386a6a6))
- Address CVE in Alpine base image ([bb3b780…](https://github.com/edgexfoundry/device-sdk-go/commit/bb3b780b49bdb40d95a497fa6cff623f428380a1))
- Set proper transform flag for Set command ([#1578](https://github.com/edgexfoundry/device-sdk-go/issues/1578)) ([97b36af…](https://github.com/edgexfoundry/device-sdk-go/commit/97b36af7baff2da76e29dc2359062fa76ab64ed3))

### 📖 Documentation

- Correct the ProtocolDriver links in README ([581c97a…](https://github.com/edgexfoundry/device-sdk-go/commit/581c97a92a3a8adbad42633b4d460162c03abb0f))
- Move API document files from openapi/v3 to openapi ([8f827bd…](https://github.com/edgexfoundry/device-sdk-go/commit/8f827bd74a4c847a61608dcbc9d76980c13265ba))

### 👷 Build

- Upgrade to go-1.23, Linter1.61.0 and Alpine 3.20 ([21d3830…](https://github.com/edgexfoundry/device-rest-go/commit/21d3830f1720a37f1895b62ff4ca15d755f1aed1))

## [v3.1.0] Napa - 2023-11-15 (Only compatible with the 3.x releases)

### ✨  Features

- *(security)* Add new AddCustomRoute method with authentication parameter ([#1475](https://github.com/edgexfoundry/device-sdk-go/issues/1475)) ([5a7d052…](https://github.com/edgexfoundry/device-sdk-go/commit/5a7d05295c20c4306666c126d05bdc304538796c))
- Add device last connected metrics ([#1515](https://github.com/edgexfoundry/device-sdk-go/issues/1515)) ([1f88585…](https://github.com/edgexfoundry/device-sdk-go/commit/1f885853b7afeee748945214d1e34d19c3162505))
- Update index files to include names for preloading check ([10694e6…](https://github.com/edgexfoundry/device-sdk-go/commit/10694e6a91a67f5fde8debfd675953a530180938))
- Allow NameFieldPath configurable ([f16fc3d…](https://github.com/edgexfoundry/device-sdk-go/commit/f16fc3d171215af7a3e26199bf520421d1739968))
- Accept Url escape in API path ([fdbdbc1…](https://github.com/edgexfoundry/device-sdk-go/commit/fdbdbc112457214c0e1c2a8eec7425427dc3eb50))
- Replace gorilla/mux router with echo ([72d350f…](https://github.com/edgexfoundry/device-sdk-go/commit/72d350f2483b97042997e4d5de4bcf3345622430))
- Add better error handling when running in hybrid mode but common config is missing ([#1480](https://github.com/edgexfoundry/device-sdk-go/issues/1480)) ([7433feb…](https://github.com/edgexfoundry/device-sdk-go/commit/7433febb356a2a6123d42b5c48f02468c3a1d213))
- Implement loading from URI for Profile, Device, & Provision Watcher files ([#1471](https://github.com/edgexfoundry/device-sdk-go/issues/1471)) ([0776c05…](https://github.com/edgexfoundry/device-sdk-go/commit/0776c05133aecd05e9cca04d73874dfbd24934b0))
- Move all the common APIs into go-mod-bootstrap ([9983415…](https://github.com/edgexfoundry/device-sdk-go/commit/9983415449d789263673552186519e6709848005))
- Adjust GetFileType() function for new secret URI format ([#1507](https://github.com/edgexfoundry/device-sdk-go/issues/1507)) ([88f0350…](https://github.com/edgexfoundry/device-sdk-go/commit/88f03503c07c02c451459ab8c2d5aba4061d83b0))
- Use WrapHandler func from go-mod-bootstrap ([ca6e439…](https://github.com/edgexfoundry/device-sdk-go/commit/ca6e439590fd3e0d96364bc20f1a7957a2569ef1))
- Update handler funcs to use Echo signatures ([93fdda3…](https://github.com/edgexfoundry/device-sdk-go/commit/93fdda3beb2e14b4e733fb76d164ead49978cb57))

### 📖 Documentation

- Replace Slack chat with GitHub discussion ([3e8c91c…](https://github.com/edgexfoundry/device-sdk-go/commit/3e8c91cf201308cb239f41d8f9ace2bd21fe3f70))


### 👷 Build

- Upgrade to go-1.21, Linter1.54.2 and Alpine 3.18 ([#1511](https://github.com/edgexfoundry/device-sdk-go/issues/1511)) ([10d83f9…](https://github.com/edgexfoundry/device-sdk-go/commit/10d83f9598abba3559abd072fcd7553d77fa119a))

### 🤖 Continuous Integration

- Add automated release workflow on tag creation ([526596b…](https://github.com/edgexfoundry/device-sdk-go/commit/526596b2a1c7dd5155db82aafa89415c2e504d1d))


### 🧪 Testing

- Add URI test files ([#1481](https://github.com/edgexfoundry/device-sdk-go/issues/1481)) ([d74039b…](https://github.com/edgexfoundry/device-sdk-go/commit/d74039b1e75348ec690936aee1c1bcbaac678fee))

## [v3.0.0] Minnesota - 2023-05-31 (Only compatible with the 3.x releases)

### Features ✨
- Add Start method to the ProtocolDriver interface ([#453fffe](https://github.com/edgexfoundry/device-sdk-go/commit/453fffe39844267b7c3d359e90be35c08d562c07))
  ```text
  BREAKING CHANGE: Added required `Start` method to the ProtocolDriver interface. 
                   This method is called after the sdk has been completely initialized.
                   This is where device service should implement post initialization code. 
  ```
- Add PatchDevice and DeviceExistsForName and refactor UpdateDeviceOperatingState([#]())
  ```text
  BREAKING CHANGE: SetDeviceOperatingState has been removed, and UpdateDeviceOperatingState has been modified to accept a models.OperatingState value.
  ```
- Using url.PathUnescape for decoding API path ([#cdd8b0f](https://github.com/edgexfoundry/device-sdk-go/commit/cdd8b0f4c44e4c2644a3f6043fa9547e0633c97a))
  ```text
  BREAKING CHANGE: Use PathUnescape for decoding API path to consist with the change from MQTT topic, the MQTT topic path will encode with url.PathEscape.
  ```
-  All ProtocolDriver interface API implementations required ([#d73da92](https://github.com/edgexfoundry/device-sdk-go/commit/d73da920b9fd3a52132699998633ff8aee3aa256))
  ```text
  BREAKING CHANGE: Move Discover(), ValidateDevice() to the ProtocolDriver interface to that they are required like the existing interfaces. 
                   This forces at a minimum an empty implementation which gives exposure to the developer that they exist.
                   Use interfaces.ProtocolDriver instead of any as parameter
  ```
- Apply provision watcher model changes ([#24d4a99](https://github.com/edgexfoundry/device-sdk-go/commit/24d4a999756b81a71ac18f620daf3eea158c1b50))
  ```text
  BREAKING CHANGE: Apply provision watcher model changes and skip the locked provision watcher.
  ```
- Change configuration file format to YAML ([#c61610b](https://github.com/edgexfoundry/device-sdk-go/commit/c61610bbbff1187c8ac68f5c7e0a10add7b75ddb))
  ```text
  BREAKING CHANGE: Configuration files are now in YAML format, Default file name is now configuration.yaml
  ```
- Change device definition file to YAML format ([#9bade9f](https://github.com/edgexfoundry/device-sdk-go/commit/9bade9fb8ad7b1ba90f4944641889e9ddeadd73b))
  ```text
  BREAKING CHANGE: Stop supporting TOML format and support YAML format for Device definition files
  ```
- Apply JWT authentication to incoming calls ([#774c203](https://github.com/edgexfoundry/device-sdk-go/commit/774c2037947f109106e5ac1f44e7b15f17fcf2b1))
  ```text
  BREAKING CHANGE: In secure mode, incoming REST API calls must have a JWT authentication token, with the exception of /api/v2/ping.
  ```
- Remove LastConnected, LastReported and UpdateLastConnected configs ([#7414f7f](https://github.com/edgexfoundry/device-sdk-go/commit/7414f7f856fa0ae762c124471c520e272c125f0c))
  ```text
  BREAKING CHANGE: Remove LastConnected, LastReported and UpdateLastConnected configs
  ```
- Consume contracts mod to update /secret DTO ([#76cf874](https://github.com/edgexfoundry/device-sdk-go/commit/76cf874e5a4c502767114360369306a3defd879c))
  ```text
  BREAKING CHANGE: secret DTO object in core contracts uses SecretName instead of Path
  ```
- Updates for common config ([#9f4fc55](https://github.com/edgexfoundry/device-sdk-go/commit/9f4fc55359c9cb479e87cde204f04957ffcda185))
  ```text
  BREAKING CHANGE: Common config settings removed from configuration file 
  ```
- Add additional service-name level in event publish topic ([#15fb7a8](https://github.com/edgexfoundry/device-sdk-go/commit/15fb7a88bc0851fefb7130faf6710b4e2cf95cf9))
  ```text
  BREAKING CHANGE: event publish topic now <prefix>/<service-name>/<profile-name>/<device-name>/<source-name>
  ```
- Replace REST device validation callback with MessageBus ([#b7afc2a](https://github.com/edgexfoundry/device-sdk-go/commit/b7afc2aba7a1c5861c6c21c42b42dc1283ea5949))
  ```text
  BREAKING CHANGE: /validate/device REST endpoint removed
  ```
- Replace REST device service callback with System Event ([#e412a44](https://github.com/edgexfoundry/device-sdk-go/commit/e412a44eb13141e6797f53827a761c4e91ea34f8))
  ```text
  BREAKING CHANGE: /callback/service REST endpoint removed
  ```
- Replace REST provision watcher callbacks with System Events([#8ea8883](https://github.com/edgexfoundry/device-sdk-go/commit/8ea8883b1ccaac0e7f122d2d9e12f4a2ee2a0f8d))
  ```text
  BREAKING CHANGE: /callback/watcher and /callback/watcher/name/{name} REST endpoints have been removed
  ```
- Replace REST device profile callback with System Events ([#09198ff](https://github.com/edgexfoundry/device-sdk-go/commit/09198ff90566170cbb480bcb988d97281c5ebb32))
  ```text
  BREAKING CHANGE: PUT /callback/profile REST endpoint has been removed
  ```
- Remove UseMessageBus config ([#482e5b9](https://github.com/edgexfoundry/device-sdk-go/commit/482e5b9a17b7b3fb4ec31d9b3676b675d900f01c))
  ```text
  BREAKING CHANGE: Removed the 'Device.UseMessageBus' config, the code for sending event via core-data REST client and the 'Clients.core-data' dependency
  ```
- Remove old metrics collection and REST/metrics endpoint ([#89d807d](https://github.com/edgexfoundry/device-sdk-go/commit/89d807d8113b45d5cd8e97f478743d732b99a48e))
  ```text
  BREAKING CHANGE: /metrics endpoint no longer available for any service
  ```
- Replace REST device callbacks with System Events ([#3f884e](https://github.com/edgexfoundry/device-sdk-go/commit/3f884ed804720b30a96228d314a94b40ef1079e0))
  ```text
  BREAKING CHANGE: The following device callback REST endpoints are removed:
                     - POST /callback/device
                     - PUT /callback/device
                     - DELETE /callback/device/name/{name}
  ```
- Remove ZeroMQ MessageBus capability ([#f8460cf](https://github.com/edgexfoundry/device-sdk-go/commit/f8460cf112e20e4666e24beb849e99a270391ebb))
  ```text
  BREAKING CHANGE: ZeroMQ MessageBus capability no longer available
  ```
- Device ProtocolProperties have typed values ([#1365](https://github.com/edgexfoundry/device-sdk-go/issues/1365)) ([#a6f9b45](https://github.com/edgexfoundry/device-sdk-go/commits/a6f9b45))
- Enhance autodiscovery to better support  multiple instances of same device service ([#1444](https://github.com/edgexfoundry/device-sdk-go/issues/1444)) ([#d75af8d](https://github.com/edgexfoundry/device-sdk-go/commits/d75af8d))
- Consume new -d/--dev Dev Mode command-line flag ([#d0b4661](https://github.com/edgexfoundry/device-sdk-go/commits/d0b4661))
- Consume SecretProvider breaking changes ([#656a0e7](https://github.com/edgexfoundry/device-sdk-go/commits/656a0e7))
- Allow regex GET command returns partial result ([#6d81477](https://github.com/edgexfoundry/device-sdk-go/commits/6d81477))
- Enable regex for executing GET command ([#efda1a8](https://github.com/edgexfoundry/device-sdk-go/commits/efda1a8))
- Support YAML format for ProvisionWatcher definition file ([#a4f7691](https://github.com/edgexfoundry/device-sdk-go/commits/a4f7691))
- Publish event with updated value for PUT command ([#420f04f](https://github.com/edgexfoundry/device-sdk-go/commits/420f04f))
- Consume watch for common Writable config changes ([#1372](https://github.com/edgexfoundry/device-sdk-go/issues/1372)) ([#84b3eaa](https://github.com/edgexfoundry/device-sdk-go/commits/84b3eaa))
- Accept URL escape for device command name and resource name ([#7be4a1d](https://github.com/edgexfoundry/device-sdk-go/commits/7be4a1d))
- Add resource, command, and device tags to reading/event ([#1297](https://github.com/edgexfoundry/device-sdk-go/issues/1297)) ([#93a0268](https://github.com/edgexfoundry/device-sdk-go/commits/93a0268))
- Implement support for ProvisionWatchersDir ([#1f95b19](https://github.com/edgexfoundry/device-sdk-go/commits/1f95b19))

### Bug Fixes 🐛

- Add UpdateDevice callback for device profile update ([#d4adbed](https://github.com/edgexfoundry/device-sdk-go/commits/d4adbed))
- Fix typo "DeviceResourece" -> "DeviceResource" in error log message ([#278dac4](https://github.com/edgexfoundry/device-sdk-go/commits/278dac4))

### Code Refactoring ♻

- Modify the numeric data type in Value Properties to pointer ([#d2a234e](https://github.com/edgexfoundry/device-sdk-go/commit/d2a234e83cee793ad02e4d30354cfd596d07cbbd))
  ```text
  BREAKING CHANGE:
     - update mask,shift,base,scale,offset to pointer
     - update maximum and minimum data type from string to float64 pointer
  ```
- Update data types in ResourceProperties ([#c3c5272](https://github.com/edgexfoundry/device-sdk-go/commit/c3c527261f00c12e33369f990b1b875843bff393))
  ```text
  BREAKING CHANGE: support mask,shift,base,scale,offset in numeric data type
  ```
- Remove global variable 'ds' in service package ([#e5f9ace](https://github.com/edgexfoundry/device-sdk-go/commit/e5f9ace8858ef16ed7fd9fdcaf664427cc73e288))
  ```text
  BREAKING CHANGE:
     - update ProtocolDriver Initialize method signature to pass
       DeviceServiceSDK interface as parameter
     - update DeviceServiceSDK interface:
       - update Getter method name to be more idiomatic
       - remove Stop() method as it should only be called by SDK
       - add AsyncValuesChannel and DiscoveredDeviceChannel getter method
       - rename AsyncReadings to AsyncReadingsEnabled
       - rename DeviceDiscovery to DeviceDiscoveryEnabled

     The DeviceServiceSDK interface is passed to ProtocolDriver as the
     only parameter in Initialize method so that developer can still access,
     mock and test with it.
  ```
- Replace internal topics from config with new constants ([#cf150bd](https://github.com/edgexfoundry/device-sdk-go/commit/cf150bd1006c505b2559d8970edbe52dd5fe7567))
  ```text
  BREAKING CHANGE: Internal topics no longer configurable, except the base topic.
  ```
- Rework Command via MessageBus for new Request API response topic ([#d7de237c](https://github.com/edgexfoundry/device-sdk-go/commit/d7de237c874a90e81091b377f7571cc3981ed0ab))
  ```text
  BREAKING CHANGE: Command via MessageBus Topic configurations have changed (Note: later commit removes topic configuration)
  ```
- Update config for message bus topic wild cards ([#73fb48f](https://github.com/edgexfoundry/device-sdk-go/commit/73fb48faf2bae20740fd473665d27f5e6eced9ca))
  ```text
  BREAKING CHANGE: use MQTT wild cards + for single level and # for multiple levels
  ```
- Use bool types for command parameters to be more consistent ([#89b19b1](https://github.com/edgexfoundry/device-sdk-go/commit/89b19b1cde0f51320ac6cf9edfa39ed75314b82b))
  ```text
  BREAKING CHANGE: ds-pushevent and ds-returnevent to use bool true/false instead of yes/no
  ```
- Update config for removal of SecretStore from services' configuration file ([#11af1f](https://github.com/edgexfoundry/device-sdk-go/commit/11af1f65973d64e418efe9ae1e7fed57ad58627f))
  ```text
  BREAKING CHANGE: SecretStore config no longer in service configuration file. 
                   Changes must be done via use of environment variable overrides of default values.
  ```
- Rework code for refactored MessageBus Configuration ([#ebb4d57](https://github.com/edgexfoundry/device-sdk-go/commit/ebb4d574cedcd34a1a0a0bbc9318b53a84a6b8a6))
  ```text
  BREAKING CHANGE: MessageQueue renamed to MessageBus and fields changed. See v3 Migration guide.
  ```
- Rename command line flags for the sake of consistency ([#4aa2fa](https://github.com/edgexfoundry/device-sdk-go/commit/4aa2fae2e06829cc5012bd5c306ed57b362ce7ee))
  ```text
  BREAKING CHANGE: renamed -c/--confdir to -cd/--configDirand -f/--file to -cf/--configFile
  ```
- Update module to v3 ([#97d52b](https://github.com/edgexfoundry/device-sdk-go/commit/97d52b643112c5c00428daae5a03f6563bf38928))
  ```text
  BREAKING CHANGE: Import paths will need to change to v3
  ```
- Consume ProvisionWatcher DTO change for ServiceName ([#1453](https://github.com/edgexfoundry/device-sdk-go/issues/1453)) ([#3563ea3](https://github.com/edgexfoundry/device-sdk-go/commits/3563ea3))
- Tweaks to private config in Device Simple ([#4b8be51](https://github.com/edgexfoundry/device-sdk-go/commits/4b8be51))

### Documentation 📖

- Update swagger to match latest changes in go-mod-core-contracts Device dto ([#05593f9](https://github.com/edgexfoundry/device-sdk-go/commits/05593f9))
- Update swagger to match latest changes in go-mod-contracts dtos common SecretRequest ([#53645a9](https://github.com/edgexfoundry/device-sdk-go/commits/53645a9))
- Replace V2 swagger file to V3 for 3.0.0 ([#05376da](https://github.com/edgexfoundry/device-sdk-go/commits/05376da))

### Build 👷

- Update to Go 1.20, Alpine 3.17 and linter v1.51.2 ([#1383](https://github.com/edgexfoundry/device-sdk-go/issues/1383)) ([#a467ed6](https://github.com/edgexfoundry/device-sdk-go/commits/a467ed6))
- Update to latest module w/o TOML package ([#bf4714d](https://github.com/edgexfoundry/device-sdk-go/commits/bf4714d))

## [v2.3.1] Levski - 2023-03-17 (Only compatible with the 2.x releases)

### Bug Fixes 🐛

- Fix device sdk cache inconsistency by registering device service prior to driver initialization ([#4d4ffc7](https://github.com/edgexfoundry/device-sdk-go/commits/4d4ffc7))

## [v2.3.0] Levski - 2022-11-09 (Only compatible with the 2.x releases)

### Features ✨

- Add metrics for count of Events and Readings sent ([#1239](https://github.com/edgexfoundry/device-sdk-go/issues/1239)) ([#5df0661](https://github.com/edgexfoundry/device-sdk-go/commits/5df0661))
- Enable service metrics ([#94ac6d2](https://github.com/edgexfoundry/device-sdk-go/commits/94ac6d2))
- Update device-simple CommandRequestTopic config ([#03803f9](https://github.com/edgexfoundry/device-sdk-go/commits/03803f9))
- Subscribe command request via internal messaging ([#aed16fd](https://github.com/edgexfoundry/device-sdk-go/commits/aed16fd))
- Add interface for accessing the Device Service SDK ([#16c2613](https://github.com/edgexfoundry/device-sdk-go/commits/16c2613))

### Bug Fixes 🐛

- Publish envelope with empty payload when ds-returnevent=no ([#6995182](https://github.com/edgexfoundry/device-sdk-go/commits/6995182))
- Fix empty device name when updating device's serviceName ([#ff1b5b0](https://github.com/edgexfoundry/device-sdk-go/commits/ff1b5b0))
- Remove redundant logic of caching profile ([#e9867ce](https://github.com/edgexfoundry/device-sdk-go/commits/e9867ce))
- Put cmd return 400 when updating empty string to NonString type ([#a1088c2](https://github.com/edgexfoundry/device-sdk-go/commits/a1088c2))
- ProvisionWatcher callbacks return BaseResponse, not 204 No Content. ([#3d985db](https://github.com/edgexfoundry/device-sdk-go/commits/3d985db))

### Code Refactoring ♻

- Use bootstrap handlers in go-mod-bootstrap and refactor device command application layer code ([#1210](https://github.com/edgexfoundry/device-sdk-go/issues/1210)) ([#e5efdee](https://github.com/edgexfoundry/device-sdk-go/commits/e5efdee))
- Update to use deepCopy of messageBusInfo to avoid external adds ([#adba0c6](https://github.com/edgexfoundry/device-sdk-go/commits/adba0c6))

### Documentation 📖

- Publish swagger to 2.3.0 ([#9d1347e](https://github.com/edgexfoundry/device-sdk-go/commits/9d1347e))

### Build 👷

- Upgrade to Go 1.18 and alpine 3.16 ([#91816c1](https://github.com/edgexfoundry/device-sdk-go/commits/91816c1))

## [v2.2.0] Kamakura - 2022-5-11 (Only compatible with the 2.x releases)

### Features ✨

- Add MaxEventSize to limit event size ([#ae5b097](https://github.com/edgexfoundry/device-sdk-go/commits/ae5b097))
- Implement ReadingUnits configuration for device profile changes ([#daeaa2d](https://github.com/edgexfoundry/device-sdk-go/commits/daeaa2d))
- Enable security hardening ([#da52579](https://github.com/edgexfoundry/device-sdk-go/commits/da52579))
- Version bump to roll-in delayed service start feature ([#7a7b6d1](https://github.com/edgexfoundry/device-sdk-go/commits/7a7b6d1))
- Implement ProtocolProperties validation mechanism ([#07054d1](https://github.com/edgexfoundry/device-sdk-go/commits/07054d1))
- Location of client service obtained from the registry ([#936332d](https://github.com/edgexfoundry/device-sdk-go/commits/936332d))
- **webserver:** Include ServiceName in Common Responses ([#402f152](https://github.com/edgexfoundry/device-sdk-go/commits/402f152))

### Bug Fixes 🐛
- Add missing Configuration interface method GetTelemetryInfo ([#f9d12fc](https://github.com/edgexfoundry/device-sdk-go/commits/f9d12fc))
- Update TestMetricsRequest to not fail when using -race ([#aa2b65f](https://github.com/edgexfoundry/device-sdk-go/commits/aa2b65f))
- Update validation API 200 response ([#7c8475a](https://github.com/edgexfoundry/device-sdk-go/commits/7c8475a))
- **configuration:** add handling for custom config on /config endpoint ([#4aeb844](https://github.com/edgexfoundry/device-sdk-go/commits/4aeb844))

### Code Refactoring ♻
- Use go-mod-bootstrap RequestLimitMiddleware for MaxRequestSize ([#b63934f](https://github.com/edgexfoundry/device-sdk-go/commits/b63934f))


### Documentation 📖
- Publish swagger to 2.2.0 ([#a109450](https://github.com/edgexfoundry/device-sdk-go/commits/a109450))
- Correct document links in README ([#588dc1c](https://github.com/edgexfoundry/device-sdk-go/commits/588dc1c))

### Build 👷
- Update to latest go-mod-messaging w/o ZMQ on windows ([#a222f54](https://github.com/edgexfoundry/device-sdk-go/commits/a222f54))
    ```
    BREAKING CHANGE:
    ZeroMQ no longer supported on native Windows for EdgeX
    MessageBus
    ```
- Updated formatting from gofmt 1.17 ([#3c2e1aa](https://github.com/edgexfoundry/device-sdk-go/commits/3c2e1aa))

### Continuous Integration 🔄
- Remove -race for unit tests for now to resolve failures in pipeline ([#a3ef393](https://github.com/edgexfoundry/device-sdk-go/commits/a3ef393))
- Go 1.17 related changes ([#20fc5d6](https://github.com/edgexfoundry/device-sdk-go/commits/20fc5d6))

## [v2.1.0] Jakarta - 2021-11-17 (Only compatible with the 2.x releases)

### Features ✨
- Support object value type in Set Command ([#801bc03](https://github.com/edgexfoundry/device-sdk-go/commits/801bc03))
- Add NewCommandValueWithOrigin function ([#c6c2082](https://github.com/edgexfoundry/device-sdk-go/commits/c6c2082))
- Support Object value type in Reading ([#1025](https://github.com/edgexfoundry/device-sdk-go/issues/1025)) ([#d245461](https://github.com/edgexfoundry/device-sdk-go/commits/d245461))

### Bug Fixes 🐛
- Stop AutoEvents if the Device is locked ([#1027](https://github.com/edgexfoundry/device-sdk-go/issues/1027)) ([#c02be29](https://github.com/edgexfoundry/device-sdk-go/commits/c02be29))
- Fix nil pointer error when executing SET command with empty value ([#0f89794](https://github.com/edgexfoundry/device-sdk-go/commits/0f89794))
- Fix device yaml  to Json  error ([#cf13810](https://github.com/edgexfoundry/device-sdk-go/commits/cf13810))
- Update all TOML to use quote and not single-quote ([#9e077e8](https://github.com/edgexfoundry/device-sdk-go/commits/9e077e8))

### Code Refactoring ♻
- Change V2 Swagger to be published with 2.0 version ([#9dee295](https://github.com/edgexfoundry/device-sdk-go/commits/9dee295))

### Documentation 📖
- Update swagger version to 2.1.0 ([#6cc4e69](https://github.com/edgexfoundry/device-sdk-go/commits/6cc4e69))
- Add apiVersion to request body example ([#1a6f6b9](https://github.com/edgexfoundry/device-sdk-go/commits/1a6f6b9))
- Remove the description about base64 encoding ([#df04f74](https://github.com/edgexfoundry/device-sdk-go/commits/df04f74))
- Update build status badge ([#da9a265](https://github.com/edgexfoundry/device-sdk-go/commits/da9a265))
- Update device-simple README and provisionwatcher example ([#76abb45](https://github.com/edgexfoundry/device-sdk-go/commits/76abb45))

### Build 👷
- Update alpine base to 3.14 ([#7fe965a](https://github.com/edgexfoundry/device-sdk-go/commits/7fe965a))

### Continuous Integration 🔄
- Remove need for CI specific Dockerfile ([#4ea8c13](https://github.com/edgexfoundry/device-sdk-go/commits/4ea8c13))

## [v2.0.0] Ireland - 2021-06-30  (Not Compatible with 1.x releases)

### Features ✨
- Enable using MessageBus as the default ([#eca11b8](https://github.com/edgexfoundry/device-sdk-go/commits/eca11b8))
- support device profile provision in json format ([#945ec1c](https://github.com/edgexfoundry/device-sdk-go/commits/945ec1c))
- add Event tagging capability ([#149daef](https://github.com/edgexfoundry/device-sdk-go/commits/149daef))
- Add secure MessageBus capability ([#57291f0](https://github.com/edgexfoundry/device-sdk-go/commits/57291f0))
- CBOR encoding http response for event with binary reading ([#0032f23](https://github.com/edgexfoundry/device-sdk-go/commits/0032f23))
- update CommandRequest.Attributes type ([#1370f96](https://github.com/edgexfoundry/device-sdk-go/commits/1370f96))
- add request size middleware for device command ([#c006079](https://github.com/edgexfoundry/device-sdk-go/commits/c006079))
- improve Event/Reading Origin logic ([#c5bee4f](https://github.com/edgexfoundry/device-sdk-go/commits/c5bee4f))
- update profile cache to reflect new profile model ([#3ff13b5](https://github.com/edgexfoundry/device-sdk-go/commits/3ff13b5))
- add capability to load devices from directory ([#e2dd6f7](https://github.com/edgexfoundry/device-sdk-go/commits/e2dd6f7))
- Updated Attribution.txt for missing crypto module ([#ed9c44c](https://github.com/edgexfoundry/device-sdk-go/commits/ed9c44c))
- Enable Registry and Config access token ([#40eaaf9](https://github.com/edgexfoundry/device-sdk-go/commits/40eaaf9))
- update example configuration to support MessageBus ([#a8316e3](https://github.com/edgexfoundry/device-sdk-go/commits/a8316e3))
- add MessageBus capablility ([#0fbf16e](https://github.com/edgexfoundry/device-sdk-go/commits/0fbf16e))
- Add support for structured custom configuration ([#0699ac1](https://github.com/edgexfoundry/device-sdk-go/commits/0699ac1))
- add GetProfileByName function API ([#691fc8b](https://github.com/edgexfoundry/device-sdk-go/commits/691fc8b))
- Upgrade core-contracts lib to use redesigned device profile ([#816](https://github.com/edgexfoundry/device-sdk-go/issues/816)) ([#a5ad5ab](https://github.com/edgexfoundry/device-sdk-go/commits/a5ad5ab))
- validate set command parameters against maximum/minimum and refactor transformer package ([#813](https://github.com/edgexfoundry/device-sdk-go/issues/813)) ([#3f52f3a](https://github.com/edgexfoundry/device-sdk-go/commits/3f52f3a))
- update CommandValue for v2 ([#dbb5e95](https://github.com/edgexfoundry/device-sdk-go/commits/dbb5e95))
- handle no readings generated gracefully ([#fd6d2d4](https://github.com/edgexfoundry/device-sdk-go/commits/fd6d2d4))
- spawn autoevents for discovered device ([#4822fad](https://github.com/edgexfoundry/device-sdk-go/commits/4822fad))
- process queryParams of url in PUT request ([#0d64304](https://github.com/edgexfoundry/device-sdk-go/commits/0d64304))
- dynamically adds profile in device callback ([#26939be](https://github.com/edgexfoundry/device-sdk-go/commits/26939be))
- update v2 cache to load data via v2 clients ([#b6352b0](https://github.com/edgexfoundry/device-sdk-go/commits/b6352b0))
- make update callback aware of service change ([#3fc6617](https://github.com/edgexfoundry/device-sdk-go/commits/3fc6617))
- make DeviceService struct updatable ([#ecc71a3](https://github.com/edgexfoundry/device-sdk-go/commits/ecc71a3))
- add update device service callback ([#9ee2204](https://github.com/edgexfoundry/device-sdk-go/commits/9ee2204))
- update v2 callback API ([#b000b54](https://github.com/edgexfoundry/device-sdk-go/commits/b000b54))
- update v2 command API ([#3d1861d](https://github.com/edgexfoundry/device-sdk-go/commits/3d1861d))
- SecretProvider for storing/retrieving secrets ([#bcd0eef](https://github.com/edgexfoundry/device-sdk-go/commits/bcd0eef))
- Updates from PR review. ([#c4cbbed](https://github.com/edgexfoundry/device-sdk-go/commits/c4cbbed))
- Modify callback func to consist with V2 API ([#98ab8c7](https://github.com/edgexfoundry/device-sdk-go/commits/98ab8c7))
- Added missing module to Attribution.txt ([#9a6397b](https://github.com/edgexfoundry/device-sdk-go/commits/9a6397b))
- Add /api/v2/secrets endpoint to store secrets ([#2c57645](https://github.com/edgexfoundry/device-sdk-go/commits/2c57645))
- Removed remote logging feature ([#1cc1ee3](https://github.com/edgexfoundry/device-sdk-go/commits/1cc1ee3))
- **v2:** prepare v2 ProvisionWatcher model cache ([#c95c718](https://github.com/edgexfoundry/device-sdk-go/commits/c95c718))
- **v2:** add v2 ProvisionWatcher callback API ([#ce664a3](https://github.com/edgexfoundry/device-sdk-go/commits/ce664a3))
### Test
- add unit tests for SDK ([#900](https://github.com/edgexfoundry/device-sdk-go/issues/900)) ([#c59eaec](https://github.com/edgexfoundry/device-sdk-go/commits/c59eaec))
- add unit tests for v2 cache ([#4879565](https://github.com/edgexfoundry/device-sdk-go/commits/4879565))
### Bug Fixes 🐛
- Fix device service update failed when startup ([#785b4de](https://github.com/edgexfoundry/device-sdk-go/commit/785b4de))
- fix default value usage in SET command ([#65fbd6f](https://github.com/edgexfoundry/device-sdk-go/commits/65fbd6f))
- modify function updateAssociatedProfile return errors ([#942](https://github.com/edgexfoundry/device-sdk-go/issues/942)) ([#13e775a](https://github.com/edgexfoundry/device-sdk-go/commits/13e775a))
- correctly update DeviceService.deviceService model ([#c48d0fd](https://github.com/edgexfoundry/device-sdk-go/commits/c48d0fd))
- use RequestLimitMiddleware on all routes ([#b69e6f6](https://github.com/edgexfoundry/device-sdk-go/commits/b69e6f6))
- add logLevel check in LoggingMiddleware ([#23eaf68](https://github.com/edgexfoundry/device-sdk-go/commits/23eaf68))
- fix deviceProfileMap refer to the same object ([#cf93359](https://github.com/edgexfoundry/device-sdk-go/commits/cf93359))
- fix set command resourceOperation mapping ([#870](https://github.com/edgexfoundry/device-sdk-go/issues/870)) ([#e9afc7c](https://github.com/edgexfoundry/device-sdk-go/commits/e9afc7c))
- add Content-Type in context for MessageEnvelope ([#2d80251](https://github.com/edgexfoundry/device-sdk-go/commits/2d80251))
- add device OperatingState check in device command API ([#db824b8](https://github.com/edgexfoundry/device-sdk-go/commits/db824b8))
- remove StopAutoEvents call during bootstrap ([#6d0d00b](https://github.com/edgexfoundry/device-sdk-go/commits/6d0d00b))
- remove hard-coded Content-Type in SDK ([#f2441cf](https://github.com/edgexfoundry/device-sdk-go/commits/f2441cf))
- sync v2 query parameter with ADR ([#9940e7e](https://github.com/edgexfoundry/device-sdk-go/commits/9940e7e))
- rm unused pkg/service function params ([#39c0334](https://github.com/edgexfoundry/device-sdk-go/commits/39c0334))
### Code Refactoring ♻
- remove redundant dic client code ([#13aef19](https://github.com/edgexfoundry/device-sdk-go/commits/13aef19))
- Change PublishTopicPrefix value to be 'edgex/events/device' ([#37f218f](https://github.com/edgexfoundry/device-sdk-go/commits/37f218f))
- remove unimplemented InitCmd/RemoveCmd configuration ([#956](https://github.com/edgexfoundry/device-sdk-go/issues/956)) ([#457577c](https://github.com/edgexfoundry/device-sdk-go/commits/457577c))
- Remove obsolete code from Add Secret ([#951](https://github.com/edgexfoundry/device-sdk-go/issues/951)) ([#93abf71](https://github.com/edgexfoundry/device-sdk-go/commits/93abf71))
- Move top level individual config settings under Device section ([#356d22d](https://github.com/edgexfoundry/device-sdk-go/commits/356d22d))
- Use common ServiceInfo struct and adjust code/configuration ([#a3cc839](https://github.com/edgexfoundry/device-sdk-go/commits/a3cc839))
    ```
    BREAKING CHANGE:
    Device Service configuration items have changed
    ```
- Update to assign and uses new Port Assignments ([#1880e37](https://github.com/edgexfoundry/device-sdk-go/commits/1880e37))
- Update for new service key names and overrides for hyphen to underscore ([#8132b7b](https://github.com/edgexfoundry/device-sdk-go/commits/8132b7b))
    ```
    BREAKING CHANGE:
    Service key names used in configuration have changed.
    ```
- rename AutoEvent.Frequency field to Interval ([#0649751](https://github.com/edgexfoundry/device-sdk-go/commits/0649751))
- replace usage of io/ioutl package ([#cb99f86](https://github.com/edgexfoundry/device-sdk-go/commits/cb99f86))
- remove unnecessary check for WriteDeviceCommand ([#e2c4148](https://github.com/edgexfoundry/device-sdk-go/commits/e2c4148))
- remove RO mapping not found warning log ([#bcdc0e5](https://github.com/edgexfoundry/device-sdk-go/commits/bcdc0e5))
- update returned type to errors.Edgex ([#eef323e](https://github.com/edgexfoundry/device-sdk-go/commits/eef323e))
- return normal error type in pkg package ([#a34d60e](https://github.com/edgexfoundry/device-sdk-go/commits/a34d60e))
- refactor MiddlewareFunc ([#fd3eb0c](https://github.com/edgexfoundry/device-sdk-go/commits/fd3eb0c))
- Replace use of BurntSushi/toml with pelletier/go-toml ([#dd0b196](https://github.com/edgexfoundry/device-sdk-go/commits/dd0b196))
- Replace file based with use of Secret Provider to get Access Tokens ([#867](https://github.com/edgexfoundry/device-sdk-go/issues/867)) ([#0004bfd](https://github.com/edgexfoundry/device-sdk-go/commits/0004bfd))
    ```
    BREAKING CHANGE:
    All Device Services running with the secure Edgex Stack now need to have the SecretStore configured, a Vault token created and run with EDGEX_SECURITY_SECRET_STORE=true.
    ```
- Switch to 2.0 Consul path ([#8efa047](https://github.com/edgexfoundry/device-sdk-go/commits/8efa047))
    ```
    BREAKING CHANGE:
    Consul configuration now under the /2.0/ path
    ```
- leverage latest profile cache ([#ff9571a](https://github.com/edgexfoundry/device-sdk-go/commits/ff9571a))
- move config struct to config package ([#5d743f6](https://github.com/edgexfoundry/device-sdk-go/commits/5d743f6))
- update bootstrap sequence ([#0526fd6](https://github.com/edgexfoundry/device-sdk-go/commits/0526fd6))
- remove v1 code and unused constants ([#3a954db](https://github.com/edgexfoundry/device-sdk-go/commits/3a954db))
- remove v2 subdirectory for SDK ([#269686d](https://github.com/edgexfoundry/device-sdk-go/commits/269686d))
- rename problematic terminology in SDK ([#819fd06](https://github.com/edgexfoundry/device-sdk-go/commits/819fd06))
- use new CommandValue function/method ([#df528cb](https://github.com/edgexfoundry/device-sdk-go/commits/df528cb))
- remove default floating encoding ([#1a909d6](https://github.com/edgexfoundry/device-sdk-go/commits/1a909d6))
- replace PUT command wording to SET command ([#6059ff5](https://github.com/edgexfoundry/device-sdk-go/commits/6059ff5))
- update logging message of simpledriver ([#1c5dcca](https://github.com/edgexfoundry/device-sdk-go/commits/1c5dcca))
- remove id map and by id method in cache ([#59cae24](https://github.com/edgexfoundry/device-sdk-go/commits/59cae24))
- upgrade SDK external function API for v2 ([#583d861](https://github.com/edgexfoundry/device-sdk-go/commits/583d861))
- use constants defined in core-contracts ([#91f06a6](https://github.com/edgexfoundry/device-sdk-go/commits/91f06a6))
- upgrade async function to handle v2 models ([#75ab748](https://github.com/edgexfoundry/device-sdk-go/commits/75ab748))
- upgrade bootstrap procoess for v2 ([#d356ded](https://github.com/edgexfoundry/device-sdk-go/commits/d356ded))
- refactor autoevents package for v2 ([#316a12a](https://github.com/edgexfoundry/device-sdk-go/commits/316a12a))
- upgrade command API to handle v2 models ([#b71bc34](https://github.com/edgexfoundry/device-sdk-go/commits/b71bc34))
- upgrade provision package to use v2 models/clients ([#62cbe10](https://github.com/edgexfoundry/device-sdk-go/commits/62cbe10))
- update clients package to use v2 client ([#801d88e](https://github.com/edgexfoundry/device-sdk-go/commits/801d88e))
- update container package for v2 ([#ef8f9ac](https://github.com/edgexfoundry/device-sdk-go/commits/ef8f9ac))
- implement new utils function for v2 ([#d7fb500](https://github.com/edgexfoundry/device-sdk-go/commits/d7fb500))
- upgrade transformer package for v2 ([#82016d5](https://github.com/edgexfoundry/device-sdk-go/commits/82016d5))
- upgrade SDK models to use v2 contract(model) ([#de17357](https://github.com/edgexfoundry/device-sdk-go/commits/de17357))
- remove v1 cache and API code ([#b998828](https://github.com/edgexfoundry/device-sdk-go/commits/b998828))
- leverage go-mod-core-contracts constants ([#839fdd7](https://github.com/edgexfoundry/device-sdk-go/commits/839fdd7))
- consume edgex v2 go-mods ([#71e5df7](https://github.com/edgexfoundry/device-sdk-go/commits/71e5df7))
- remove ValueType enumeration in Go-SDK ([#ba4bab1](https://github.com/edgexfoundry/device-sdk-go/commits/ba4bab1))
- Make SDK a V2 Go Module ([#e04106a](https://github.com/edgexfoundry/device-sdk-go/commits/e04106a))
- Fixup to change all occurances of `secrets' to `secret` ([#24db2b5](https://github.com/edgexfoundry/device-sdk-go/commits/24db2b5))
### Documentation 📖
- Add badges to readme ([#c8eb33a](https://github.com/edgexfoundry/device-sdk-go/commits/c8eb33a))
- update v2 API swagger file ([#97747a4](https://github.com/edgexfoundry/device-sdk-go/commits/97747a4))
- **v2:** update schema to be consistent with edgex-go ([#15155f8](https://github.com/edgexfoundry/device-sdk-go/commits/15155f8))
### Build 👷
- update snap build to use go1.16 via a build-snap ([#bb5fdaa](https://github.com/edgexfoundry/device-sdk-go/commits/bb5fdaa))
- upgrade Golang to 1.16 ([#9773bb9](https://github.com/edgexfoundry/device-sdk-go/commits/9773bb9))
- update snap build to use Golang 1.16 ([#6dcb0bd](https://github.com/edgexfoundry/device-sdk-go/commits/6dcb0bd))
### Continuous Integration 🔄
- update docker image name ([#a051116](https://github.com/edgexfoundry/device-sdk-go/commits/a051116))
- update build files to support zmq dependency ([#0b2e4f3](https://github.com/edgexfoundry/device-sdk-go/commits/0b2e4f3))
- standardize dockerfiles ([#4682622](https://github.com/edgexfoundry/device-sdk-go/commits/4682622))

<a name="v1.4.0"></a>
## [v1.4.0] - 2021-01-08
### Features ✨
- update DS in metadata to reflect config change ([#a5cd81b](https://github.com/edgexfoundry/device-sdk-go/commits/a5cd81b))
- add a comment to explain why use buffer for sending event ([#e273dc8](https://github.com/edgexfoundry/device-sdk-go/commits/e273dc8))
- autoevent adds buffer for sending event to coredata ([#1c85f20](https://github.com/edgexfoundry/device-sdk-go/commits/1c85f20))
- Make the numeric value type allows overflow and NaN ([#1c75f93](https://github.com/edgexfoundry/device-sdk-go/commits/1c75f93))
- remove AdminState check for callback api route ([#b6110b2](https://github.com/edgexfoundry/device-sdk-go/commits/b6110b2))
- add v1 deviceService callback handler ([#4cd2f77](https://github.com/edgexfoundry/device-sdk-go/commits/4cd2f77))
- **sdk:** Implement Device AutoEvents SDK APIs ([#e78e4a6](https://github.com/edgexfoundry/device-sdk-go/commits/e78e4a6))
### Refactor
- remove github.com/pkg/errors from Attribution.txt ([#06df777](https://github.com/edgexfoundry/device-sdk-go/commits/06df777))
### Bug Fixes 🐛
- use pointer on Executor to correctly update it ([#2bf6939](https://github.com/edgexfoundry/device-sdk-go/commits/2bf6939))
### Documentation 📖
- update release note for 1.4.0 release ([#00c0363](https://github.com/edgexfoundry/device-sdk-go/commits/00c0363))
### Build 👷
- **deps:** bump gopkg.in/yaml.v2 from 2.3.0 to 2.4.0 ([#7e6026a](https://github.com/edgexfoundry/device-sdk-go/commits/7e6026a))
- **deps:** bump github.com/gorilla/mux from 1.7.1 to 1.8.0 ([#ea97e15](https://github.com/edgexfoundry/device-sdk-go/commits/ea97e15))
- **deps:** bump github.com/edgexfoundry/go-mod-bootstrap ([#4691506](https://github.com/edgexfoundry/device-sdk-go/commits/4691506))
### Continuous Integration 🔄
- add semantic.yml for commit linting, update PR template to latest ([#ea09293](https://github.com/edgexfoundry/device-sdk-go/commits/ea09293))

<a name="v1.3.0"></a>
## [v1.3.0] - 2020-11-11
### Features ✨
- implement Device SDK v2 common API ([#fa7ee70](https://github.com/edgexfoundry/device-sdk-go/commits/fa7ee70))
- implement v2 command API ([#619](https://github.com/edgexfoundry/device-sdk-go/issues/619)) ([#abee7d5](https://github.com/edgexfoundry/device-sdk-go/commits/abee7d5))
- implement v2 discovery API ([#e0816de](https://github.com/edgexfoundry/device-sdk-go/commits/e0816de))
- **sdk:** implement v2 callback api ([#a3969cc](https://github.com/edgexfoundry/device-sdk-go/commits/a3969cc))
- **sdk:** prepare v2 cache for v2 callback api ([#b24cbb4](https://github.com/edgexfoundry/device-sdk-go/commits/b24cbb4))
- **sdk:** expose RegistryClient registryClient was already available on DeviceService, just made it public ([#096110c](https://github.com/edgexfoundry/device-sdk-go/commits/096110c))
- **sdk:** start auto-discovery upon service startup ([#9693139](https://github.com/edgexfoundry/device-sdk-go/commits/9693139))
### Bug Fixes 🐛
- pass ctx into client initialization function ([#388ecb7](https://github.com/edgexfoundry/device-sdk-go/commits/388ecb7))
- improve device discovery flow and whitelist logic ([#0e0b041](https://github.com/edgexfoundry/device-sdk-go/commits/0e0b041))
- fix autoevent panic during callback ([#4c965aa](https://github.com/edgexfoundry/device-sdk-go/commits/4c965aa))
- fix device autodiscovery behavior ([#4e5d2bb](https://github.com/edgexfoundry/device-sdk-go/commits/4e5d2bb))
- Create a buffer to handle the AsyncValues ([#925ab5b](https://github.com/edgexfoundry/device-sdk-go/commits/925ab5b))
- update CommandExists to only check Device Commands ([#7a69a11](https://github.com/edgexfoundry/device-sdk-go/commits/7a69a11))
- allow startup duration/interval to be overridden via env vars ([#ea2f983](https://github.com/edgexfoundry/device-sdk-go/commits/ea2f983))
- **sdk:** prevent creating duplicate autoevent executor ([#36fb88b](https://github.com/edgexfoundry/device-sdk-go/commits/36fb88b))
### Code Refactoring ♻
- update LogLevel for pushing event to coredata ([#05d1f85](https://github.com/edgexfoundry/device-sdk-go/commits/05d1f85))
- update dockerfile to appropriately use ENTRYPOINT and CMD, closes[#569](https://github.com/edgexfoundry/device-sdk-go/issues/569) ([#6d7be66](https://github.com/edgexfoundry/device-sdk-go/commits/6d7be66))
- **sdk:** fix requested changes in PR ([#97d2159](https://github.com/edgexfoundry/device-sdk-go/commits/97d2159))
- **sdk:** release lock after discovery complete ([#3b7bdf6](https://github.com/edgexfoundry/device-sdk-go/commits/3b7bdf6))
### Documentation 📖
- addition of version and LTS refs to README per arch's meeting ([#bced3e5](https://github.com/edgexfoundry/device-sdk-go/commits/bced3e5))
### Build 👷
- add dependabot.yml ([#623c4a4](https://github.com/edgexfoundry/device-sdk-go/commits/623c4a4))
- upgrade to use Go 1.15 ([#9448c71](https://github.com/edgexfoundry/device-sdk-go/commits/9448c71))
- **deps:** bump gopkg.in/yaml.v2 from 2.2.8 to 2.3.0 ([#bf2f1c2](https://github.com/edgexfoundry/device-sdk-go/commits/bf2f1c2))
- **deps:** bump github.com/OneOfOne/xxhash from 1.2.6 to 1.2.8 ([#e4c69b4](https://github.com/edgexfoundry/device-sdk-go/commits/e4c69b4))
- **deps:** bump github.com/edgexfoundry/go-mod-bootstrap ([#497e2f1](https://github.com/edgexfoundry/device-sdk-go/commits/497e2f1))
- **deps:** bump github.com/edgexfoundry/go-mod-registry ([#a563bd0](https://github.com/edgexfoundry/device-sdk-go/commits/a563bd0))
- **deps:** bump github.com/edgexfoundry/go-mod-bootstrap ([#d1e6a07](https://github.com/edgexfoundry/device-sdk-go/commits/d1e6a07))

<a name="v1.2.3"></a>
## [v1.2.3] - 2020-07-21
### Code Refactoring ♻
- Remove client monitoring ([#57f0453](https://github.com/edgexfoundry/device-sdk-go/commits/57f0453))
### Documentation 📖
- **README.md:** clarity for readers; fix [#537](https://github.com/edgexfoundry/device-sdk-go/issues/537) ([#d265bc9](https://github.com/edgexfoundry/device-sdk-go/commits/d265bc9))

<a name="v1.2.2"></a>
## [v1.2.2] - 2020-06-11

<a name="v1.2.1"></a>
## [v1.2.1] - 2020-05-13
### Features ✨
- replace --serviceName with --instance ([#7ed50e3](https://github.com/edgexfoundry/device-sdk-go/commits/7ed50e3))
- add ability for service name override ([#b7ceb80](https://github.com/edgexfoundry/device-sdk-go/commits/b7ceb80))
### Bug Fixes 🐛
- Upgrade the Go modules to include the fixes ([#d7b2a7f](https://github.com/edgexfoundry/device-sdk-go/commits/d7b2a7f))
- fix device service config stem in consul ([#82cb5f2](https://github.com/edgexfoundry/device-sdk-go/commits/82cb5f2))
- Update to use go-mod-bootstrap to fix issue with override un-done ([#bf5510e](https://github.com/edgexfoundry/device-sdk-go/commits/bf5510e))

<a name="v1.2.0"></a>
## [v1.2.0] - 2020-04-27
### Features ✨
- **environment:** Allow uppercase environment overrides ([#9151334](https://github.com/edgexfoundry/device-sdk-go/commits/9151334))
### Bug
- **MediaType:** Update to latest go-mod-contracts for PropertyValue.MediaType fix ([#84f383f](https://github.com/edgexfoundry/device-sdk-go/commits/84f383f))
### Bug Fixes 🐛
- add update profile logic in device callback ([#2966048](https://github.com/edgexfoundry/device-sdk-go/commits/2966048))
### Documentation 📖
- update release notes ([#1850fbb](https://github.com/edgexfoundry/device-sdk-go/commits/1850fbb))
### Build 👷
- Update relevant files in device-sdk-go for Go 1.13. Closes [#440](https://github.com/edgexfoundry/device-sdk-go/issues/440). ([#23bc509](https://github.com/edgexfoundry/device-sdk-go/commits/23bc509))
### Reverts
- Encode the float as base64 string by LB


<a name="v1.1.2"></a>
## [v1.1.2] - 2020-02-11
### Bug
- **MediaType:** Fuji - Update to latest go-mod-contracts for PropertyValue.MediaType fix ([#1218022](https://github.com/edgexfoundry/device-sdk-go/commits/1218022))

<a name="v1.1.1"></a>
## [v1.1.1] - 2019-12-06

<a name="v1.1.0"></a>
## [v1.1.0] - 2019-11-11
### Features ✨
- **server:** Add API to allow adding additional routes to internal webserver ([#6872dfc](https://github.com/edgexfoundry/device-sdk-go/commits/6872dfc))
### Bug
- **query-params:** Fix panic if attributes isn't initalized ([#dac6f04](https://github.com/edgexfoundry/device-sdk-go/commits/dac6f04))
### Feature
- **query-params:** Handle QueryParams from EdgeX Command Service ([#b172f45](https://github.com/edgexfoundry/device-sdk-go/commits/b172f45))
### Bootstrap
- fix OS signal handling ([#2eb7c7a](https://github.com/edgexfoundry/device-sdk-go/commits/2eb7c7a))

<a name="v1.0.0"></a>
## [v1.0.0] - 2019-06-25

<a name="0.7.1"></a>
## 0.7.1 - 2018-12-10
### Async
- add gdoc ([#e65a7be](https://github.com/edgexfoundry/device-sdk-go/commits/e65a7be))
### Dstore
- minor comments update ([#b112f46](https://github.com/edgexfoundry/device-sdk-go/commits/b112f46))

[Unreleased]: https://github.com/edgexfoundry/device-sdk-go/compare/x.y.z...HEAD
[x.y.z]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.4.0...x.y.z
[v1.4.0]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.2.3...v1.3.0
[v1.2.3]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.2.2...v1.2.3
[v1.2.2]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.2.1...v1.2.2
[v1.2.1]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.2.0...v1.2.1
[v1.2.0]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.1.2...v1.2.0
[v1.1.2]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.1.1...v1.1.2
[v1.1.1]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/edgexfoundry/device-sdk-go/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/edgexfoundry/device-sdk-go/compare/0.7.1...v1.0.0
