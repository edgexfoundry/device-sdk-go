//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package callback

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/google/uuid"
)

func handleService(method string, id string, dic *di.Container) common.AppError {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	switch method {
	case http.MethodPut:
		handleUpdateService(ctx, id, dic)
	default:
		lc.Error(fmt.Sprintf("Invalid service method type: %s", method))
		appErr := common.NewBadRequestError("Invalid service method", nil)
		return appErr
	}

	return nil
}

func handleUpdateService(ctx context.Context, _ string, dic *di.Container) common.AppError {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dsc := container.MetadataDeviceServiceClientFrom(dic.Get)
	ds := container.DeviceServiceFrom(dic.Get)

	// v1 DeviceServiceClient doesn't support DeviceServiceForId. Use name instead
	// for the minimum development effort purpose. (assuming device service name won't
	// be updated)
	service, err := dsc.DeviceServiceForName(ctx, ds.Name)
	if err != nil {
		appErr := common.NewBadRequestError(err.Error(), err)
		lc.Error(fmt.Sprintf("Cannot find DeviceService %s from Core Metadata: %v", ds.Name, err))
		return appErr
	}

	ds.AdminState = service.AdminState
	dic.Update(di.ServiceConstructorMap{
		container.DeviceServiceName: func(get di.Get) interface{} {
			return ds
		},
	})

	return nil
}
