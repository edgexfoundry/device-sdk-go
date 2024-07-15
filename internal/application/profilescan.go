//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
)

type profileScanLocker struct {
	mux     sync.Mutex
	busyMap map[string]bool
}

var locker = profileScanLocker{busyMap: make(map[string]bool)}

func ProfileScanWrapper(busy chan bool, ps interfaces.ProfileScan, req sdkModels.ProfileScanRequest, dic *di.Container) {
	locker.mux.Lock()
	b := locker.busyMap[req.DeviceName]
	busy <- b
	if b {
		locker.mux.Unlock()
		return
	}
	locker.busyMap[req.DeviceName] = true
	locker.mux.Unlock()

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) //nolint: staticcheck

	lc.Debugf("profile scan triggered with device name '%s' and profile name '%s'", req.DeviceName, req.ProfileName)
	profile, err := ps.ProfileScan(req)
	if err != nil {
		lc.Errorf("failed to trigger profile scan: %v", err.Error())
	}
	// Add profile to metadata
	profileReq := requests.NewDeviceProfileRequest(dtos.FromDeviceProfileModelToDTO(profile))
	_, err = dpc.Add(ctx, []requests.DeviceProfileRequest{profileReq})
	if err != nil {
		lc.Errorf("failed to add device profile '%s': %v", profile.Name, err)
	}
	// Update device
	deviceReq := requests.NewUpdateDeviceRequest(dtos.UpdateDevice{Name: &req.DeviceName, ProfileName: &profile.Name})
	_, err = dc.Update(ctx, []requests.UpdateDeviceRequest{deviceReq})
	if err != nil {
		lc.Errorf("failed to update device '%s' with profile '%s': %v", req.DeviceName, profile.Name, err)
	}

	// ReleaseLock
	locker.mux.Lock()
	locker.busyMap[req.DeviceName] = false
	locker.mux.Unlock()
}
