// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package startup

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgexfoundry/device-sdk-go"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	UseRegistry   string `short:"r" long:"registry" description:"Indicates the service should use the registry and provide the registry url." optional:"true" optional-value:"LOAD_FROM_FILE"`
	ConfProfile   string `short:"p" long:"profile" description:"Specify a profile other than default."`
	ConfDir       string `short:"c" long:"confdir" description:"Specify an alternate configuration directory."`
	OverwriteConf bool   `short:"o" long:"overwrite" description:"Overwrite configuration in the registry."`
}

var opts Options

// Bootstrap starts the Device Service in a default way
func Bootstrap(serviceName string, serviceVersion string, driver dsModels.ProtocolDriver) {
	//flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // clean up existing flag defined by other code
	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	device.SetOverwriteConfig(opts.OverwriteConf)
	if err := startService(serviceName, serviceVersion, driver); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func startService(serviceName string, serviceVersion string, driver dsModels.ProtocolDriver) error {
	s, err := device.NewService(serviceName, serviceVersion, opts.ConfProfile, opts.ConfDir, opts.UseRegistry, driver)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Calling service.Start.\n")
	errChan := make(chan error, 2)
	listenForInterrupt(errChan)
	go s.Start(errChan)

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
