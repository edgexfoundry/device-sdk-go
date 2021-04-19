# Go Device Service SDK
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-sdk-go/job/master/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-sdk-go/job/master/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/device-sdk-go/branch/master/graph/badge.svg?token=NoUXyBZgt6)](https://codecov.io/gh/edgexfoundry/device-sdk-go) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/device-sdk-go)](https://goreportcard.com/report/github.com/edgexfoundry/device-sdk-go) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-sdk-go?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/device-sdk-go/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-sdk-go?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/device-sdk-go)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/device-sdk-go) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/device-sdk-go)](https://github.com/edgexfoundry/device-sdk-go/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/device-sdk-go)](https://github.com/edgexfoundry/device-sdk-go/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/device-sdk-go-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/device-sdk-go)](https://github.com/edgexfoundry/device-sdk-go/commits)


## Overview

This repository is a set of Go packages that can be used to build Go-based [device services](https://docs.edgexfoundry.org/1.2/microservices/device/Ch-DeviceServices/) for use within the EdgeX framework.

## Usage

Developers can make their own device service by implementing the [`ProtocolDriver`](https://github.com/edgexfoundry/device-sdk-go/blob/master/pkg/models/protocoldriver.go) interface for their desired IoT protocol, and the `main` function to start the Device Service. To implement the `main` function, the [`startup`](https://github.com/edgexfoundry/device-sdk-go/tree/master/pkg/startup) package can be optionally leveraged, or developers can write customized bootstrap code by themselves.

Please see the provided [simple device service](https://github.com/edgexfoundry/device-sdk-go/tree/master/example) as an example, included in this repository.

## Command Line Options

The following command line options are available

```text
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

## Float value encoding

In EdgeX, float values have two kinds of encoding, [Base64](#base64), and [scientific notation (`eNotation`)](#scientific-notation-e-notation).

> When EdgeX is given (or returns) a float32 or float64 value as a string, the format of the string is by default a base64 encoded little-endian of the float32 or float64 value, but the `floatEncoding` attribute relating to the value may instead specify `eNotation` in which case the representation is a decimal with exponent (eg `1.234e-5`)

The above quote is from the official EdgeX device service requirements document, viewable [on Google Docs here](https://docs.google.com/document/d/1aMIQ0kb46VE5eeCpDlaTg8PP29-DBSBTlgeWrv6LuYk), under the "Device readings" section.

### base64

Currently, the [C device service SDK](https://github.com/edgexfoundry/device-sdk-c) converts float values to [little-endian](https://en.wikipedia.org/wiki/Endianness) binary, which is consistent with the [EdgeX device service specifications](https://docs.google.com/document/d/1aMIQ0kb46VE5eeCpDlaTg8PP29-DBSBTlgeWrv6LuYk). However, the Go device service SDK converts float values to big-endian binary. This inconsistency is due to the fact that the device service specifications changed in the [EdgeX Fuji release](https://www.edgexfoundry.org/release-1-1-fuji/whats-new/) - to track the status of this, please review issue [#457](https://github.com/edgexfoundry/device-sdk-go/issues/457).

In the device profile ([example here](https://github.com/edgexfoundry/device-sdk-go/blob/master/example/cmd/device-simple/res/Simple-Driver.yaml)), configure a [profile property](https://docs.edgexfoundry.org/1.2/microservices/device/profile/Ch-DeviceProfileRef/#profileproperty) with [property values](https://docs.edgexfoundry.org/1.2/microservices/device/profile/Ch-DeviceProfileRef/#propertyvalue) as follows:

```yaml
- name: "Temperature"
  description: "Temperature value"
  properties:
    value: { type: "FLOAT64", readWrite: "RW", floatEncoding: "Base64" }
    units: { type: "String", readWrite: "R", defaultValue: "degrees Celsius" }
```

### Scientific Notation (e-notation)

The SDK will convert incoming string values with [scientific notation (aka e-notation)](https://en.wikipedia.org/wiki/Scientific_notation) representation to float values. To enable this, the `floatEncoding` field should be set to the value `eNotation`, detailed below.

In the device profile ([example here](https://github.com/edgexfoundry/device-sdk-go/blob/master/example/cmd/device-simple/res/Simple-Driver.yaml)), configure a [profile property](https://docs.edgexfoundry.org/1.2/microservices/device/profile/Ch-DeviceProfileRef/#profileproperty) with [property values](https://docs.edgexfoundry.org/1.2/microservices/device/profile/Ch-DeviceProfileRef/#propertyvalue) as follows:

```yaml
- name: "Temperature"
  description: "Temperature value"
  properties:
    value: { type: "FLOAT64", readWrite: "RW", floatEncoding: "eNotation" }
    units: { type: "String", readWrite: "R", defaultValue: "degrees Celsius" }
```

## Community

- Chat: [https://edgexfoundry.slack.com](https://edgexfoundry.slack.com)
- Mailing lists: [https://lists.edgexfoundry.org/mailman/listinfo](https://lists.edgexfoundry.org/mailman/listinfo)

## License

[Apache-2.0](LICENSE)

## Versioning

Please refer to the EdgeX Foundry [versioning policy](https://wiki.edgexfoundry.org/pages/viewpage.action?pageId=21823969) for information on how EdgeX services are released and how EdgeX services are compatible with one another.  Specifically, device services (and the associated SDK), application services (and the associated app functions SDK), and client tools (like the EdgeX CLI and UI) can have independent minor releases, but these services must be compatible with the latest major release of EdgeX.

## Long Term Support

Please refer to the EdgeX Foundry [LTS policy](https://wiki.edgexfoundry.org/display/FA/Long+Term+Support) for information on support of EdgeX releases. The EdgeX community does not offer support on any non-LTS release outside of the latest release.