// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/google/uuid"
)

type discoveryLocker struct {
	busy bool
	id   string
	mux  sync.Mutex
}

var locker discoveryLocker

func TransformHandler(requestMap map[string]string, lc logger.LoggingClient) (map[string]string, common.AppError) {
	lc.Info(fmt.Sprintf("service: transform request: transformData: %s", requestMap["transformData"]))
	return requestMap, nil
}

func DiscoveryHandler(w http.ResponseWriter, discovery models.ProtocolDiscovery, lc logger.LoggingClient) {
	locker.mux.Lock()
	if locker.id == "" {
		locker.id = uuid.New().String()
	}
	locker.mux.Unlock()

	if w != nil {
		msg := fmt.Sprintf("Discovery triggered or already running, id = %s", locker.id)
		w.WriteHeader(http.StatusAccepted) //status=202
		_, _ = io.WriteString(w, msg)
	}

	locker.mux.Lock()
	defer locker.mux.Unlock()
	if locker.busy {
		lc.Info(fmt.Sprintf("Device discovery process is running, id = %s", locker.id))
		return
	}
	locker.busy = true
	lc.Info(fmt.Sprintf("Device discovery triggered"))

	go discovery.Discover()
}

func ReleaseLock() string {
	var id string
	locker.mux.Lock()
	id = locker.id
	locker.id = ""
	locker.busy = false
	locker.mux.Unlock()

	return id
}
