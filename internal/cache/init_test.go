// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/mock"
)

func TestInitCache(t *testing.T) {
	serviceName := "init-cache-test"
	lc := logger.NewClientStdOut("device-sdk-test", false, "DEBUG")
	vdc := &mock.ValueDescriptorMock{}
	dc := &mock.DeviceClientMock{}
	pwc := &mock.ProvisionWatcherClientMock{}
	InitCache(serviceName, lc, vdc, dc, pwc)

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())

	vdsBeforeAddingToCache, _ := vdc.ValueDescriptors(ctx)
	if vl := len(ValueDescriptors().All()); vl != len(vdsBeforeAddingToCache) {
		t.Errorf("the expected number of valuedescriptors in cache is %d but got: %d:", len(vdsBeforeAddingToCache), vl)
	}

	dsBeforeAddingToCache, _ := dc.DevicesForServiceByName(ctx, serviceName)
	if dl := len(Devices().All()); dl != len(dsBeforeAddingToCache) {
		t.Errorf("the expected number of devices in cache is %d but got: %d:", len(dsBeforeAddingToCache), dl)
	}

	pMap := make(map[string]contract.DeviceProfile, len(dsBeforeAddingToCache)*2)
	for _, d := range dsBeforeAddingToCache {
		pMap[d.Profile.Name] = d.Profile
	}
	if pl := len(Profiles().All()); pl != len(pMap) {
		t.Errorf("the expected number of device profiles in cache is %d but got: %d:", len(pMap), pl)
	} else {
		psFromCache := Profiles().All()
		for _, p := range psFromCache {
			assert.Equal(t, pMap[p.Name], p)
		}
	}
}
