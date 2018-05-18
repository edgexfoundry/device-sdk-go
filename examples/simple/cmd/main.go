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

func init() {
	fmt.Fprintf(os.Stdout, "Init called\n")

	flags.configPath = flag.String("config", "./configuration.json", "simple configuration file")
}

func main() {

	flag.Parse()

	if err := startService(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func startService() error {
	s := simple.SimpleDriver{}

	d, err := service.New("device-simple")
	if err != nil {
		return err
	}

	// TODO: create a ProtocolHandler implementation
	// and pass to d.Init()
	if err := d.Init(flags.configPath, &s); err != nil {
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
