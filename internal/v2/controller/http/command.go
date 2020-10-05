// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/v2/application"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

const SDKPostEventReserved = "ds-postevent"
const SDKReturnEventReserved = "ds-returnevent"

func (c *V2HttpController) Command(writer http.ResponseWriter, request *http.Request) {
	var reserved url.Values
	vars := mux.Vars(request)
	ds := container.DeviceServiceFrom(c.dic.Get)
	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)

	err := checkServiceLocked(request, ds.AdminState)
	if err != nil {
		c.sendError(writer, request, edgexErr.KindServiceLocked, "service locked", err, sdkCommon.APIV2NameCommandRoute, correlationID)
		return
	}

	// read request body for PUT command
	body, err := readBodyAsString(request)
	if err != nil {
		c.sendError(writer, request, edgexErr.KindServerError, "failed to read request body", err, sdkCommon.APIV2NameCommandRoute, correlationID)
		return
	}
	// filter out the SDK reserved parameters and save the result for GET command
	if len(body) == 0 {
		body, reserved = sdkCommon.V2FilterQueryParams(request.URL.RawQuery, c.lc)
	}

	event, edgexErr := c.CommandHandler(request.Method, vars, body, correlationID)
	if edgexErr != nil {
		c.sendEdgexError(writer, request, edgexErr, sdkCommon.APIV2NameCommandRoute, correlationID)
		return
	}

	// push to core and return event in http response based on SDK reserved query parameters
	if ok, exist := reserved[SDKPostEventReserved]; exist && ok[0] == "yes" {
		go sdkCommon.SendEvent(event, c.lc, container.CoredataEventClientFrom(c.dic.Get))
	}
	if ok, exist := reserved[SDKReturnEventReserved]; !exist || ok[0] == "yes" {
		c.returnEvent(writer, request, event, container.CoredataEventClientFrom(c.dic.Get), sdkCommon.APIV2NameCommandRoute, correlationID)
	}
}

func (c *V2HttpController) CommandHandler(method string, vars map[string]string, body string, id string) (event *dsModels.Event, err edgexErr.EdgeX) {
	var device contract.Device
	deviceKey := vars[sdkCommon.IdVar]

	defer func() {
		if err == nil {
			go sdkCommon.UpdateLastConnected(
				device.Name,
				container.ConfigurationFrom(c.dic.Get),
				bootstrapContainer.LoggingClientFrom(c.dic.Get),
				container.MetadataDeviceClientFrom(c.dic.Get))
		}
	}()

	// check provided device exists
	device, exist := cache.Devices().ForId(deviceKey)
	if !exist {
		deviceKey = vars[sdkCommon.NameVar]
		device, exist = cache.Devices().ForName(deviceKey)
		if !exist {
			return nil, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "device not found", nil)
		}
	}

	// check device's AdminState
	if device.AdminState == contract.Locked {
		return nil, edgexErr.NewCommonEdgeX(edgexErr.KindServiceLocked, fmt.Sprintf("device %s locked", device.Name), nil)
	}

	cmd := vars[sdkCommon.CommandVar]
	helper := application.NewCommandHelper(&device, nil, id, cmd, body, c.dic)
	cmdExists, _ := cache.Profiles().CommandExists(device.Profile.Name, cmd, method)
	if !cmdExists {
		dr, drExists := cache.Profiles().DeviceResource(device.Profile.Name, cmd)
		if !drExists {
			return nil, edgexErr.NewCommonEdgeX(edgexErr.KindEntityDoesNotExist, "command not found", nil)
		}

		helper = application.NewCommandHelper(&device, &dr, id, cmd, body, c.dic)
		if method == http.MethodGet {
			return application.ReadDeviceResource(helper)
		} else {
			return nil, application.WriteDeviceResource(helper)
		}
	} else {
		if method == http.MethodGet {
			return application.ReadCommand(helper)
		} else {
			return nil, application.WriteCommand(helper)
		}
	}
}

func readBodyAsString(req *http.Request) (string, error) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	if len(body) == 0 && req.Method == http.MethodPut {
		return "", errors.New("no request body provided for PUT command")
	}

	return string(body), nil
}
