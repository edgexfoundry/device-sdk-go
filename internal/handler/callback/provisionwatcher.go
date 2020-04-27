// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package callback

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/google/uuid"
)

func handleProvisionWatcher(method string, id string) common.AppError {
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	switch method {
	case http.MethodPost:
		handleAddProvisionWatcher(ctx, id)
	case http.MethodPut:
		handleUpdateProvisionWatcher(ctx, id)
	case http.MethodDelete:
		handleDeleteProvisionWatcher(id)
	default:
		common.LoggingClient.Error(fmt.Sprintf("Invalid provisionwatcher method type: %s", method))
		appErr := common.NewBadRequestError("Invalid provisionwatcher method", nil)
		return appErr
	}

	return nil
}

func handleAddProvisionWatcher(ctx context.Context, id string) common.AppError {
	pw, err := common.ProvisionWatcherClient.ProvisionWatcher(ctx, id)
	if err != nil {
		appErr := common.NewBadRequestError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Cannot find provisionwatcher %s in Core Metadata: %v", id, err))
		return appErr
	}

	err = cache.ProvisionWatchers().Add(pw)
	if err == nil {
		common.LoggingClient.Info(fmt.Sprintf("Added provisionwatcher %s", id))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Cannot add provisionwatcher %s: %v", id, err.Error()))
		return appErr
	}

	return nil
}

func handleUpdateProvisionWatcher(ctx context.Context, id string) common.AppError {
	pw, err := common.ProvisionWatcherClient.ProvisionWatcher(ctx, id)
	if err != nil {
		appErr := common.NewBadRequestError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Cannot find provisionwatcher %s in Core Metadata: %v", id, err))
		return appErr
	}

	err = cache.ProvisionWatchers().Update(pw)
	if err == nil {
		common.LoggingClient.Info(fmt.Sprintf("Updated provisionwatcher %s", id))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Cannot update provisionwatcher %s: %v", id, err.Error()))
		return appErr
	}

	return nil
}

func handleDeleteProvisionWatcher(id string) common.AppError {
	err := cache.ProvisionWatchers().Remove(id)
	if err == nil {
		common.LoggingClient.Info(fmt.Sprintf("Removed provisionwatcher %s", id))
	} else {
		appErr := common.NewServerError(err.Error(), err)
		common.LoggingClient.Error(fmt.Sprintf("Cannot remove provisionwatcher %s: %v", id, err.Error()))
		return appErr
	}

	return nil
}
