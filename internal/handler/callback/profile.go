// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package callback

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/google/uuid"
)

func handleProfile(method string, id string, dic *di.Container) common.AppError {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	if method == http.MethodPut {
		dpc := container.MetadataDeviceProfileClientFrom(dic.Get)
		ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
		profile, err := dpc.DeviceProfile(ctx, id)
		if err != nil {
			appErr := common.NewBadRequestError(err.Error(), err)
			lc.Error(fmt.Sprintf("Cannot find the device profile %s from Core Metadata: %v", id, err))
			return appErr
		}

		err = cache.Profiles().Update(profile)
		if err == nil {
			provision.CreateDescriptorsFromProfile(
				&profile,
				lc,
				container.GeneralClientFrom(dic.Get),
				container.CoredataValueDescriptorClientFrom(dic.Get))
			lc.Info(fmt.Sprintf("Updated device profile %s", id))
			devices := cache.Devices().All()
			driver := container.ProtocolDriverFrom(dic.Get)
			for _, d := range devices {
				if d.Profile.Name == profile.Name {
					d.Profile = profile
					_ = cache.Devices().Update(d)
					err := driver.UpdateDevice(d.Name, d.Protocols, d.AdminState)
					if err != nil {
						lc.Error(fmt.Sprintf("Failed to update device in protocoldriver: %s", err))
					}
				}
			}
		} else {
			appErr := common.NewServerError(err.Error(), err)
			lc.Error(fmt.Sprintf("Couldn't update device profile %s: %v", id, err.Error()))
			return appErr
		}
	} else {
		lc.Error(fmt.Sprintf("Invalid device profile method: %s", method))
		appErr := common.NewBadRequestError("Invalid device profile method", nil)
		return appErr
	}

	return nil
}
