<a name="unreleased"></a>
## [Unreleased]

### Build
- **go.mod:** Add running go mod tidy to `make test`
- **makefile:** allow building in gopath by setting GO111MODULE=on

### Docs
- **readme:** Address unknown type issue from getting started section
- **readme:** Updated sample code in readme
- **contributing:** Document suggested format for commits
- **readme:** fix typo "rigistry" -> "registry"
- **toc:** Adding a Table of Contents

### Feat
- **Context:** Add useful edgex clients to expose them for pipeline functions and internal use.
- **Filter:** Pass all events/reading through if no filter values are set
- **configurable:** Expose MarkAsPushed
- **contracts:** Update to latest Core Contracts for new Command APIs
- **contracts:** Update to latest Core Contracts for new Command APIs
- **coredata:** Provide API to push to core-data
- **examples:** Add example to demonstrate using TargetType
- **mqtt:** Support to pass MQTT key/pair as byte array
- **profile:** Add environment override for profile command line argument
- **runtime:** Support types other than Event to be sent to function pipeline
- **runtime:** Store and Forward core implementation in runtime package.
- **store:** Redid Mongo integration tests.
- **store:** Added error test cases.
- **store:** add abstraction for StoredObject.
- **store:** Explicitly return values, fix missing imports on test.
- **store:** Address PR feedback.
- **store:** add mongo driver
- **store:** Updated to remove all indexing by ObjectID.
- **store:** Added contract validation and tests.
- **store:** Added Redis driver.
- **store:** Refactored validateContract().
- **store:** Add mock inplementation for unit testing.
- **transforms:** Add ability to persist data for later retry to export functions
- **webserver:** Expose webserver to enable developer to add routes.
- **webserver:** Docs and tests for webserver use
- **core-data:** MarkAsPushed is now available as a standalone function
- **version:** Add /api/version endpoint to SDK

### Fix
- **TargetType:** Make copy of target type before using it.
- **configuration:** Utilize protocol from [Service] configuration
- **configuration:** Check Interval is now respected
- **logging:** When trace is enabled, log message for topic subscription is correct
- **pushtocore:** error not returned to pipeline
- **trigger:** Return error to HTTP trigger response
- **webserver:** Timeout wasn't be used
- **CommandClient:** Use proper API Route for Command Client
- **log-filename:** filename specified in configuration.toml was not being respected

### Perf
- **db.redis:** Denormalize AppServiceKey out of store object to optimize update

### Refactor
- Ensure test names are consistent with function names
- **sdk:** Refactor to use New func pattern instead of helper functions

### BREAKING CHANGE

Pipeline functions used in the SetPipeline() now need to be created with the provided New…() functions.
