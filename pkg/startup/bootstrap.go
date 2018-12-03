// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package startup

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgexfoundry/device-sdk-go"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
)

var (
	confProfile string
	confDir     string
	useRegistry bool
)

// Bootstrap the Device Service in a default way
func Bootstrap(serviceName string, serviceVersion string, driver ds_models.ProtocolDriver) {
	//flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // clean up existing flag defined by other code
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use the registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry.")
	flag.StringVar(&confProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&confProfile, "p", "", "Specify a profile other than default.")
	flag.StringVar(&confDir, "confdir", "", "Specify an alternate configuration directory.")
	flag.StringVar(&confDir, "c", "", "Specify an alternate configuration directory.")
	flag.Parse()

	if err := startService(serviceName, serviceVersion, driver); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func startService(serviceName string, serviceVersion string, driver ds_models.ProtocolDriver) error {
	s, err := device.NewService(serviceName, serviceVersion, confProfile, confDir, useRegistry, driver)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Calling service.Start.\n")

	errChan := make(chan error, 2)
	listenForInterrupt(errChan)

	err = s.Start(errChan)
	if err != nil {
		return err
	}

	err = <-errChan
	fmt.Fprintf(os.Stdout, "Terminating: %v.\n", err)

	return s.Stop(false)
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
