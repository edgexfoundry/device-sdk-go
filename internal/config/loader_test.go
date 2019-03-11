// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-registry"
	"github.com/edgexfoundry/go-mod-registry/pkg/factory"
)

func TestCheckConsulUpReturnErrorOnTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second * 11)
		}))
	defer ts.Close()

	url := strings.Split(ts.URL, ":")
	host := strings.Split(url[1], "//")[1]
	port, err := strconv.Atoi(url[2])
	if err != nil {
		fmt.Println(err.Error())
	}

	config := &common.Config{}
	config.Registry.Host = host
	config.Registry.Port = port
	config.Registry.FailLimit = 1
	config.Registry.FailWaitTime = 0

	registryConfig := registry.Config{
		Host: host,
		Port: port,
		Type: "consul",
	}

	RegistryClient, err = factory.NewRegistryClient(registryConfig)
	if err != nil {
		t.Error("failed to create new registry client")
	}

	err = checkRegistryUp(config)
	if err == nil {
		t.Fatal("Error should be raised")
	}

	if err.Error() != "can't get connection to Registry" {
		t.Error("Wrong error message", err.Error())
	}
}

func TestCheckConsulUpReturnErrorOnBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
	defer ts.Close()

	url := strings.Split(ts.URL, ":")
	host := strings.Split(url[1], "//")[1]
	port, err := strconv.Atoi(url[2])
	if err != nil {
		fmt.Println(err.Error())
	}

	config := &common.Config{}
	config.Registry.Host = host
	config.Registry.Port = port
	config.Registry.FailLimit = 1
	config.Registry.FailWaitTime = 0

	registryConfig := registry.Config{
		Host: host,
		Port: port,
		Type: "consul",
	}

	RegistryClient, err = factory.NewRegistryClient(registryConfig)
	if err != nil {
		t.Error("failed to create new registry client")
	}

	err = checkRegistryUp(config)
	if err == nil {
		t.Error("Error should be raised")
	}

	if err.Error() != "can't get connection to Registry" {
		t.Error("Wrong error message ", err.Error())
	}
}

func TestLoadDriverConfigFromFile(t *testing.T) {
	expectedProperty1 := "Protocol"
	expectedValue1 := "tcp"
	expectedProperty2 := "Port"
	expectedValue2 := "1883"

	config, err := loadConfigFromFile("", "./test")

	if err != nil {
		t.Errorf("Fail to load config from file, %v", err)
	} else if val, ok := config.Driver[expectedProperty1]; ok != true || val != expectedValue1 {
		t.Errorf("Unexpected test result, '%s' should be exist and value shoud be '%s'", expectedProperty1, expectedValue1)
	} else if val, ok := config.Driver[expectedProperty2]; ok != true || val != expectedValue2 {
		t.Errorf("Unexpected test result, '%s' should be exist and value shoud be '%s'", expectedProperty2, expectedValue2)
	}
}
