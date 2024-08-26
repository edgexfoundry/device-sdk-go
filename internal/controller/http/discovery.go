// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/application"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/controller/http/correlation"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func (c *RestController) Discovery(e echo.Context) error {
	request := e.Request()
	writer := e.Response()
	ds := container.DeviceServiceFrom(c.dic.Get)
	if ds.AdminState == models.Locked {
		err := errors.NewCommonEdgeX(errors.KindServiceLocked, "service locked", nil)
		return c.sendEdgexError(writer, request, err, common.ApiDiscoveryRoute)
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	if !configuration.Device.Discovery.Enabled {
		err := errors.NewCommonEdgeX(errors.KindServiceUnavailable, "device discovery disabled", nil)
		return c.sendEdgexError(writer, request, err, common.ApiDiscoveryRoute)
	}

	driver := container.ProtocolDriverFrom(c.dic.Get)

	// Use correlation id as request id since there is no request body
	requestId := correlation.IdFromContext(request.Context())
	go func() {
		c.lc.Infof("Discovery triggered. Correlation Id: %s", requestId)
		autodiscovery.DiscoveryWrapper(driver, request.Context(), c.dic)
		c.lc.Infof("Discovery end. Correlation Id: %s", requestId)
	}()

	response := commonDTO.NewBaseResponse(requestId, "Device Discovery is triggered.", http.StatusAccepted)
	return c.sendResponse(writer, request, common.ApiDiscoveryRoute, response, http.StatusAccepted)
}

func (c *RestController) ProfileScan(e echo.Context) error {
	request := e.Request()
	writer := e.Response()
	if request.Body != nil {
		defer func() { _ = request.Body.Close() }()
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "Failed to read request body", err)
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScan)
	}

	ps := container.ProfileScanFrom(c.dic.Get)
	if ps == nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindNotImplemented, "Profile scan is not implemented", nil)
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScan)
	}

	req, edgexErr := profileScanValidation(body, request.Context(), c.dic)
	if edgexErr != nil {
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScan)
	}

	busy := make(chan bool)
	go func() {
		c.lc.Infof("Profile scanning is triggered. Correlation Id: %s", req.RequestId)
		application.ProfileScanWrapper(busy, ps, req, request.Context(), c.dic)
		c.lc.Infof("Profile scanning is end. Correlation Id: %s", req.RequestId)
	}()
	b := <-busy
	if b {
		edgexErr := errors.NewCommonEdgeX(errors.KindStatusConflict, fmt.Sprintf("Another profile scan process for %s is currently running", req.DeviceName), nil)
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScan)
	}

	response := commonDTO.NewBaseResponse(req.RequestId, "Device ProfileScan is triggered.", http.StatusAccepted)
	return c.sendResponse(writer, request, common.ApiProfileScan, response, http.StatusAccepted)
}

func profileScanValidation(request []byte, ctx context.Context, dic *di.Container) (sdkModels.ProfileScanRequest, errors.EdgeX) {
	var r sdkModels.ProfileScanRequest
	// check device service AdminState
	ds := container.DeviceServiceFrom(dic.Get)
	if ds.AdminState == models.Locked {
		return r, errors.NewCommonEdgeX(errors.KindServiceLocked, "service locked", nil)
	}

	// parse request payload
	var req requests.ProfileScanRequest
	err := req.UnmarshalJSON(request)
	if err != nil {
		return r, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to parse request body", err)
	}

	// check requested device exists
	device, ok := cache.Devices().ForName(req.DeviceName)
	if !ok {
		return r, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device %s not found", req.DeviceName), nil)
	}

	// check profile should not exist
	if len(req.ProfileName) > 0 {
		if _, exist := cache.Profiles().ForName(req.ProfileName); exist {
			return r, errors.NewCommonEdgeX(errors.KindStatusConflict, fmt.Sprintf("profile name %s is duplicated", req.ProfileName), nil)
		}
	} else {
		req.ProfileName = fmt.Sprintf("%s_profile_%d", req.DeviceName, time.Now().UnixMilli())
	}

	requestId := req.RequestId
	if len(requestId) == 0 {
		// Use correlation id as request id if request id is not provided
		requestId = correlation.IdFromContext(ctx)
	}

	r = sdkModels.ProfileScanRequest{
		BaseRequest: commonDTO.BaseRequest{
			Versionable: commonDTO.NewVersionable(),
			RequestId:   requestId,
		},
		DeviceName:  req.DeviceName,
		ProfileName: req.ProfileName,
		Options:     req.Options,
		Protocols:   device.Protocols,
	}

	return r, nil
}
