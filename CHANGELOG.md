<a name="v1.2.0"></a>
## [v1.2.0] - (Compatible w/ Geneva Release)

### Fix
- fixed log message formatting ([#378](https://github.com/rsdmike/app-functions-sdk-go/issues/378))

### Feat
- Add ability to Filter functions to reverse the logic to filter out specified names ([#375](https://github.com/rsdmike/app-functions-sdk-go/issues/375))

### Fix
- Allow overrides that have empty/blank value ([#374](https://github.com/rsdmike/app-functions-sdk-go/issues/374))

### Docs
- update changelog


<a name="v1.1.0"></a>

## [v1.1.0] - (Compatible w/ Geneva Release)

### CI
- github actions experiment ([#366](https://github.com/edgex-foundry/app-functions-sdk-go/issues/366)) [#78b69fc](https://github.com/edgexfoundry/app-functions-sdk-go/commits/78b69fccff132480e5dc738eccde30bbfd5ef5b0)
- allow merge in git history [#62cc162](https://github.com/edgexfoundry/app-functions-sdk-go/commits/62cc162d7a9565c8c9827baaffab1d3e9628bdb6)
- improve conventional commit conformance [#1f63c5f](https://github.com/edgexfoundry/app-functions-sdk-go/commits/1f63c5f5d0e61ef50c2bbddc3e556641e31ff154)
- **jenkins:** remove sandbox file [#531f52b](https://github.com/edgexfoundry/app-functions-sdk-go/commits/531f52b4381da6723911255756d2043c09f46967)
- **VERSION:** Remove VERSION file [#9d74176](https://github.com/edgexfoundry/app-functions-sdk-go/commits/9d74176de65cced5dc65e01930123706a15314ed)

### Docs
- update links to point to v1.2 [#d3c62bb](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d3c62bb0b2cb6c75edf21c1623d18afa109a5ab6)
- move docs to edgex-docs [#76095f3](https://github.com/edgexfoundry/app-functions-sdk-go/commits/76095f36402a5893d3e7af21fe2b4ade8fe7e65c)
- **pr-template:** remove contribution guidelines from PR checklist since commitlint checks this [#4321bad](https://github.com/edgexfoundry/app-functions-sdk-go/commits/4321bad254b0905604ca39c3df66b8d8a7d95151)
- adding batch to TOC [#9695d7b](https://github.com/edgexfoundry/app-functions-sdk-go/commits/9695d7bb5901f08c445b9014a90b9b078cea46c0)
- batch documentation created [#2d51189](https://github.com/edgexfoundry/app-functions-sdk-go/commits/2d51189fa629ad78044fc66389e7a0442e685e44)
- Update PR Template based on feedback [#b1a1b0b](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b1a1b0b7e61563fdda23ce2b3478d1955d5cef25)
- Add webserver usage to ToC [#7ea3b5e](https://github.com/edgexfoundry/app-functions-sdk-go/commits/7ea3b5e1f2b52204b9972f2454cc89d92bbc0924)
- **swagger:** add swagger annotations to generate spec from code [#8e83cab](https://github.com/edgexfoundry/app-functions-sdk-go/commits/8e83cab64cbf4f6a7a610872fa8f352b4d61cb57)
- PR Template [#ec47f61](https://github.com/edgexfoundry/app-functions-sdk-go/commits/ec47f61b6231b2268726c56099f94afbc31eb82a)
- changelog information [#75cbd94](https://github.com/edgexfoundry/app-functions-sdk-go/commits/75cbd945d14ece47bebecd709d3e73255009479f)

### Refactor
- Change serviceName override to be ServiceKey ([#365](https://github.com/edgex-foundry/app-functions-sdk-go/issues/365)) [#85cb718](https://github.com/edgexfoundry/app-functions-sdk-go/commits/85cb718c22fb25b94368edfbbdf8d4014ad727d3)
- **CBOR:** Replace ugorji/go with fixmacker/cbor [#93f855c](https://github.com/edgexfoundry/app-functions-sdk-go/commits/93f855c6736eb14c851865d8166d33e0344a0483)
- **tests:** Fix order of expected vs actual and other clean up [#c0ff507](https://github.com/edgexfoundry/app-functions-sdk-go/commits/c0ff50735eafe8a1d3dfec01d5cf1e4124356200)
- **sdk:** Add MQTT MessageBus Support [#9cc961e](https://github.com/edgexfoundry/app-functions-sdk-go/commits/9cc961ee7ceeab3faf703295cfef30be34b3ae57)
- Updated to use latest core-contracts changes [#7c6633a](https://github.com/edgexfoundry/app-functions-sdk-go/commits/7c6633a2868114e37fef054a208cf630c75b1f80)
- **examples:** Move examples out of SDK into new app-service-examples repo [#ed9e796](https://github.com/edgexfoundry/app-functions-sdk-go/commits/ed9e796467ebb858d59026c8af8ce7293f0bc0af)
- formatted code [#c6dcc18](https://github.com/edgexfoundry/app-functions-sdk-go/commits/c6dcc18fb5a73f41d6b9c1322f8ab7d8cabefffb)
- **sdk:** Update usage of NewSecretClient to use the latest go-mod-secrets [#8b11b1f](https://github.com/edgexfoundry/app-functions-sdk-go/commits/8b11b1f6b8ef8f69042bfbc965d57af4231be745)

### Build
- Updated to latest go-mod-core-contract for bug fix ([#364](https://github.com/edgex-foundry/app-functions-sdk-go/issues/364)) [#aceb24c](https://github.com/edgexfoundry/app-functions-sdk-go/commits/aceb24cef6d482b079ee6328f01d7d2766dd808c)
- **go.mod:** update dependencies [#2da5c5e](https://github.com/edgexfoundry/app-functions-sdk-go/commits/2da5c5e8df9cee839100b8f9db9efae0b68b0a79)
- update go version to 1.13 [#b26dc8a](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b26dc8a22007af8f4aad8b56a097349ac5cf0a37)
- Update relevant files in app-functions-sdk-go for Go 1.13. Close [#280](https://github.com/edgex-foundry/app-functions-sdk-go/issues/280) [#0123828](https://github.com/edgexfoundry/app-functions-sdk-go/commits/0123828641f15b12bd083c6274e0ea3a56407108)
- **Jenkinsfile:** Pipeline changes for Geneva release [#5de66a3](https://github.com/edgexfoundry/app-functions-sdk-go/commits/5de66a37ffe2f039fb8234c8bc86b9b5b4aae7ac)
- **Attribution:** Add missing Attribution.txt file and update makefile test target [#6f1a755](https://github.com/edgexfoundry/app-functions-sdk-go/commits/6f1a755d295b39b1b875bbc897be0f7096c04b6e)
- **go.mod:** Add running go mod tidy to `make test` [#d24fbcd](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d24fbcddb21966812ed31c57c85bd14d181406b3)
- **makefile:** allow building in gopath by setting GO111MODULE=on [#d11277d](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d11277d5f38fef38c096b023f070fb808676fd41)

### Fix
- Add more sleep time to Batch and Send unit test to fix ARM CI failures ([#361](https://github.com/edgex-foundry/app-functions-sdk-go/issues/361)) [#2c4cbff](https://github.com/edgexfoundry/app-functions-sdk-go/commits/2c4cbffe10a0d09bfbe679fb680918d3d34391a2)
- Use correct parameter key name for MQTTSecretSend AuthMode in configurable pipeline ([#358](https://github.com/edgex-foundry/app-functions-sdk-go/issues/358)) [#b47159d](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b47159de68f9b2b082cb855b11b78af65f3baea7)
- Implement smarter configuration update processing ([#354](https://github.com/edgex-foundry/app-functions-sdk-go/issues/354)) [#678d12a](https://github.com/edgexfoundry/app-functions-sdk-go/commits/678d12aee65b98b12cab7ea8d7bc0b6019a8dc80)
- Added longer sleep to fix intermittent unit test failure on ARM ([#352](https://github.com/edgex-foundry/app-functions-sdk-go/issues/352)) [#65b44ef](https://github.com/edgexfoundry/app-functions-sdk-go/commits/65b44efd1334d51050433c739a090a460d433904)
- JSONLogic now runs rules everytime insted of 1st time [#e83dc16](https://github.com/edgexfoundry/app-functions-sdk-go/commits/e83dc16c714ff806052072b14cc5f609f337ef7a
- **retry loop:** Wrap version check and DB connection is a retry loop instead of sending an error ([#345](https://github.com/edgex-foundry/app-functions-sdk-go/issues/345)) [#1bfa060](https://github.com/edgexfoundry/app-functions-sdk-go/commits/1bfa060a7faf4a975bdc949262c5a9b7f5a0b108)
- Use credentials from Database config if not found in InsecureSecrets [#5c97927](https://github.com/edgexfoundry/app-functions-sdk-go/commits/5c97927956d1d0b933f73c35dbcc06c3652c8a35)
- Remove code that returns empty credentials for Redis [#bd9dac5](https://github.com/edgexfoundry/app-functions-sdk-go/commits/bd9dac5d163e84f492055640cd1d69b80365eaff)
- Handle deprecated edgex_service env variable [#9e68ba5](https://github.com/edgexfoundry/app-functions-sdk-go/commits/9e68ba58953551d317afb2ac875dca7841cf53a9)
- **profile:** Set profile properly in service's service key when env override used [#f6dd20a](https://github.com/edgexfoundry/app-functions-sdk-go/commits/f6dd20a966180cb0215facf4a8ee4a04eafda3dc)
- **SecretClient:** Initialization of secret client retry logic [#ba62973](https://github.com/edgexfoundry/app-functions-sdk-go/commits/ba629738dce9a205f3a0d626e1e52673d75cd415)
- **SecurityProvider:** Make initialization of secret clients optional [#4b86353](https://github.com/edgexfoundry/app-functions-sdk-go/commits/4b86353ccd0cf476c521e5f7397cd6fd3dcc5a1c)
- **batch:** 2nd batch hanging in count mode [#3879fbb](https://github.com/edgexfoundry/app-functions-sdk-go/commits/3879fbbfb1cb5079791212876c9fef1773f159d4)
- **go.mod:** Removed wrong version of ZMQ package used. [#4bd3797](https://github.com/edgexfoundry/app-functions-sdk-go/commits/4bd3797f7e3c90230197f799574c788ee38a690b)
- **trigger:** invoke connect on initialization [#b5a07d6](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b5a07d60f827f2d5781e28fb3efa304f8b87fe90)
- **StoreForward:** Add missing retrieval of DB credentials from Vault [#e2e81ce](https://github.com/edgexfoundry/app-functions-sdk-go/commits/e2e81ce02b764d0fa53497138e45614d92e2d2d3)
- **urlclient:** Update contracts version to fix bug in URLClient. [#a8ba403](https://github.com/edgexfoundry/app-functions-sdk-go/commits/a8ba403fc1b983cbcdc92313387c54f8e8ef721b)
- **README:** Fix example code in README to not panic if LoggingClient not initialized [#a7f6acd](https://github.com/edgexfoundry/app-functions-sdk-go/commits/a7f6acd5f208df2c37dabf583c70d5e947e47c29)

### Test
- fix race condition in batch tests [#87f21c6](https://github.com/edgexfoundry/app-functions-sdk-go/commits/87f21c67bb661aedd7b21d02aadfca8cffa5c3fc)
- fix timing issue with Batch transform test [#701e960](https://github.com/edgexfoundry/app-functions-sdk-go/commits/701e9602d8650aa65a71d33bff5ea1f4c5bba0ef)


### Feat
- Add ability for command-line and environment override of service name ([#356](https://github.com/edgex-foundry/app-functions-sdk-go/issues/356)) [#dcb01ac](https://github.com/edgexfoundry/app-functions-sdk-go/commits/dcb01ac20ebe7429154ec8a60478a159f5f2a3e7)
- Integrate with new redis streams message bus implementation [#6fcbfc4](https://github.com/edgexfoundry/app-functions-sdk-go/commits/6fcbfc4cce97834a8ff210bf562632a32b5e2bea)
- **bootstrap:** Integrate go-mod-bootstrap for common bootstraping [#1034e84](https://github.com/edgexfoundry/app-functions-sdk-go/commits/1034e843966f0cfc729a3c3cb90a33a655031538)
- **configurable:** add mqtt secret support [#d9433ed](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d9433ed29dd497db688285e6fd6ec8a1607b049d)
- **mqtt:** add security provider support for mqtt connection [#9695290](https://github.com/edgexfoundry/app-functions-sdk-go/commits/96952909b8f75930d3783bd82df42d74fbcd53f7)
- **configurable:** support secrets for http export [#3358642](https://github.com/edgexfoundry/app-functions-sdk-go/commits/3358642d990a778e7a2658256a443ec35db6da73)
- **configurable:** add JSONLogic [#e05bd13](https://github.com/edgexfoundry/app-functions-sdk-go/commits/e05bd13fcf2a28a4310cc12c5ea32597be2008fe)
- **configurable:** add batch functions [#3ef7d39](https://github.com/edgexfoundry/app-functions-sdk-go/commits/3ef7d39f8f6b264ec371b2a4968901b46d76af58
- **security:** Add second SecurityStore client for service specific secrets [#204e3ef](https://github.com/edgexfoundry/app-functions-sdk-go/commits/204e3ef09e8796444a678ead5b27635cec4a7f53)
- **http-export:** add support for auth token in header [#311414e](https://github.com/edgexfoundry/app-functions-sdk-go/commits/311414eaf1a4185adc5b6aaf0b173f66aff4865f)
- **jsonlogic:** add filter feature of jsonlogic [#9637eb0](https://github.com/edgexfoundry/app-functions-sdk-go/commits/9637eb052d817c64201bcef4324904801906da99)
- **core contracts:** Upgrade to latest Core Contracts for Reading enhancements [#a93dbb5](https://github.com/edgexfoundry/app-functions-sdk-go/commits/a93dbb5f393ad872a26564e88def9b0b0b791046)
- **sdk:** Add full path to secrets api route [#9f72141](https://github.com/edgexfoundry/app-functions-sdk-go/commits/9f721413d517c233cbb164c246d07643b770f419)
- **sdk:** Implement StoreSecrets in app functions SDK [#1f7dc12](https://github.com/edgexfoundry/app-functions-sdk-go/commits/1f7dc121a04dc62f5ce31519e5cb9620ca218a18)
- **sdk:** Add support for insecure secrets for when running non secure mode [#ad238fe](https://github.com/edgexfoundry/app-functions-sdk-go/commits/ad238fe4badd48867bb5a76ffc88ff73656e202c)
- **sdk:** Add helper function to SDK to get string slice from App Settings. [#f83b325](https://github.com/edgexfoundry/app-functions-sdk-go/commits/f83b325c589d2c04e3b00ce06b10276304bb31e6)
- **appsdk:** Add support for HTTPs on REST trigger [#b594893](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b594893db049378d5034dd414c5e3bbed4c0d3e6)
- **appsdk:** Add support for HTTPs on REST trigger [#b9ccbab](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b9ccbab44b015e43655a570c79d36fff832ade4d)
- **configuration:** Add overwrite option for force local settings into Registry [#7b6318d](https://github.com/edgexfoundry/app-functions-sdk-go/commits/7b6318d55c9a8868e61ffcdf541d5815e2b7ace8)
- **version:** Validate that SDK's major version matches Core Service's major version [#d91fdf1](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d91fdf17dd37f85215f4a27c223eb05c05aa796c)
- **appsdk:** Change configuration intervals to duration strings [#e80ce9a](https://github.com/edgexfoundry/app-functions-sdk-go/commits/e80ce9a2aeb0ea62d44827946635d5d3529cd9b2)
- **MqttSend:** Add SkipCertVerify setting and refactor MqttSend [#2c25a52](https://github.com/edgexfoundry/app-functions-sdk-go/commits/2c25a52df271937deb2ce7b21c670b95a6607cf8)
- **appsdk:** Appsdk changes for Store and Forward. [#211efe4](https://github.com/edgexfoundry/app-functions-sdk-go/commits/211efe43393dd1aadfeb65555a9b69ec5eb8d223)
- batch and send [#1a44398](https://github.com/edgexfoundry/app-functions-sdk-go/commits/1a443984c10091fe4c084316853c9d717e25e0ac)

### BREAKING CHANGE

Inserting preceding "-" when replacing `<profile>` in the service key has been removed so the use is more flexible.  The only service using the <profile> replacement text is app-service-configurable which will be updated to add the "-" in the initial service key.




<a name="v1.0.0"></a>

## v1.0.0 - (Compatible w/ Fuji Release)

### Build
- **go.mod:** Add running go mod tidy to `make test`
- **makefile:** allow building in gopath by setting GO111MODULE=on

### Docs
- **readme:** Address unknown type issue from getting started section [#a6b9976](https://github.com/edgexfoundry/app-functions-sdk-go/commits/a6b9976a8029fb227ec8ab9e9a5a2e745c83c1de)
- **readme:** Updated sample code in readme [#2fbe312](https://github.com/edgexfoundry/app-functions-sdk-go/commits/2fbe3123321487567d34c8c8fd295a346a559566)
- **contributing:** Document suggested format for commits [#b264877](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b264877bf3a113685abc8fca4110303f95cd0eb0)
- fix typo "rigistry" -> "registry" in README [#0cce673](https://github.com/edgexfoundry/app-functions-sdk-go/commits/0cce673dae29c5d9de18e2899f894809b72caff2)
- **toc:** Adding a Table of Contents [#08620d2](https://github.com/edgexfoundry/app-functions-sdk-go/commits/08620d2539279ac837dbf77a5d8672f5dc054bb8)

### Feat
- **Context:** Add useful edgex clients to expose them for pipeline functions and internal use. [#29978f0](https://github.com/edgexfoundry/app-functions-sdk-go/commits/29978f0e7e085c3ad14e955610225d18367530c7)
- **Filter:** Pass all events/reading through if no filter values are set [#ad8e2ed](https://github.com/edgexfoundry/app-functions-sdk-go/commits/ad8e2eda94786750aa928f8846ddc8b2f23e52fb)
- **configurable:** Expose MarkAsPushed [#d86d0a0](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d86d0a046f06accc1ced5363522322ba635f3bd9)
- **contracts:** Update to latest Core Contracts for new Command APIs
- **coredata:** Provide API to push to core-data [#d18e9d2](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d18e9d28c3811c6373fc24ec26c1dba087cc85a5)
- **coredata:** MarkAsPushed is now available as a standalone function [#fdc4f0e](https://github.com/edgexfoundry/app-functions-sdk-go/commits/fdc4f0eb500919d2b15c96eaaeee9bd036852801)
- **examples:** Add example to demonstrate using TargetType [#1b9758f](https://github.com/edgexfoundry/app-functions-sdk-go/commits/1b9758f9fdd80524672e62b1b1fc8d1d8638556c)
- **mqtt:** Support to pass MQTT key/pair as byte array [#985c91b](https://github.com/edgexfoundry/app-functions-sdk-go/commits/985c91b7bdd71a4941081652db9c721adfdb6fbd)
- **profile:** Add environment override for profile command line argument [#c75d2ca](https://github.com/edgexfoundry/app-functions-sdk-go/commits/c75d2ca311cf01d1f41677fa0461fda4c8db8bae)
- **runtime:** Support types other than Event to be sent to function pipeline [#ee6cf0e](https://github.com/edgexfoundry/app-functions-sdk-go/commits/ee6cf0e21a9a699882e6a01b26d773195b013f72)
- **runtime:** Store and Forward core implementation in runtime package. [#1d28cc9](https://github.com/edgexfoundry/app-functions-sdk-go/commits/1d28cc9bd8a247a7649066d0c55a5a134e66f123)
- **store:** Redid Mongo integration tests. [#132f2fc](https://github.com/edgexfoundry/app-functions-sdk-go/commits/132f2fc6f4de49d68cf0c11dce4df70e348c6e87)
- **store:** Added error test cases. [#52e7605](https://github.com/edgexfoundry/app-functions-sdk-go/commits/52e7605fc4441dcc92fe0ed59fea3cfc451d31d8)
- **store:** add abstraction for StoredObject. [#b8d7b6a](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b8d7b6a90b290be782bcedcdc1a11b24f1766e04)
- **store:** Explicitly return values, fix missing imports on test. [#93fbaa2](https://github.com/edgexfoundry/app-functions-sdk-go/commits/93fbaa2d69efba7cf596371006466f1b17db5b0c)
- **store:** Address PR feedback. [#8ab3aba](https://github.com/edgexfoundry/app-functions-sdk-go/commits/8ab3abafc29e5da856c46e96cc8941b2fd1f88e2)
- **store:** add mongo driver [#48f9171](https://github.com/edgexfoundry/app-functions-sdk-go/commits/48f9171211f3338585fff306c1493774909d5532)
- **store:** Updated to remove all indexing by ObjectID. [#01c114b](https://github.com/edgexfoundry/app-functions-sdk-go/commits/01c114b485bb9493c6be2bdb0b1b4950d3597e29)
- **store:** Added contract validation and tests. [#354adfb](https://github.com/edgexfoundry/app-functions-sdk-go/commits/354adfb60b029955ee0ad293beaa4ca574546151)
- **store:** Added Redis driver. [#4f8ef02](https://github.com/edgexfoundry/app-functions-sdk-go/commits/4f8ef02dbad0203677097bdf826f956ba7d3c588)
- **store:** Refactored validateContract(). [#50b0712](https://github.com/edgexfoundry/app-functions-sdk-go/commits/50b07120bc40e26d4e13092e236023e6717c2f3d)
- **store:** Add mock implementation for unit testing. [#5cd4eaf](https://github.com/edgexfoundry/app-functions-sdk-go/commits/5cd4eaf22e75ca0ba577fb34a09e242a7372a8ae)
- **transforms:** Add ability to persist data for later retry to export functions [#351bbc2](https://github.com/edgexfoundry/app-functions-sdk-go/commits/351bbc2f3ff13edf291811f9d4f9b643fe0854b5)
- **webserver:** Expose webserver to enable developer to add routes. [#e48170e](https://github.com/edgexfoundry/app-functions-sdk-go/commits/e48170e57bb375ab978c6266843c60df94fa4b73)
- **webserver:** Docs and tests for webserver use [#3d5ac67](https://github.com/edgexfoundry/app-functions-sdk-go/commits/3d5ac6749ff069c2cd905994871ae8c92a7345fd)
- **version:** Add /api/version endpoint to SDK [#d9fdfd0](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d9fdfd09028a403bf62761950bbc03eff6d6bc21)
- **contracts:** Update to latest Core Contracts for new Command APIs [#e818c23](https://github.com/edgexfoundry/app-functions-sdk-go/commits/e818c23731875fa45f8f6598bb2c3f6ae1c80292)

### Fix
- **TargetType:** Make copy of target type before using it. [#069304b](https://github.com/edgexfoundry/app-functions-sdk-go/commits/069304bcec0a60cc9c000379542aca65b36f3a6a)
- **configuration:** Utilize protocol from [Service] configuration [#c6bec4a](https://github.com/edgexfoundry/app-functions-sdk-go/commits/c6bec4a610f184f503ca929c6ec075fd6db56f67)
- **configuration:** Check Interval is now respected [#06a310f](https://github.com/edgexfoundry/app-functions-sdk-go/commits/06a310fdb705a51e152ae599b648e3286d927c8c)
- **logging:** When trace is enabled, log message for topic subscription is correct [#ebe38a9](https://github.com/edgexfoundry/app-functions-sdk-go/commits/ebe38a9fd19ecbc8c3bc7d4f109250e41776d855)
- **pushtocore:** error not returned to pipeline [#61a3c1b](https://github.com/edgexfoundry/app-functions-sdk-go/commits/61a3c1b4d7db5ded7e8f409b6c8edc02696a5dc6)
- **trigger:** Return error to HTTP trigger response [#af60e79](https://github.com/edgexfoundry/app-functions-sdk-go/commits/af60e79f3432236f7a98f43ea89ea2b643aae75e)
- **webserver:** Timeout wasn't be used [#df39230](https://github.com/edgexfoundry/app-functions-sdk-go/commits/df392302bd55c96e3c2e9d8b883d17dfa3708593)
- **CommandClient:** Use proper API Route for Command Client [#b76f85c](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b76f85ca6626ae8c2dd3ae95a83275a7daa3a6c0)
- **log-filename:** filename specified in configuration.toml was not being respected [#9007019](https://github.com/edgexfoundry/app-functions-sdk-go/commits/90070193dda29078c1b818f547bad7950025c40d)

### Perf
- **db.redis:** Denormalize AppServiceKey out of store object to optimize update  [#d065621](https://github.com/edgexfoundry/app-functions-sdk-go/commits/d065621b3fa2f76f22643455dd78c9a9425decaf)

### Refactor
- Ensure test names are consistent with function names [#b1e3b13](https://github.com/edgexfoundry/app-functions-sdk-go/commits/b1e3b13ca867f15606ad7cd8076f209d05e2766d)
- **sdk:** Refactor to use New func pattern instead of helper functions [#105f120](https://github.com/edgexfoundry/app-functions-sdk-go/commits/105f1202652f1599f82630f1a6bb0ea0cd0584f2)

### BREAKING CHANGE

Pipeline functions used in the SetPipeline() now need to be created with the provided New…() functions.
`/trigger` endpoint now follows standard edgex convention. It is now `/api/v1/trigger`
HTTPPost and MQTTSend no longer automatically call MarkAsPushed upon success. It is upon the developer to ensure the method is called appropriately.
Pipeline functions used in the SetPipeline() now need to be created with the provided New…() functions.


