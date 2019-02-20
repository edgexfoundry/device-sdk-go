// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type AddressableClientMock struct {
}

func (AddressableClientMock) Addressable(id string, ctx context.Context) (models.Addressable, error) {
	panic("implement me")
}

func (AddressableClientMock) Add(addr *models.Addressable, ctx context.Context) (string, error) {
	return "5b977c62f37ba10e36673802", nil
}

func (AddressableClientMock) AddressableForName(name string, ctx context.Context) (models.Addressable, error) {
	var addressable = models.Addressable{Id: "5b977c62f37ba10e36673802", Name: name}
	var err error = nil
	if name == "" {
		err = types.NewErrServiceClient(http.StatusNotFound, nil)
	}

	return addressable, err
}

func (AddressableClientMock) Update(addr models.Addressable, ctx context.Context) error {
	return nil
}

func (AddressableClientMock) Delete(id string, ctx context.Context) error {
	return nil
}
