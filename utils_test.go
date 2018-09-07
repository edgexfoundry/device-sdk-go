// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package device

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestConsulClientReturnErrorOnTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second * 11)
		}))
	defer ts.Close()

	url := strings.Split(ts.URL, ":")
	host := url[0] + ":" + url[1]
	port, err := strconv.Atoi(url[2])
	if err != nil {
		fmt.Println(err.Error())
	}

	config := &Config{}
	config.Registry.Host = host
	config.Registry.Port = port
	config.Registry.FailLimit = 1
	config.Registry.FailWaitTime = 0

	consul, err := getConsulClient(config)
	if consul != nil || err == nil {
		t.Error("Error should be raised")
	}

	if err.Error() != "Cannot get connection to Consul" {
		t.Error("Wrong error message")
	}
}

func TestConsulClientReturnErrorOnBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
	defer ts.Close()

	url := strings.Split(ts.URL, ":")
	host := url[0] + ":" + url[1]
	port, err := strconv.Atoi(url[2])
	if err != nil {
		fmt.Println(err.Error())
	}

	config := &Config{}
	config.Registry.Host = host
	config.Registry.Port = port
	config.Registry.FailLimit = 1
	config.Registry.FailWaitTime = 0

	consul, err := getConsulClient(config)
	if consul != nil || err == nil {
		t.Error("Error should be raised")
	}

	if err.Error() != "Bad response from Consul service" {
		t.Error("Wrong error message")
	}
}
