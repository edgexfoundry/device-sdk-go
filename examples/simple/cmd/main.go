// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package provides a simple example of a device service.
//
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tonyespy/gxds/examples/simple"
	"github.com/tonyespy/gxds/service"
)

var flags struct {
	configPath *string
}

func main() {
	var useRegistry bool
	var profile string
	var confDir string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use the registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry.")
	flag.StringVar(&profile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profile, "p", "", "Specify a profile other than default.")
	flag.StringVar(&confDir, "confdir", "", "Specify an alternate configuration directory.")
	flag.StringVar(&confDir, "c", "", "Specify an alternate configuration directory.")
	flag.Parse()

	if err := startService(useRegistry, profile, confDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func startService(useRegistry bool, profile string, confDir string) error {
	s := simple.SimpleDriver{}

	d, err := service.New("device-simple")
	if err != nil {
		return err
	}

	if err := d.Init(useRegistry, profile, confDir, &s); err != nil {
		return err
	}

	d.Version = "0.1"

	d.Start()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-ch:
		fmt.Fprintf(os.Stderr, "Exiting on %s signal.\n", sig)
	}

	return d.Stop()
}
