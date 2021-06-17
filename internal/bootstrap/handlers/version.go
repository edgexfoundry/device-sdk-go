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
	"strings"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	clients "github.com/edgexfoundry/go-mod-core-contracts/v2/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDtos "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

const (
	CorePreReleaseVersion = "master"
	CoreDeveloperVersion  = "0.0.0"
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
		logger.Errorf("SDK version is malformed: version=%s", internal.SDKVersion)
		return false
	}

	sdkVersionParts[VersionMajorIndex] = strings.Replace(sdkVersionParts[VersionMajorIndex], "v", "", 1)
	if sdkVersionParts[VersionMajorIndex] == "0" {
		logger.Infof("Skipping version compatibility check for SDK Beta version or running in debugger: version=%s",
			internal.SDKVersion)
		return true
	}

	// Using Core Metadata for Version Check since Core Data can now be optional.
	// Core Metadata will never be optional.
	val, ok := config.Clients[common.CoreMetaDataServiceKey]
	if !ok {
		logger.Error("Unable to get version of Core Metadata: Core Metadata missing from Clients configuration")
		return false
	}

	client := clients.NewCommonClient(val.Url())

	var response commonDtos.VersionResponse
	var err error
	for startupTimer.HasNotElapsed() {
		if response, err = client.Version(context.Background()); err != nil {
			logger.Warnf("Unable to get version of Core Metadata: %s", err.Error())
			startupTimer.SleepForInterval()
			continue
		}
		break
	}

	if err != nil {
		logger.Errorf("Unable to get version of Core Metadata after retries: %s", err.Error())
		return false
	}

	coreVersion := response.Version

	if coreVersion == CorePreReleaseVersion {
		logger.Infof("Skipping version compatibility check for Core Services Pre-release version: version=%s", coreVersion)
		return true
	}

	if coreVersion == CoreDeveloperVersion {
		logger.Infof("Skipping version compatibility check for Core Services Developer version: version=%s", coreVersion)
		return true
	}

	// Core Service version is reported as "{major}.{minor}.{patch}"
	coreVersionParts := strings.Split(coreVersion, ".")
	if len(coreVersionParts) < 3 {
		logger.Errorf("Core Services version is malformed: version=%s", coreVersion)
		return false
	}

	// Do Major versions match?
	if coreVersionParts[0] == sdkVersionParts[0] {
		logger.Debugf("Confirmed Core Services version (%s) is compatible with SDK's version (%s)",
			coreVersion,
			internal.SDKVersion)
		return true
	}

	logger.Errorf("Core Services version (%s) is not compatible with SDK's version(%s)",
		coreVersion,
		internal.SDKVersion)

	return false
}
