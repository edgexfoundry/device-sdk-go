// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package callback

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	"github.com/google/uuid"
)

func handleProfile(method string, id string) common.AppError {
	if method == http.MethodPut {
		ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
		profile, err := common.DeviceProfileClient.DeviceProfile(ctx, id)
		if err != nil {
			appErr := common.NewBadRequestError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Cannot find the device profile %s from Core Metadata: %v", id, err))
			return appErr
		}

		err = cache.Profiles().Update(profile)
		if err == nil {
			provision.CreateDescriptorsFromProfile(&profile)
			common.LoggingClient.Info(fmt.Sprintf("Updated device profile %s", id))
			devices := cache.Devices().All()
			for _, d := range devices {
				if d.Profile.Name == profile.Name {
					d.Profile = profile
					_ = cache.Devices().Update(d)
					err := common.Driver.UpdateDevice(d.Name, d.Protocols, d.AdminState)
					if err != nil {
						common.LoggingClient.Error(fmt.Sprintf("Failed to update device in protocoldriver: %s", err))
					}
				}
			}
		} else {
			appErr := common.NewServerError(err.Error(), err)
			common.LoggingClient.Error(fmt.Sprintf("Couldn't update device profile %s: %v", id, err.Error()))
			return appErr
		}
	} else if method == http.MethodDelete{
		cache.Profiles().Remove(id)
	}else {
		common.LoggingClient.Error(fmt.Sprintf("Invalid device profile method: %s", method))
		appErr := common.NewBadRequestError("Invalid device profile method", nil)
		return appErr
	}

	return nil
}
