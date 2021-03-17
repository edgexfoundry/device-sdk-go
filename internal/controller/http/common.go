// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/telemetry"
)

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *RestController) Ping(writer http.ResponseWriter, request *http.Request) {
	response := common.NewPingResponse()
	c.sendResponse(writer, request, v2.ApiPingRoute, response, http.StatusOK)
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *RestController) Version(writer http.ResponseWriter, request *http.Request) {
	response := common.NewVersionSdkResponse(sdkCommon.ServiceVersion, sdkCommon.SDKVersion)
	c.sendResponse(writer, request, v2.ApiVersionRoute, response, http.StatusOK)
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *RestController) Config(writer http.ResponseWriter, request *http.Request) {
	response := common.NewConfigResponse(container.ConfigurationFrom(c.dic.Get))
	c.sendResponse(writer, request, v2.ApiVersionRoute, response, http.StatusOK)
}

// Metrics handles the request to the /metrics endpoint, memory and cpu utilization stats
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *RestController) Metrics(writer http.ResponseWriter, request *http.Request) {
	telem := telemetry.NewSystemUsage()
	metrics := common.Metrics{
		MemAlloc:       telem.Memory.Alloc,
		MemFrees:       telem.Memory.Frees,
		MemLiveObjects: telem.Memory.LiveObjects,
		MemMallocs:     telem.Memory.Mallocs,
		MemSys:         telem.Memory.Sys,
		MemTotalAlloc:  telem.Memory.TotalAlloc,
		CpuBusyAvg:     uint8(telem.CpuBusyAvg),
	}

	response := common.NewMetricsResponse(metrics)
	c.sendResponse(writer, request, v2.ApiMetricsRoute, response, http.StatusOK)
}

// Secret handles the request to add Device Service exclusive secret to the Secret Store
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *RestController) Secret(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		_ = request.Body.Close()
	}()

	provider := bootstrapContainer.SecretProviderFrom(c.dic.Get)
	secretRequest := common.SecretRequest{}
	err := json.NewDecoder(request.Body).Decode(&secretRequest)
	if err != nil {
		edgexError := errors.NewCommonEdgeX(errors.KindContractInvalid, "JSON decode failed", err)
		c.sendEdgexError(writer, request, edgexError, sdkCommon.APIV2SecretRoute)
		return
	}

	path, secret := c.prepareSecret(secretRequest)

	if err := provider.StoreSecrets(path, secret); err != nil {
		edgexError := errors.NewCommonEdgeX(errors.KindServerError, "Storing secret failed", err)
		c.sendEdgexError(writer, request, edgexError, sdkCommon.APIV2SecretRoute)
		return
	}

	response := common.NewBaseResponse(secretRequest.RequestId, "", http.StatusCreated)
	c.sendResponse(writer, request, sdkCommon.APIV2SecretRoute, response, http.StatusCreated)
}

func (c *RestController) prepareSecret(request common.SecretRequest) (string, map[string]string) {
	var secretKVs = make(map[string]string)
	for _, secret := range request.SecretData {
		secretKVs[secret.Key] = secret.Value
	}

	path := strings.TrimSpace(request.Path)
	config := container.ConfigurationFrom(c.dic.Get)

	// add '/' in the full URL path if it's not already at the end of the base path or sub path
	if !strings.HasSuffix(config.SecretStore.Path, "/") && !strings.HasPrefix(path, "/") {
		path = "/" + path
	} else if strings.HasSuffix(config.SecretStore.Path, "/") && strings.HasPrefix(path, "/") {
		// remove extra '/' in the full URL path because secret store's (Vault) APIs don't handle extra '/'.
		path = path[1:]
	}

	return path, secretKVs
}
