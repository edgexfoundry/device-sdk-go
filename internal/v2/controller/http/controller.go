// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/device-sdk-go/internal/cache"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/provision"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// V2HttpController controller for V2 REST APIs
type V2HttpController struct {
	dic *di.Container
	lc  logger.LoggingClient
}

// NewV2HttpController creates and initializes an V2HttpController
func NewV2HttpController(dic *di.Container) *V2HttpController {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	return &V2HttpController{
		dic: dic,
		lc:  lc,
	}
}

func updateSpecifiedProfile(profileName string, dic *di.Container) error {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	gc := container.GeneralClientFrom(dic.Get)
	vdc := container.CoredataValueDescriptorClientFrom(dic.Get)
	dpc := container.MetadataDeviceProfileClientFrom(dic.Get)

	profile, err := dpc.DeviceProfileForName(context.Background(), profileName)
	if err != nil {
		return err
	}

	_, exist := cache.Profiles().ForName(profileName)
	if exist == false {
		err = cache.Profiles().Add(profile)
		if err == nil {
			provision.CreateDescriptorsFromProfile(&profile, lc, gc, vdc)
			lc.Info(fmt.Sprintf("Added device profile: %s", profileName))
		} else {
			return err
		}
	} else {
		err := cache.Profiles().Update(profile)
		if err != nil {
			lc.Warn(fmt.Sprintf("Unable to update profile %s in cache, using the original one", profileName))
		}
	}

	return nil
}

func checkServiceLocked(request *http.Request, locked contract.AdminState) error {
	if locked == contract.Locked {
		return fmt.Errorf("service is locked; %s %s", request.Method, request.URL)
	}

	return nil
}

// sendResponse puts together the response packet for the V2 API
func (c *V2HttpController) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	api string,
	response interface{},
	statusCode int) {

	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)

	writer.Header().Set(sdkCommon.CorrelationHeader, correlationID)
	writer.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	if response != nil {
		data, err := json.Marshal(response)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(data)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), clients.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (c *V2HttpController) sendError(
	writer http.ResponseWriter,
	request *http.Request,
	errKind edgexErr.ErrKind,
	message string,
	err error,
	api string,
	requestID string) {
	edgexErr := edgexErr.NewCommonEdgeX(errKind, message, err)
	c.sendEdgexError(writer, request, edgexErr, api, requestID)
}

func (c *V2HttpController) sendEdgexError(
	writer http.ResponseWriter,
	request *http.Request,
	err edgexErr.EdgeX,
	api string,
	requestID string) {
	c.lc.Error(err.Error())
	c.lc.Debug(err.DebugMessages())
	response := common.NewBaseResponse(requestID, err.Message(), err.Code())
	c.sendResponse(writer, request, api, response, err.Code())
}

func (c *V2HttpController) sendEventResponse(
	writer http.ResponseWriter,
	request *http.Request,
	event *dsModels.Event,
	ec coredata.EventClient,
	api string,
	requestID string) {
	if event != nil {
		if event.HasBinaryValue() {
			// Encode response as application/CBOR.
			if len(event.EncodedEvent) <= 0 {
				var err error
				event.EncodedEvent, err = ec.MarshalEvent(event.Event)
				if err != nil {
					c.lc.Error("DeviceCommand: error encoding event", "device", event.Device, "error", err)
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			writer.Header().Set(clients.ContentType, clients.ContentTypeCBOR)
			_, err := writer.Write(event.EncodedEvent)
			if err != nil {
				c.lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), clients.CorrelationHeader, requestID)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			c.sendResponse(writer, request, api, event, http.StatusOK)
		}
	}
}
