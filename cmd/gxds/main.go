// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/tonyespy/gxds/service"
)

var flags struct {
	configPath *string
}

func init() {
	fmt.Fprintf(os.Stdout, "Init called\n")

	flags.configPath = flag.String("config", "./service-config.json", "service configuration file")
}

func main() {

	flag.Parse()

	if err := startService(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func startService() error {

	d, err := service.New()
	if err != nil {
		return err
	}

	// TODO: create a ProtocolHandler implementation
	// and pass to d.Init()
	if err := d.Init(flags.configPath, nil); err != nil {
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
