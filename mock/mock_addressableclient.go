//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type AddressableClientMock struct {
}

func (AddressableClientMock) Add(addr *models.Addressable) (string, error) {
	return "5b977c62f37ba10e36673802", nil
}

func (AddressableClientMock) AddressableForName(name string) (models.Addressable, error) {
	var addressable = models.Addressable{Id: bson.ObjectIdHex("5b977c62f37ba10e36673802"), Name: name}
	var err error = nil
	if name == "" {
		err = types.NewErrServiceClient(http.StatusNotFound, []byte("Item not found"))
	}

	return addressable, err
}
