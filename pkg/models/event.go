// -- Mode: Go; indent-tabs-mode: t --
//
// Copyright (C) 2019 Intel Ltd
//
// SPDX-License-Identifier: Apache-2.0
package models

import (
	"bytes"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/ugorji/go/codec"
)

type Event struct {
	contract.Event
	EncodedEvent []byte
}

// HasBinaryValue confirms whether an event contains one or more
// readings populated with a BinaryValue payload.
func (e Event) HasBinaryValue() bool {
	if len(e.Readings) > 0 {
		for r := range e.Readings {
			if len(e.Readings[r].BinaryValue) > 0 {
				return true
			}
		}
	}
	return false
}

// TODO: Add as method of dsModels.event or contract.event/client
func (e Event) EncodeBinaryEvent(ev *contract.Event) ([]byte, error) {
	buf := new(bytes.Buffer)
	hCbor := new(codec.CborHandle)
	enc := codec.NewEncoder(buf, hCbor)
	err := enc.Encode(ev)
	if err == nil {
		return buf.Bytes(), nil
	}
	return nil, err
}
