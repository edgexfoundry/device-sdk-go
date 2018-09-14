//
// Copyright (C) 2018 IOTech
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"errors"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type AddressableClientMock struct {
}

func (AddressableClientMock) Add(addr *models.Addressable) (string, error) {
	return "", nil
}

func (AddressableClientMock) AddressableForName(name string) (models.Addressable, error) {
	var addressable = models.Addressable{Name: name}
	var err error = nil
	if name == "" {
		err = errors.New("addressable not exist")
	}

	return addressable, err
}
