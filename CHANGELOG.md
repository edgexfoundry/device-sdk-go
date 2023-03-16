
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
- ProvisionWatcher callbacks return BaseResponse, not 204 No Content ([#3d985db](https://github.com/edgexfoundry/device-sdk-go/commits/3d985db))

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
