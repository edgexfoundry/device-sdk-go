// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"testing"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/mock"
)

func TestInitCache(t *testing.T) {
	common.ValueDescriptorClient = &mock.ValueDescriptorMock{}
	common.DeviceClient = &mock.DeviceClientMock{}
	InitCache()

	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())

	vdsBeforeAddingToCache, _ := common.ValueDescriptorClient.ValueDescriptors(ctx)
	if vl := len(ValueDescriptors().All()); vl != len(vdsBeforeAddingToCache) {
		t.Errorf("the expected number of valuedescriptors in cache is %d but got: %d:", len(vdsBeforeAddingToCache), vl)
	}

	dsBeforeAddingToCache, _ := common.DeviceClient.DevicesForServiceByName(ctx, common.ServiceName)
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
