// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
// This package provides a simple example of a device service.
//
package main

import (
	"flag"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go/common"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgexfoundry/device-sdk-go"
	configLoader "github.com/edgexfoundry/device-sdk-go/config"
	"github.com/edgexfoundry/device-sdk-go/driver"
)

var flags struct {
	configPath *string
}

func main() {
	var useRegistry bool
	var profile string
	var confDir string

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // clean up existing flag defined by other code
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use the registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry.")
	flag.StringVar(&profile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profile, "p", "", "Specify a profile other than default.")
	flag.StringVar(&confDir, "confdir", "", "Specify an alternate configuration directory.")
	flag.StringVar(&confDir, "c", "", "Specify an alternate configuration directory.")
	flag.Parse()

	config, err := configLoader.LoadConfig(useRegistry, profile, confDir)
	if err = startService(useRegistry, profile, config); err != nil {
		fmt.Fprintf(os.Stderr, "error loading config file: %v\n", err)
		os.Exit(1)
	}

	if err = startService(useRegistry, profile, config); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func startService(useRegistry bool, profile string, config *common.Config) error {
	sd := driver.SimpleDriver{}

	s, err := device.NewService(&sd)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Calling service.Start.\n")

	if err := s.Start(config); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Setting up signals.\n")

	// TODO: this code never executes!

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-ch:
		fmt.Fprintf(os.Stderr, "Exiting on %s signal.\n", sig)
	}

	return s.Stop(false)
}
