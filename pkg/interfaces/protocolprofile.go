// Copyright (C) 2024 IOTech Ltd

package interfaces

import (
	sdkModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// ProfileScan is a low-level device-specific interface implemented
// by device services that support dynamic profile scan.
type ProfileScan interface {
	// ProfileScan triggers specific device to discover device profile.
	ProfileScan(req sdkModels.ProfileScanRequest) (model.DeviceProfile, error)
}
