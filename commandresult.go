// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package defines an interface used to build an EdgeX Foundry device
// service.  This interace provides an asbstraction layer for the device
// or protocol specific logic of a device service.
//
// TODO:
// * Determine if gxds should define separate 'Handler' and 'Driver'
//   protocol interfaces?  Can they be combined?
//
// * Investigate changing calling signatures to leverage std Go
//   interfaces, such as Reader/Writer, ...
//
package gxds

import "github.com/edgexfoundry/edgex-go/core/domain/models"

type CommandResult struct {
     RO     *models.ResourceOperation
     Result string
}
