//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
)

const (
	CorePreReleaseVersion = "master"
	CoreDeveloperVersion  = "0.0.0"
	CoreServiceVersionKey = "version"
	VersionMajorIndex     = 0
)

// VersionValidator contains references to dependencies required by the Version Validation bootstrap implementation.
type VersionValidator struct {
	skipVersionCheck bool
	sdkVersion       string
}

// NewVersionValidator create a new instance of VersionValidator
func NewVersionValidator(skip bool, sdkVersion string) *VersionValidator {
	return &VersionValidator{
		skipVersionCheck: skip,
		sdkVersion:       sdkVersion,
	}
}

// BootstrapHandler verifies that Core Services major version matches this SDK's major version
func (vv *VersionValidator) BootstrapHandler(
	_ context.Context,
	_ *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	logger := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	if vv.skipVersionCheck {
		logger.Info("Skipping core service version compatibility check")
		return true
	}

	// SDK version is set via the SemVer TAG at build time
	// and has the format "v{major}.{minor}.{patch}[-dev.{build}]"
	sdkVersionParts := strings.Split(vv.sdkVersion, ".")
	if len(sdkVersionParts) < 3 {
		logger.Error("SDK version is malformed", "version", internal.SDKVersion)
		return false
	}

	sdkVersionParts[VersionMajorIndex] = strings.Replace(sdkVersionParts[VersionMajorIndex], "v", "", 1)
	if sdkVersionParts[VersionMajorIndex] == "0" {
		logger.Info(
			"Skipping version compatibility check for SDK Beta version or running in debugger",
			"version",
			internal.SDKVersion)
		return true
	}

	url := config.Clients[CoreDataClientName].Url() + clients.ApiVersionRoute
	var data []byte
	var err error
	for startupTimer.HasNotElapsed() {
		if data, err = clients.GetRequestWithURL(context.Background(), url); err != nil {
			logger.Warn("Unable to get version of Core Services", "error", err)
			startupTimer.SleepForInterval()
			continue
		}
		break
	}

	if err != nil {
		logger.Warn("Unable to get version of Core Services after retries", "error", err)
		return false
	}

	versionJson := map[string]string{}
	err = json.Unmarshal(data, &versionJson)
	if err != nil {
		logger.Error("Unable to un-marshal Core Services version data", "error", err)
		return false
	}

	coreVersion, ok := versionJson[CoreServiceVersionKey]
	if !ok {
		logger.Error(fmt.Sprintf("Core Services version data missing '%s' information", CoreServiceVersionKey))
		return false
	}

	if coreVersion == CorePreReleaseVersion {
		logger.Info(
			"Skipping version compatibility check for Core Pre-release version",
			"version",
			coreVersion)
		return true
	}

	if coreVersion == CoreDeveloperVersion {
		logger.Info(
			"Skipping version compatibility check for Core Developer version",
			"version",
			coreVersion)
		return true
	}

	// Core Service version is reported as "{major}.{minor}.{patch}"
	coreVersionParts := strings.Split(coreVersion, ".")
	if len(coreVersionParts) < 3 {
		logger.Error("Core Services version is malformed", "version", coreVersion)
		return false
	}

	// Do Major versions match?
	if coreVersionParts[0] == sdkVersionParts[0] {
		logger.Debug(
			fmt.Sprintf("Confirmed Core Services version (%s) is compatible with SDK's version (%s)",
				coreVersion,
				internal.SDKVersion))
		return true
	}

	logger.Error(
		fmt.Sprintf("Core services version (%s) is not compatible with SDK's version(%s)",
			coreVersion,
			internal.SDKVersion))

	return false
}
