// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type AddressableClientMock struct {
}

func (AddressableClientMock) Addressable(id string) (models.Addressable, error) {
	panic("implement me")
}

func (AddressableClientMock) Add(addr *models.Addressable) (string, error) {
	return "5b977c62f37ba10e36673802", nil
}

func (AddressableClientMock) AddressableForName(name string) (models.Addressable, error) {
	var addressable = models.Addressable{Id: "5b977c62f37ba10e36673802", Name: name}
	var err error = nil
	if name == "" {
		err = types.NewErrServiceClient(http.StatusNotFound, nil)
	}

	return addressable, err
}

func (AddressableClientMock) Update(addr models.Addressable) error {
	return nil
}

func (AddressableClientMock) Delete(id string) error {
	return nil
}
